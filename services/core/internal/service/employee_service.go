package service

import (
	"context"
	"io"
	"strings"

	"github.com/google/uuid"
	"github.com/zu-seguridad/core/internal/model"
	"github.com/zu-seguridad/core/internal/repository"
)

type EmployeeService struct {
	repo           *repository.EmployeeRepository
	rabbit         *repository.RabbitMQ
	syntheticPilot bool
}

func NewEmployeeService(repo *repository.EmployeeRepository, rabbit *repository.RabbitMQ, syntheticPilot ...bool) *EmployeeService {
	var synthetic bool
	if len(syntheticPilot) > 0 {
		synthetic = syntheticPilot[0]
	}
	return &EmployeeService{repo: repo, rabbit: rabbit, syntheticPilot: synthetic}
}

func (s *EmployeeService) List(ctx context.Context, companyID, tenantID uuid.UUID, page, limit int) ([]model.Employee, int64, error) {
	if !s.syntheticPilot {
		return nil, 0, ErrSensitiveDataProcessingDisabled
	}
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 50
	}
	return s.repo.FindByCompany(ctx, companyID, tenantID, page, limit, true)
}

func (s *EmployeeService) GetByID(ctx context.Context, id, companyID, tenantID uuid.UUID) (*model.Employee, error) {
	if !s.syntheticPilot {
		return nil, ErrSensitiveDataProcessingDisabled
	}
	return s.repo.FindByID(ctx, id, companyID, tenantID, true)
}

func (s *EmployeeService) Create(ctx context.Context, input CreateEmployeeInput, companyID, tenantID uuid.UUID) (*model.Employee, error) {
	if !s.syntheticPilot {
		return nil, ErrSensitiveDataProcessingDisabled
	}
	if err := ValidateSyntheticEmployeeInput(input); err != nil {
		return nil, err
	}
	employee := &model.Employee{
		TenantID: tenantID, CompanyID: companyID, IsSynthetic: true,
		FirstName: strings.TrimSpace(input.FirstName), LastName: strings.TrimSpace(input.LastName),
		DocumentType: "SYN", DocumentNumber: strings.TrimSpace(input.DocumentNumber),
		Email: strings.TrimSpace(input.Email), Position: strings.TrimSpace(input.Position),
		Department: strings.TrimSpace(input.Department), ContractType: model.ContractTypeService,
		Status: model.EmployeeStatusActive,
	}
	if err := s.repo.Create(ctx, employee); err != nil {
		return nil, err
	}
	return employee, nil
}

// Bulk import is intentionally closed: an uploaded CSV cannot be reliably
// classified as synthetic and could contain worker personal data.
func (s *EmployeeService) BulkImportFromCSV(context.Context, io.Reader, uuid.UUID, uuid.UUID) (int, []string, error) {
	return 0, nil, ErrSensitiveDataProcessingDisabled
}

func (s *EmployeeService) Update(ctx context.Context, id, companyID, tenantID uuid.UUID, input UpdateEmployeeInput) (*model.Employee, error) {
	if !s.syntheticPilot {
		return nil, ErrSensitiveDataProcessingDisabled
	}
	if input.Phone != "" || (input.Email != "" && !isReservedSyntheticEmail(input.Email)) {
		return nil, ErrSyntheticDataRequired
	}
	if input.Position != "" && !isSyntheticLabel(input.Position) {
		return nil, ErrSyntheticDataRequired
	}
	if input.Department != "" && !isSyntheticLabel(input.Department) {
		return nil, ErrSyntheticDataRequired
	}
	employee, err := s.repo.FindByID(ctx, id, companyID, tenantID, true)
	if err != nil {
		return nil, err
	}
	if input.Position != "" {
		employee.Position = strings.TrimSpace(input.Position)
	}
	if input.Department != "" {
		employee.Department = strings.TrimSpace(input.Department)
	}
	if input.Email != "" {
		employee.Email = strings.TrimSpace(input.Email)
	}
	if err := s.repo.Update(ctx, employee); err != nil {
		return nil, err
	}
	return employee, nil
}

func (s *EmployeeService) Delete(ctx context.Context, id, companyID, tenantID uuid.UUID) error {
	if !s.syntheticPilot {
		return ErrSensitiveDataProcessingDisabled
	}
	return s.repo.Delete(ctx, id, companyID, tenantID, true)
}

type CreateEmployeeInput struct {
	FirstName      string `json:"first_name" validate:"required"`
	LastName       string `json:"last_name" validate:"required"`
	DocumentType   string `json:"document_type" validate:"required"`
	DocumentNumber string `json:"document_number" validate:"required"`
	Email          string `json:"email" validate:"email"`
	Phone          string `json:"phone"`
	Position       string `json:"position" validate:"required"`
	Department     string `json:"department"`
	ContractType   string `json:"contract_type"`
}

type UpdateEmployeeInput struct {
	Position   string `json:"position"`
	Department string `json:"department"`
	Phone      string `json:"phone"`
	Email      string `json:"email"`
}

func ValidateSyntheticEmployeeInput(input CreateEmployeeInput) error {
	if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(input.FirstName)), "persona") ||
		(!strings.HasPrefix(strings.ToLower(strings.TrimSpace(input.LastName)), "sintético") && !strings.HasPrefix(strings.ToLower(strings.TrimSpace(input.LastName)), "prueba")) ||
		(strings.ToUpper(strings.TrimSpace(input.DocumentType)) != "SYN" && strings.ToUpper(strings.TrimSpace(input.DocumentType)) != "TEST") ||
		!strings.HasPrefix(strings.ToUpper(strings.TrimSpace(input.DocumentNumber)), "SYN-") ||
		!isSyntheticLabel(input.Position) || input.Phone != "" ||
		(input.Department != "" && !isSyntheticLabel(input.Department)) ||
		(input.Email != "" && !isReservedSyntheticEmail(input.Email)) {
		return ErrSyntheticDataRequired
	}
	return nil
}

func isReservedSyntheticEmail(value string) bool {
	email := strings.ToLower(strings.TrimSpace(value))
	return strings.HasSuffix(email, "@zu360.test") || strings.HasSuffix(email, ".invalid")
}

func isSyntheticLabel(value string) bool {
	label := strings.ToLower(strings.TrimSpace(value))
	return strings.HasPrefix(label, "demostración") || strings.HasPrefix(label, "prueba") || strings.HasPrefix(label, "sintético")
}
