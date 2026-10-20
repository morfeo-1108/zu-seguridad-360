package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/zu-seguridad/core/internal/model"
	"gorm.io/gorm"
)

type RiskService struct{ db *gorm.DB }

func NewRiskService(db *gorm.DB) *RiskService { return &RiskService{db: db} }

type CreateRiskInput struct {
	Process        string `json:"process" validate:"required"`
	Area           string `json:"area" validate:"required"`
	HazardType     string `json:"hazard_type" validate:"required"`
	HazardSource   string `json:"hazard_source" validate:"required"`
	HealthEffect   string `json:"health_effect" validate:"required"`
	WorkersExposed int    `json:"workers_exposed"`
	ND             int    `json:"nd" validate:"required,min=1,max=10"`
	NE             int    `json:"ne" validate:"required,min=1,max=4"`
	NC             int    `json:"nc" validate:"required,min=10,max=100"`
	CurrentControl string `json:"current_control"`
}

func (s *RiskService) CreateEntry(ctx context.Context, input CreateRiskInput, companyID, tenantID uuid.UUID) (*model.RiskEntry, error) {
	entry := &model.RiskEntry{
		TenantID:       tenantID,
		CompanyID:      companyID,
		Process:        input.Process,
		Area:           input.Area,
		HazardType:     input.HazardType,
		HazardSource:   input.HazardSource,
		HealthEffect:   input.HealthEffect,
		WorkersExposed: input.WorkersExposed,
		ND:             input.ND,
		NE:             input.NE,
		NC:             input.NC,
		CurrentControl: input.CurrentControl,
	}
	entry.CalculateNR() // GTC 45: ND×NE=NP; NP×NC=NR
	nextReview := nextReviewFor(entry.Acceptability)
	entry.NextReviewAt = &nextReview
	return entry, s.db.WithContext(ctx).Create(entry).Error
}

// nextReviewFor sets a review cadence by risk level: critical risks get
// reviewed far sooner than acceptable ones. There is no single universal
// GTC-45 review interval — this is a reasonable, documented default, not a
// cited normative requirement.
func nextReviewFor(level model.RiskLevel) time.Time {
	now := time.Now().UTC()
	switch level {
	case model.RiskLevelCritical:
		return now.AddDate(0, 3, 0)
	case model.RiskLevelNotAcceptable:
		return now.AddDate(0, 6, 0)
	default:
		return now.AddDate(1, 0, 0)
	}
}

func (s *RiskService) ListNeedingReview(ctx context.Context, companyID, tenantID uuid.UUID, days int) ([]model.RiskEntry, error) {
	cutoff := time.Now().UTC().AddDate(0, 0, days)
	var entries []model.RiskEntry
	err := s.db.WithContext(ctx).
		Where("company_id = ? AND tenant_id = ? AND next_review_at IS NOT NULL AND next_review_at <= ?", companyID, tenantID, cutoff).
		Order("next_review_at ASC").Find(&entries).Error
	return entries, err
}

func (s *RiskService) GetMatrix(ctx context.Context, companyID, tenantID uuid.UUID) ([]model.RiskEntry, error) {
	var entries []model.RiskEntry
	err := s.db.WithContext(ctx).
		Where("company_id = ? AND tenant_id = ?", companyID, tenantID).
		Order("nr DESC").Find(&entries).Error
	return entries, err
}

func (s *RiskService) GetHeatMap(ctx context.Context, companyID, tenantID uuid.UUID) (map[string]interface{}, error) {
	entries, err := s.GetMatrix(ctx, companyID, tenantID)
	if err != nil {
		return nil, err
	}

	heatmap := map[string]interface{}{
		"critical":       []model.RiskEntry{},
		"not_acceptable": []model.RiskEntry{},
		"acceptable":     []model.RiskEntry{},
		"summary": map[string]int{
			"total": len(entries), "critical": 0, "not_acceptable": 0, "acceptable": 0,
		},
	}

	summary := heatmap["summary"].(map[string]int)
	for _, e := range entries {
		switch e.Acceptability {
		case model.RiskLevelCritical:
			heatmap["critical"] = append(heatmap["critical"].([]model.RiskEntry), e)
			summary["critical"]++
		case model.RiskLevelNotAcceptable:
			heatmap["not_acceptable"] = append(heatmap["not_acceptable"].([]model.RiskEntry), e)
			summary["not_acceptable"]++
		default:
			heatmap["acceptable"] = append(heatmap["acceptable"].([]model.RiskEntry), e)
			summary["acceptable"]++
		}
	}
	return heatmap, nil
}
