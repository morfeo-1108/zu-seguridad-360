package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/zu-seguridad/core/internal/model"
	"github.com/zu-seguridad/core/internal/repository"
	"gorm.io/gorm"
)

type TrainingService struct {
	db             *gorm.DB
	rabbit         *repository.RabbitMQ
	syntheticPilot bool
}

func NewTrainingService(db *gorm.DB, rabbit *repository.RabbitMQ, syntheticPilot ...bool) *TrainingService {
	var synthetic bool
	if len(syntheticPilot) > 0 {
		synthetic = syntheticPilot[0]
	}
	return &TrainingService{db: db, rabbit: rabbit, syntheticPilot: synthetic}
}

type CreateTrainingInput struct {
	Title          string  `json:"title" validate:"required,min=5"`
	NormRef        string  `json:"norm_ref"`
	Category       string  `json:"category" validate:"required"`
	DurationHours  float64 `json:"duration_hours"`
	ValidityMonths int     `json:"validity_months"`
	InstructorName string  `json:"instructor_name"`
	ScheduledAt    string  `json:"scheduled_at"`
	EvaluationMin  float64 `json:"evaluation_min_score"`
}

type RegisterAttendanceInput struct {
	EmployeeIDs []string           `json:"employee_ids" validate:"required"`
	Scores      map[string]float64 `json:"scores"`
}

func ValidateSyntheticTrainingInput(input CreateTrainingInput) error {
	if !isSyntheticLabel(input.Title) || !isSyntheticLabel(input.Category) || input.NormRef != "" ||
		input.DurationHours <= 0 || input.DurationHours > 8 || input.ValidityMonths < 1 || input.ValidityMonths > 36 ||
		(input.InstructorName != "" && !isSyntheticLabel(input.InstructorName)) ||
		input.EvaluationMin != 0 {
		return ErrSyntheticDataRequired
	}
	return nil
}

func (s *TrainingService) Create(ctx context.Context, input CreateTrainingInput, companyID, tenantID uuid.UUID) (*model.Training, error) {
	if !s.syntheticPilot {
		return nil, ErrSensitiveDataProcessingDisabled
	}
	if err := ValidateSyntheticTrainingInput(input); err != nil {
		return nil, err
	}
	var scheduledAt *time.Time
	if strings.TrimSpace(input.ScheduledAt) != "" {
		parsed, err := time.Parse(time.RFC3339, input.ScheduledAt)
		if err != nil {
			parsed, err = time.Parse("2006-01-02", input.ScheduledAt)
		}
		if err != nil {
			return nil, fmt.Errorf("scheduled_at must be a valid test date")
		}
		scheduledAt = &parsed
	}
	training := &model.Training{
		TenantID: tenantID, CompanyID: companyID, IsSynthetic: true,
		Title: strings.TrimSpace(input.Title), Category: strings.TrimSpace(input.Category),
		DurationHours: input.DurationHours, ValidityMonths: input.ValidityMonths, InstructorName: strings.TrimSpace(input.InstructorName),
		ScheduledAt: scheduledAt, Status: model.TrainingStatusScheduled,
		EvaluationMin: 0, Metadata: model.JSONB{"synthetic": true, "official_validity": false},
	}
	if err := s.db.WithContext(ctx).Create(training).Error; err != nil {
		return nil, err
	}
	return training, nil
}

func (s *TrainingService) List(ctx context.Context, companyID, tenantID uuid.UUID) ([]model.Training, error) {
	if !s.syntheticPilot {
		return nil, ErrSensitiveDataProcessingDisabled
	}
	var trainings []model.Training
	err := s.db.WithContext(ctx).Where("company_id = ? AND tenant_id = ? AND is_synthetic = true", companyID, tenantID).
		Order("created_at DESC").Find(&trainings).Error
	return trainings, err
}

func (s *TrainingService) Complete(ctx context.Context, trainingID, companyID, tenantID uuid.UUID, input RegisterAttendanceInput) ([]model.Certificate, error) {
	if !s.syntheticPilot {
		return nil, ErrSensitiveDataProcessingDisabled
	}
	if len(input.EmployeeIDs) == 0 || len(input.EmployeeIDs) > 100 {
		return nil, ErrSyntheticDataRequired
	}
	employeeIDs := make([]uuid.UUID, 0, len(input.EmployeeIDs))
	seen := make(map[uuid.UUID]struct{}, len(input.EmployeeIDs))
	for _, idText := range input.EmployeeIDs {
		id, err := uuid.Parse(idText)
		if err != nil || id == uuid.Nil {
			return nil, ErrSyntheticDataRequired
		}
		if _, duplicate := seen[id]; duplicate {
			return nil, ErrSyntheticDataRequired
		}
		seen[id] = struct{}{}
		employeeIDs = append(employeeIDs, id)
		if score, ok := input.Scores[idText]; ok && (score < 0 || score > 100) {
			return nil, ErrSyntheticDataRequired
		}
	}

	var certificates []model.Certificate
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var training model.Training
		if err := tx.Where("id = ? AND company_id = ? AND tenant_id = ? AND is_synthetic = true", trainingID, companyID, tenantID).First(&training).Error; err != nil {
			return err
		}
		var employeeCount int64
		if err := tx.Model(&model.Employee{}).Where("id IN ? AND company_id = ? AND tenant_id = ? AND is_synthetic = true AND deleted_at IS NULL", employeeIDs, companyID, tenantID).Count(&employeeCount).Error; err != nil {
			return err
		}
		if employeeCount != int64(len(employeeIDs)) {
			return ErrSyntheticDataRequired
		}
		now := time.Now().UTC()
		training.Status = model.TrainingStatusCompleted
		training.CompletedAt = &now
		training.AttendeeCount = len(employeeIDs)
		training.PassCount = 0 // No official pass/fail threshold is evaluated in synthetic mode.
		training.Metadata = model.JSONB{"synthetic": true, "official_validity": false, "attendance_simulated": true}
		if err := tx.Save(&training).Error; err != nil {
			return err
		}
		for _, employeeID := range employeeIDs {
			score := input.Scores[employeeID.String()]
			certificates = append(certificates, model.Certificate{
				TenantID: tenantID, TrainingID: training.ID, EmployeeID: employeeID,
				CompanyID: companyID, IsSynthetic: true, Folio: "SYN-" + strings.ReplaceAll(uuid.NewString(), "-", ""),
				Score: score, IssuedAt: now, ExpiresAt: time.Time{}, IsValid: false,
			})
		}
		return tx.Create(&certificates).Error
	})
	if err != nil {
		return nil, err
	}
	return certificates, nil
}

func (s *TrainingService) GetExpiringCertificates(context.Context, uuid.UUID, *uuid.UUID, int) ([]model.Certificate, error) {
	if !s.syntheticPilot {
		return nil, ErrSensitiveDataProcessingDisabled
	}
	// Synthetic completion records are explicitly invalid and have no expiry.
	return []model.Certificate{}, nil
}

func (s *TrainingService) VerifyCertificate(context.Context, string) (*model.Certificate, error) {
	return nil, ErrSensitiveDataProcessingDisabled
}
