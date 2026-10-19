package repository

import (
	"context"
	"github.com/google/uuid"
	"github.com/zu-seguridad/core/internal/model"
	"gorm.io/gorm"
	"time"
)

type IncidentRepository struct{ db *gorm.DB }

func NewIncidentRepository(db *gorm.DB) *IncidentRepository { return &IncidentRepository{db: db} }

func (r *IncidentRepository) FindByCompany(ctx context.Context, companyID, tenantID uuid.UUID, syntheticOnly ...bool) ([]model.Incident, error) {
	var incidents []model.Incident
	query := r.db.WithContext(ctx).
		Preload("Employee").
		Where("company_id = ? AND tenant_id = ?", companyID, tenantID)
	if len(syntheticOnly) > 0 && syntheticOnly[0] {
		query = query.Where("is_synthetic = true")
	}
	err := query.Order("occurred_at DESC").Find(&incidents).Error
	return incidents, err
}

func (r *IncidentRepository) FindByID(ctx context.Context, id, companyID, tenantID uuid.UUID, syntheticOnly ...bool) (*model.Incident, error) {
	var i model.Incident
	query := r.db.WithContext(ctx).Preload("Employee").Where("id = ? AND company_id = ? AND tenant_id = ?", id, companyID, tenantID)
	if len(syntheticOnly) > 0 && syntheticOnly[0] {
		query = query.Where("is_synthetic = true")
	}
	err := query.First(&i).Error
	return &i, err
}

func (r *IncidentRepository) EmployeeBelongsToCompany(ctx context.Context, employeeID, companyID, tenantID uuid.UUID, syntheticOnly ...bool) (bool, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&model.Employee{}).Where("id = ? AND company_id = ? AND tenant_id = ?", employeeID, companyID, tenantID)
	if len(syntheticOnly) > 0 && syntheticOnly[0] {
		query = query.Where("is_synthetic = true")
	}
	err := query.Count(&count).Error
	return count == 1, err
}

func (r *IncidentRepository) Create(ctx context.Context, i *model.Incident) error {
	return r.db.WithContext(ctx).Create(i).Error
}

func (r *IncidentRepository) Update(ctx context.Context, i *model.Incident) error {
	return r.db.WithContext(ctx).Save(i).Error
}

func (r *IncidentRepository) FindOpenByTenant(ctx context.Context, tenantID uuid.UUID, companyID *uuid.UUID, syntheticOnly ...bool) ([]model.Incident, error) {
	var incidents []model.Incident
	query := r.db.WithContext(ctx).
		Preload("Company").Preload("Employee").
		Where("tenant_id = ? AND status = 'open'", tenantID)
	if companyID != nil {
		query = query.Where("company_id = ?", *companyID)
	}
	if len(syntheticOnly) > 0 && syntheticOnly[0] {
		query = query.Where("is_synthetic = true")
	}
	err := query.Order("occurred_at DESC").Find(&incidents).Error
	return incidents, err
}

// IndicatorsIF calcula Índice de Frecuencia = (AT * 1,000,000) / HHT
func (r *IncidentRepository) CountByPeriod(ctx context.Context, companyID uuid.UUID, from, to time.Time) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Incident{}).
		Where("company_id = ? AND type = 'accidente_trabajo' AND occurred_at BETWEEN ? AND ?", companyID, from, to).
		Count(&count).Error
	return count, err
}
