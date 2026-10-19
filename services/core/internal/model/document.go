package model

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type DocumentStatus string
type DocumentFormat string

const (
	DocStatusDraft     DocumentStatus = "draft"
	DocStatusReview    DocumentStatus = "review"
	DocStatusSigning   DocumentStatus = "signing"
	DocStatusPublished DocumentStatus = "published"
	DocStatusArchived  DocumentStatus = "archived"

	FormatPDF   DocumentFormat = "pdf"
	FormatWord  DocumentFormat = "docx"
	FormatExcel DocumentFormat = "xlsx"
	FormatPPT   DocumentFormat = "pptx"
)

type Document struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantID    uuid.UUID      `gorm:"type:uuid;not null;index" json:"tenant_id"`
	CompanyID   uuid.UUID      `gorm:"type:uuid;not null;index" json:"company_id"`
	TemplateKey string         `gorm:"size:100;not null" json:"template_key"`
	Title       string         `gorm:"size:255;not null" json:"title"`
	Category    string         `gorm:"size:100" json:"category"`
	Format      DocumentFormat `gorm:"size:10;default:'pdf'" json:"format"`
	Status      DocumentStatus `gorm:"size:20;default:'draft'" json:"status"`
	Version     int            `gorm:"default:1" json:"version"`
	FileURL     string         `gorm:"size:500" json:"file_url"`
	FileSize    int64          `json:"file_size"`
	QRCode      string         `gorm:"size:500" json:"qr_code"`
	ChecksumSHA string         `gorm:"size:64" json:"checksum_sha"`
	Signers     JSONB          `gorm:"type:jsonb;default:'[]'" json:"signers"`
	SignedAt    *time.Time     `json:"signed_at"`
	PublishedAt *time.Time     `json:"published_at"`
	ExpiresAt   *time.Time     `gorm:"index" json:"expires_at"`
	GeneratedBy uuid.UUID      `gorm:"type:uuid" json:"generated_by"`
	AIGenerated bool           `gorm:"default:false" json:"ai_generated"`
	NormRefs    JSONB          `gorm:"type:jsonb;default:'[]'" json:"norm_refs"`
	Metadata    JSONB          `gorm:"type:jsonb" json:"metadata"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Document) TableName() string { return "documents" }
