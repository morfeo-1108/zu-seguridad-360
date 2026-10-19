package model

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type EmployeeStatus string
type ContractType string

const (
	EmployeeStatusActive     EmployeeStatus = "active"
	EmployeeStatusInactive   EmployeeStatus = "inactive"
	EmployeeStatusOnLeave    EmployeeStatus = "on_leave"
	EmployeeStatusContractor EmployeeStatus = "contractor"

	ContractTypeIndefinite ContractType = "indefinite"
	ContractTypeFixed      ContractType = "fixed"
	ContractTypeService    ContractType = "service"
)

type Employee struct {
	ID                 uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantID           uuid.UUID      `gorm:"type:uuid;not null;index" json:"tenant_id"`
	CompanyID          uuid.UUID      `gorm:"type:uuid;not null;index" json:"company_id"`
	IsSynthetic        bool           `gorm:"not null;default:false;index" json:"is_synthetic"`
	FirstName          string         `gorm:"not null;size:100" json:"first_name"`
	LastName           string         `gorm:"not null;size:100" json:"last_name"`
	DocumentType       string         `gorm:"size:10" json:"document_type"`
	DocumentNumber     string         `gorm:"size:20" json:"document_number"`
	Email              string         `gorm:"size:255" json:"email"`
	Phone              string         `gorm:"size:20" json:"phone"`
	Position           string         `gorm:"size:255" json:"position"`
	Department         string         `gorm:"size:100" json:"department"`
	RiskProfile        string         `gorm:"size:50" json:"risk_profile"`
	ContractType       ContractType   `gorm:"size:20" json:"contract_type"`
	StartDate          *time.Time     `json:"start_date"`
	Status             EmployeeStatus `gorm:"default:'active'" json:"status"`
	FaceEncodingURL    string         `gorm:"size:500" json:"face_encoding_url"`
	ActiveRestrictions JSONBArray     `gorm:"type:jsonb;default:'[]'" json:"active_restrictions"`
	EPPRequired        JSONBArray     `gorm:"type:jsonb;default:'[]'" json:"epp_required"`
	Metadata           JSONB          `gorm:"type:jsonb" json:"metadata"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Company Company `gorm:"foreignKey:CompanyID" json:"company,omitempty"`
}

func (Employee) TableName() string { return "employees" }

func (e *Employee) FullName() string {
	return e.FirstName + " " + e.LastName
}
