package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/zu-seguridad/core/internal/model"
	"gorm.io/gorm"
)

type DocumentService struct{ db *gorm.DB }

func NewDocumentService(db *gorm.DB) *DocumentService { return &DocumentService{db: db} }

// CreateDocumentInput records the metadata of a file already uploaded to
// Media (services/media) — this service never receives file bytes itself,
// only the URL/checksum Media returned. Category "foto" is how company
// photos are distinguished from regular documents; both live in the same
// table since neither needs a different schema.
type CreateDocumentInput struct {
	Title       string `json:"title" validate:"required"`
	Category    string `json:"category"`
	Format      string `json:"format"`
	FileURL     string `json:"file_url" validate:"required"`
	FileSize    int64  `json:"file_size"`
	ChecksumSHA string `json:"checksum_sha"`
}

func (s *DocumentService) Create(ctx context.Context, input CreateDocumentInput, companyID, tenantID, generatedBy uuid.UUID) (*model.Document, error) {
	title := strings.TrimSpace(input.Title)
	if title == "" {
		return nil, fmt.Errorf("el título es obligatorio")
	}
	fileURL := strings.TrimSpace(input.FileURL)
	if fileURL == "" {
		return nil, fmt.Errorf("file_url es obligatorio — sube el archivo a Media primero")
	}
	format := model.DocumentFormat(strings.ToLower(strings.TrimSpace(input.Format)))
	if format == "" {
		format = model.FormatPDF
	}

	now := time.Now().UTC()
	doc := &model.Document{
		TenantID: tenantID, CompanyID: companyID, TemplateKey: "upload",
		Title: title, Category: strings.TrimSpace(input.Category), Format: format,
		Status: model.DocStatusPublished, FileURL: fileURL, FileSize: input.FileSize,
		ChecksumSHA: input.ChecksumSHA, GeneratedBy: generatedBy, PublishedAt: &now,
	}
	return doc, s.db.WithContext(ctx).Create(doc).Error
}

func (s *DocumentService) List(ctx context.Context, companyID, tenantID uuid.UUID, category string) ([]model.Document, error) {
	q := s.db.WithContext(ctx).Where("company_id = ? AND tenant_id = ?", companyID, tenantID)
	if category != "" {
		q = q.Where("category = ?", category)
	}
	var docs []model.Document
	err := q.Order("created_at DESC").Find(&docs).Error
	return docs, err
}

func (s *DocumentService) Delete(ctx context.Context, id, companyID, tenantID uuid.UUID) error {
	result := s.db.WithContext(ctx).Where("id = ? AND company_id = ? AND tenant_id = ?", id, companyID, tenantID).
		Delete(&model.Document{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
