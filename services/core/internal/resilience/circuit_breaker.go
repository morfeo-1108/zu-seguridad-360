package resilience

import (
	"errors"
	"sync"
	"time"
)

// CircuitBreaker evita fallos en cascada. Si un servicio downstream (ai, documents)
// empieza a fallar, el breaker "abre" y deja de llamarlo, devolviendo error rápido
// en lugar de colgar goroutines esperando timeouts. Tras un cooldown, prueba de nuevo.
//
// Estados: Closed (normal) → Open (falla, rechaza) → HalfOpen (probando recuperación)
type State int

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

var ErrCircuitOpen = errors.New("circuit breaker abierto: servicio downstream no disponible")

type CircuitBreaker struct {
	mu               sync.Mutex
	name             string
	state            State
	failureCount     int
	failureThreshold int           // fallos consecutivos para abrir
	cooldown         time.Duration // tiempo antes de probar half-open
	lastFailure      time.Time
	halfOpenMax      int // llamadas de prueba en half-open
	halfOpenCount    int
}

func New(name string, threshold int, cooldown time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		name: name, state: StateClosed,
		failureThreshold: threshold, cooldown: cooldown, halfOpenMax: 3,
	}
}

// Execute corre fn protegida por el breaker.
func (cb *CircuitBreaker) Execute(fn func() error) error {
	cb.mu.Lock()
	switch cb.state {
	case StateOpen:
		if time.Since(cb.lastFailure) > cb.cooldown {
			cb.state = StateHalfOpen
			cb.halfOpenCount = 0
		} else {
			cb.mu.Unlock()
			return ErrCircuitOpen
		}
	case StateHalfOpen:
		if cb.halfOpenCount >= cb.halfOpenMax {
			cb.mu.Unlock()
			return ErrCircuitOpen
		}
		cb.halfOpenCount++
	}
	cb.mu.Unlock()

	err := fn()

	cb.mu.Lock()
	defer cb.mu.Unlock()
	if err != nil {
		cb.failureCount++
		cb.lastFailure = time.Now()
		if cb.failureCount >= cb.failureThreshold {
			cb.state = StateOpen
		}
		return err
	}
	// Éxito: si estaba probando, cerrar
	if cb.state == StateHalfOpen {
		cb.state = StateClosed
	}
	cb.failureCount = 0
	return nil
}

func (cb *CircuitBreaker) State() State { return cb.state }
