package flight

import (
	"math"
	"sort"
)

// InspectionTarget es un punto/zona que el dron debe inspeccionar.
type InspectionTarget struct {
	ID          string   `json:"id"`
	ZoneName    string   `json:"zone_name"`
	Center      GeoPoint `json:"center"`
	InspectAlt  float64  `json:"inspect_alt"`  // altitud de inspección
	GimbalPitch float64  `json:"gimbal_pitch"` // ángulo de cámara (-90 = abajo)
	HoverSecs   int      `json:"hover_secs"`   // tiempo de captura por punto
	Priority    int      `json:"priority"`     // mayor = primero
}

// Waypoint es un punto de la ruta ejecutable con acciones.
type Waypoint struct {
	Seq         int      `json:"seq"`
	Point       GeoPoint `json:"point"`
	SpeedMS     float64  `json:"speed_ms"`
	GimbalPitch float64  `json:"gimbal_pitch"`
	Action      string   `json:"action"` // "transit" | "inspect" | "hover_capture"
	HoverSecs   int      `json:"hover_secs"`
	TargetID    string   `json:"target_id,omitempty"`
}

// MissionPlan es la ruta completa que ejecutará el piloto autónomo.
type MissionPlan struct {
	Home             GeoPoint   `json:"home"`
	Waypoints        []Waypoint `json:"waypoints"`
	TotalDistanceM   float64    `json:"total_distance_m"`
	EstimatedSecs    int        `json:"estimated_secs"`
	EstimatedBattery int        `json:"estimated_battery_pct"` // batería estimada a consumir
	CruiseAlt        float64    `json:"cruise_alt"`
}

// PlanMission genera la ruta óptima de inspección.
// Estrategia tipo piloto: ordena objetivos por prioridad, luego optimiza el
// recorrido por vecino más cercano (nearest-neighbor) para minimizar distancia,
// con perfil de altitud de crucero seguro entre puntos.
func PlanMission(home GeoPoint, targets []InspectionTarget, env FlightEnvelope) MissionPlan {
	cruiseAlt := 40.0
	if cruiseAlt > env.MaxAltitudeAGL {
		cruiseAlt = env.MaxAltitudeAGL - 10
	}

	// 1. Prioridad primero, luego optimización de ruta dentro de cada nivel
	sort.SliceStable(targets, func(i, j int) bool {
		return targets[i].Priority > targets[j].Priority
	})

	// 2. Nearest-neighbor desde home para minimizar trayecto
	ordered := nearestNeighborRoute(home, targets)

	// 3. Construir waypoints con acciones de inspección
	plan := MissionPlan{Home: home, CruiseAlt: cruiseAlt}
	seq := 0
	var totalDist float64

	// Despegue + ascenso a crucero
	takeoff := GeoPoint{Lat: home.Lat, Lng: home.Lng, Alt: cruiseAlt}
	plan.Waypoints = append(plan.Waypoints, Waypoint{
		Seq: seq, Point: takeoff, SpeedMS: 2, GimbalPitch: 0, Action: "transit",
	})
	seq++
	cursor := takeoff

	for _, t := range ordered {
		// Transitar a la vertical del objetivo en altitud de crucero
		transit := GeoPoint{Lat: t.Center.Lat, Lng: t.Center.Lng, Alt: cruiseAlt}
		totalDist += Haversine(cursor, transit)
		plan.Waypoints = append(plan.Waypoints, Waypoint{
			Seq: seq, Point: transit, SpeedMS: env.MaxGroundSpeedMS, GimbalPitch: 0, Action: "transit",
		})
		seq++
		// Descender a altitud de inspección y capturar
		inspect := GeoPoint{Lat: t.Center.Lat, Lng: t.Center.Lng, Alt: t.InspectAlt}
		totalDist += math.Abs(cruiseAlt - t.InspectAlt)
		plan.Waypoints = append(plan.Waypoints, Waypoint{
			Seq: seq, Point: inspect, SpeedMS: 1.5, GimbalPitch: t.GimbalPitch,
			Action: "hover_capture", HoverSecs: t.HoverSecs, TargetID: t.ID,
		})
		seq++
		cursor = inspect
	}

	// Regreso a home (RTH planificado)
	rthClimb := GeoPoint{Lat: cursor.Lat, Lng: cursor.Lng, Alt: cruiseAlt}
	totalDist += math.Abs(cruiseAlt - cursor.Alt)
	plan.Waypoints = append(plan.Waypoints, Waypoint{Seq: seq, Point: rthClimb, SpeedMS: 2, Action: "transit"})
	seq++
	totalDist += Haversine(rthClimb, GeoPoint{Lat: home.Lat, Lng: home.Lng, Alt: cruiseAlt})
	plan.Waypoints = append(plan.Waypoints, Waypoint{
		Seq: seq, Point: GeoPoint{Lat: home.Lat, Lng: home.Lng, Alt: cruiseAlt},
		SpeedMS: env.MaxGroundSpeedMS, Action: "transit",
	})

	plan.TotalDistanceM = totalDist
	// Estimación: tiempo de tránsito + hovers
	transitSecs := int(totalDist / env.MaxGroundSpeedMS)
	hoverSecs := 0
	for _, t := range ordered {
		hoverSecs += t.HoverSecs
	}
	plan.EstimatedSecs = transitSecs + hoverSecs
	// ~1% batería por cada 30m + 1% por cada 30s de hover
	plan.EstimatedBattery = int(totalDist/30.0) + hoverSecs/30 + 5
	return plan
}

func nearestNeighborRoute(start GeoPoint, targets []InspectionTarget) []InspectionTarget {
	remaining := make([]InspectionTarget, len(targets))
	copy(remaining, targets)
	route := make([]InspectionTarget, 0, len(targets))
	cursor := start
	for len(remaining) > 0 {
		bestIdx, bestDist := 0, math.MaxFloat64
		for i, t := range remaining {
			d := Haversine(cursor, t.Center)
			if d < bestDist {
				bestIdx, bestDist = i, d
			}
		}
		next := remaining[bestIdx]
		route = append(route, next)
		cursor = next.Center
		remaining = append(remaining[:bestIdx], remaining[bestIdx+1:]...)
	}
	return route
}
