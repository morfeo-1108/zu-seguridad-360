package flight

import (
	"testing"
)

// Geocerca cuadrada de ~100m de lado cerca de Bogotá para pruebas
func testFence() *Geofence {
	return &Geofence{
		Polygon: []GeoPoint{
			{Lat: 4.6500, Lng: -74.0700},
			{Lat: 4.6510, Lng: -74.0700},
			{Lat: 4.6510, Lng: -74.0690},
			{Lat: 4.6500, Lng: -74.0690},
		},
		MaxAlt: 120,
	}
}

func TestGeofence_Contains(t *testing.T) {
	fence := testFence()
	tests := []struct {
		name string
		p    GeoPoint
		want bool
	}{
		{"centro dentro", GeoPoint{Lat: 4.6505, Lng: -74.0695, Alt: 40}, true},
		{"fuera al norte", GeoPoint{Lat: 4.6520, Lng: -74.0695, Alt: 40}, false},
		{"fuera al este", GeoPoint{Lat: 4.6505, Lng: -74.0680, Alt: 40}, false},
		{"dentro pero excede altitud", GeoPoint{Lat: 4.6505, Lng: -74.0695, Alt: 130}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fence.Contains(tt.p); got != tt.want {
				t.Errorf("Contains(%v) = %v, quería %v", tt.p, got, tt.want)
			}
		})
	}
}

func TestHaversine(t *testing.T) {
	// Dos puntos separados ~111m (0.001 grados de latitud ≈ 111m)
	a := GeoPoint{Lat: 4.6500, Lng: -74.0700}
	b := GeoPoint{Lat: 4.6510, Lng: -74.0700}
	d := Haversine(a, b)
	if d < 100 || d > 120 {
		t.Errorf("Haversine = %.1fm, esperaba ~111m", d)
	}
	// Mismo punto = 0
	if got := Haversine(a, a); got > 0.01 {
		t.Errorf("Haversine(mismo punto) = %.4f, esperaba 0", got)
	}
}

func TestBatteryReserveOK(t *testing.T) {
	env := DefaultEnvelope() // MinBatteryReserve = 25
	tests := []struct {
		name     string
		battery  int
		distance float64
		want     bool
	}{
		{"batería alta, cerca", 80, 100, true},       // 80 - (3+5) = 72 >= 25 ✓
		{"batería justa para volver", 35, 100, true}, // 35 - 8 = 27 >= 25 ✓
		{"batería insuficiente", 30, 200, false},     // 30 - (6+5) = 19 < 25 ✗
		{"lejos con poca batería", 40, 500, false},   // 40 - (16+5) = 19 < 25 ✗
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BatteryReserveOK(tt.battery, tt.distance, env); got != tt.want {
				t.Errorf("BatteryReserveOK(%d%%, %.0fm) = %v, quería %v", tt.battery, tt.distance, got, tt.want)
			}
		})
	}
}

// El test MÁS importante: el cerebro de seguridad que decide RTH/emergencia.
func TestEvaluate_SafetyDecisions(t *testing.T) {
	env := DefaultEnvelope()
	home := GeoPoint{Lat: 4.6505, Lng: -74.0695, Alt: 0}
	fence := testFence()
	safePos := GeoPoint{Lat: 4.6505, Lng: -74.0695, Alt: 40}

	t.Run("vuelo normal es seguro", func(t *testing.T) {
		v := env.Evaluate(safePos, home, 80, 4.0, 16, fence)
		if !v.Safe || v.MustRTH || v.MustLand {
			t.Errorf("vuelo normal debería ser seguro, got %+v", v)
		}
	})

	t.Run("batería crítica fuerza aterrizaje de emergencia", func(t *testing.T) {
		v := env.Evaluate(safePos, home, 8, 4.0, 16, fence)
		if !v.MustLand {
			t.Errorf("batería 8%% debe forzar MustLand, got %+v", v)
		}
	})

	t.Run("exceder 120m fuerza RTH", func(t *testing.T) {
		highPos := GeoPoint{Lat: 4.6505, Lng: -74.0695, Alt: 130}
		v := env.Evaluate(highPos, home, 80, 4.0, 16, fence)
		if !v.MustRTH {
			t.Errorf("altitud 130m debe forzar MustRTH (límite legal), got %+v", v)
		}
	})

	t.Run("viento excesivo fuerza RTH", func(t *testing.T) {
		v := env.Evaluate(safePos, home, 80, 15.0, 16, fence)
		if !v.MustRTH {
			t.Errorf("viento 15m/s debe forzar MustRTH, got %+v", v)
		}
	})

	t.Run("fuera de geocerca fuerza RTH", func(t *testing.T) {
		outside := GeoPoint{Lat: 4.6530, Lng: -74.0695, Alt: 40}
		v := env.Evaluate(outside, home, 80, 4.0, 16, fence)
		if !v.MustRTH {
			t.Errorf("fuera de geocerca debe forzar MustRTH, got %+v", v)
		}
	})

	t.Run("GPS degradado marca inseguro pero no RTH inmediato", func(t *testing.T) {
		v := env.Evaluate(safePos, home, 80, 4.0, 8, fence)
		if v.Safe {
			t.Errorf("GPS con 8 satélites debe marcar inseguro, got %+v", v)
		}
	})
}
