package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/zu-seguridad/core/internal/model"
	"github.com/zu-seguridad/core/internal/repository"
	"gorm.io/gorm"
)

type MedicalExamService struct {
	db     *gorm.DB
	rabbit *repository.RabbitMQ
}

func NewMedicalExamService(db *gorm.DB, rabbit *repository.RabbitMQ) *MedicalExamService {
	return &MedicalExamService{db: db, rabbit: rabbit}
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

func (s *MedicalExamService) Schedule(context.Context, ScheduleExamInput, uuid.UUID, uuid.UUID) (*model.MedicalExam, error) {
	return nil, ErrSensitiveDataProcessingDisabled
}

func (s *MedicalExamService) ProcessOCR(context.Context, uuid.UUID, map[string]interface{}) error {
	return ErrSensitiveDataProcessingDisabled
}

func (s *MedicalExamService) Review(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, ReviewExamInput) (*model.MedicalExam, error) {
	return nil, ErrSensitiveDataProcessingDisabled
}

func (s *MedicalExamService) GetByEmployee(context.Context, uuid.UUID, uuid.UUID) ([]model.MedicalExam, error) {
	return nil, ErrSensitiveDataProcessingDisabled
}

func (s *MedicalExamService) GetPendingReview(context.Context, uuid.UUID) ([]model.MedicalExam, error) {
	return nil, ErrSensitiveDataProcessingDisabled
}

func (s *MedicalExamService) GetExpiring(context.Context, uuid.UUID, int) ([]model.MedicalExam, error) {
	return nil, ErrSensitiveDataProcessingDisabled
}
