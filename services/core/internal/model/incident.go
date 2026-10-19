package model

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type IncidentType string
type IncidentSeverity string
type IncidentStatus string

const (
	IncidentTypeAT       IncidentType = "accidente_trabajo"
	IncidentTypeIncident IncidentType = "incidente"
	IncidentTypeEL       IncidentType = "enfermedad_laboral"

	SeverityLeve     IncidentSeverity = "leve"
	SeverityModerado IncidentSeverity = "moderado"
	SeverityGrave    IncidentSeverity = "grave"
	SeverityFatal    IncidentSeverity = "fatal"

	IncidentStatusOpen        IncidentStatus = "open"
	IncidentStatusUnderReview IncidentStatus = "under_review"
	IncidentStatusClosed      IncidentStatus = "closed"
	IncidentStatusLegalHold   IncidentStatus = "legal_hold"
)

type Incident struct {
	ID                uuid.UUID        `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantID          uuid.UUID        `gorm:"type:uuid;not null;index" json:"tenant_id"`
	CompanyID         uuid.UUID        `gorm:"type:uuid;not null;index" json:"company_id"`
	IsSynthetic       bool             `gorm:"not null;default:false;index" json:"is_synthetic"`
	EmployeeID        *uuid.UUID       `gorm:"type:uuid;index" json:"employee_id"`
	ReportedBy        uuid.UUID        `gorm:"type:uuid" json:"reported_by"`
	Type              IncidentType     `gorm:"size:30;not null" json:"type"`
	Severity          IncidentSeverity `gorm:"size:20" json:"severity"`
	Status            IncidentStatus   `gorm:"size:20;default:'open'" json:"status"`
	OccurredAt        time.Time        `gorm:"not null" json:"occurred_at"`
	Location          string           `gorm:"size:500" json:"location"`
	GPSLat            *float64         `json:"gps_lat"`
	GPSLng            *float64         `json:"gps_lng"`
	Description       string           `gorm:"type:text;not null" json:"description"`
	InjuredCount      int              `gorm:"default:0" json:"injured_count"`
	InjuredBodyPart   string           `gorm:"size:100" json:"injured_body_part"`
	InjuryNature      string           `gorm:"size:100" json:"injury_nature"`
	CausalTree        JSONB            `gorm:"type:jsonb" json:"causal_tree"`
	CorrectiveActions []interface{}    `gorm:"type:jsonb;serializer:json;default:'[]'" json:"corrective_actions"`
	FURATData         JSONB            `gorm:"type:jsonb" json:"furat_data"`
	FURATSentAt       *time.Time       `json:"furat_sent_at"`
	ARLReportedAt     *time.Time       `json:"arl_reported_at"`
	EvidenceURLs      JSONBArray       `gorm:"type:jsonb;default:'[]'" json:"evidence_urls"`
	LegalHold         bool             `gorm:"default:false" json:"legal_hold"`
	InvestigatedBy    *uuid.UUID       `gorm:"type:uuid" json:"investigated_by"`
	InvestigatedAt    *time.Time       `json:"investigated_at"`
	ClosedBy          *uuid.UUID       `gorm:"type:uuid" json:"closed_by"`
	ClosedAt          *time.Time       `json:"closed_at"`
	CreatedAt         time.Time        `json:"created_at"`
	UpdatedAt         time.Time        `json:"updated_at"`
	DeletedAt         gorm.DeletedAt   `gorm:"index" json:"-"`

	Company  Company   `gorm:"foreignKey:CompanyID" json:"company,omitempty"`
	Employee *Employee `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
}

func (Incident) TableName() string { return "incidents" }
