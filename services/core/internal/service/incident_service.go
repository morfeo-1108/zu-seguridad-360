package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/zu-seguridad/core/internal/model"
	"github.com/zu-seguridad/core/internal/repository"
)

type IncidentService struct {
	repo           *repository.IncidentRepository
	rabbit         *repository.RabbitMQ
	syntheticPilot bool
}

func NewIncidentService(repo *repository.IncidentRepository, rabbit *repository.RabbitMQ, syntheticPilot ...bool) *IncidentService {
	var synthetic bool
	if len(syntheticPilot) > 0 {
		synthetic = syntheticPilot[0]
	}
	return &IncidentService{repo: repo, rabbit: rabbit, syntheticPilot: synthetic}
}

func (s *IncidentService) Create(ctx context.Context, input CreateIncidentInput, companyID, tenantID, reportedBy uuid.UUID) (*model.Incident, error) {
	if !s.syntheticPilot {
		return nil, ErrSensitiveDataProcessingDisabled
	}
	if strings.TrimSpace(input.Type) != string(model.IncidentTypeIncident) || input.InjuredCount != 0 || input.GPSLat != nil || input.GPSLng != nil || len(input.EvidenceURLs) != 0 {
		return nil, ErrSyntheticDataRequired
	}
	if !strings.HasPrefix(strings.ToUpper(strings.TrimSpace(input.Location)), "SITIO SINTÉTICO") || !strings.HasPrefix(strings.ToUpper(strings.TrimSpace(input.Description)), "SIMULACIÓN:") {
		return nil, ErrSyntheticDataRequired
	}
	occurredAt, err := time.Parse(time.RFC3339, input.OccurredAt)
	if err != nil {
		occurredAt, err = time.Parse("2006-01-02", input.OccurredAt)
	}
	if err != nil || occurredAt.After(time.Now().Add(time.Minute)) {
		return nil, fmt.Errorf("occurred_at must be a valid test timestamp")
	}
	var employeeID *uuid.UUID
	if input.EmployeeID != "" {
		parsed, parseErr := uuid.Parse(input.EmployeeID)
		if parseErr != nil || parsed == uuid.Nil {
			return nil, ErrSyntheticDataRequired
		}
		belongs, checkErr := s.repo.EmployeeBelongsToCompany(ctx, parsed, companyID, tenantID, true)
		if checkErr != nil || !belongs {
			return nil, ErrSyntheticDataRequired
		}
		employeeID = &parsed
	}
	incident := &model.Incident{
		TenantID: tenantID, CompanyID: companyID, IsSynthetic: true, EmployeeID: employeeID,
		ReportedBy: reportedBy, Type: model.IncidentTypeIncident, Severity: model.SeverityLeve,
		Status: model.IncidentStatusOpen, OccurredAt: occurredAt,
		Location: strings.TrimSpace(input.Location), Description: strings.TrimSpace(input.Description),
		InjuredCount: 0,
	}
	if err := s.repo.Create(ctx, incident); err != nil {
		return nil, err
	}
	return incident, nil
}

func (s *IncidentService) GetByCompany(ctx context.Context, companyID, tenantID uuid.UUID) ([]model.Incident, error) {
	if !s.syntheticPilot {
		return nil, ErrSensitiveDataProcessingDisabled
	}
	return s.repo.FindByCompany(ctx, companyID, tenantID, true)
}

func (s *IncidentService) GetByID(ctx context.Context, id, companyID, tenantID uuid.UUID) (*model.Incident, error) {
	if !s.syntheticPilot {
		return nil, ErrSensitiveDataProcessingDisabled
	}
	return s.repo.FindByID(ctx, id, companyID, tenantID, true)
}

func (s *IncidentService) GetOpenByTenant(ctx context.Context, tenantID uuid.UUID, companyID *uuid.UUID) ([]model.Incident, error) {
	if !s.syntheticPilot {
		return nil, ErrSensitiveDataProcessingDisabled
	}
	if companyID == nil || *companyID == uuid.Nil {
		return nil, ErrSyntheticDataRequired
	}
	return s.repo.FindOpenByTenant(ctx, tenantID, companyID, true)
}

func (s *IncidentService) UpdateInvestigation(ctx context.Context, id, companyID, tenantID, investigatedBy uuid.UUID, _ UpdateInvestigationInput) (*model.Incident, error) {
	if !s.syntheticPilot {
		return nil, ErrSensitiveDataProcessingDisabled
	}
	incident, err := s.repo.FindByID(ctx, id, companyID, tenantID, true)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	incident.Status = model.IncidentStatusUnderReview
	incident.InvestigatedBy = &investigatedBy
	incident.InvestigatedAt = &now
	incident.CausalTree = model.JSONB{"simulation": true, "root_cause": "SIMULACIÓN: causa de prueba"}
	incident.CorrectiveActions = []interface{}{map[string]interface{}{"action": "PRUEBA: acción demostrativa", "status": "simulated"}}
	if err := s.repo.Update(ctx, incident); err != nil {
		return nil, err
	}
	return incident, nil
}

func (s *IncidentService) GenerateFURAT(ctx context.Context, id, companyID, tenantID uuid.UUID) (map[string]interface{}, error) {
	if !s.syntheticPilot {
		return nil, ErrSensitiveDataProcessingDisabled
	}
	if _, err := s.repo.FindByID(ctx, id, companyID, tenantID, true); err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"status": "simulated_preview", "synthetic": true, "official": false,
		"notice": "Vista demostrativa; no es un FURAT oficial y no se transmite a ARL ni autoridad.",
	}, nil
}

func (s *IncidentService) CloseSynthetic(ctx context.Context, id, companyID, tenantID, closedBy uuid.UUID) (*model.Incident, error) {
	if !s.syntheticPilot {
		return nil, ErrSensitiveDataProcessingDisabled
	}
	incident, err := s.repo.FindByID(ctx, id, companyID, tenantID, true)
	if err != nil {
		return nil, err
	}
	if incident.LegalHold || incident.Status == model.IncidentStatusLegalHold {
		return nil, fmt.Errorf("incident is on legal hold")
	}
	if incident.Status == model.IncidentStatusClosed {
		return incident, nil
	}
	if incident.Status != model.IncidentStatusOpen && incident.Status != model.IncidentStatusUnderReview {
		return nil, fmt.Errorf("incident cannot be closed from its current state")
	}
	now := time.Now().UTC()
	incident.Status = model.IncidentStatusClosed
	incident.ClosedAt = &now
	incident.ClosedBy = &closedBy
	if err := s.repo.Update(ctx, incident); err != nil {
		return nil, err
	}
	return incident, nil
}

type CreateIncidentInput struct {
	Type         string   `json:"type" validate:"required,oneof=accidente_trabajo incidente enfermedad_laboral"`
	OccurredAt   string   `json:"occurred_at" validate:"required"`
	Location     string   `json:"location" validate:"required"`
	Description  string   `json:"description" validate:"required,min=20"`
	InjuredCount int      `json:"injured_count"`
	EmployeeID   string   `json:"employee_id"`
	GPSLat       *float64 `json:"gps_lat"`
	GPSLng       *float64 `json:"gps_lng"`
	EvidenceURLs []string `json:"evidence_urls"`
}

type UpdateInvestigationInput struct {
	CausalTree        map[string]interface{} `json:"causal_tree"`
	CorrectiveActions []interface{}          `json:"corrective_actions"`
}
