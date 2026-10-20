package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/zu-seguridad/core/internal/model"
	"github.com/zu-seguridad/core/internal/repository"
	"gorm.io/gorm"
)

type CompanySummary struct {
	CompanyID      uuid.UUID `json:"company_id"`
	Name           string    `json:"name"`
	IsSynthetic    bool      `json:"is_synthetic"`
	ActiveEmployees int64    `json:"active_employees"`
	OpenIncidents  int64     `json:"open_incidents"`
	CriticalRisks  int64     `json:"critical_risks"`
	PendingLegal   int64     `json:"pending_legal_requirements"`
}

// PortfolioService answers "how is my whole portfolio of companies doing",
// a single query per metric across every company the professional has
// membership to — not N client-side round trips per company.
type PortfolioService struct {
	db         *gorm.DB
	companyRepo *repository.CompanyRepository
}

func NewPortfolioService(db *gorm.DB, companyRepo *repository.CompanyRepository) *PortfolioService {
	return &PortfolioService{db: db, companyRepo: companyRepo}
}

func (s *PortfolioService) Summary(ctx context.Context, userID, tenantID uuid.UUID) ([]CompanySummary, error) {
	companies, err := s.companyRepo.FindAllForUser(ctx, userID, tenantID)
	if err != nil {
		return nil, err
	}
	if len(companies) == 0 {
		return []CompanySummary{}, nil
	}

	companyIDs := make([]uuid.UUID, len(companies))
	for i, c := range companies {
		companyIDs[i] = c.ID
	}

	employeeCounts := map[uuid.UUID]int64{}
	var employeeRows []struct {
		CompanyID uuid.UUID
		Count     int64
	}
	if err := s.db.WithContext(ctx).Model(&model.Employee{}).
		Select("company_id, count(*) as count").
		Where("company_id IN ? AND tenant_id = ? AND status = 'active'", companyIDs, tenantID).
		Group("company_id").Scan(&employeeRows).Error; err != nil {
		return nil, err
	}
	for _, r := range employeeRows {
		employeeCounts[r.CompanyID] = r.Count
	}

	incidentCounts := map[uuid.UUID]int64{}
	var incidentRows []struct {
		CompanyID uuid.UUID
		Count     int64
	}
	if err := s.db.WithContext(ctx).Model(&model.Incident{}).
		Select("company_id, count(*) as count").
		Where("company_id IN ? AND tenant_id = ? AND status = 'open'", companyIDs, tenantID).
		Group("company_id").Scan(&incidentRows).Error; err != nil {
		return nil, err
	}
	for _, r := range incidentRows {
		incidentCounts[r.CompanyID] = r.Count
	}

	riskCounts := map[uuid.UUID]int64{}
	var riskRows []struct {
		CompanyID uuid.UUID
		Count     int64
	}
	if err := s.db.WithContext(ctx).Model(&model.RiskEntry{}).
		Select("company_id, count(*) as count").
		Where("company_id IN ? AND tenant_id = ? AND acceptability = ?", companyIDs, tenantID, model.RiskLevelCritical).
		Group("company_id").Scan(&riskRows).Error; err != nil {
		return nil, err
	}
	for _, r := range riskRows {
		riskCounts[r.CompanyID] = r.Count
	}

	legalCounts := map[uuid.UUID]int64{}
	var legalRows []struct {
		CompanyID uuid.UUID
		Count     int64
	}
	if err := s.db.WithContext(ctx).Model(&model.LegalRequirement{}).
		Select("company_id, count(*) as count").
		Where("company_id IN ? AND tenant_id = ? AND compliance_status = ?", companyIDs, tenantID, model.LegalStatusPending).
		Group("company_id").Scan(&legalRows).Error; err != nil {
		return nil, err
	}
	for _, r := range legalRows {
		legalCounts[r.CompanyID] = r.Count
	}

	summaries := make([]CompanySummary, len(companies))
	for i, c := range companies {
		summaries[i] = CompanySummary{
			CompanyID: c.ID, Name: c.Name, IsSynthetic: c.IsSynthetic,
			ActiveEmployees: employeeCounts[c.ID], OpenIncidents: incidentCounts[c.ID],
			CriticalRisks: riskCounts[c.ID], PendingLegal: legalCounts[c.ID],
		}
	}
	return summaries, nil
}
