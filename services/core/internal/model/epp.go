package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// EPPItem is a company's catalog entry for a type of personal protective
// equipment (helmet, harness, gloves, ...), with how often it must be renewed.
type EPPItem struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	TenantID     uuid.UUID      `gorm:"type:uuid;not null;index" json:"tenant_id"`
	CompanyID    uuid.UUID      `gorm:"type:uuid;not null;index" json:"company_id"`
	Name         string         `gorm:"not null;size:255" json:"name"`
	Category     string         `gorm:"size:100" json:"category"`
	LifespanDays int            `gorm:"not null" json:"lifespan_days"` // periodicidad de renovación
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (EPPItem) TableName() string { return "epp_items" }

func (i *EPPItem) BeforeCreate(tx *gorm.DB) error {
	if i.ID == uuid.Nil {
		i.ID = uuid.New()
	}
	return nil
}

// EPPDelivery records one delivery of an item to a worker. "Confirmed" is an
// audit trail of acknowledgment (who recorded it, when) — it is not a drawn
// or cryptographic signature; this system has no e-signature capture
// infrastructure, and faking one would misrepresent what actually happened.
type EPPDelivery struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	TenantID     uuid.UUID      `gorm:"type:uuid;not null;index" json:"tenant_id"`
	CompanyID    uuid.UUID      `gorm:"type:uuid;not null;index" json:"company_id"`
	EmployeeID   uuid.UUID      `gorm:"type:uuid;not null;index" json:"employee_id"`
	ItemID       uuid.UUID      `gorm:"type:uuid;not null;index" json:"item_id"`
	Quantity     int            `gorm:"not null;default:1" json:"quantity"`
	Condition    string         `gorm:"size:30;default:'nuevo'" json:"condition"`
	DeliveredAt  time.Time      `json:"delivered_at"`
	ExpiresAt    *time.Time     `gorm:"index" json:"expires_at"`
	ConfirmedBy  *uuid.UUID     `gorm:"type:uuid" json:"confirmed_by"`
	ConfirmedAt  *time.Time     `json:"confirmed_at"`
	ReturnedAt   *time.Time     `json:"returned_at"`
	ReturnReason string         `gorm:"size:255" json:"return_reason"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	Item EPPItem `gorm:"foreignKey:ItemID" json:"item,omitempty"`
}

func (EPPDelivery) TableName() string { return "epp_deliveries" }

func (d *EPPDelivery) BeforeCreate(tx *gorm.DB) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	return nil
}
