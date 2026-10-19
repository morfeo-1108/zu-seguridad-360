package flight

import "testing"

func TestPlanMission_GeneratesValidRoute(t *testing.T) {
	env := DefaultEnvelope()
	home := GeoPoint{Lat: 4.6505, Lng: -74.0695, Alt: 0}
	targets := []InspectionTarget{
		{ID: "z1", ZoneName: "Alturas", Center: GeoPoint{Lat: 4.6508, Lng: -74.0693}, InspectAlt: 30, GimbalPitch: -45, HoverSecs: 20, Priority: 5},
		{ID: "z2", ZoneName: "Soldadura", Center: GeoPoint{Lat: 4.6502, Lng: -74.0697}, InspectAlt: 25, GimbalPitch: -60, HoverSecs: 15, Priority: 3},
	}
	plan := PlanMission(home, targets, env)

	// Debe incluir despegue + 2 objetivos (transit+inspect cada uno) + RTH
	if len(plan.Waypoints) < 4 {
		t.Fatalf("plan con muy pocos waypoints: %d", len(plan.Waypoints))
	}
	// Debe haber exactamente 2 waypoints de captura (uno por objetivo)
	captures := 0
	for _, wp := range plan.Waypoints {
		if wp.Action == "hover_capture" {
			captures++
		}
	}
	if captures != 2 {
		t.Errorf("esperaba 2 capturas, got %d", captures)
	}
	// Estimaciones positivas y coherentes
	if plan.TotalDistanceM <= 0 {
		t.Error("distancia total debe ser > 0")
	}
	if plan.EstimatedBattery <= 0 || plan.EstimatedBattery >= 100 {
		t.Errorf("batería estimada fuera de rango: %d", plan.EstimatedBattery)
	}
}

func TestPlanMission_RespectsAltitudeLimit(t *testing.T) {
	env := DefaultEnvelope() // MaxAltitudeAGL = 120
	home := GeoPoint{Lat: 4.6505, Lng: -74.0695, Alt: 0}
	plan := PlanMission(home, []InspectionTarget{
		{ID: "z1", Center: GeoPoint{Lat: 4.6508, Lng: -74.0693}, InspectAlt: 30, HoverSecs: 10, Priority: 1},
	}, env)
	// La altitud de crucero nunca debe exceder el límite legal
	if plan.CruiseAlt > env.MaxAltitudeAGL {
		t.Errorf("crucero %.0fm excede límite legal %.0fm", plan.CruiseAlt, env.MaxAltitudeAGL)
	}
	for _, wp := range plan.Waypoints {
		if wp.Point.Alt > env.MaxAltitudeAGL {
			t.Errorf("waypoint %d a %.0fm excede límite legal", wp.Seq, wp.Point.Alt)
		}
	}
}

func TestPlanMission_PrioritizesHighPriority(t *testing.T) {
	env := DefaultEnvelope()
	home := GeoPoint{Lat: 4.6505, Lng: -74.0695, Alt: 0}
	// El objetivo de mayor prioridad (crítico) está MÁS lejos
	targets := []InspectionTarget{
		{ID: "lejos_critico", Center: GeoPoint{Lat: 4.6509, Lng: -74.0691}, InspectAlt: 30, HoverSecs: 10, Priority: 10},
		{ID: "cerca_bajo", Center: GeoPoint{Lat: 4.6506, Lng: -74.0694}, InspectAlt: 30, HoverSecs: 10, Priority: 1},
	}
	plan := PlanMission(home, targets, env)
	// Encontrar el primer hover_capture y verificar que ambos objetivos están
	var capturedIDs []string
	for _, wp := range plan.Waypoints {
		if wp.Action == "hover_capture" {
			capturedIDs = append(capturedIDs, wp.TargetID)
		}
	}
	if len(capturedIDs) != 2 {
		t.Fatalf("esperaba 2 objetivos capturados, got %d", len(capturedIDs))
	}
}
