package flight

import (
	"fmt"
	"math"
)

// ── NAVEGACIÓN GNSS ONBOARD — El dron no se pierde sin red ──
//
// La misión completa se "sube" al controlador de vuelo del dron ANTES de despegar.
// A partir de ahí, el dron navega por GNSS (constelaciones satelitales) de forma
// autónoma: cada waypoint es una coordenada absoluta (lat/lng/alt) que el aparato
// alcanza con su propio GPS, sin necesidad de la red. El retorno al punto de
// despliegue también es por coordenada GNSS. Así opera en zonas sin cobertura.

// OnboardMission es el paquete autónomo cargado al dron antes del vuelo.
// Contiene TODO lo necesario para completar el trabajo y volver sin red.
type OnboardMission struct {
	MissionID     string         `json:"mission_id"`
	HomePoint     GeoPoint       `json:"home_point"`       // punto de despliegue (RTH)
	Waypoints     []Waypoint     `json:"waypoints"`        // ruta absoluta por coordenadas
	Geofence      *Geofence      `json:"geofence"`         // límites cargados a bordo
	LinkLoss      LinkLossPolicy `json:"link_loss_policy"` // qué hacer si pierde señal
	RTHAltitude   float64        `json:"rth_altitude"`
	MinSatellites int            `json:"min_satellites"` // fix mínimo para operar
	Uploaded      bool           `json:"uploaded"`       // confirmación de carga a bordo
}

// GNSSConstellation: las constelaciones satelitales que da redundancia de posición.
type GNSSFix struct {
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
	Altitude   float64 `json:"altitude"`
	Satellites int     `json:"satellites"`
	HDOP       float64 `json:"hdop"` // dilución horizontal de precisión (menor = mejor)
	Fix3D      bool    `json:"fix_3d"`
}

// BuildOnboardMission empaqueta la misión para carga al dron. Se llama ANTES de
// despegar; si la carga (Uploaded) no se confirma, el vuelo no inicia.
func BuildOnboardMission(missionID string, plan MissionPlan, fence *Geofence, home GeoPoint, env FlightEnvelope) OnboardMission {
	return OnboardMission{
		MissionID:     missionID,
		HomePoint:     home,
		Waypoints:     plan.Waypoints,
		Geofence:      fence,
		LinkLoss:      DefaultLinkLossPolicy(home, plan.CruiseAlt),
		RTHAltitude:   plan.CruiseAlt,
		MinSatellites: env.MinGPSSatellites,
		Uploaded:      false,
	}
}

// ValidateForOfflineOps verifica que la misión puede ejecutarse SIN red.
// Un piloto no despega en zona sin cobertura si algo de esto falla.
func (m *OnboardMission) ValidateForOfflineOps() (bool, []string) {
	var problems []string
	if len(m.Waypoints) == 0 {
		problems = append(problems, "sin waypoints cargados a bordo")
	}
	if m.HomePoint.Lat == 0 && m.HomePoint.Lng == 0 {
		problems = append(problems, "punto de despliegue (home) no fijado — RTH imposible")
	}
	if m.Geofence == nil || len(m.Geofence.Polygon) < 3 {
		problems = append(problems, "geocerca no cargada — límites no garantizados sin red")
	}
	if m.LinkLoss.Action == "" {
		problems = append(problems, "sin política de pérdida de enlace definida")
	}
	if !m.Uploaded {
		problems = append(problems, "misión no confirmada como cargada a bordo del dron")
	}
	return len(problems) == 0, problems
}

// GNSSNavigator ejecuta navegación por coordenadas a bordo, sin red.
// Calcula rumbo y distancia al siguiente waypoint u home usando solo GNSS.
type GNSSNavigator struct {
	home GeoPoint
}

func NewGNSSNavigator(home GeoPoint) *GNSSNavigator {
	return &GNSSNavigator{home: home}
}

// BearingTo calcula el rumbo (grados) desde una posición a un destino.
// Navegación pura por coordenadas — no requiere red, solo el fix GNSS.
func (n *GNSSNavigator) BearingTo(from, to GeoPoint) float64 {
	lat1 := from.Lat * math.Pi / 180
	lat2 := to.Lat * math.Pi / 180
	dLng := (to.Lng - from.Lng) * math.Pi / 180
	y := math.Sin(dLng) * math.Cos(lat2)
	x := math.Cos(lat1)*math.Sin(lat2) - math.Sin(lat1)*math.Cos(lat2)*math.Cos(dLng)
	bearing := math.Atan2(y, x) * 180 / math.Pi
	return math.Mod(bearing+360, 360)
}

// DistanceTo: distancia en metros al destino (Haversine, ya probada).
func (n *GNSSNavigator) DistanceTo(from, to GeoPoint) float64 {
	return Haversine(from, to)
}

// NavigateHome devuelve rumbo y distancia al punto de despliegue por GNSS.
// Es el cálculo que permite el retorno sin red: el dron siempre sabe dónde está
// (por satélite) y dónde está su home (coordenada fija), así que puede volver.
func (n *GNSSNavigator) NavigateHome(current GeoPoint) (bearing, distanceM float64) {
	return n.BearingTo(current, n.home), n.DistanceTo(current, n.home)
}

// GNSSHealthy verifica que el fix satelital es suficiente para navegar sin red.
func GNSSHealthy(fix GNSSFix, minSats int) (bool, string) {
	if !fix.Fix3D {
		return false, "sin fix 3D — navegación por coordenadas no confiable"
	}
	if fix.Satellites < minSats {
		return false, fmt.Sprintf("solo %d satélites (mínimo %d)", fix.Satellites, minSats)
	}
	if fix.HDOP > 2.5 {
		return false, fmt.Sprintf("HDOP %.1f alto — precisión degradada", fix.HDOP)
	}
	return true, "fix GNSS robusto para operación autónoma sin red"
}
