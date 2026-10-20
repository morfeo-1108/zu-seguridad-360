package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/zu-seguridad/core/internal/model"
)

type Alert struct {
	Category    string    `json:"category"` // medical_exam | epp | risk_review | committee | legal_requirement
	Severity    string    `json:"severity"` // critical | warning
	Title       string    `json:"title"`
	Detail      string    `json:"detail"`
	DueAt       time.Time `json:"due_at"`
	ReferenceID uuid.UUID `json:"reference_id"`
}

// AlertService composes the "próximos vencimientos" views that already
// existed separately per module (medical exam/EPP expiry, risk review
// cadence, committee period end, pending legal evaluations) into one feed.
// It does not introduce a new source of truth or a background scheduler —
// no module here sends notifications on its own; this is a pull-based
// aggregation a dashboard (or, later, a cron that calls it) can read.
type AlertService struct {
	medical *MedicalExamService
	epp     *EPPService
	risk    *RiskService
	legal   *LegalRequirementService
}

func NewAlertService(medical *MedicalExamService, epp *EPPService, risk *RiskService, legal *LegalRequirementService) *AlertService {
	return &AlertService{medical: medical, epp: epp, risk: risk, legal: legal}
}

func (s *AlertService) GetAll(ctx context.Context, companyID, tenantID uuid.UUID) ([]Alert, error) {
	var alerts []Alert

	exams, err := s.medical.GetExpiring(ctx, tenantID, companyID, 30)
	if err != nil && err != ErrSensitiveDataProcessingDisabled {
		return nil, err
	}
	for _, e := range exams {
		if e.ExpiresAt == nil {
			continue
		}
		alerts = append(alerts, Alert{
			Category: "medical_exam", Severity: severityFor(*e.ExpiresAt),
			Title: "Examen médico por vencer", Detail: string(e.Type), DueAt: *e.ExpiresAt, ReferenceID: e.ID,
		})
	}

	deliveries, err := s.epp.ListExpiring(ctx, companyID, tenantID, 30)
	if err != nil {
		return nil, err
	}
	for _, d := range deliveries {
		if d.ExpiresAt == nil {
			continue
		}
		title := "EPP por renovar"
		if d.Item.Name != "" {
			title = "Renovar " + d.Item.Name
		}
		alerts = append(alerts, Alert{
			Category: "epp", Severity: severityFor(*d.ExpiresAt),
			Title: title, Detail: d.Condition, DueAt: *d.ExpiresAt, ReferenceID: d.ID,
		})
	}

	risks, err := s.risk.ListNeedingReview(ctx, companyID, tenantID, 30)
	if err != nil {
		return nil, err
	}
	for _, r := range risks {
		if r.NextReviewAt == nil {
			continue
		}
		alerts = append(alerts, Alert{
			Category: "risk_review", Severity: severityFor(*r.NextReviewAt),
			Title: "Revisar riesgo: " + r.Process, Detail: r.HazardSource, DueAt: *r.NextReviewAt, ReferenceID: r.ID,
		})
	}

	legalReqs, err := s.legal.List(ctx, companyID, tenantID)
	if err != nil {
		return nil, err
	}
	for _, l := range legalReqs {
		if l.ComplianceStatus == model.LegalStatusPending {
			alerts = append(alerts, Alert{
				Category: "legal_requirement", Severity: "warning",
				Title: "Requisito legal sin evaluar", Detail: l.NormReference, DueAt: time.Now().UTC(), ReferenceID: l.ID,
			})
		}
	}

	return alerts, nil
}

func severityFor(due time.Time) string {
	if due.Before(time.Now().UTC().AddDate(0, 0, 7)) {
		return "critical"
	}
	return "warning"
}
