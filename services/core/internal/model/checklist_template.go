package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ChecklistTemplate is a reusable list of checklist items (just code +
// description, no compliant/notes — those only make sense once a specific
// inspection is performed) that a professional can save once and apply to
// many inspections, instead of retyping the same checklist every time.
// Items shape: {code, description}.
type ChecklistTemplate struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	TenantID    uuid.UUID      `gorm:"type:uuid;not null;index" json:"tenant_id"`
	CompanyID   uuid.UUID      `gorm:"type:uuid;not null;index" json:"company_id"`
	Name        string         `gorm:"not null;size:255" json:"name"`
	Category    string         `gorm:"size:100" json:"category"`
	Items       JSONBArray     `gorm:"type:jsonb;not null;default:'[]'" json:"items"`
	CreatedBy   uuid.UUID      `gorm:"type:uuid;not null" json:"created_by"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (ChecklistTemplate) TableName() string { return "checklist_templates" }

func (t *ChecklistTemplate) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}
