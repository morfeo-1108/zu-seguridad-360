package repository

import (
	"context"
	"github.com/google/uuid"
	"github.com/zu-seguridad/core/internal/model"
	"gorm.io/gorm"
)

type EmployeeRepository struct{ db *gorm.DB }

func NewEmployeeRepository(db *gorm.DB) *EmployeeRepository {
	return &EmployeeRepository{db: db}
}

func (r *EmployeeRepository) FindByCompany(ctx context.Context, companyID, tenantID uuid.UUID, page, limit int, syntheticOnly ...bool) ([]model.Employee, int64, error) {
	var employees []model.Employee
	var total int64
	q := r.db.WithContext(ctx).Where("company_id = ? AND tenant_id = ?", companyID, tenantID)
	if len(syntheticOnly) > 0 && syntheticOnly[0] {
		q = q.Where("is_synthetic = true")
	}
	q.Model(&model.Employee{}).Count(&total)
	err := q.Offset((page - 1) * limit).Limit(limit).Order("last_name ASC").Find(&employees).Error
	return employees, total, err
}

func (r *EmployeeRepository) FindByID(ctx context.Context, id, companyID, tenantID uuid.UUID, syntheticOnly ...bool) (*model.Employee, error) {
	var e model.Employee
	query := r.db.WithContext(ctx).Where("id = ? AND company_id = ? AND tenant_id = ?", id, companyID, tenantID)
	if len(syntheticOnly) > 0 && syntheticOnly[0] {
		query = query.Where("is_synthetic = true")
	}
	err := query.First(&e).Error
	return &e, err
}

func (r *EmployeeRepository) FindByDocument(ctx context.Context, docNumber string) (*model.Employee, error) {
	var e model.Employee
	err := r.db.WithContext(ctx).Where("document_number = ?", docNumber).First(&e).Error
	return &e, err
}

func (r *EmployeeRepository) Create(ctx context.Context, e *model.Employee) error {
	return r.db.WithContext(ctx).Create(e).Error
}

func (r *EmployeeRepository) BulkCreate(ctx context.Context, employees []model.Employee) error {
	return r.db.WithContext(ctx).CreateInBatches(employees, 100).Error
}

func (r *EmployeeRepository) Update(ctx context.Context, e *model.Employee) error {
	return r.db.WithContext(ctx).Save(e).Error
}

func (r *EmployeeRepository) Delete(ctx context.Context, id, companyID, tenantID uuid.UUID, syntheticOnly ...bool) error {
	query := r.db.WithContext(ctx).Where("id = ? AND company_id = ? AND tenant_id = ?", id, companyID, tenantID)
	if len(syntheticOnly) > 0 && syntheticOnly[0] {
		query = query.Where("is_synthetic = true")
	}
	return query.Delete(&model.Employee{}).Error
}

func (r *EmployeeRepository) FindWithExpiringCerts(ctx context.Context, tenantID uuid.UUID, days int) ([]model.Employee, error) {
	var employees []model.Employee
	err := r.db.WithContext(ctx).
		Joins("JOIN certificates c ON c.employee_id = employees.id").
		Where("employees.tenant_id = ? AND c.expires_at <= NOW() + INTERVAL '? days' AND c.is_valid = true", tenantID, days).
		Distinct("employees.*").Find(&employees).Error
	return employees, err
}

func (r *EmployeeRepository) CountByCompany(ctx context.Context, companyID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Employee{}).
		Where("company_id = ? AND status = 'active'", companyID).Count(&count).Error
	return count, err
}
