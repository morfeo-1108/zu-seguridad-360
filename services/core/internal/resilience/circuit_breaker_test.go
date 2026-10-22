package resilience

import (
	"errors"
	"testing"
	"time"
)

func TestCircuitBreaker_OpensAfterThreshold(t *testing.T) {
	cb := New("test", 3, 100*time.Millisecond)
	failing := func() error { return errors.New("downstream caído") }

	// Las primeras 3 fallas mantienen intentando (Closed)
	for i := 0; i < 3; i++ {
		if err := cb.Execute(failing); err == nil {
			t.Fatalf("iteración %d: esperaba error", i)
		}
	}
	// Tras 3 fallas consecutivas, el circuito abre y rechaza rápido
	if cb.State() != StateOpen {
		t.Errorf("tras %d fallas el circuito debe estar Open, got %v", 3, cb.State())
	}
	if err := cb.Execute(failing); !errors.Is(err, ErrCircuitOpen) {
		t.Errorf("circuito abierto debe devolver ErrCircuitOpen, got %v", err)
	}
}

func TestCircuitBreaker_RecoversAfterCooldown(t *testing.T) {
	cb := New("test", 2, 50*time.Millisecond)
	failing := func() error { return errors.New("falla") }
	healthy := func() error { return nil }

	// Abrir el circuito
	_ = cb.Execute(failing)
	_ = cb.Execute(failing)
	if cb.State() != StateOpen {
		t.Fatal("circuito debería estar abierto")
	}

	// Esperar cooldown → pasa a HalfOpen y una llamada exitosa lo cierra
	time.Sleep(60 * time.Millisecond)
	if err := cb.Execute(healthy); err != nil {
		t.Errorf("tras cooldown, llamada sana debería pasar, got %v", err)
	}
	if cb.State() != StateClosed {
		t.Errorf("éxito en HalfOpen debe cerrar el circuito, got %v", cb.State())
	}
}

func TestCircuitBreaker_SuccessResetsFailures(t *testing.T) {
	cb := New("test", 3, 100*time.Millisecond)
	failing := func() error { return errors.New("falla") }
	healthy := func() error { return nil }

	_ = cb.Execute(failing)
	_ = cb.Execute(failing)
	// Un éxito antes de llegar al umbral resetea el contador
	_ = cb.Execute(healthy)
	_ = cb.Execute(failing)
	_ = cb.Execute(failing)
	// Solo 2 fallas consecutivas tras el reset: sigue cerrado
	if cb.State() == StateOpen {
		t.Error("un éxito debió resetear el contador de fallas")
	}
}
