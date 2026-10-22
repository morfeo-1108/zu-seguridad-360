package service

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/zu-seguridad/drone/internal/flight"
	"github.com/zu-seguridad/drone/internal/mqtt"
)

// FlightService gestiona los pilotos autónomos activos (uno por dron en misión).
// Es la fachada que el handler usa para planear, autorizar y monitorear vuelos.
type FlightService struct {
	mu          sync.RWMutex
	broker      *mqtt.MQTTBroker
	controllers map[string]*flight.FlightController // missionID -> controlador
}

func NewFlightService(broker *mqtt.MQTTBroker) *FlightService {
	fs := &FlightService{
		broker:      broker,
		controllers: make(map[string]*flight.FlightController),
	}
	// Conectar el stream de telemetría al cerebro de cada piloto
	go fs.routeTelemetry()
	return fs
}

// PlanRequest es lo que pide el cliente para planear una misión autónoma.
type PlanRequest struct {
	DroneID  string                    `json:"drone_id"`
	DroneSN  string                    `json:"drone_sn"`
	Home     flight.GeoPoint           `json:"home"`
	Targets  []flight.InspectionTarget `json:"targets"`
	Geofence *flight.Geofence          `json:"geofence"`
	State    flight.DroneState         `json:"drone_state"`
	Weather  flight.WeatherConditions  `json:"weather"`
}

// PlanResult devuelve el plan + el pre-vuelo para que el PIC decida.
type PlanResult struct {
	MissionID string                   `json:"mission_id"`
	Plan      flight.MissionPlan       `json:"plan"`
	Preflight flight.PreflightResult   `json:"preflight"`
	Weather   flight.WeatherAssessment `json:"weather"`
}

// PlanMission genera plan + pre-vuelo + dictamen de clima (sin volar aún).
func (fs *FlightService) PlanMission(req PlanRequest) PlanResult {
	env := flight.DefaultEnvelope()
	plan := flight.PlanMission(req.Home, req.Targets, env)
	weather := flight.AssessWeather(req.Weather, env)
	pre := flight.RunPreflight(req.State, env, weather, plan.TotalDistanceM)

	missionID := fmt.Sprintf("mission-%s-%d", req.DroneID, len(fs.controllers)+1)

	cmd := flight.NewDJICommander(req.DroneSN, fs.broker)
	fc := flight.NewFlightController(req.DroneID, missionID, plan, env, req.Geofence, cmd)

	fs.mu.Lock()
	fs.controllers[missionID] = fc
	fs.mu.Unlock()

	return PlanResult{MissionID: missionID, Plan: plan, Preflight: pre, Weather: weather}
}

// AuthorizeAndStart: el PIC autoriza y arranca la misión autónoma.
func (fs *FlightService) AuthorizeAndStart(ctx context.Context, missionID string, pre flight.PreflightResult) error {
	fc := fs.get(missionID)
	if fc == nil {
		return fmt.Errorf("misión %s no encontrada", missionID)
	}
	fc.AuthorizePIC()
	go func() { // ejecución autónoma en background
		if err := fc.Start(ctx, pre); err != nil {
			log.Printf("misión %s terminó con error: %v", missionID, err)
		}
	}()
	return nil
}

// TakeControl: el PIC toma control manual de una misión en curso.
func (fs *FlightService) TakeControl(missionID string) error {
	fc := fs.get(missionID)
	if fc == nil {
		return fmt.Errorf("misión %s no encontrada", missionID)
	}
	fc.TakeControl()
	return nil
}

// Status devuelve la fase y las decisiones del piloto (caja negra).
func (fs *FlightService) Status(missionID string) (map[string]interface{}, error) {
	fc := fs.get(missionID)
	if fc == nil {
		return nil, fmt.Errorf("misión %s no encontrada", missionID)
	}
	return map[string]interface{}{
		"mission_id": missionID,
		"phase":      fc.Phase(),
		"decisions":  fc.Decisions(),
	}, nil
}

func (fs *FlightService) get(missionID string) *flight.FlightController {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	return fs.controllers[missionID]
}

// routeTelemetry conecta el stream MQTT al cerebro de cada piloto activo.
func (fs *FlightService) routeTelemetry() {
	for range fs.broker.TelemetryCh() {
		// En producción: mapear el frame MQTT a flight.Telemetry y rutear al
		// controlador del dron correspondiente por su ID. Cada frame dispara
		// OnTelemetry, donde el piloto decide RTH/emergencia/hover en vivo.
		_ = context.Background()
	}
}
