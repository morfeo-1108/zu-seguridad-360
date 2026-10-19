package model

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type RiskLevel string

const (
	RiskLevelAcceptable    RiskLevel = "aceptable"
	RiskLevelNotAcceptable RiskLevel = "no_aceptable"
	RiskLevelCritical      RiskLevel = "critico"
)

type RiskEntry struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantID       uuid.UUID `gorm:"type:uuid;not null;index" json:"tenant_id"`
	CompanyID      uuid.UUID `gorm:"type:uuid;not null;index" json:"company_id"`
	MatrixVersion  int       `gorm:"default:1" json:"matrix_version"`
	Process        string    `gorm:"size:255" json:"process"`
	Area           string    `gorm:"size:255" json:"area"`
	HazardType     string    `gorm:"size:100" json:"hazard_type"`
	HazardSource   string    `gorm:"size:500" json:"hazard_source"`
	HealthEffect   string    `gorm:"size:500" json:"health_effect"`
	WorkersExposed int       `json:"workers_exposed"`
	// GTC 45: ND×NE=NP; NP×NC=NR
	ND             int            `gorm:"column:nd" json:"nd"` // Nivel Deficiencia
	NE             int            `gorm:"column:ne" json:"ne"` // Nivel Exposición
	NP             int            `gorm:"column:np" json:"np"` // Nivel Probabilidad = ND×NE
	NC             int            `gorm:"column:nc" json:"nc"` // Nivel Consecuencia
	NR             int            `gorm:"column:nr" json:"nr"` // Nivel Riesgo = NP×NC
	Acceptability  RiskLevel      `gorm:"size:20" json:"acceptability"`
	Controls       JSONB          `gorm:"type:jsonb;default:'[]'" json:"controls"`
	CurrentControl string         `gorm:"size:500" json:"current_control"`
	ActionPlan     JSONB          `gorm:"type:jsonb" json:"action_plan"`
	AssessedBy     *uuid.UUID     `gorm:"type:uuid" json:"assessed_by"`
	AssessedAt     *time.Time     `json:"assessed_at"`
	NextReviewAt   *time.Time     `json:"next_review_at"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

func (RiskEntry) TableName() string { return "risk_entries" }

// CalculateNR aplica metodología GTC 45
func (r *RiskEntry) CalculateNR() {
	r.NP = r.ND * r.NE
	r.NR = r.NP * r.NC
	switch {
	case r.NR >= 600 || (r.NR >= 150 && r.NC == 100):
		r.Acceptability = RiskLevelCritical
	case r.NR >= 70:
		r.Acceptability = RiskLevelNotAcceptable
	default:
		r.Acceptability = RiskLevelAcceptable
	}
}
