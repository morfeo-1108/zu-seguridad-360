package flight

import (
	"context"
	"testing"
	"time"
)

func TestLinkMonitor_StateTransitions(t *testing.T) {
	lm := NewLinkMonitor()
	lm.intermittentAfter = 50 * time.Millisecond
	lm.lostAfter = 120 * time.Millisecond

	if lm.State() != LinkOK {
		t.Error("enlace recién iniciado debe estar OK")
	}
	time.Sleep(70 * time.Millisecond)
	if lm.State() != LinkIntermittent {
		t.Errorf("tras umbral intermitente debe ser intermittent, got %s", lm.State())
	}
	time.Sleep(70 * time.Millisecond)
	if lm.State() != LinkLost {
		t.Errorf("tras umbral de pérdida debe ser lost, got %s", lm.State())
	}
	// Un heartbeat lo recupera
	lm.Heartbeat()
	if lm.State() != LinkOK {
		t.Error("heartbeat debe recuperar el enlace a OK")
	}
}

func TestOnboardMission_ValidateForOfflineOps(t *testing.T) {
	home := GeoPoint{Lat: 4.65, Lng: -74.07, Alt: 0}
	fence := testFence()
	plan := testPlan()

	m := BuildOnboardMission("m1", plan, fence, home, DefaultEnvelope())
	// Sin confirmar carga, NO debe validar para ops sin red
	ok, problems := m.ValidateForOfflineOps()
	if ok {
		t.Error("misión no cargada no debe validar para ops offline")
	}
	found := false
	for _, p := range problems {
		if contains(p, "cargada a bordo") {
			found = true
		}
	}
	if !found {
		t.Error("debe reportar que falta confirmar carga a bordo")
	}

	// Confirmada la carga, debe validar
	m.Uploaded = true
	ok, problems = m.ValidateForOfflineOps()
	if !ok {
		t.Errorf("misión completa y cargada debe validar, problemas: %v", problems)
	}
}

func TestOnboardMission_RequiresHomePoint(t *testing.T) {
	// Sin home point, el RTH es imposible → no valida
	plan := testPlan()
	m := BuildOnboardMission("m1", plan, testFence(), GeoPoint{}, DefaultEnvelope())
	m.Uploaded = true
	ok, problems := m.ValidateForOfflineOps()
	if ok {
		t.Error("sin home point no debe validar")
	}
	_ = problems
}

func TestGNSSNavigator_NavigateHome(t *testing.T) {
	home := GeoPoint{Lat: 4.6500, Lng: -74.0700, Alt: 0}
	nav := NewGNSSNavigator(home)
	// Dron al norte de home
	current := GeoPoint{Lat: 4.6600, Lng: -74.0700, Alt: 40}
	bearing, dist := nav.NavigateHome(current)
	// Debe apuntar al sur (~180°) y ~1.1km
	if bearing < 170 || bearing > 190 {
		t.Errorf("rumbo a home debe ser ~180° (sur), got %.0f", bearing)
	}
	if dist < 1000 || dist > 1200 {
		t.Errorf("distancia a home debe ser ~1.1km, got %.0fm", dist)
	}
}

func TestGNSSHealthy(t *testing.T) {
	good := GNSSFix{Satellites: 16, HDOP: 0.9, Fix3D: true}
	if ok, _ := GNSSHealthy(good, 12); !ok {
		t.Error("fix robusto debe ser saludable")
	}
	noFix := GNSSFix{Satellites: 16, HDOP: 0.9, Fix3D: false}
	if ok, _ := GNSSHealthy(noFix, 12); ok {
		t.Error("sin fix 3D no debe ser saludable")
	}
	fewSats := GNSSFix{Satellites: 6, HDOP: 0.9, Fix3D: true}
	if ok, _ := GNSSHealthy(fewSats, 12); ok {
		t.Error("pocos satélites no debe ser saludable")
	}
}

func TestFlightController_LinkLossContinuesMissionOnboard(t *testing.T) {
	cmd := &mockCommander{}
	fc := NewFlightController("d1", "m1", testPlan(), DefaultEnvelope(), testFence(), cmd)
	home := testPlan().Home
	m := BuildOnboardMission("m1", testPlan(), testFence(), home, DefaultEnvelope())
	m.LinkLoss.Action = ActionContinueMission
	m.Uploaded = true
	fc.LoadOnboardMission(&m)
	fc.transition(PhaseEnroute)

	// Forzar enlace perdido
	fc.linkMonitor.lastHeartbeat = time.Now().Add(-30 * time.Second)
	goodFix := GNSSFix{Latitude: 4.6505, Longitude: -74.0695, Altitude: 40, Satellites: 16, HDOP: 0.8, Fix3D: true}

	decision := fc.EvaluateOnboardLinkLoss(context.Background(), goodFix)
	if decision == nil || decision.Action != ActionContinueMission {
		t.Fatalf("con política continue_mission y buen GNSS, debe continuar la misión onboard, got %v", decision)
	}
	if decision.NavMode != "gnss_onboard" {
		t.Error("la navegación debe ser por GNSS onboard, sin red")
	}
}

func TestFlightController_LinkLossRTHByGNSS(t *testing.T) {
	cmd := &mockCommander{}
	fc := NewFlightController("d1", "m1", testPlan(), DefaultEnvelope(), testFence(), cmd)
	home := testPlan().Home
	m := BuildOnboardMission("m1", testPlan(), testFence(), home, DefaultEnvelope())
	m.LinkLoss.Action = ActionReturnToHome
	m.Uploaded = true
	fc.LoadOnboardMission(&m)
	fc.transition(PhaseEnroute)

	fc.linkMonitor.lastHeartbeat = time.Now().Add(-30 * time.Second)
	goodFix := GNSSFix{Latitude: 4.6530, Longitude: -74.0720, Altitude: 40, Satellites: 15, HDOP: 1.0, Fix3D: true}

	decision := fc.EvaluateOnboardLinkLoss(context.Background(), goodFix)
	if decision == nil || decision.Action != ActionReturnToHome {
		t.Fatalf("política RTH debe disparar retorno, got %v", decision)
	}
	if !cmd.calledWith("rth") {
		t.Error("debe comandar RTH al dron")
	}
}

func TestFlightController_LinkLossBadGNSSLandsInPlace(t *testing.T) {
	cmd := &mockCommander{}
	fc := NewFlightController("d1", "m1", testPlan(), DefaultEnvelope(), testFence(), cmd)
	home := testPlan().Home
	m := BuildOnboardMission("m1", testPlan(), testFence(), home, DefaultEnvelope())
	m.Uploaded = true
	fc.LoadOnboardMission(&m)
	fc.transition(PhaseEnroute)

	fc.linkMonitor.lastHeartbeat = time.Now().Add(-30 * time.Second)
	// GNSS degradado: sin fix confiable, RTH no es seguro
	badFix := GNSSFix{Satellites: 5, HDOP: 4.0, Fix3D: false}

	decision := fc.EvaluateOnboardLinkLoss(context.Background(), badFix)
	if decision == nil || decision.Action != ActionLandInPlace {
		t.Fatalf("enlace perdido + GNSS malo debe aterrizar en sitio, got %v", decision)
	}
	if !cmd.calledWith("emergency_land") {
		t.Error("debe comandar aterrizaje")
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || indexOf(s, sub) >= 0)
}
func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
