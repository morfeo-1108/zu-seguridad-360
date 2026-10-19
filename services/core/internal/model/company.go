package model

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type CompanyStatus string
type RiskClass string

const (
	CompanyStatusActive    CompanyStatus = "active"
	CompanyStatusInactive  CompanyStatus = "inactive"
	CompanyStatusSuspended CompanyStatus = "suspended"

	RiskClassI   RiskClass = "I"
	RiskClassII  RiskClass = "II"
	RiskClassIII RiskClass = "III"
	RiskClassIV  RiskClass = "IV"
	RiskClassV   RiskClass = "V"
)

type Company struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantID       uuid.UUID      `gorm:"type:uuid;not null;index" json:"tenant_id"`
	IsSynthetic    bool           `gorm:"not null;default:false;index" json:"is_synthetic"`
	Name           string         `gorm:"not null;size:255" json:"name"`
	NIT            string         `gorm:"uniqueIndex;size:20" json:"nit"`
	CIIU           string         `gorm:"size:10" json:"ciiu"`
	RiskClass      RiskClass      `gorm:"size:5" json:"risk_class"`
	ARL            string         `gorm:"size:100" json:"arl"`
	IPS            string         `gorm:"size:100" json:"ips"`
	LegalRep       string         `gorm:"size:255" json:"legal_rep"`
	LegalRepEmail  string         `gorm:"size:255" json:"legal_rep_email"`
	LegalRepPhone  string         `gorm:"size:20" json:"legal_rep_phone"`
	Address        string         `gorm:"size:500" json:"address"`
	City           string         `gorm:"size:100" json:"city"`
	Department     string         `gorm:"size:100" json:"department"`
	EmployeeCount  int            `gorm:"default:0" json:"employee_count"`
	RiskScore      float64        `gorm:"default:0" json:"risk_score"`
	Status         CompanyStatus  `gorm:"default:'active'" json:"status"`
	PlanID         *uuid.UUID     `gorm:"type:uuid" json:"plan_id"`
	OnboardingStep int            `gorm:"default:0" json:"onboarding_step"`
	OnboardingDone bool           `gorm:"default:false" json:"onboarding_done"`
	DroneAssigned  *uuid.UUID     `gorm:"type:uuid" json:"drone_assigned"`
	RUTFileURL     string         `gorm:"size:500" json:"rut_file_url"`
	LogoURL        string         `gorm:"size:500" json:"logo_url"`
	Metadata       JSONB          `gorm:"type:jsonb" json:"metadata"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Employees []Employee `gorm:"foreignKey:CompanyID" json:"employees,omitempty"`
}

func (Company) TableName() string { return "companies" }

// JSONB helper
type JSONB map[string]interface{}

func (c *Company) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}
