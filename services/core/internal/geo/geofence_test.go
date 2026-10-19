package geo

import "testing"

func TestPointInPolygon(t *testing.T) {
	square := []interface{}{
		[]interface{}{4.6500, -74.0700},
		[]interface{}{4.6510, -74.0700},
		[]interface{}{4.6510, -74.0690},
		[]interface{}{4.6500, -74.0690},
	}
	tests := []struct {
		name     string
		lat, lng float64
		want     bool
	}{
		{"centro dentro", 4.6505, -74.0695, true},
		{"fuera al norte", 4.6520, -74.0695, false},
		{"fuera al oeste", 4.6505, -74.0710, false},
		{"fuera al sur", 4.6490, -74.0695, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PointInPolygon(tt.lat, tt.lng, square); got != tt.want {
				t.Errorf("PointInPolygon(%.4f,%.4f) = %v, quería %v", tt.lat, tt.lng, got, tt.want)
			}
		})
	}
}

func TestPointInPolygon_Degenerate(t *testing.T) {
	line := []interface{}{
		[]interface{}{4.65, -74.07},
		[]interface{}{4.66, -74.07},
	}
	if PointInPolygon(4.655, -74.07, line) {
		t.Error("polígono de 2 puntos no contiene nada")
	}
}
