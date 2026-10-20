package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/zu-seguridad/core/internal/model"
	"gorm.io/gorm"
)

type ChecklistTemplateService struct{ db *gorm.DB }

func NewChecklistTemplateService(db *gorm.DB) *ChecklistTemplateService {
	return &ChecklistTemplateService{db: db}
}

type ChecklistTemplateItemInput struct {
	Code        string `json:"code" validate:"required"`
	Description string `json:"description" validate:"required"`
}

type CreateChecklistTemplateInput struct {
	Name     string                       `json:"name" validate:"required"`
	Category string                       `json:"category"`
	Items    []ChecklistTemplateItemInput `json:"items" validate:"required,min=1"`
}

func (s *ChecklistTemplateService) Create(ctx context.Context, input CreateChecklistTemplateInput, companyID, tenantID, createdBy uuid.UUID) (*model.ChecklistTemplate, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, fmt.Errorf("el nombre de la plantilla es obligatorio")
	}
	if len(input.Items) == 0 {
		return nil, fmt.Errorf("la plantilla debe tener al menos un ítem")
	}
	items := model.JSONBArray{}
	for _, item := range input.Items {
		code := strings.TrimSpace(item.Code)
		description := strings.TrimSpace(item.Description)
		if code == "" || description == "" {
			return nil, fmt.Errorf("cada ítem requiere code y description")
		}
		items = append(items, map[string]interface{}{"code": code, "description": description})
	}

	template := &model.ChecklistTemplate{
		TenantID: tenantID, CompanyID: companyID, Name: name,
		Category: strings.TrimSpace(input.Category), Items: items, CreatedBy: createdBy,
	}
	return template, s.db.WithContext(ctx).Create(template).Error
}

func (s *ChecklistTemplateService) List(ctx context.Context, companyID, tenantID uuid.UUID) ([]model.ChecklistTemplate, error) {
	var templates []model.ChecklistTemplate
	err := s.db.WithContext(ctx).Where("company_id = ? AND tenant_id = ?", companyID, tenantID).
		Order("name ASC").Find(&templates).Error
	return templates, err
}

func (s *ChecklistTemplateService) GetByID(ctx context.Context, id, companyID, tenantID uuid.UUID) (*model.ChecklistTemplate, error) {
	var template model.ChecklistTemplate
	err := s.db.WithContext(ctx).Where("id = ? AND company_id = ? AND tenant_id = ?", id, companyID, tenantID).First(&template).Error
	if err != nil {
		return nil, err
	}
	return &template, nil
}

func (s *ChecklistTemplateService) Delete(ctx context.Context, id, companyID, tenantID uuid.UUID) error {
	result := s.db.WithContext(ctx).Where("id = ? AND company_id = ? AND tenant_id = ?", id, companyID, tenantID).
		Delete(&model.ChecklistTemplate{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
