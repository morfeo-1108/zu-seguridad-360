package flight

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// FlightPhase es el estado del vuelo autónomo. La máquina de estados replica
// exactamente la secuencia mental de un piloto profesional.
type FlightPhase string

const (
	PhaseIdle       FlightPhase = "idle"
	PhasePreflight  FlightPhase = "preflight" // verificando checklist
	PhaseAuthWait   FlightPhase = "auth_wait" // esperando autorización del PIC
	PhaseArmed      FlightPhase = "armed"
	PhaseTakeoff    FlightPhase = "takeoff"
	PhaseEnroute    FlightPhase = "enroute"    // transitando a un objetivo
	PhaseInspecting FlightPhase = "inspecting" // capturando en un punto
	PhaseRTH        FlightPhase = "rth"        // regresando a casa
	PhaseLanding    FlightPhase = "landing"
	PhaseLanded     FlightPhase = "landed"
	PhaseEmergency  FlightPhase = "emergency" // aterrizaje de emergencia
	PhaseAborted    FlightPhase = "aborted"
)

// Telemetry es el frame de telemetría en vivo desde el dron (vía MQTT).
type Telemetry struct {
	Position    GeoPoint  `json:"position"`
	BatteryPct  int       `json:"battery_pct"`
	WindMS      float64   `json:"wind_ms"`
	GPSSats     int       `json:"gps_sats"`
	GroundSpeed float64   `json:"ground_speed"`
	Heading     float64   `json:"heading"`
	Timestamp   time.Time `json:"timestamp"`
}

// PilotDecision es una decisión tomada por el piloto autónomo, registrada
// para auditoría (igual que la caja negra de un avión).
type PilotDecision struct {
	At       time.Time   `json:"at"`
	Phase    FlightPhase `json:"phase"`
	Decision string      `json:"decision"`
	Reason   string      `json:"reason"`
	Critical bool        `json:"critical"`
}

// FlightController es el piloto autónomo: ejecuta el plan y decide en vivo.
type FlightController struct {
	mu        sync.RWMutex
	missionID string
	droneID   string
	phase     FlightPhase
	plan      MissionPlan
	env       FlightEnvelope
	fence     *Geofence
	home      GeoPoint
	currentWP int
	decisions []PilotDecision

	// Autoridad legal: el PIC humano puede tomar control en cualquier momento.
	picAuthorized bool
	picOverride   bool

	// Comandos hacia el dron (los implementa el adaptador DJI/MQTT).
	commander DroneCommander
	onEvent   func(PilotDecision)

	// Resiliencia de enlace / operación sin red
	linkMonitor       *LinkMonitor
	onboardMission    *OnboardMission
	navigator         *GNSSNavigator
	linkFailsafeFired bool
}

// DroneCommander abstrae el control real del dron (DJI SDK / MQTT).
// El controlador decide; el commander ejecuta en el hardware.
type DroneCommander interface {
	Arm(ctx context.Context) error
	Takeoff(ctx context.Context, alt float64) error
	GotoWaypoint(ctx context.Context, wp Waypoint) error
	SetGimbal(ctx context.Context, pitch float64) error
	CapturePhoto(ctx context.Context) error
	ReturnToHome(ctx context.Context) error
	Land(ctx context.Context) error
	EmergencyLand(ctx context.Context) error
	Hover(ctx context.Context) error
}

func NewFlightController(droneID, missionID string, plan MissionPlan, env FlightEnvelope, fence *Geofence, cmd DroneCommander) *FlightController {
	return &FlightController{
		missionID: missionID, droneID: droneID, phase: PhaseIdle,
		plan: plan, env: env, fence: fence, home: plan.Home,
		commander: cmd, decisions: []PilotDecision{},
		linkMonitor: NewLinkMonitor(),
		navigator:   NewGNSSNavigator(plan.Home),
	}
}

// LoadOnboardMission carga la misión completa al dron ANTES de despegar.
// Sin esto confirmado, el vuelo en zona sin red no se autoriza.
func (fc *FlightController) LoadOnboardMission(m *OnboardMission) {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	fc.onboardMission = m
	fc.log("misión cargada a bordo", "ruta, geocerca y política de enlace en el dron — listo para operar sin red", false)
}

// LinkStatus expone el estado del enlace y segundos sin contacto (bitácora).
func (fc *FlightController) LinkStatus() (LinkState, float64) {
	return fc.linkMonitor.State(), fc.linkMonitor.SecondsSinceContact()
}

func (fc *FlightController) OnEvent(f func(PilotDecision)) { fc.onEvent = f }

func (fc *FlightController) log(decision, reason string, critical bool) {
	d := PilotDecision{At: time.Now(), Phase: fc.phase, Decision: decision, Reason: reason, Critical: critical}
	fc.decisions = append(fc.decisions, d)
	if fc.onEvent != nil {
		fc.onEvent(d)
	}
}

func (fc *FlightController) Phase() FlightPhase {
	fc.mu.RLock()
	defer fc.mu.RUnlock()
	return fc.phase
}

func (fc *FlightController) Decisions() []PilotDecision {
	fc.mu.RLock()
	defer fc.mu.RUnlock()
	out := make([]PilotDecision, len(fc.decisions))
	copy(out, fc.decisions)
	return out
}

// AuthorizePIC: el piloto al mando humano autoriza la misión autónoma.
// Requisito legal Aerocivil: hay un PIC responsable que habilita el vuelo.
func (fc *FlightController) AuthorizePIC() {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	fc.picAuthorized = true
	fc.log("autorización PIC recibida", "piloto al mando habilitó la misión", false)
}

// TakeControl: el PIC interrumpe la autonomía y toma control manual.
func (fc *FlightController) TakeControl() {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	fc.picOverride = true
	fc.log("PIC toma control manual", "override humano activado", true)
}

// Start ejecuta la misión completa de forma autónoma.
func (fc *FlightController) Start(ctx context.Context, pre PreflightResult) error {
	fc.mu.Lock()
	fc.phase = PhasePreflight
	fc.mu.Unlock()

	// 1. Pre-vuelo
	if !pre.Cleared {
		fc.transition(PhaseAborted)
		fc.log("vuelo abortado en pre-vuelo", pre.Reason, true)
		return fmt.Errorf("pre-vuelo no superado: %s", pre.Reason)
	}
	fc.log("pre-vuelo superado", "todas las verificaciones críticas OK", false)

	// 2. Autorización del PIC (legal)
	fc.transition(PhaseAuthWait)
	if !fc.picAuthorized {
		fc.log("esperando autorización PIC", "la misión no inicia sin piloto al mando", false)
		return fmt.Errorf("pendiente de autorización del PIC")
	}

	// 3. Armado y despegue
	fc.transition(PhaseArmed)
	if err := fc.commander.Arm(ctx); err != nil {
		return fc.abort(ctx, "fallo al armar: "+err.Error())
	}
	fc.transition(PhaseTakeoff)
	fc.log("despegue", "ascenso a altitud de crucero", false)
	if err := fc.commander.Takeoff(ctx, fc.plan.CruiseAlt); err != nil {
		return fc.abort(ctx, "fallo en despegue: "+err.Error())
	}

	// 4. Ejecutar waypoints
	for i, wp := range fc.plan.Waypoints {
		if fc.isOverridden() {
			fc.log("misión cedida al PIC", "control manual en curso", true)
			return nil
		}
		fc.mu.Lock()
		fc.currentWP = i
		fc.mu.Unlock()

		switch wp.Action {
		case "hover_capture":
			fc.transition(PhaseInspecting)
			fc.commander.GotoWaypoint(ctx, wp)
			fc.commander.SetGimbal(ctx, wp.GimbalPitch)
			fc.commander.CapturePhoto(ctx)
			fc.log(fmt.Sprintf("inspección punto %s", wp.TargetID), "captura para análisis de visión IA", false)
		default:
			fc.transition(PhaseEnroute)
			fc.commander.GotoWaypoint(ctx, wp)
		}
	}

	// 5. RTH planificado y aterrizaje
	fc.transition(PhaseRTH)
	fc.log("regreso a casa", "misión completa, RTH planificado", false)
	fc.commander.ReturnToHome(ctx)
	fc.transition(PhaseLanding)
	fc.commander.Land(ctx)
	fc.transition(PhaseLanded)
	fc.log("aterrizaje completado", "misión finalizada con éxito", false)
	return nil
}

// OnTelemetry es el "cerebro en vivo": corre en cada frame de telemetría y
// decide como un piloto vigilante. Es la pieza que hace que el sistema vuele
// con criterio profesional, no solo siga waypoints ciegamente.
func (fc *FlightController) OnTelemetry(ctx context.Context, t Telemetry) {
	// Cada frame recibido = el enlace está vivo. El monitor lo registra.
	fc.linkMonitor.Heartbeat()

	fc.mu.RLock()
	phase := fc.phase
	overridden := fc.picOverride
	fc.mu.RUnlock()

	if overridden || phase == PhaseLanded || phase == PhaseIdle {
		return
	}

	verdict := fc.env.Evaluate(t.Position, fc.home, t.BatteryPct, t.WindMS, t.GPSSats, fc.fence)

	if verdict.MustLand {
		fc.transition(PhaseEmergency)
		fc.log("ATERRIZAJE DE EMERGENCIA", joinReasons(verdict.Violations), true)
		fc.commander.EmergencyLand(ctx)
		return
	}
	if verdict.MustRTH && phase != PhaseRTH && phase != PhaseLanding {
		fc.transition(PhaseRTH)
		fc.log("RTH AUTÓNOMO", "decisión de seguridad: "+joinReasons(verdict.Violations), true)
		fc.commander.ReturnToHome(ctx)
		return
	}
	if !verdict.Safe && len(verdict.Violations) > 0 {
		// Anomalía no fatal: el piloto reduce y vigila
		fc.commander.Hover(ctx)
		fc.log("hover de precaución", joinReasons(verdict.Violations), false)
	}
}

func (fc *FlightController) transition(p FlightPhase) {
	fc.mu.Lock()
	fc.phase = p
	fc.mu.Unlock()
}

func (fc *FlightController) isOverridden() bool {
	fc.mu.RLock()
	defer fc.mu.RUnlock()
	return fc.picOverride
}

func (fc *FlightController) abort(ctx context.Context, reason string) error {
	fc.transition(PhaseAborted)
	fc.log("misión abortada", reason, true)
	fc.commander.ReturnToHome(ctx)
	return fmt.Errorf("%s", reason)
}

func joinReasons(rs []string) string {
	out := ""
	for i, r := range rs {
		if i > 0 {
			out += "; "
		}
		out += r
	}
	return out
}

// EvaluateOnboardLinkLoss representa la lógica de failsafe que corre A BORDO del
// dron (no en el servidor). Se invoca desde el bucle de control del aparato con
// su propio reloj y GNSS. Es lo que garantiza que el dron NO se pierde ni se
// queda inerte cuando cae la red: decide con la política cargada antes del vuelo.
//
// En el despliegue real, esta lógica se ejecuta en el Onboard SDK / firmware del
// dron. Aquí la modelamos y probamos para que el comportamiento sea el correcto.
func (fc *FlightController) EvaluateOnboardLinkLoss(ctx context.Context, fix GNSSFix) *LinkLossDecision {
	state := fc.linkMonitor.State()
	if state != LinkLost {
		return nil // enlace OK o intermitente: el dron sigue su misión onboard
	}

	fc.mu.Lock()
	if fc.linkFailsafeFired {
		fc.mu.Unlock()
		return nil // ya se actuó
	}
	fc.linkFailsafeFired = true
	mission := fc.onboardMission
	fc.mu.Unlock()

	// Sin fix GNSS suficiente, el retorno por coordenadas no es confiable → aterrizar.
	if ok, _ := GNSSHealthy(fix, fc.env.MinGPSSatellites); !ok {
		fc.log("FAILSAFE: enlace y GNSS degradados", "aterrizaje controlado en sitio", true)
		fc.commander.EmergencyLand(ctx)
		return &LinkLossDecision{Action: ActionLandInPlace, NavMode: "gnss_onboard",
			Reason: "enlace perdido y fix GNSS insuficiente para RTH seguro"}
	}

	policy := DefaultLinkLossPolicy(fc.home, fc.plan.CruiseAlt)
	if mission != nil {
		policy = mission.LinkLoss
	}
	decision := policy.OnLinkLost()

	// El dron ejecuta la política con su GNSS, sin red.
	switch decision.Action {
	case ActionContinueMission:
		// Sigue la ruta onboard; al terminar hará RTH por coordenadas.
		fc.log("ENLACE PERDIDO — continúa misión onboard", decision.Reason, true)
	case ActionReturnToHome:
		fc.transition(PhaseRTH)
		fc.commander.ReturnToHome(ctx)
		bearing, dist := fc.navigator.NavigateHome(GeoPoint{Lat: fix.Latitude, Lng: fix.Longitude, Alt: fix.Altitude})
		fc.log("ENLACE PERDIDO — RTH por GNSS",
			joinReasons([]string{decision.Reason, fmt.Sprintf("rumbo %.0f°, %.0fm a home", bearing, dist)}), true)
	case ActionHoverThenRTH:
		fc.commander.Hover(ctx)
		fc.log("ENLACE PERDIDO — hover y luego RTH", decision.Reason, true)
	default:
		fc.commander.EmergencyLand(ctx)
		fc.log("ENLACE PERDIDO — aterrizaje en sitio", decision.Reason, true)
	}
	return &decision
}
