package flight

import (
	"context"
	"sync"
	"testing"
)

// mockCommander registra los comandos enviados al "dron" para verificarlos.
type mockCommander struct {
	mu      sync.Mutex
	calls   []string
	failArm bool
}

func (m *mockCommander) record(cmd string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = append(m.calls, cmd)
}
func (m *mockCommander) calledWith(cmd string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, c := range m.calls {
		if c == cmd {
			return true
		}
	}
	return false
}
func (m *mockCommander) Arm(ctx context.Context) error {
	m.record("arm")
	if m.failArm {
		return context.Canceled
	}
	return nil
}
func (m *mockCommander) Takeoff(ctx context.Context, alt float64) error {
	m.record("takeoff")
	return nil
}
func (m *mockCommander) GotoWaypoint(ctx context.Context, wp Waypoint) error {
	m.record("goto")
	return nil
}
func (m *mockCommander) SetGimbal(ctx context.Context, pitch float64) error {
	m.record("gimbal")
	return nil
}
func (m *mockCommander) CapturePhoto(ctx context.Context) error { m.record("capture"); return nil }
func (m *mockCommander) ReturnToHome(ctx context.Context) error { m.record("rth"); return nil }
func (m *mockCommander) Land(ctx context.Context) error         { m.record("land"); return nil }
func (m *mockCommander) EmergencyLand(ctx context.Context) error {
	m.record("emergency_land")
	return nil
}
func (m *mockCommander) Hover(ctx context.Context) error { m.record("hover"); return nil }

func testPlan() MissionPlan {
	home := GeoPoint{Lat: 4.6505, Lng: -74.0695, Alt: 0}
	return PlanMission(home, []InspectionTarget{
		{ID: "z1", Center: GeoPoint{Lat: 4.6508, Lng: -74.0693}, InspectAlt: 30, GimbalPitch: -45, HoverSecs: 10, Priority: 1},
	}, DefaultEnvelope())
}

func TestFlightController_RequiresPICAuthorization(t *testing.T) {
	cmd := &mockCommander{}
	fc := NewFlightController("d1", "m1", testPlan(), DefaultEnvelope(), nil, cmd)
	pre := PreflightResult{Cleared: true}

	// Sin autorización del PIC, la misión NO debe armar ni despegar
	err := fc.Start(context.Background(), pre)
	if err == nil {
		t.Error("sin autorización PIC, Start debería fallar")
	}
	if cmd.calledWith("arm") || cmd.calledWith("takeoff") {
		t.Error("no debe armar/despegar sin autorización del PIC")
	}
	if fc.Phase() != PhaseAuthWait {
		t.Errorf("fase debería ser auth_wait, got %s", fc.Phase())
	}
}

func TestFlightController_AbortsOnFailedPreflight(t *testing.T) {
	cmd := &mockCommander{}
	fc := NewFlightController("d1", "m1", testPlan(), DefaultEnvelope(), nil, cmd)
	fc.AuthorizePIC()
	pre := PreflightResult{Cleared: false, Reason: "batería baja"}

	err := fc.Start(context.Background(), pre)
	if err == nil {
		t.Error("pre-vuelo fallido debe abortar")
	}
	if fc.Phase() != PhaseAborted {
		t.Errorf("fase debería ser aborted, got %s", fc.Phase())
	}
	if cmd.calledWith("takeoff") {
		t.Error("no debe despegar con pre-vuelo fallido")
	}
}

func TestFlightController_CompletesAuthorizedMission(t *testing.T) {
	cmd := &mockCommander{}
	fc := NewFlightController("d1", "m1", testPlan(), DefaultEnvelope(), nil, cmd)
	fc.AuthorizePIC()
	pre := PreflightResult{Cleared: true}

	err := fc.Start(context.Background(), pre)
	if err != nil {
		t.Fatalf("misión autorizada debería completar, got %v", err)
	}
	// Debe haber ejecutado la secuencia completa
	for _, expected := range []string{"arm", "takeoff", "capture", "rth", "land"} {
		if !cmd.calledWith(expected) {
			t.Errorf("misión completa debe ejecutar '%s'", expected)
		}
	}
	if fc.Phase() != PhaseLanded {
		t.Errorf("fase final debería ser landed, got %s", fc.Phase())
	}
}

func TestFlightController_EmergencyLandOnCriticalBattery(t *testing.T) {
	cmd := &mockCommander{}
	home := GeoPoint{Lat: 4.6505, Lng: -74.0695, Alt: 0}
	fc := NewFlightController("d1", "m1", testPlan(), DefaultEnvelope(), nil, cmd)
	// Forzar fase de vuelo
	fc.transition(PhaseEnroute)

	// Telemetría con batería crítica (8%)
	fc.OnTelemetry(context.Background(), Telemetry{
		Position:   GeoPoint{Lat: 4.6505, Lng: -74.0695, Alt: 40},
		BatteryPct: 8, WindMS: 4, GPSSats: 16,
	})
	_ = home
	if !cmd.calledWith("emergency_land") {
		t.Error("batería crítica debe disparar aterrizaje de emergencia")
	}
	if fc.Phase() != PhaseEmergency {
		t.Errorf("fase debería ser emergency, got %s", fc.Phase())
	}
}

func TestFlightController_AutonomousRTHOnLowBattery(t *testing.T) {
	cmd := &mockCommander{}
	fc := NewFlightController("d1", "m1", testPlan(), DefaultEnvelope(), nil, cmd)
	fc.transition(PhaseEnroute)

	// Batería en reserva mínima estando lejos → RTH autónomo
	fc.OnTelemetry(context.Background(), Telemetry{
		Position:   GeoPoint{Lat: 4.6530, Lng: -74.0720, Alt: 40},
		BatteryPct: 22, WindMS: 4, GPSSats: 16,
	})
	if !cmd.calledWith("rth") {
		t.Error("batería en reserva debe disparar RTH autónomo")
	}
}

func TestFlightController_PICOverrideStopsAutonomy(t *testing.T) {
	cmd := &mockCommander{}
	fc := NewFlightController("d1", "m1", testPlan(), DefaultEnvelope(), nil, cmd)
	fc.transition(PhaseEnroute)
	fc.TakeControl() // PIC toma control

	// Tras override, la telemetría no debe disparar acciones autónomas
	fc.OnTelemetry(context.Background(), Telemetry{
		Position:   GeoPoint{Lat: 4.6505, Lng: -74.0695, Alt: 40},
		BatteryPct: 8, WindMS: 4, GPSSats: 16,
	})
	if cmd.calledWith("emergency_land") {
		t.Error("con PIC en control manual, el sistema no debe actuar autónomamente")
	}
}

func TestFlightController_RecordsDecisions(t *testing.T) {
	cmd := &mockCommander{}
	fc := NewFlightController("d1", "m1", testPlan(), DefaultEnvelope(), nil, cmd)
	fc.AuthorizePIC()
	fc.Start(context.Background(), PreflightResult{Cleared: true})

	// La caja negra debe tener decisiones registradas
	decisions := fc.Decisions()
	if len(decisions) == 0 {
		t.Error("la caja negra debe registrar decisiones del piloto")
	}
}
