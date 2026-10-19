// Package geo contiene utilidades geométricas puras (sin dependencias externas)
// para reconocimiento de zona. Aislado para ser testeable de forma independiente.
package geo

// PointInPolygon determina si un punto (lat,lng) está dentro de un polígono
// definido como [][lat,lng], usando ray-casting. Base del reconocimiento de
// zona por geocerca GPS.
func PointInPolygon(lat, lng float64, polygon []interface{}) bool {
	n := len(polygon)
	if n < 3 {
		return false
	}
	inside := false
	j := n - 1
	for i := 0; i < n; i++ {
		pi, _ := polygon[i].([]interface{})
		pj, _ := polygon[j].([]interface{})
		if len(pi) < 2 || len(pj) < 2 {
			j = i
			continue
		}
		xi, _ := toFloat(pi[0])
		yi, _ := toFloat(pi[1])
		xj, _ := toFloat(pj[0])
		yj, _ := toFloat(pj[1])
		if ((yi > lng) != (yj > lng)) && (lat < (xj-xi)*(lng-yi)/(yj-yi)+xi) {
			inside = !inside
		}
		j = i
	}
	return inside
}

func toFloat(v interface{}) (float64, bool) {
	if f, ok := v.(float64); ok {
		return f, true
	}
	return 0, false
}
