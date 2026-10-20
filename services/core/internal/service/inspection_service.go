package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/zu-seguridad/core/internal/model"
	"gorm.io/gorm"
)

type InspectionService struct{ db *gorm.DB }

func NewInspectionService(db *gorm.DB) *InspectionService { return &InspectionService{db: db} }

type ChecklistItemInput struct {
	Code        string `json:"code" validate:"required"`
	Description string `json:"description" validate:"required"`
	Compliant   bool   `json:"compliant"`
	Notes       string `json:"notes"`
}

type CreateInspectionInput struct {
	Title       string               `json:"title" validate:"required"`
	Area        string               `json:"area" validate:"required"`
	PerformedAt string               `json:"performed_at" validate:"required"`
	Checklist   []ChecklistItemInput `json:"checklist" validate:"required,min=1"`
}

// CreateEntry computes compliance_pct from the checklist and auto-generates
// one action-plan item per non-compliant finding ("Generación de planes de
// acción desde los hallazgos" — the plan is derived, never hand-typed
// separately, so it can't drift from what was actually found non-compliant).
func (s *InspectionService) Create(ctx context.Context, input CreateInspectionInput, companyID, tenantID, inspectorID uuid.UUID) (*model.Inspection, error) {
	if len(input.Checklist) == 0 {
		return nil, fmt.Errorf("el checklist debe tener al menos un ítem")
	}
	performedAt, err := time.Parse(time.RFC3339, input.PerformedAt)
	if err != nil {
		performedAt, err = time.Parse("2006-01-02", input.PerformedAt)
		if err != nil {
			return nil, fmt.Errorf("performed_at debe ser RFC3339 o YYYY-MM-DD")
		}
	}

	checklist := model.JSONBArray{}
	actionPlan := model.JSONBArray{}
	compliant := 0
	for _, item := range input.Checklist {
		checklist = append(checklist, map[string]interface{}{
			"code": item.Code, "description": item.Description, "compliant": item.Compliant, "notes": item.Notes,
		})
		if item.Compliant {
			compliant++
		} else {
			actionPlan = append(actionPlan, map[string]interface{}{
				"item_code": item.Code, "description": "Corregir hallazgo: " + item.Description,
				"responsible": "", "due_date": nil, "status": "pendiente",
			})
		}
	}
	compliancePct := float64(compliant) / float64(len(input.Checklist)) * 100

	inspection := &model.Inspection{
		TenantID: tenantID, CompanyID: companyID, Title: input.Title, Area: input.Area,
		InspectorID: inspectorID, PerformedAt: performedAt, Checklist: checklist,
		CompliancePct: compliancePct, Status: model.InspectionStatusCompleted, ActionPlan: actionPlan,
	}
	return inspection, s.db.WithContext(ctx).Create(inspection).Error
}

func (s *InspectionService) List(ctx context.Context, companyID, tenantID uuid.UUID) ([]model.Inspection, error) {
	var inspections []model.Inspection
	err := s.db.WithContext(ctx).Where("company_id = ? AND tenant_id = ?", companyID, tenantID).
		Order("performed_at DESC").Find(&inspections).Error
	return inspections, err
}

func (s *InspectionService) GetByID(ctx context.Context, id, companyID, tenantID uuid.UUID) (*model.Inspection, error) {
	var inspection model.Inspection
	err := s.db.WithContext(ctx).Where("id = ? AND company_id = ? AND tenant_id = ?", id, companyID, tenantID).First(&inspection).Error
	if err != nil {
		return nil, err
	}
	return &inspection, nil
}

// Approve is the inspection's approval flow: a second professional confirms
// the findings and action plan stand. It does not re-verify the checklist —
// approval is attribution (who signed off, when), not recomputation.
func (s *InspectionService) Approve(ctx context.Context, id, companyID, tenantID, approvedBy uuid.UUID) (*model.Inspection, error) {
	now := time.Now().UTC()
	result := s.db.WithContext(ctx).Model(&model.Inspection{}).
		Where("id = ? AND company_id = ? AND tenant_id = ?", id, companyID, tenantID).
		Updates(map[string]interface{}{"status": model.InspectionStatusApproved, "approved_by": approvedBy, "approved_at": now})
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected != 1 {
		return nil, gorm.ErrRecordNotFound
	}
	return s.GetByID(ctx, id, companyID, tenantID)
}
