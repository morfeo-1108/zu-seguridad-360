package model

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type ExamType string
type AptitudeResult string

const (
	ExamTypeIngreso  ExamType = "ingreso"
	ExamTypePeriodic ExamType = "periodico"
	ExamTypeRetiro   ExamType = "retiro"

	AptitudeFit           AptitudeResult = "apto"
	AptitudeFitRestricted AptitudeResult = "apto_con_restriccion"
	AptitudeUnfit         AptitudeResult = "no_apto"
	AptitudePending       AptitudeResult = "pendiente_revision"
)

type MedicalExam struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantID       uuid.UUID      `gorm:"type:uuid;not null;index" json:"tenant_id"`
	CompanyID      uuid.UUID      `gorm:"type:uuid;not null;index" json:"company_id"`
	EmployeeID     uuid.UUID      `gorm:"type:uuid;not null;index" json:"employee_id"`
	Type           ExamType       `gorm:"size:20;not null" json:"type"`
	IPSName        string         `gorm:"size:255" json:"ips_name"`
	IPSLicense     string         `gorm:"size:50" json:"ips_license"`
	DoctorName     string         `gorm:"size:255" json:"doctor_name"`
	DoctorLicense  string         `gorm:"size:50" json:"doctor_license"`
	ScheduledAt    *time.Time     `json:"scheduled_at"`
	PerformedAt    *time.Time     `json:"performed_at"`
	ExpiresAt      *time.Time     `gorm:"index" json:"expires_at"`
	AptitudeResult AptitudeResult `gorm:"size:30;default:'pendiente_revision'" json:"aptitude_result"`
	Restrictions   JSONBArray     `gorm:"type:jsonb;default:'[]'" json:"restrictions"`
	OCRRawData     JSONB          `gorm:"type:jsonb" json:"ocr_raw_data"`
	OCRConfidence  float64        `json:"ocr_confidence"`
	DocumentURL    string         `gorm:"size:500" json:"document_url"`
	ReviewedBy     *uuid.UUID     `gorm:"type:uuid" json:"reviewed_by"`
	ReviewedAt     *time.Time     `json:"reviewed_at"`
	ReviewNotes    string         `gorm:"type:text" json:"review_notes"`
	// Storage reference only. No retention period is asserted by this model;
	// retention and legal-hold rules require an approved policy before use.
	S3WORMKey string         `gorm:"size:500" json:"s3_worm_key"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Employee Employee `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
}

func (MedicalExam) TableName() string { return "medical_exams" }
