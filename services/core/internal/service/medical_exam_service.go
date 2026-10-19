package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/zu-seguridad/core/internal/model"
	"github.com/zu-seguridad/core/internal/repository"
	"gorm.io/gorm"
)

type MedicalExamService struct {
	db             *gorm.DB
	rabbit         *repository.RabbitMQ
	syntheticPilot bool
}

func NewMedicalExamService(db *gorm.DB, rabbit *repository.RabbitMQ, syntheticPilot ...bool) *MedicalExamService {
	var synthetic bool
	if len(syntheticPilot) > 0 {
		synthetic = syntheticPilot[0]
	}
	return &MedicalExamService{db: db, rabbit: rabbit, syntheticPilot: synthetic}
}

type ScheduleExamInput struct {
	EmployeeID  string `json:"employee_id" validate:"required"`
	Type        string `json:"type" validate:"required,oneof=ingreso periodico retiro"`
	IPSName     string `json:"ips_name" validate:"required"`
	ScheduledAt string `json:"scheduled_at" validate:"required"`
}

type ReviewExamInput struct {
	AptitudeResult string   `json:"aptitude_result" validate:"required,oneof=apto apto_con_restriccion no_apto"`
	Restrictions   []string `json:"restrictions"`
	ReviewNotes    string   `json:"review_notes"`
	ExpiresAt      string   `json:"expires_at"`
}

// Schedule creates a medical exam record scoped to the company bound to the caller's
// active token. It never accepts OCR/diagnostic data directly; ProcessOCR remains
// disabled because medical-document OCR requires its own privacy review.
func (s *MedicalExamService) Schedule(ctx context.Context, input ScheduleExamInput, companyID, tenantID uuid.UUID) (*model.MedicalExam, error) {
	if !s.syntheticPilot {
		return nil, ErrSensitiveDataProcessingDisabled
	}
	employeeID, err := uuid.Parse(input.EmployeeID)
	if err != nil {
		return nil, fmt.Errorf("employee_id inválido")
	}
	scheduledAt, err := time.Parse(time.RFC3339, input.ScheduledAt)
	if err != nil {
		return nil, fmt.Errorf("scheduled_at debe ser RFC3339")
	}

	exam := &model.MedicalExam{
		TenantID:    tenantID,
		CompanyID:   companyID,
		EmployeeID:  employeeID,
		Type:        model.ExamType(input.Type),
		IPSName:     input.IPSName,
		ScheduledAt: &scheduledAt,
	}

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Table("employees").
			Where("id = ? AND tenant_id = ? AND company_id = ? AND is_synthetic = true", employeeID, tenantID, companyID).
			Count(&count).Error; err != nil {
			return err
		}
		if count != 1 {
			return gorm.ErrRecordNotFound
		}
		return tx.Create(exam).Error
	})
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("el trabajador no pertenece a la empresa activa")
		}
		return nil, err
	}
	return exam, nil
}

// ProcessOCR remains disabled: medical-document OCR requires its own privacy and
// retention review, independent of the synthetic pilot gate.
func (s *MedicalExamService) ProcessOCR(context.Context, uuid.UUID, map[string]interface{}) error {
	return ErrSensitiveDataProcessingDisabled
}

func (s *MedicalExamService) Review(ctx context.Context, examID, tenantID, companyID, reviewedBy uuid.UUID, input ReviewExamInput) (*model.MedicalExam, error) {
	if !s.syntheticPilot {
		return nil, ErrSensitiveDataProcessingDisabled
	}
	updates := map[string]interface{}{
		"aptitude_result": input.AptitudeResult,
		"review_notes":    input.ReviewNotes,
		"reviewed_by":     reviewedBy,
		"reviewed_at":     time.Now().UTC(),
	}
	restrictions := model.JSONBArray{}
	for _, r := range input.Restrictions {
		restrictions = append(restrictions, r)
	}
	updates["restrictions"] = restrictions
	if input.ExpiresAt != "" {
		expiresAt, err := time.Parse(time.RFC3339, input.ExpiresAt)
		if err != nil {
			return nil, fmt.Errorf("expires_at debe ser RFC3339")
		}
		updates["expires_at"] = expiresAt
	}

	result := s.db.WithContext(ctx).Model(&model.MedicalExam{}).
		Where("id = ? AND tenant_id = ? AND company_id = ?", examID, tenantID, companyID).
		Updates(updates)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected != 1 {
		return nil, gorm.ErrRecordNotFound
	}

	var exam model.MedicalExam
	if err := s.db.WithContext(ctx).Where("id = ? AND tenant_id = ? AND company_id = ?", examID, tenantID, companyID).First(&exam).Error; err != nil {
		return nil, err
	}
	return &exam, nil
}

func (s *MedicalExamService) GetByEmployee(ctx context.Context, employeeID, tenantID, companyID uuid.UUID) ([]model.MedicalExam, error) {
	if !s.syntheticPilot {
		return nil, ErrSensitiveDataProcessingDisabled
	}
	var exams []model.MedicalExam
	err := s.db.WithContext(ctx).
		Where("employee_id = ? AND tenant_id = ? AND company_id = ?", employeeID, tenantID, companyID).
		Order("scheduled_at DESC").Find(&exams).Error
	return exams, err
}

func (s *MedicalExamService) GetPendingReview(ctx context.Context, tenantID, companyID uuid.UUID) ([]model.MedicalExam, error) {
	if !s.syntheticPilot {
		return nil, ErrSensitiveDataProcessingDisabled
	}
	var exams []model.MedicalExam
	err := s.db.WithContext(ctx).
		Where("tenant_id = ? AND company_id = ? AND aptitude_result = ?", tenantID, companyID, model.AptitudePending).
		Order("scheduled_at ASC").Find(&exams).Error
	return exams, err
}

func (s *MedicalExamService) GetExpiring(ctx context.Context, tenantID, companyID uuid.UUID, days int) ([]model.MedicalExam, error) {
	if !s.syntheticPilot {
		return nil, ErrSensitiveDataProcessingDisabled
	}
	cutoff := time.Now().UTC().AddDate(0, 0, days)
	var exams []model.MedicalExam
	err := s.db.WithContext(ctx).
		Where("tenant_id = ? AND company_id = ? AND expires_at IS NOT NULL AND expires_at <= ?", tenantID, companyID, cutoff).
		Order("expires_at ASC").Find(&exams).Error
	return exams, err
}
