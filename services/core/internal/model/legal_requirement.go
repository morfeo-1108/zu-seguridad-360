package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LegalComplianceStatus string

const (
	LegalStatusCompliant    LegalComplianceStatus = "cumple"
	LegalStatusPartial      LegalComplianceStatus = "parcial"
	LegalStatusNonCompliant LegalComplianceStatus = "no_cumple"
	LegalStatusPending      LegalComplianceStatus = "pendiente_evaluacion"
)

// LegalRequirement is one entry in a company's legal compliance matrix —
// a specific law/decree/resolution article the company must track, distinct
// from the PHVA cycle's ComplianceStandard (which tracks ISO/Res.0312
// clauses inside the Planear-Hacer-Verificar-Actuar structure). This is the
// flatter "qué normas me aplican y en qué estoy" list competitors call the
// legal/requirements matrix.
type LegalRequirement struct {
	ID               uuid.UUID             `gorm:"type:uuid;primaryKey" json:"id"`
	TenantID         uuid.UUID             `gorm:"type:uuid;not null;index" json:"tenant_id"`
	CompanyID        uuid.UUID             `gorm:"type:uuid;not null;index" json:"company_id"`
	NormReference    string                `gorm:"not null;size:255" json:"norm_reference"` // "Decreto 1072 de 2015, Art. 2.2.4.6.8"
	Title            string                `gorm:"type:text;not null" json:"title"`
	// No gorm "default:true": the handler always sets this from the request
	// body (applicable can legitimately be false), and a default on a bool
	// column makes GORM silently overwrite an explicit false with it on INSERT.
	Applicable       bool                  `json:"applicable"`
	ComplianceStatus LegalComplianceStatus `gorm:"size:30;default:'pendiente_evaluacion'" json:"compliance_status"`
	EvidenceRef      string                `gorm:"type:text" json:"evidence_ref"`
	Notes            string                `gorm:"type:text" json:"notes"`
	ReviewedBy       *uuid.UUID            `gorm:"type:uuid" json:"reviewed_by"`
	ReviewedAt       *time.Time            `json:"reviewed_at"`
	CreatedAt        time.Time             `json:"created_at"`
	UpdatedAt        time.Time             `json:"updated_at"`
	DeletedAt        gorm.DeletedAt        `gorm:"index" json:"-"`
}

func (LegalRequirement) TableName() string { return "legal_requirements" }

func (r *LegalRequirement) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}
