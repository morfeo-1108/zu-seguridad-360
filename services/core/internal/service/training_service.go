package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	qrcode "github.com/skip2/go-qrcode"
	"github.com/zu-seguridad/core/internal/model"
	"github.com/zu-seguridad/core/internal/repository"
	"gorm.io/gorm"
)

type TrainingService struct {
	db             *gorm.DB
	rabbit         *repository.RabbitMQ
	syntheticPilot bool
	publicAppURL   string
}

func NewTrainingService(db *gorm.DB, rabbit *repository.RabbitMQ, syntheticPilot bool, publicAppURL string) *TrainingService {
	return &TrainingService{db: db, rabbit: rabbit, syntheticPilot: syntheticPilot, publicAppURL: publicAppURL}
}

func (s *TrainingService) verifyURL(folio string) string {
	return strings.TrimRight(s.publicAppURL, "/") + "/verify/" + folio
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
			folio := "SYN-" + strings.ReplaceAll(uuid.NewString(), "-", "")
			certificates = append(certificates, model.Certificate{
				TenantID: tenantID, TrainingID: training.ID, EmployeeID: employeeID,
				CompanyID: companyID, IsSynthetic: true, Folio: folio,
				Score: score, IssuedAt: now, ExpiresAt: time.Time{}, IsValid: false,
				QRCode: s.verifyURL(folio),
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

// CertificateVerification is the deliberately minimized shape returned by the
// public verification endpoint: it answers "is this folio real and valid",
// not a dump of the underlying employee/training/company rows.
type CertificateVerification struct {
	Folio          string     `json:"folio"`
	EmployeeName   string     `json:"employee_name"`
	TrainingTitle  string     `json:"training_title"`
	Category       string     `json:"category"`
	IssuedAt       time.Time  `json:"issued_at"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
	IsValid        bool       `json:"is_valid"`
	IsSynthetic    bool       `json:"is_synthetic"`
	RevokedAt      *time.Time `json:"revoked_at,omitempty"`
	ValidityNotice string     `json:"validity_notice"`
}

func (s *TrainingService) VerifyCertificate(ctx context.Context, folio string) (*CertificateVerification, error) {
	if !s.syntheticPilot {
		return nil, ErrSensitiveDataProcessingDisabled
	}
	folio = strings.TrimSpace(folio)
	if folio == "" {
		return nil, gorm.ErrRecordNotFound
	}
	var cert model.Certificate
	if err := s.db.WithContext(ctx).Preload("Employee").Preload("Training").
		Where("folio = ?", folio).First(&cert).Error; err != nil {
		return nil, err
	}

	notice := "Certificado válido."
	if cert.IsSynthetic {
		notice = "Registro de datos sintéticos del piloto ZU360: NO es un certificado oficial ni tiene validez legal."
	} else if cert.RevokedAt != nil {
		notice = "Certificado revocado."
	} else if !cert.IsValid {
		notice = "Certificado marcado como no válido."
	}

	var expiresAt *time.Time
	if !cert.ExpiresAt.IsZero() {
		expiresAt = &cert.ExpiresAt
	}

	return &CertificateVerification{
		Folio:          cert.Folio,
		EmployeeName:   strings.TrimSpace(cert.Employee.FirstName + " " + cert.Employee.LastName),
		TrainingTitle:  cert.Training.Title,
		Category:       cert.Training.Category,
		IssuedAt:       cert.IssuedAt,
		ExpiresAt:      expiresAt,
		IsValid:        cert.IsValid && cert.RevokedAt == nil,
		IsSynthetic:    cert.IsSynthetic,
		RevokedAt:      cert.RevokedAt,
		ValidityNotice: notice,
	}, nil
}

// CertificateQRPNG renders the QR image pointing at the public verification
// page for a given folio. It re-checks the folio exists first so the image
// endpoint can't be used to generate QR codes for arbitrary, unissued folios.
func (s *TrainingService) CertificateQRPNG(ctx context.Context, folio string) ([]byte, error) {
	if !s.syntheticPilot {
		return nil, ErrSensitiveDataProcessingDisabled
	}
	folio = strings.TrimSpace(folio)
	var count int64
	if err := s.db.WithContext(ctx).Model(&model.Certificate{}).Where("folio = ?", folio).Count(&count).Error; err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return qrcode.Encode(s.verifyURL(folio), qrcode.Medium, 320)
}
