package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/zu-seguridad/core/internal/model"
	"gorm.io/gorm"
)

type LegalRequirementService struct{ db *gorm.DB }

func NewLegalRequirementService(db *gorm.DB) *LegalRequirementService {
	return &LegalRequirementService{db: db}
}

type CreateLegalRequirementInput struct {
	NormReference string `json:"norm_reference" validate:"required"`
	Title         string `json:"title" validate:"required"`
	Applicable    bool   `json:"applicable"`
	EvidenceRef   string `json:"evidence_ref"`
	Notes         string `json:"notes"`
}

func (s *LegalRequirementService) Create(ctx context.Context, input CreateLegalRequirementInput, companyID, tenantID uuid.UUID) (*model.LegalRequirement, error) {
	req := &model.LegalRequirement{
		TenantID: tenantID, CompanyID: companyID, NormReference: input.NormReference, Title: input.Title,
		Applicable: input.Applicable, EvidenceRef: input.EvidenceRef, Notes: input.Notes,
		ComplianceStatus: model.LegalStatusPending,
	}
	return req, s.db.WithContext(ctx).Create(req).Error
}

func (s *LegalRequirementService) List(ctx context.Context, companyID, tenantID uuid.UUID) ([]model.LegalRequirement, error) {
	var requirements []model.LegalRequirement
	err := s.db.WithContext(ctx).Where("company_id = ? AND tenant_id = ?", companyID, tenantID).
		Order("norm_reference ASC").Find(&requirements).Error
	return requirements, err
}

type UpdateComplianceInput struct {
	ComplianceStatus string `json:"compliance_status" validate:"required,oneof=cumple parcial no_cumple pendiente_evaluacion"`
	EvidenceRef      string `json:"evidence_ref"`
	Notes            string `json:"notes"`
}

// UpdateCompliance is the human review step: a professional records the
// current compliance status with supporting evidence. Nothing here is
// inferred automatically.
func (s *LegalRequirementService) UpdateCompliance(ctx context.Context, id, companyID, tenantID, reviewedBy uuid.UUID, input UpdateComplianceInput) (*model.LegalRequirement, error) {
	now := time.Now().UTC()
	result := s.db.WithContext(ctx).Model(&model.LegalRequirement{}).
		Where("id = ? AND company_id = ? AND tenant_id = ?", id, companyID, tenantID).
		Updates(map[string]interface{}{
			"compliance_status": input.ComplianceStatus, "evidence_ref": input.EvidenceRef, "notes": input.Notes,
			"reviewed_by": reviewedBy, "reviewed_at": now,
		})
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected != 1 {
		return nil, gorm.ErrRecordNotFound
	}
	var req model.LegalRequirement
	if err := s.db.WithContext(ctx).Where("id = ? AND company_id = ? AND tenant_id = ?", id, companyID, tenantID).First(&req).Error; err != nil {
		return nil, err
	}
	return &req, nil
}
