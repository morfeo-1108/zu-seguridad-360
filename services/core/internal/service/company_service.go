package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/zu-seguridad/core/internal/model"
	"github.com/zu-seguridad/core/internal/repository"
	"math"
	"net/http"
	"strings"
	"time"
)

var ErrCompanyQuotaExceeded = fmt.Errorf("adding this company would exceed your plan's company quota")

type CompanyService struct {
	repo           *repository.CompanyRepository
	rabbit         *repository.RabbitMQ
	syntheticPilot bool
	billingURL     string
	httpClient     *http.Client
}

func NewCompanyService(repo *repository.CompanyRepository, rabbit *repository.RabbitMQ, billingURL string, syntheticPilot ...bool) *CompanyService {
	var synthetic bool
	if len(syntheticPilot) > 0 {
		synthetic = syntheticPilot[0]
	}
	return &CompanyService{
		repo: repo, rabbit: rabbit, syntheticPilot: synthetic, billingURL: billingURL,
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}
}

// checkCompanyQuota asks billing whether tenantID has room for one more
// company under their active plan. The synthetic pilot has no subscription
// concept, so this is skipped entirely in that mode.
func (s *CompanyService) checkCompanyQuota(ctx context.Context, tenantID uuid.UUID, bearerToken string) error {
	if s.syntheticPilot || s.billingURL == "" {
		return nil
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.billingURL+"/v1/tenants/"+tenantID.String()+"/can-add-company", nil)
	if err != nil {
		return fmt.Errorf("billing quota check request failed: %w", err)
	}
	if bearerToken != "" {
		req.Header.Set("Authorization", bearerToken)
	}
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("billing service unavailable: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("billing quota check failed with status %d", resp.StatusCode)
	}
	var body struct {
		Allowed bool `json:"allowed"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return fmt.Errorf("billing quota check returned an invalid response: %w", err)
	}
	if !body.Allowed {
		return ErrCompanyQuotaExceeded
	}
	return nil
}

func (s *CompanyService) List(ctx context.Context, tenantID uuid.UUID) ([]model.Company, error) {
	return nil, fmt.Errorf("user identity required; use ListForUser")
}

func (s *CompanyService) ListForUser(ctx context.Context, userID, tenantID uuid.UUID) ([]model.Company, error) {
	return s.repo.FindAllForUser(ctx, userID, tenantID)
}

func (s *CompanyService) CreateForUser(ctx context.Context, input CreateCompanyInput, tenantID, userID uuid.UUID, bearerToken string) (*model.Company, error) {
	if err := s.checkCompanyQuota(ctx, tenantID, bearerToken); err != nil {
		return nil, err
	}
	if s.syntheticPilot {
		if !strings.HasPrefix(strings.ToUpper(strings.TrimSpace(input.Name)), "ZU360 PILOTO SINTÉTICO") || !strings.HasPrefix(strings.ToUpper(strings.TrimSpace(input.NIT)), "TEST-NIT-") {
			return nil, fmt.Errorf("synthetic pilot requires a marked test company name and TEST-NIT identifier")
		}
		if strings.ToUpper(strings.TrimSpace(input.CIIU)) != "TEST" || !isSyntheticARL(input.ARL) {
			return nil, ErrSyntheticDataRequired
		}
		email := strings.ToLower(strings.TrimSpace(input.LegalRepEmail))
		if email != "" && !strings.HasSuffix(email, "@zu360.test") && !strings.HasSuffix(email, ".invalid") {
			return nil, fmt.Errorf("synthetic pilot accepts only reserved test email domains")
		}
		if input.LegalRep != "" && !strings.HasPrefix(strings.ToLower(input.LegalRep), "persona sintética") && !strings.HasPrefix(strings.ToLower(input.LegalRep), "prueba") {
			return nil, fmt.Errorf("synthetic pilot requires a clearly synthetic legal representative")
		}
		if input.Address != "" || input.City != "" || input.Department != "" || input.LegalRepPhone != "" {
			return nil, fmt.Errorf("address, location, and phone fields are unavailable in the synthetic pilot")
		}
	}
	if tenantID == uuid.Nil || userID == uuid.Nil || strings.TrimSpace(input.Name) == "" || strings.TrimSpace(input.NIT) == "" {
		return nil, fmt.Errorf("company name, NIT, tenant, and creating user are required")
	}
	if _, err := s.repo.FindByNIT(ctx, input.NIT); err == nil {
		return nil, fmt.Errorf("company NIT already exists")
	}
	company := &model.Company{
		TenantID: tenantID, IsSynthetic: s.syntheticPilot, Name: strings.TrimSpace(input.Name),
		NIT: strings.TrimSpace(input.NIT), CIIU: input.CIIU, ARL: input.ARL,
		LegalRep: input.LegalRep, LegalRepEmail: input.LegalRepEmail,
		LegalRepPhone: input.LegalRepPhone, Address: input.Address, City: input.City,
		Department: input.Department, Status: model.CompanyStatusActive, OnboardingStep: 1,
	}
	if s.syntheticPilot {
		company.RiskClass = model.RiskClass("TEST")
	} else {
		company.RiskClass = s.assignRiskClass(input.CIIU)
	}
	if err := s.repo.CreateForUser(ctx, company, userID); err != nil {
		return nil, err
	}
	// Synthetic companies do not publish business payloads to RabbitMQ.
	if !s.syntheticPilot {
		s.publishEvent("zu.companies", "company.created", map[string]interface{}{"company_id": company.ID.String(), "tenant_id": tenantID.String()})
	}
	return company, nil
}

func (s *CompanyService) GetByID(ctx context.Context, id, tenantID uuid.UUID) (*model.Company, error) {
	company, err := s.repo.FindByID(ctx, id, tenantID)
	if err != nil {
		return nil, err
	}
	if s.syntheticPilot && !company.IsSynthetic {
		return nil, ErrSyntheticDataRequired
	}
	return company, nil
}

func (s *CompanyService) Create(ctx context.Context, input CreateCompanyInput, tenantID uuid.UUID) (*model.Company, error) {
	return nil, fmt.Errorf("creating user identity is required; use CreateForUser")
}

func (s *CompanyService) Update(ctx context.Context, id, tenantID uuid.UUID, input UpdateCompanyInput) (*model.Company, error) {
	if s.syntheticPilot {
		if input.Name != "" && !strings.HasPrefix(strings.ToUpper(strings.TrimSpace(input.Name)), "ZU360 PILOTO SINTÉTICO") {
			return nil, fmt.Errorf("synthetic pilot requires a marked test company name")
		}
		if input.ARL != "" && !isSyntheticARL(input.ARL) {
			return nil, ErrSyntheticDataRequired
		}
		if input.LegalRep != "" && !strings.HasPrefix(strings.ToLower(input.LegalRep), "persona sintética") && !strings.HasPrefix(strings.ToLower(input.LegalRep), "prueba") {
			return nil, fmt.Errorf("synthetic pilot requires a clearly synthetic legal representative")
		}
		email := strings.ToLower(strings.TrimSpace(input.LegalRepEmail))
		if email != "" && !strings.HasSuffix(email, "@zu360.test") && !strings.HasSuffix(email, ".invalid") {
			return nil, fmt.Errorf("synthetic pilot accepts only reserved test email domains")
		}
		if input.LegalRepPhone != "" || input.Address != "" {
			return nil, fmt.Errorf("address and phone fields are unavailable in the synthetic pilot")
		}
	}
	company, err := s.repo.FindByID(ctx, id, tenantID)
	if err != nil {
		return nil, fmt.Errorf("empresa no encontrada")
	}
	if s.syntheticPilot && !company.IsSynthetic {
		return nil, ErrSyntheticDataRequired
	}

	if input.Name != "" {
		company.Name = input.Name
	}
	if input.ARL != "" {
		company.ARL = input.ARL
	}
	if input.LegalRep != "" {
		company.LegalRep = input.LegalRep
	}
	if input.LegalRepEmail != "" {
		company.LegalRepEmail = input.LegalRepEmail
	}
	if input.Address != "" {
		company.Address = input.Address
	}
	if input.LogoURL != "" {
		company.LogoURL = input.LogoURL
	}

	return company, s.repo.Update(ctx, company)
}

func (s *CompanyService) Delete(ctx context.Context, id, tenantID uuid.UUID) error {
	if s.syntheticPilot {
		company, err := s.repo.FindByID(ctx, id, tenantID)
		if err != nil {
			return err
		}
		if !company.IsSynthetic {
			return ErrSyntheticDataRequired
		}
	}
	return s.repo.Delete(ctx, id, tenantID)
}

// CalculateRiskScore calcula el score de riesgo global (0-100, mayor = mejor)
func (s *CompanyService) CalculateRiskScore(ctx context.Context, id, tenantID uuid.UUID) (float64, error) {
	if s.syntheticPilot {
		return 0, ErrSensitiveDataProcessingDisabled
	}
	company, err := s.repo.FindByID(ctx, id, tenantID)
	if err != nil {
		return 0, err
	}

	// Factores ponderados
	score := 100.0

	// Penalizar por incidentes (peso 30%)
	// Penalizar por vencimientos (peso 25%)
	// Penalizar por cumplimiento Res.0312 (peso 35%)
	// Penalizar por clase de riesgo ARL (peso 10%)
	riskPenalty := map[model.RiskClass]float64{
		model.RiskClassV: 10, model.RiskClassIV: 7,
		model.RiskClassIII: 4, model.RiskClassII: 2, model.RiskClassI: 0,
	}
	score -= riskPenalty[company.RiskClass]
	score = math.Max(0, math.Min(100, score))

	s.repo.UpdateRiskScore(ctx, id, score)
	return score, nil
}

func (s *CompanyService) assignRiskClass(ciiu string) model.RiskClass {
	// Clasificación por código CIIU según tabla ARL colombiana
	highRisk := map[string]bool{"0510": true, "0520": true, "1610": true, "2410": true, "4111": true, "4290": true}
	medRisk := map[string]bool{"1311": true, "2310": true, "2599": true, "3110": true}
	if highRisk[ciiu] {
		return model.RiskClassV
	}
	if medRisk[ciiu] {
		return model.RiskClassIII
	}
	return model.RiskClassII
}

func (s *CompanyService) publishEvent(exchange, key string, payload map[string]interface{}) {
	data, _ := json.Marshal(payload)
	s.rabbit.Publish(exchange, key, data)
}

func isSyntheticARL(value string) bool {
	arl := strings.ToLower(strings.TrimSpace(value))
	return strings.HasPrefix(arl, "arl sintética") || strings.HasPrefix(arl, "prueba") || strings.HasPrefix(arl, "demostración") || strings.HasPrefix(arl, "sintético")
}

// DTOs
type CreateCompanyInput struct {
	Name          string `json:"name" validate:"required,min=3,max=255"`
	NIT           string `json:"nit" validate:"required"`
	CIIU          string `json:"ciiu" validate:"required"`
	ARL           string `json:"arl" validate:"required"`
	LegalRep      string `json:"legal_rep" validate:"required"`
	LegalRepEmail string `json:"legal_rep_email" validate:"required,email"`
	LegalRepPhone string `json:"legal_rep_phone"`
	Address       string `json:"address"`
	City          string `json:"city"`
	Department    string `json:"department"`
}

type UpdateCompanyInput struct {
	Name          string `json:"name"`
	ARL           string `json:"arl"`
	LegalRep      string `json:"legal_rep"`
	LegalRepEmail string `json:"legal_rep_email"`
	LegalRepPhone string `json:"legal_rep_phone"`
	Address       string `json:"address"`
	LogoURL       string `json:"logo_url"`
}
