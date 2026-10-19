package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/zu-seguridad/core/internal/model"
	"gorm.io/gorm"
)

type EvidenceService struct {
	db *gorm.DB
}

func NewEvidenceService(db *gorm.DB) *EvidenceService {
	return &EvidenceService{db: db}
}

// Record only accepts references that belong to the authenticated tenant/company.
func (s *EvidenceService) Record(ctx context.Context, ev *model.EvidenceRecord) error {
	if s.db == nil || ev == nil || ev.TenantID == uuid.Nil || ev.CompanyID == uuid.Nil || ev.EmployeeID == uuid.Nil {
		return gorm.ErrRecordNotFound
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Table("employees").Where("id = ? AND tenant_id = ? AND company_id = ?", ev.EmployeeID, ev.TenantID, ev.CompanyID).Count(&count).Error; err != nil {
			return err
		}
		if count != 1 {
			return gorm.ErrRecordNotFound
		}
		if ev.ProjectID != nil {
			count = 0
			if err := tx.Table("projects").Where("id = ? AND tenant_id = ? AND company_id = ?", *ev.ProjectID, ev.TenantID, ev.CompanyID).Count(&count).Error; err != nil {
				return err
			}
			if count != 1 {
				return gorm.ErrRecordNotFound
			}
		}
		if ev.ZoneID != nil {
			count = 0
			q := tx.Table("zones").Where("id = ? AND tenant_id = ?", *ev.ZoneID, ev.TenantID)
			if ev.ProjectID != nil {
				q = q.Where("project_id = ?", *ev.ProjectID)
			}
			if err := q.Count(&count).Error; err != nil {
				return err
			}
			if count != 1 {
				return gorm.ErrRecordNotFound
			}
		}
		if ev.CapturedAt.IsZero() {
			ev.CapturedAt = time.Now().UTC()
		}
		return tx.Create(ev).Error
	})
}

func (s *EvidenceService) EmployeeTimeline(ctx context.Context, tenantID, companyID, employeeID uuid.UUID, limit int) ([]model.EvidenceRecord, error) {
	var records []model.EvidenceRecord
	q := s.db.WithContext(ctx).Where("tenant_id = ? AND company_id = ? AND employee_id = ?", tenantID, companyID, employeeID).Order("captured_at DESC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	return records, q.Find(&records).Error
}

func (s *EvidenceService) EmployeeEvidenceByType(ctx context.Context, tenantID, companyID, employeeID uuid.UUID, etype model.EvidenceType) ([]model.EvidenceRecord, error) {
	var records []model.EvidenceRecord
	err := s.db.WithContext(ctx).
		Where("tenant_id = ? AND company_id = ? AND employee_id = ? AND type = ?", tenantID, companyID, employeeID, etype).
		Order("captured_at DESC").Find(&records).Error
	return records, err
}

func (s *EvidenceService) ProjectEvidence(ctx context.Context, tenantID, companyID, projectID uuid.UUID, from, to time.Time) ([]model.EvidenceRecord, error) {
	var records []model.EvidenceRecord
	err := s.db.WithContext(ctx).
		Where("tenant_id = ? AND company_id = ? AND project_id = ? AND captured_at BETWEEN ? AND ?", tenantID, companyID, projectID, from, to).
		Order("captured_at DESC").Find(&records).Error
	return records, err
}

func (s *EvidenceService) ComplianceProof(ctx context.Context, tenantID, companyID, employeeID uuid.UUID, checklistCode string, date time.Time) (*model.EvidenceRecord, error) {
	var record model.EvidenceRecord
	start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	end := start.Add(24 * time.Hour)
	err := s.db.WithContext(ctx).
		Where("tenant_id = ? AND company_id = ? AND employee_id = ? AND checklist_code = ? AND captured_at >= ? AND captured_at < ? AND compliant = true",
			tenantID, companyID, employeeID, checklistCode, start, end).
		Order("captured_at DESC").First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (s *EvidenceService) Stats(ctx context.Context, tenantID, companyID, employeeID uuid.UUID) (map[string]interface{}, error) {
	var total, compliant int64
	q := s.db.WithContext(ctx).Model(&model.EvidenceRecord{}).Where("tenant_id = ? AND company_id = ? AND employee_id = ?", tenantID, companyID, employeeID)
	if err := q.Count(&total).Error; err != nil {
		return nil, err
	}
	if err := s.db.WithContext(ctx).Model(&model.EvidenceRecord{}).
		Where("tenant_id = ? AND company_id = ? AND employee_id = ? AND compliant = true", tenantID, companyID, employeeID).Count(&compliant).Error; err != nil {
		return nil, err
	}
	rate := 100.0
	if total > 0 {
		rate = float64(compliant) / float64(total) * 100
	}
	return map[string]interface{}{
		"total_evidences": total,
		"compliant":       compliant,
		"compliance_rate": rate,
	}, nil
}
