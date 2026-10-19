package model

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type TrainingStatus string

const (
	TrainingStatusScheduled  TrainingStatus = "scheduled"
	TrainingStatusInProgress TrainingStatus = "in_progress"
	TrainingStatusCompleted  TrainingStatus = "completed"
	TrainingStatusCancelled  TrainingStatus = "cancelled"
)

type Training struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantID       uuid.UUID      `gorm:"type:uuid;not null;index" json:"tenant_id"`
	CompanyID      uuid.UUID      `gorm:"type:uuid;not null;index" json:"company_id"`
	IsSynthetic    bool           `gorm:"not null;default:false;index" json:"is_synthetic"`
	Title          string         `gorm:"not null;size:255" json:"title"`
	NormRef        string         `gorm:"size:100" json:"norm_ref"` // Populate only after normative review.
	Category       string         `gorm:"size:100" json:"category"`
	DurationHours  float64        `gorm:"default:8" json:"duration_hours"`
	ValidityMonths int            `gorm:"default:12" json:"validity_months"`
	InstructorName string         `gorm:"size:255" json:"instructor_name"`
	ScheduledAt    *time.Time     `json:"scheduled_at"`
	CompletedAt    *time.Time     `json:"completed_at"`
	Status         TrainingStatus `gorm:"default:'scheduled'" json:"status"`
	AttendeeCount  int            `gorm:"default:0" json:"attendee_count"`
	PassCount      int            `gorm:"default:0" json:"pass_count"`
	ContentURL     string         `gorm:"size:500" json:"content_url"`
	EvaluationMin  float64        `gorm:"default:70" json:"evaluation_min_score"`
	Metadata       JSONB          `gorm:"type:jsonb" json:"metadata"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

type Certificate struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"tenant_id"`
	TrainingID  uuid.UUID  `gorm:"type:uuid;not null" json:"training_id"`
	EmployeeID  uuid.UUID  `gorm:"type:uuid;not null;index" json:"employee_id"`
	CompanyID   uuid.UUID  `gorm:"type:uuid;not null;index" json:"company_id"`
	IsSynthetic bool       `gorm:"not null;default:false;index" json:"is_synthetic"`
	Folio       string     `gorm:"uniqueIndex;size:36" json:"folio"` // UUID v4
	Score       float64    `json:"score"`
	IssuedAt    time.Time  `json:"issued_at"`
	ExpiresAt   time.Time  `gorm:"index" json:"expires_at"`
	PDFURL      string     `gorm:"size:500" json:"pdf_url"`
	QRCode      string     `gorm:"size:500" json:"qr_code"`
	IsValid     bool       `gorm:"default:true" json:"is_valid"`
	RevokedAt   *time.Time `json:"revoked_at"`
	CreatedAt   time.Time  `json:"created_at"`

	Training Training `gorm:"foreignKey:TrainingID" json:"training,omitempty"`
	Employee Employee `gorm:"foreignKey:EmployeeID" json:"employee,omitempty"`
}

func (Training) TableName() string    { return "trainings" }
func (Certificate) TableName() string { return "certificates" }
