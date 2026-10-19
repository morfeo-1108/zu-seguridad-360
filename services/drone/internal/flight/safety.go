package flight

import (
	"math"
)

// FlightEnvelope define los límites operacionales que el piloto autónomo
// jamás cruza. Equivale al criterio de un piloto profesional conservador:
// ante la duda, no se vuela; siempre se reserva batería para volver.
type FlightEnvelope struct {
	MaxAltitudeAGL      float64 // 120 m — límite legal Aerocivil RAC 91
	MaxDistanceFromHome float64 // m — radio operacional
	MaxWindSpeedMS      float64 // tolerancia de viento del equipo
	MinBatteryReserve   int     // % mínimo para abortar y volver (RTH)
	MinGPSSatellites    int     // calidad mínima de fix GPS
	MaxGroundSpeedMS    float64 // velocidad máxima en inspección
	MinVisibilityM      float64 // visibilidad mínima
}

// DefaultEnvelope: parámetros conservadores tipo piloto profesional.
func DefaultEnvelope() FlightEnvelope {
	return FlightEnvelope{
		MaxAltitudeAGL:      120, // límite legal Colombia
		MaxDistanceFromHome: 500,
		MaxWindSpeedMS:      10, // ~36 km/h, límite prudente para DJI Mavic
		MinBatteryReserve:   25, // nunca baja de 25% sin estar volviendo
		MinGPSSatellites:    12, // fix robusto
		MaxGroundSpeedMS:    5,  // lento = imágenes estables para visión IA
		MinVisibilityM:      500,
	}
}

type GeoPoint struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
	Alt float64 `json:"alt"` // metros AGL
}

// Geofence: polígono del que el dron no puede salir (zonas de la obra).
type Geofence struct {
	Polygon []GeoPoint `json:"polygon"`
	MaxAlt  float64    `json:"max_alt"`
}

// Contains verifica si un punto está dentro de la geocerca (ray-casting).
func (g Geofence) Contains(p GeoPoint) bool {
	n := len(g.Polygon)
	if n < 3 {
		return false
	}
	inside := false
	j := n - 1
	for i := 0; i < n; i++ {
		pi, pj := g.Polygon[i], g.Polygon[j]
		if ((pi.Lng > p.Lng) != (pj.Lng > p.Lng)) &&
			(p.Lat < (pj.Lat-pi.Lat)*(p.Lng-pi.Lng)/(pj.Lng-pi.Lng)+pi.Lat) {
			inside = !inside
		}
		j = i
	}
	return inside && p.Alt <= g.MaxAlt
}

// haversine: distancia en metros entre dos puntos GPS.
func Haversine(a, b GeoPoint) float64 {
	const R = 6371000.0
	lat1, lat2 := a.Lat*math.Pi/180, b.Lat*math.Pi/180
	dLat := (b.Lat - a.Lat) * math.Pi / 180
	dLng := (b.Lng - a.Lng) * math.Pi / 180
	h := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1)*math.Cos(lat2)*math.Sin(dLng/2)*math.Sin(dLng/2)
	return R * 2 * math.Atan2(math.Sqrt(h), math.Sqrt(1-h))
}

// BatteryReserveOK: ¿alcanza la batería para terminar Y volver con margen?
// Cálculo que hace mentalmente un piloto: energía para volver a home + reserva.
func BatteryReserveOK(currentPct int, distanceToHome float64, env FlightEnvelope) bool {
	// Estimación: ~1% de batería por cada 30m de regreso + 5% de margen de aterrizaje
	returnCost := int(distanceToHome/30.0) + 5
	return currentPct-returnCost >= env.MinBatteryReserve
}

// SafetyVerdict consolida el estado de seguridad del vuelo en curso.
type SafetyVerdict struct {
	Safe       bool     `json:"safe"`
	MustRTH    bool     `json:"must_rth"`  // debe volver a casa ya
	MustLand   bool     `json:"must_land"` // aterrizaje de emergencia
	Violations []string `json:"violations"`
}

// Evaluate: el "cerebro de seguridad" que corre en cada ciclo de telemetría.
func (env FlightEnvelope) Evaluate(pos GeoPoint, home GeoPoint, battery int, windMS float64, sats int, fence *Geofence) SafetyVerdict {
	v := SafetyVerdict{Safe: true}
	dist := Haversine(pos, home)

	if pos.Alt > env.MaxAltitudeAGL {
		v.Safe = false
		v.MustRTH = true
		v.Violations = append(v.Violations, "altitud excede límite legal 120m AGL")
	}
	if dist > env.MaxDistanceFromHome {
		v.Safe = false
		v.MustRTH = true
		v.Violations = append(v.Violations, "distancia excede radio operacional")
	}
	if windMS > env.MaxWindSpeedMS {
		v.Safe = false
		v.MustRTH = true
		v.Violations = append(v.Violations, "viento excede tolerancia del equipo")
	}
	if !BatteryReserveOK(battery, dist, env) {
		v.Safe = false
		v.MustRTH = true
		v.Violations = append(v.Violations, "batería en reserva mínima para regreso")
	}
	if battery <= 10 {
		v.Safe = false
		v.MustLand = true
		v.Violations = append(v.Violations, "batería crítica: aterrizaje de emergencia")
	}
	if sats < env.MinGPSSatellites {
		v.Safe = false
		v.Violations = append(v.Violations, "fix GPS degradado")
	}
	if fence != nil && !fence.Contains(pos) {
		v.Safe = false
		v.MustRTH = true
		v.Violations = append(v.Violations, "fuera de geocerca")
	}
	return v
}
