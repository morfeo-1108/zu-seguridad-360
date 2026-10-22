// Package service tracks a client company's SG-SST onboarding from the data
// that actually exists in Core. A step is "done" only when its record exists;
// nothing is marked done by a timer or by the service deciding so. The first
// time each step is seen done is persisted, which makes onboarding duration a
// measured number rather than a promise.
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Step struct {
	Key         string `json:"key"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Link        string `json:"link"`
	// Query counts rows in Core that satisfy the step for (tenant, company).
	// Blocked steps have no query and explain why in BlockedReason.
	query         string
	BlockedReason string `json:"blocked_reason,omitempty"`
}

// Core tables read here are granted to this service's role column-by-column
// (SELECT only); see scripts/provision-service.sh.
var Steps = []Step{
	{Key: "company_profile", Title: "Ficha de la empresa", Description: "NIT, CIIU, clase de riesgo y ARL registrados.", Link: "/dashboard/companies",
		query: `SELECT count(*) FROM public.companies WHERE id = @company AND tenant_id = @tenant AND deleted_at IS NULL AND coalesce(nit,'') <> '' AND coalesce(ciiu,'') <> '' AND coalesce(risk_class,'') <> '' AND coalesce(arl,'') <> ''`},
	{Key: "legal_matrix", Title: "Matriz legal", Description: "Al menos un requisito legal aplicable identificado.", Link: "/dashboard/legal-matrix",
		query: `SELECT count(*) FROM public.legal_requirements WHERE company_id = @company AND tenant_id = @tenant AND deleted_at IS NULL`},
	{Key: "risk_matrix", Title: "Matriz de peligros (GTC 45)", Description: "Al menos un peligro identificado y valorado.", Link: "/dashboard/risks",
		query: `SELECT count(*) FROM public.risk_entries WHERE company_id = @company AND tenant_id = @tenant AND deleted_at IS NULL`},
	{Key: "epp_catalog", Title: "Catálogo de EPP", Description: "Elementos de protección personal registrados.", Link: "/dashboard/epp",
		query: `SELECT count(*) FROM public.epp_items WHERE company_id = @company AND tenant_id = @tenant AND deleted_at IS NULL`},
	{Key: "committee", Title: "COPASST o Vigía", Description: "Comité o vigía conformado para el periodo.", Link: "/dashboard/committees",
		query: `SELECT count(*) FROM public.committees WHERE company_id = @company AND tenant_id = @tenant AND deleted_at IS NULL`},
	{Key: "phva", Title: "Ciclo PHVA iniciado", Description: "Sistema de gestión inicializado sobre al menos un marco normativo.", Link: "/dashboard/management",
		query: `SELECT count(*) FROM public.management_systems WHERE company_id = @company AND tenant_id = @tenant`},
	{Key: "first_inspection", Title: "Primera inspección", Description: "Una inspección de seguridad registrada.", Link: "/dashboard/inspections",
		query: `SELECT count(*) FROM public.inspections WHERE company_id = @company AND tenant_id = @tenant AND deleted_at IS NULL`},
	{Key: "documents", Title: "Documentación base", Description: "Al menos un documento del SG-SST cargado o generado.", Link: "/dashboard/documents",
		query: `SELECT count(*) FROM public.documents WHERE company_id = @company AND tenant_id = @tenant AND deleted_at IS NULL`},
	{Key: "workers", Title: "Trabajadores", Description: "Directorio de trabajadores y exámenes médicos.", Link: "/dashboard/employees",
		BlockedReason: "Bloqueado hasta cerrar la compuerta de datos reales de trabajadores (dato sensible, Ley 1581 de 2012)."},
}

type Onboarding struct {
	TenantID    string          `gorm:"type:uuid;primaryKey" json:"tenant_id"`
	CompanyID   string          `gorm:"type:uuid;primaryKey" json:"company_id"`
	StartedAt   time.Time       `gorm:"not null" json:"started_at"`
	CompletedAt *time.Time      `json:"completed_at"`
	StepDoneAt  json.RawMessage `gorm:"type:jsonb;not null;default:'{}'" json:"-"` // step key → first time seen done
	UpdatedAt   time.Time       `json:"updated_at"`
}

func (Onboarding) TableName() string { return "onboardings" }

type StepStatus struct {
	Step
	Status string     `json:"status"` // done | pending | blocked
	Count  int64      `json:"count"`
	DoneAt *time.Time `json:"done_at,omitempty"`
}

type Progress struct {
	CompanyID    string       `json:"company_id"`
	StartedAt    time.Time    `json:"started_at"`
	CompletedAt  *time.Time   `json:"completed_at"`
	Done         int          `json:"done"`
	Total        int          `json:"total"` // excludes blocked steps
	Steps        []StepStatus `json:"steps"`
	ElapsedHours float64      `json:"elapsed_hours"`
}

type Publisher interface {
	Publish(ctx context.Context, exchange, key string, body []byte) error
}

type Tracker struct {
	db  *gorm.DB
	pub Publisher
	now func() time.Time
}

func NewTracker(db *gorm.DB, pub Publisher) *Tracker {
	return &Tracker{db: db, pub: pub, now: time.Now}
}

func (t *Tracker) Migrate() error { return t.db.AutoMigrate(&Onboarding{}) }

// Start registers a company's onboarding (idempotent) — called on Core's
// company.created event, or lazily on first read for companies that predate
// this service.
func (t *Tracker) Start(ctx context.Context, tenantID, companyID string, at time.Time) error {
	if at.IsZero() {
		at = t.now()
	}
	return t.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).
		Create(&Onboarding{TenantID: tenantID, CompanyID: companyID, StartedAt: at, StepDoneAt: json.RawMessage("{}")}).Error
}

// companyVisible confirms the company exists in Core under that tenant, so a
// forged company id in an event cannot create rows for someone else.
func (t *Tracker) companyVisible(ctx context.Context, tenantID, companyID string) (bool, error) {
	var n int64
	err := t.db.WithContext(ctx).Raw(`SELECT count(*) FROM public.companies WHERE id = ? AND tenant_id = ? AND deleted_at IS NULL`, companyID, tenantID).Scan(&n).Error
	return n == 1, err
}

// Recompute evaluates every step against Core, records first-done times and
// emits a one-time notification when the last step completes.
func (t *Tracker) Recompute(ctx context.Context, tenantID, companyID string) (*Progress, error) {
	ok, err := t.companyVisible(ctx, tenantID, companyID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	if err := t.Start(ctx, tenantID, companyID, time.Time{}); err != nil {
		return nil, err
	}
	var row Onboarding
	if err := t.db.WithContext(ctx).Where("tenant_id = ? AND company_id = ?", tenantID, companyID).First(&row).Error; err != nil {
		return nil, err
	}
	doneAt := map[string]time.Time{}
	_ = json.Unmarshal(row.StepDoneAt, &doneAt)

	p := &Progress{CompanyID: companyID, StartedAt: row.StartedAt}
	now := t.now()
	for _, s := range Steps {
		st := StepStatus{Step: s, Status: "pending"}
		if s.query == "" {
			st.Status = "blocked"
			p.Steps = append(p.Steps, st)
			continue
		}
		p.Total++
		if err := t.db.WithContext(ctx).Raw(s.query, map[string]interface{}{"tenant": tenantID, "company": companyID}).Scan(&st.Count).Error; err != nil {
			return nil, fmt.Errorf("step %s: %w", s.Key, err)
		}
		if st.Count > 0 {
			st.Status = "done"
			p.Done++
			if _, seen := doneAt[s.Key]; !seen {
				doneAt[s.Key] = now
			}
		}
		if at, seen := doneAt[s.Key]; seen && st.Status == "done" {
			at := at
			st.DoneAt = &at
		}
		p.Steps = append(p.Steps, st)
	}

	justCompleted := false
	if p.Done == p.Total && row.CompletedAt == nil {
		row.CompletedAt = &now
		justCompleted = true
	}
	raw, _ := json.Marshal(doneAt)
	if err := t.db.WithContext(ctx).Model(&Onboarding{}).Where("tenant_id = ? AND company_id = ?", tenantID, companyID).
		Updates(map[string]interface{}{"step_done_at": raw, "completed_at": row.CompletedAt, "updated_at": now}).Error; err != nil {
		return nil, err
	}
	p.CompletedAt = row.CompletedAt
	end := now
	if p.CompletedAt != nil {
		end = *p.CompletedAt
	}
	p.ElapsedHours = end.Sub(p.StartedAt).Hours()
	if justCompleted {
		t.notifyCompleted(ctx, tenantID, companyID, p)
	}
	return p, nil
}

func (t *Tracker) notifyCompleted(ctx context.Context, tenantID, companyID string, p *Progress) {
	var c struct{ Name string }
	t.db.WithContext(ctx).Raw(`SELECT name FROM public.companies WHERE id = ? AND tenant_id = ?`, companyID, tenantID).Scan(&c)
	body, _ := json.Marshal(map[string]interface{}{
		"kind": "onboarding.completed", "tenant_id": tenantID, "company_id": companyID,
		"channels": []string{"inapp"}, "priority": "normal", "link": "/dashboard",
		"title":   "Onboarding completo: " + c.Name,
		"message": fmt.Sprintf("Los %d pasos base del SG-SST quedaron registrados en %.1f días.", p.Total, p.ElapsedHours/24),
	})
	if err := t.pub.Publish(ctx, "zu.notifications", "notify.onboarding.completed", body); err != nil {
		log.Printf("onboarding completion notice not published for %s: %v", companyID, err)
	}
}

// RecomputeAll refreshes every onboarding still in progress.
func (t *Tracker) RecomputeAll(ctx context.Context) {
	var rows []Onboarding
	if err := t.db.WithContext(ctx).Where("completed_at IS NULL").Find(&rows).Error; err != nil {
		log.Printf("recompute list failed: %v", err)
		return
	}
	for _, r := range rows {
		if _, err := t.Recompute(ctx, r.TenantID, r.CompanyID); err != nil {
			log.Printf("recompute %s failed: %v", r.CompanyID, err)
		}
	}
}

// Portfolio returns one progress per company of the tenant that has an
// onboarding record (the professional's whole client list).
func (t *Tracker) Portfolio(ctx context.Context, tenantID string) ([]Progress, error) {
	var ids []string
	if err := t.db.WithContext(ctx).Raw(`SELECT id FROM public.companies WHERE tenant_id = ? AND deleted_at IS NULL ORDER BY created_at`, tenantID).Scan(&ids).Error; err != nil {
		return nil, err
	}
	out := make([]Progress, 0, len(ids))
	for _, id := range ids {
		p, err := t.Recompute(ctx, tenantID, id)
		if err != nil {
			return nil, err
		}
		out = append(out, *p)
	}
	return out, nil
}
