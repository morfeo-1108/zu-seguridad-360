package model

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type DroneStatus string
type MissionStatus string

const (
	DroneStatusAvailable   DroneStatus = "available"
	DroneStatusMission     DroneStatus = "in_mission"
	DroneStatusCharging    DroneStatus = "charging"
	DroneStatusMaintenance DroneStatus = "maintenance"
	DroneStatusOffline     DroneStatus = "offline"

	MissionStatusPending   MissionStatus = "pending"
	MissionStatusActive    MissionStatus = "active"
	MissionStatusCompleted MissionStatus = "completed"
	MissionStatusFailed    MissionStatus = "failed"
)

type Drone struct {
	ID              uuid.UUID              `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantID        uuid.UUID              `gorm:"type:uuid;not null" json:"tenant_id"`
	SerialNumber    string                 `gorm:"uniqueIndex;size:50" json:"serial_number"`
	Model           string                 `gorm:"size:100" json:"model"` // DJI Mavic 3E, DJI M30T, etc.
	FirmwareVer     string                 `gorm:"size:50" json:"firmware_version"`
	AssignedCompany *uuid.UUID             `gorm:"type:uuid" json:"assigned_company"`
	Status          DroneStatus            `gorm:"default:'available'" json:"status"`
	BatteryLevel    int                    `gorm:"default:100" json:"battery_level"`
	GPSLat          float64                `json:"gps_lat"`
	GPSLng          float64                `json:"gps_lng"`
	GPSAlt          float64                `json:"gps_alt"`
	LastHeartbeat   *time.Time             `json:"last_heartbeat"`
	TotalFlightHrs  float64                `gorm:"default:0" json:"total_flight_hours"`
	NextMaintenance *time.Time             `json:"next_maintenance"`
	Metadata        map[string]interface{} `gorm:"serializer:json" json:"metadata"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
	DeletedAt       gorm.DeletedAt         `gorm:"index" json:"-"`
}

type DroneInspectionReport struct {
	ID           uuid.UUID       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	DroneID      uuid.UUID       `gorm:"type:uuid" json:"drone_id"`
	CompanyID    uuid.UUID       `gorm:"type:uuid" json:"company_id"`
	MissionID    uuid.UUID       `gorm:"type:uuid" json:"mission_id"`
	EmployeeID   *uuid.UUID      `gorm:"type:uuid" json:"employee_id"`
	Timestamp    time.Time       `json:"timestamp"`
	GPSLat       float64         `json:"gps_lat"`
	GPSLng       float64         `json:"gps_lng"`
	EPPCompliant bool            `json:"epp_compliant"`
	EPPDetails   map[string]bool `gorm:"serializer:json" json:"epp_details"` // {helmet: true, vest: false}
	PostureScore float64         `json:"posture_score"`
	Detections   []string        `gorm:"serializer:json" json:"detections"` // fall, smoke, fire, crowd
	PhotoURL     string          `gorm:"size:500" json:"photo_url"`
	VideoClipURL string          `gorm:"size:500" json:"video_clip_url"`
	AlertSent    bool            `gorm:"default:false" json:"alert_sent"`
	ReviewedBy   *uuid.UUID      `gorm:"type:uuid" json:"reviewed_by"`
	CreatedAt    time.Time       `json:"created_at"`
}

func (Drone) TableName() string                 { return "drones" }
func (DroneInspectionReport) TableName() string { return "drone_inspection_reports" }
