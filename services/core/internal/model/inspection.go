package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type InspectionStatus string

const (
	InspectionStatusCompleted InspectionStatus = "completada"
	InspectionStatusApproved  InspectionStatus = "aprobada"
)

// Inspection records a structured checklist-based safety inspection.
// Checklist and ActionPlan are JSON arrays of objects
// ({code, description, compliant, notes} and
// {item_code, description, responsible, due_date, status} respectively) —
// kept as flexible JSON rather than normalized tables since checklists vary
// per inspection type and this mirrors the project's existing JSONBArray
// convention (evidence, risk controls).
type Inspection struct {
	ID            uuid.UUID        `gorm:"type:uuid;primaryKey" json:"id"`
	TenantID      uuid.UUID        `gorm:"type:uuid;not null;index" json:"tenant_id"`
	CompanyID     uuid.UUID        `gorm:"type:uuid;not null;index" json:"company_id"`
	Title         string           `gorm:"not null;size:255" json:"title"`
	Area          string           `gorm:"size:255" json:"area"`
	InspectorID   uuid.UUID        `gorm:"type:uuid;not null" json:"inspector_id"`
	PerformedAt   time.Time        `json:"performed_at"`
	Checklist     JSONBArray       `gorm:"type:jsonb;not null;default:'[]'" json:"checklist"`
	CompliancePct float64          `json:"compliance_pct"`
	Status        InspectionStatus `gorm:"size:20;default:'completada'" json:"status"`
	ApprovedBy    *uuid.UUID       `gorm:"type:uuid" json:"approved_by"`
	ApprovedAt    *time.Time       `json:"approved_at"`
	ActionPlan    JSONBArray       `gorm:"type:jsonb;default:'[]'" json:"action_plan"`
	CreatedAt     time.Time        `json:"created_at"`
	UpdatedAt     time.Time        `json:"updated_at"`
	DeletedAt     gorm.DeletedAt   `gorm:"index" json:"-"`
}

func (Inspection) TableName() string { return "inspections" }

func (i *Inspection) BeforeCreate(tx *gorm.DB) error {
	if i.ID == uuid.Nil {
		i.ID = uuid.New()
	}
	return nil
}
