package flight

import (
	"time"
)

// ── RESILIENCIA DE ENLACE — Operación sin red / señal intermitente ──
//
// PRINCIPIO DE INGENIERÍA (cómo vuela un piloto profesional en zona sin cobertura):
// El dron NO depende del enlace de red para completar su trabajo. La misión se
// carga COMPLETA a bordo antes de despegar; el dron la ejecuta de forma autónoma
// usando su GNSS (GPS/GLONASS/Galileo/BeiDou) aunque pierda el enlace de control.
// La red solo transporta telemetría y comandos de conveniencia — nunca es un
// punto único de falla para la seguridad ni para el retorno.
//
// Esto refleja el comportamiento real de un DJI en modo waypoint: la ruta vive
// en el controlador de vuelo del aparato, no en el servidor.

// LinkState describe el estado del enlace de control/datos.
type LinkState string

const (
	LinkOK           LinkState = "ok"           // enlace estable
	LinkIntermittent LinkState = "intermittent" // señal débil/entrecortada
	LinkLost         LinkState = "lost"         // enlace perdido (timeout superado)
)

// LinkMonitor vigila la salud del enlace por heartbeat/última telemetría.
// Cuando el enlace se degrada, el dron NO se detiene: sigue su misión onboard.
type LinkMonitor struct {
	lastHeartbeat     time.Time
	intermittentAfter time.Duration // sin heartbeat → intermitente
	lostAfter         time.Duration // sin heartbeat → perdido (dispara failsafe)
}

func NewLinkMonitor() *LinkMonitor {
	return &LinkMonitor{
		lastHeartbeat:     time.Now(),
		intermittentAfter: 3 * time.Second,
		lostAfter:         15 * time.Second, // tolerancia antes de failsafe de enlace
	}
}

// Heartbeat registra que llegó telemetría (el enlace está vivo).
func (lm *LinkMonitor) Heartbeat() { lm.lastHeartbeat = time.Now() }

// State evalúa el estado del enlace según el tiempo sin heartbeat.
func (lm *LinkMonitor) State() LinkState {
	elapsed := time.Since(lm.lastHeartbeat)
	switch {
	case elapsed >= lm.lostAfter:
		return LinkLost
	case elapsed >= lm.intermittentAfter:
		return LinkIntermittent
	default:
		return LinkOK
	}
}

// SecondsSinceContact expone cuánto lleva sin enlace (para bitácora/decisión).
func (lm *LinkMonitor) SecondsSinceContact() float64 {
	return time.Since(lm.lastHeartbeat).Seconds()
}

// ── ESCALERA DE FAILSAFES POR PÉRDIDA DE ENLACE ──
//
// Un piloto profesional configura QUÉ hace el dron si pierde señal, ANTES de
// volar. La política se carga a bordo. Opciones en orden de autonomía:

// LinkLossAction es la política de failsafe ante pérdida de enlace.
type LinkLossAction string

const (
	// ContinueMission: el dron TERMINA su misión cargada onboard y luego RTH.
	// Ideal en zona sin cobertura: la falta de red no interrumpe el trabajo.
	ActionContinueMission LinkLossAction = "continue_mission"
	// ReturnToHome: regresa inmediatamente al punto de despliegue por GNSS.
	ActionReturnToHome LinkLossAction = "return_to_home"
	// Hover: mantiene posición un tiempo y luego RTH si no recupera enlace.
	ActionHoverThenRTH LinkLossAction = "hover_then_rth"
	// Land: aterriza en el sitio (solo si RTH no es seguro).
	ActionLandInPlace LinkLossAction = "land_in_place"
)

// LinkLossPolicy define el comportamiento ante pérdida de enlace, cargado a bordo.
type LinkLossPolicy struct {
	Action             LinkLossAction `json:"action"`
	HoverSecsBeforeRTH int            `json:"hover_secs_before_rth"` // para hover_then_rth
	// Punto de despliegue (home) para RTH por coordenadas GNSS — SIEMPRE fijado.
	HomePoint   GeoPoint `json:"home_point"`
	RTHAltitude float64  `json:"rth_altitude"` // altura de crucero segura para volver
}

// DefaultLinkLossPolicy: la política conservadora recomendada para zona sin red.
// Termina la misión onboard (ya está a bordo) y regresa por GNSS — no se pierde.
func DefaultLinkLossPolicy(home GeoPoint, rthAlt float64) LinkLossPolicy {
	return LinkLossPolicy{
		Action:      ActionContinueMission,
		HomePoint:   home,
		RTHAltitude: rthAlt,
	}
}

// LinkLossDecision es lo que el dron hace al confirmarse pérdida de enlace.
type LinkLossDecision struct {
	Action  LinkLossAction `json:"action"`
	Reason  string         `json:"reason"`
	NavMode string         `json:"nav_mode"` // siempre "gnss_onboard": navega por satélite sin red
}

// OnLinkLost resuelve qué hacer cuando el enlace se pierde, según la política
// cargada a bordo. Clave: la navegación es SIEMPRE por GNSS onboard, sin depender
// de la red. El dron no se pierde porque su posición viene de los satélites.
func (p LinkLossPolicy) OnLinkLost() LinkLossDecision {
	d := LinkLossDecision{NavMode: "gnss_onboard"}
	switch p.Action {
	case ActionContinueMission:
		d.Action = ActionContinueMission
		d.Reason = "enlace perdido — completando misión cargada a bordo por GNSS, luego RTH"
	case ActionReturnToHome:
		d.Action = ActionReturnToHome
		d.Reason = "enlace perdido — RTH inmediato al punto de despliegue por coordenadas GNSS"
	case ActionHoverThenRTH:
		d.Action = ActionHoverThenRTH
		d.Reason = "enlace perdido — hover en espera de recuperación, luego RTH por GNSS"
	default:
		d.Action = ActionLandInPlace
		d.Reason = "enlace perdido y RTH no seguro — aterrizaje controlado en sitio"
	}
	return d
}
