package model

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type CertificationType string

const (
	CertAlturas      CertificationType = "trabajo_alturas"
	CertConfinados   CertificationType = "espacios_confinados"
	CertPrimAuxilios CertificationType = "primeros_auxilios"
	CertExtintores   CertificationType = "extintores"
	CertElectrico    CertificationType = "riesgo_electrico"
	CertQuimicos     CertificationType = "quimicos_peligrosos"
	CertInduccion    CertificationType = "induccion_sst"
)

type Certification struct {
	ID             uuid.UUID         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantID       uuid.UUID         `gorm:"type:uuid;not null;index" json:"tenant_id"`
	CompanyID      uuid.UUID         `gorm:"type:uuid;not null;index" json:"company_id"`
	EmployeeID     uuid.UUID         `gorm:"type:uuid;not null;index" json:"employee_id"`
	Type           CertificationType `gorm:"size:50" json:"type"`
	FolioUUID      string            `gorm:"uniqueIndex;size:50" json:"folio_uuid"`
	IssuedAt       time.Time         `json:"issued_at"`
	ExpiresAt      time.Time         `json:"expires_at"`
	IssuedBy       string            `gorm:"size:255" json:"issued_by"`
	TrainingID     *uuid.UUID        `gorm:"type:uuid" json:"training_id"`
	CertificateURL string            `gorm:"size:500" json:"certificate_url"`
	QRCode         string            `gorm:"size:1000" json:"qr_code"`
	IsActive       bool              `gorm:"default:true" json:"is_active"`
	NormReference  string            `gorm:"size:200" json:"norm_reference"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
	DeletedAt      gorm.DeletedAt    `gorm:"index" json:"-"`
}

func (Certification) TableName() string { return "certifications" }
