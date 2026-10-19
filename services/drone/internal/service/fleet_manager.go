package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/zu-seguridad/drone/internal/model"
	"github.com/zu-seguridad/drone/internal/mqtt"
	"gorm.io/gorm"
	"log"
	"time"
)

// ErrNoDB indica que el servicio corre en modo degradado sin persistencia.
var ErrNoDB = errors.New("base de datos no disponible (modo degradado)")

type FleetManager struct {
	db     *gorm.DB
	broker *mqtt.MQTTBroker
	drones map[string]*model.Drone // cache en memoria
}

func NewFleetManager(db *gorm.DB, broker *mqtt.MQTTBroker) *FleetManager {
	fm := &FleetManager{db: db, broker: broker, drones: make(map[string]*model.Drone)}
	go fm.consumeTelemetry()
	go fm.consumeVisionEvents()
	go fm.healthMonitor()
	return fm
}

func (fm *FleetManager) AssignDroneToCompany(ctx context.Context, droneID, companyID uuid.UUID) error {
	if fm.db == nil {
		return ErrNoDB
	}
	return fm.db.WithContext(ctx).Model(&model.Drone{}).
		Where("id = ?", droneID).
		Updates(map[string]interface{}{"assigned_company": companyID, "status": model.DroneStatusAvailable}).Error
}

func (fm *FleetManager) ScheduleMission(ctx context.Context, companyID uuid.UUID, missionType string, waypoints []map[string]float64) (string, error) {
	if fm.db == nil {
		return "", ErrNoDB
	}
	// Encontrar dron disponible para la empresa
	var drone model.Drone
	if err := fm.db.WithContext(ctx).
		Where("assigned_company = ? AND status = 'available' AND battery_level >= 30", companyID).
		First(&drone).Error; err != nil {
		return "", fmt.Errorf("no hay drones disponibles para la empresa (batería mínima 30%%)")
	}

	missionID := uuid.New().String()

	// Enviar comando de misión via MQTT
	command := map[string]interface{}{
		"mission_id":   missionID,
		"mission_type": missionType,
		"waypoints":    waypoints,
		"company_id":   companyID.String(),
		"issued_at":    time.Now().UTC(),
		"config": map[string]interface{}{
			"altitude_m":            30,
			"speed_ms":              5,
			"camera_mode":           "continuous",
			"vision_ai":             true,
			"return_on_low_battery": 20,
		},
	}

	if err := fm.broker.SendMissionCommand(drone.ID.String(), command); err != nil {
		return "", fmt.Errorf("error enviando comando al dron: %v", err)
	}

	log.Printf("🚁 Misión %s despachada → Dron %s → Empresa %s", missionID, drone.SerialNumber, companyID)
	return missionID, nil
}

func (fm *FleetManager) GetFleetStatus(ctx context.Context, tenantID uuid.UUID) ([]model.Drone, error) {
	if fm.db == nil {
		return nil, ErrNoDB
	}
	var drones []model.Drone
	err := fm.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Find(&drones).Error
	return drones, err
}

func (fm *FleetManager) GetDroneByID(ctx context.Context, id uuid.UUID) (*model.Drone, error) {
	if fm.db == nil {
		return nil, ErrNoDB
	}
	var d model.Drone
	err := fm.db.WithContext(ctx).Where("id = ?", id).First(&d).Error
	return &d, err
}

// consumeTelemetry procesa telemetría en tiempo real via MQTT
func (fm *FleetManager) consumeTelemetry() {
	for telemetry := range fm.broker.TelemetryCh() {
		// Actualizar cache en memoria (no escribir a BD en cada telemetría — demasiado volumen)
		drone, ok := fm.drones[telemetry.DroneID]
		if !ok {
			drone = &model.Drone{}
			fm.drones[telemetry.DroneID] = drone
		}
		drone.BatteryLevel = telemetry.BatteryLevel
		drone.GPSLat = telemetry.GPSLat
		drone.GPSLng = telemetry.GPSLng
		drone.GPSAlt = telemetry.GPSAlt

		now := time.Now()
		drone.LastHeartbeat = &now

		// Alerta si batería baja
		if telemetry.BatteryLevel <= 15 {
			log.Printf("⚠️  BATERÍA BAJA: Dron %s — %d%%", telemetry.DroneID, telemetry.BatteryLevel)
		}
	}
}

// consumeVisionEvents procesa eventos de visión IA del dron
func (fm *FleetManager) consumeVisionEvents() {
	for event := range fm.broker.VisionCh() {
		log.Printf("👁️  Vision AI [%s] Confianza: %.0f%% → %s", event.EventType, event.Confidence*100, event.CompanyID)

		// Guardar reporte de inspección
		report := &model.DroneInspectionReport{
			CompanyID:  uuid.MustParse(event.CompanyID),
			Timestamp:  event.Timestamp,
			PhotoURL:   event.PhotoURL,
			Detections: []string{event.EventType},
		}

		// EPP violations: alerta inmediata
		if event.EventType == "epp_violation" {
			report.EPPCompliant = false
			report.AlertSent = true
			log.Printf("🚨 EPP VIOLATION detectado en empresa %s — Confianza: %.0f%%", event.CompanyID, event.Confidence*100)
		} else if event.EventType == "epp_compliant" {
			report.EPPCompliant = true
		} else if event.EventType == "fall_detected" || event.EventType == "fire" || event.EventType == "smoke" {
			log.Printf("🚨 EMERGENCIA detectada: %s en empresa %s", event.EventType, event.CompanyID)
		}

		_ = report // En producción: fm.db.Create(report)
	}
}

// healthMonitor verifica latidos cada 60s
func (fm *FleetManager) healthMonitor() {
	ticker := time.NewTicker(60 * time.Second)
	for range ticker.C {
		for id, drone := range fm.drones {
			if drone.LastHeartbeat == nil {
				continue
			}
			if time.Since(*drone.LastHeartbeat) > 2*time.Minute {
				log.Printf("⚠️  Dron %s sin señal desde %s", id, drone.LastHeartbeat.Format("15:04:05"))
			}
		}
	}
}

func (fm *FleetManager) publishAlert(exchange, key string, payload map[string]interface{}) {
	data, _ := json.Marshal(payload)
	log.Printf("Alert published: %s/%s — %s", exchange, key, string(data))
}

// GetInspectionReports devuelve los reportes de inspección de una empresa.
func (fm *FleetManager) GetInspectionReports(ctx context.Context, companyID uuid.UUID) ([]model.DroneInspectionReport, error) {
	if fm.db == nil {
		return nil, ErrNoDB
	}
	var reports []model.DroneInspectionReport
	err := fm.db.WithContext(ctx).
		Where("company_id = ?", companyID).
		Order("timestamp DESC").Limit(100).Find(&reports).Error
	return reports, err
}
