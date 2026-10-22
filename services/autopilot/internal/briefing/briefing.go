// Package briefing builds the morning briefing of an SST professional from
// Core's records. It is read-only: it never creates, changes or approves a
// record — it lists what needs the professional's attention today, and each
// item links to where the human acts.
package briefing

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Item struct {
	CompanyID   string `json:"company_id"`
	CompanyName string `json:"company_name"`
	Kind        string `json:"kind"`
	Severity    string `json:"severity"` // high | medium | low
	Text        string `json:"text"`
	Basis       string `json:"basis,omitempty"` // normative or service-rule basis, shown to the user
	Link        string `json:"link"`
}

type Briefing struct {
	ID          string          `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantID    string          `gorm:"type:uuid;not null;uniqueIndex:idx_briefing_day,priority:1" json:"tenant_id"`
	UserID      string          `gorm:"type:uuid;not null;uniqueIndex:idx_briefing_day,priority:2" json:"user_id"`
	Day         string          `gorm:"type:varchar(10);not null;uniqueIndex:idx_briefing_day,priority:3" json:"day"` // YYYY-MM-DD, America/Bogota
	Companies   int             `json:"companies"`
	ItemsJSON   json.RawMessage `gorm:"column:items;type:jsonb;not null" json:"items"`
	GeneratedAt time.Time       `json:"generated_at"`
	NotifiedAt  *time.Time      `json:"notified_at"`
}

func (Briefing) TableName() string { return "briefings" }

type Publisher interface {
	Publish(ctx context.Context, exchange, key string, body []byte) error
}

type Service struct {
	db  *gorm.DB
	pub Publisher
	now func() time.Time
	// OnlyTenant limits the daily run to one tenant (staged rollout / QA).
	OnlyTenant string
}

func New(db *gorm.DB, pub Publisher) *Service { return &Service{db: db, pub: pub, now: time.Now} }

func (s *Service) Migrate() error { return s.db.AutoMigrate(&Briefing{}) }

var Bogota = func() *time.Location {
	l, err := time.LoadLocation("America/Bogota")
	if err != nil {
		return time.FixedZone("COT", -5*3600)
	}
	return l
}()

type company struct {
	ID, Name string
}

type professional struct {
	ID, TenantID, Email, FirstName string
}

// companiesOf lists the active companies a professional is a member of.
func (s *Service) companiesOf(ctx context.Context, tenantID, userID string) ([]company, error) {
	var out []company
	err := s.db.WithContext(ctx).Raw(`
		SELECT c.id, c.name FROM public.companies c
		JOIN public.company_memberships m ON m.company_id = c.id AND m.tenant_id = c.tenant_id
		WHERE m.user_id = ? AND m.tenant_id = ? AND m.is_active AND c.deleted_at IS NULL
		ORDER BY c.name`, userID, tenantID).Scan(&out).Error
	return out, err
}

type check struct {
	kind, severity, link, basis string
	sql                         string
	text                        func(n int64) string
}

// Each check is a count over one company. Thresholds are deliberately plain
// and visible in the text so the professional can judge them.
var checks = []check{
	{kind: "inspection_unapproved", severity: "high", link: "/dashboard/inspections",
		sql:  `SELECT count(*) FROM public.inspections WHERE company_id = @c AND tenant_id = @t AND deleted_at IS NULL AND status = 'completada' AND jsonb_array_length(action_plan) > 0`,
		text: func(n int64) string { return fmt.Sprintf("%d inspección(es) con hallazgos sin revisar ni aprobar", n) }},
	{kind: "sla_breached", severity: "high", link: "/dashboard/inspections", basis: "Plazo de servicio ZU 360",
		sql:  `SELECT count(*) FROM sla.events WHERE company_id = @c AND tenant_id = @t AND closed_at IS NULL AND breached_at IS NOT NULL`,
		text: func(n int64) string { return fmt.Sprintf("%d plazo(s) de atención vencido(s)", n) }},
	{kind: "epp_expiring", severity: "medium", link: "/dashboard/epp",
		sql: `SELECT count(*) FROM public.epp_deliveries WHERE company_id = @c AND tenant_id = @t AND deleted_at IS NULL AND returned_at IS NULL AND expires_at IS NOT NULL AND expires_at < now() + interval '15 days'`,
		text: func(n int64) string {
			return fmt.Sprintf("%d entrega(s) de EPP vencida(s) o por vencer en 15 días", n)
		}},
	{kind: "copasst_no_meeting", severity: "medium", link: "/dashboard/committees", basis: "Res. 2013 de 1986, art. 7: el COPASST se reúne al menos una vez al mes",
		sql: `SELECT count(*) FROM public.committees k WHERE k.company_id = @c AND k.tenant_id = @t AND k.deleted_at IS NULL AND lower(k.type) LIKE '%copasst%'
		      AND now() BETWEEN k.period_start AND k.period_end
		      AND NOT EXISTS (SELECT 1 FROM public.committee_meetings m WHERE m.committee_id = k.id AND m.deleted_at IS NULL AND m.held_at > now() - interval '30 days')`,
		text: func(n int64) string { return "COPASST sin reunión registrada en los últimos 30 días" }},
	{kind: "committee_period_ending", severity: "low", link: "/dashboard/committees",
		sql: `SELECT count(*) FROM public.committees WHERE company_id = @c AND tenant_id = @t AND deleted_at IS NULL AND period_end BETWEEN now() AND now() + interval '30 days'`,
		text: func(n int64) string {
			return "Periodo del comité termina en menos de 30 días: preparar la nueva conformación"
		}},
	{kind: "legal_noncompliant", severity: "medium", link: "/dashboard/legal-matrix",
		sql: `SELECT count(*) FROM public.legal_requirements WHERE company_id = @c AND tenant_id = @t AND deleted_at IS NULL AND applicable AND (reviewed_at IS NULL OR lower(coalesce(compliance_status,'')) IN ('no_cumple','no cumple','incumple','parcial'))`,
		text: func(n int64) string {
			return fmt.Sprintf("%d requisito(s) legal(es) aplicable(s) incumplido(s) o sin revisión", n)
		}},
	{kind: "objective_deadline", severity: "low", link: "/dashboard/management",
		sql: `SELECT count(*) FROM public.sst_objectives o JOIN public.management_systems ms ON ms.id = o.system_id
		      WHERE ms.company_id = @c AND o.tenant_id = @t AND o.deadline BETWEEN now() AND now() + interval '30 days' AND lower(coalesce(o.status,'')) NOT IN ('cumplido','completado','done')`,
		text: func(n int64) string { return fmt.Sprintf("%d objetivo(s) del SG-SST con fecha límite en 30 días", n) }},
}

func (s *Service) itemsFor(ctx context.Context, tenantID string, c company) ([]Item, error) {
	var items []Item
	for _, ch := range checks {
		var n int64
		if err := s.db.WithContext(ctx).Raw(ch.sql, map[string]interface{}{"c": c.ID, "t": tenantID}).Scan(&n).Error; err != nil {
			return nil, fmt.Errorf("%s: %w", ch.kind, err)
		}
		if n > 0 {
			items = append(items, Item{CompanyID: c.ID, CompanyName: c.Name, Kind: ch.kind, Severity: ch.severity, Text: ch.text(n), Basis: ch.basis, Link: ch.link})
		}
	}
	return items, nil
}

// Build computes (and stores, replacing today's) the briefing of one user.
func (s *Service) Build(ctx context.Context, tenantID, userID string) (*Briefing, []Item, error) {
	companies, err := s.companiesOf(ctx, tenantID, userID)
	if err != nil {
		return nil, nil, err
	}
	items := []Item{}
	for _, c := range companies {
		ci, err := s.itemsFor(ctx, tenantID, c)
		if err != nil {
			return nil, nil, err
		}
		items = append(items, ci...)
	}
	raw, _ := json.Marshal(items)
	b := &Briefing{TenantID: tenantID, UserID: userID, Day: s.now().In(Bogota).Format("2006-01-02"),
		Companies: len(companies), ItemsJSON: raw, GeneratedAt: s.now()}
	err = s.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "tenant_id"}, {Name: "user_id"}, {Name: "day"}},
		DoUpdates: clause.AssignmentColumns([]string{"companies", "items", "generated_at"}),
	}).Create(b).Error
	if err != nil {
		return nil, nil, err
	}
	s.db.WithContext(ctx).Where("tenant_id = ? AND user_id = ? AND day = ?", tenantID, userID, b.Day).First(b)
	return b, items, nil
}

// RunDaily builds and notifies the briefing of every active SST professional.
// A briefing already notified today is not sent twice (restart-safe).
func (s *Service) RunDaily(ctx context.Context) (sent int, err error) {
	var pros []professional
	q := `SELECT id, tenant_id, email, coalesce(first_name,'') AS first_name FROM public.users
		WHERE role = 'professional_sst' AND is_active AND deleted_at IS NULL AND (? = '' OR tenant_id::text = ?)`
	if err := s.db.WithContext(ctx).Raw(q, s.OnlyTenant, s.OnlyTenant).Scan(&pros).Error; err != nil {
		return 0, err
	}
	for _, p := range pros {
		b, items, err := s.Build(ctx, p.TenantID, p.ID)
		if err != nil {
			log.Printf("briefing for %s failed: %v", p.ID, err)
			continue
		}
		if b.NotifiedAt != nil || b.Companies == 0 {
			continue
		}
		if err := s.notify(ctx, p, b, items); err != nil {
			log.Printf("briefing notify for %s failed: %v", p.ID, err)
			continue
		}
		now := s.now()
		s.db.WithContext(ctx).Model(&Briefing{}).Where("id = ?", b.ID).Update("notified_at", now)
		sent++
	}
	return sent, nil
}

func (s *Service) notify(ctx context.Context, p professional, b *Briefing, items []Item) error {
	lines := make([]string, 0, len(items))
	high := 0
	for _, it := range items {
		line := it.CompanyName + ": " + it.Text
		if it.Basis != "" {
			line += " (" + it.Basis + ")"
		}
		lines = append(lines, line)
		if it.Severity == "high" {
			high++
		}
	}
	title := fmt.Sprintf("Briefing del %s: %d pendiente(s) en %d empresa(s)", b.Day, len(items), b.Companies)
	msg := "Sin pendientes detectados en los registros de su cartera."
	if len(items) > 0 {
		msg = fmt.Sprintf("%d pendiente(s), %d de prioridad alta. Abra el briefing para ver el detalle por empresa.", len(items), high)
	}
	body, _ := json.Marshal(map[string]interface{}{
		"kind": "briefing.morning", "tenant_id": p.TenantID, "user_id": p.ID, "email": p.Email,
		"channels": []string{"inapp", "email"}, "priority": "normal", "link": "/dashboard",
		"title": title, "message": msg,
		"mail_template": "morning_briefing",
		"variables":     map[string]string{"date": b.Day, "companies": fmt.Sprint(b.Companies)},
		"items":         lines,
	})
	return s.pub.Publish(ctx, "zu.notifications", "notify.briefing.morning", body)
}

func (s *Service) Today(ctx context.Context, tenantID, userID string) (*Briefing, error) {
	var b Briefing
	err := s.db.WithContext(ctx).Where("tenant_id = ? AND user_id = ?", tenantID, userID).Order("day DESC").First(&b).Error
	return &b, err
}
