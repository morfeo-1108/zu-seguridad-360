// Package service tracks response deadlines (SLA) opened and closed by domain
// events from Core, and raises a notification when one is missed.
package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Kinds of deadline tracked today. Each one is opened and closed by a pair of
// Core events; nothing here is opened by hand through an API.
const KindInspectionReview = "inspection.review"

var kindLabels = map[string]string{
	KindInspectionReview: "Revisión de inspección con hallazgos",
}

type Event struct {
	ID            string     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	SourceEventID string     `gorm:"type:uuid;uniqueIndex;not null" json:"-"`
	TenantID      string     `gorm:"type:uuid;not null;index:idx_sla_scope,priority:1" json:"tenant_id"`
	CompanyID     string     `gorm:"type:uuid;not null;index:idx_sla_scope,priority:2" json:"company_id"`
	CompanyName   string     `gorm:"size:255" json:"company_name"`
	Kind          string     `gorm:"size:64;not null" json:"kind"`
	ResourceID    string     `gorm:"type:uuid;not null;index" json:"resource_id"`
	ResourceLabel string     `gorm:"size:255" json:"resource_label"`
	NotifyUserID  string     `gorm:"type:uuid" json:"-"`
	NotifyEmail   string     `gorm:"size:255" json:"-"`
	OpenedAt      time.Time  `gorm:"not null" json:"opened_at"`
	DueAt         time.Time  `gorm:"not null;index" json:"due_at"`
	ClosedAt      *time.Time `json:"closed_at"`
	BreachedAt    *time.Time `json:"breached_at"`
}

func (Event) TableName() string { return "events" }

// CoreEvent is the payload Core publishes on zu.companies.
type CoreEvent struct {
	EventID       string    `json:"event_id"`
	Type          string    `json:"type"`
	OccurredAt    time.Time `json:"occurred_at"`
	TenantID      string    `json:"tenant_id"`
	CompanyID     string    `json:"company_id"`
	CompanyName   string    `json:"company_name"`
	ResourceID    string    `json:"resource_id"`
	ResourceLabel string    `json:"resource_label"`
	ActorUserID   string    `json:"actor_user_id"`
	ActorEmail    string    `json:"actor_email"`
}

var ErrMalformed = errors.New("malformed event")

type Publisher interface {
	Publish(ctx context.Context, exchange, key string, body []byte) error
}

type Monitor struct {
	db      *gorm.DB
	pub     Publisher
	targets map[string]time.Duration
	now     func() time.Time
}

func NewMonitor(db *gorm.DB, pub Publisher, inspectionReview time.Duration) *Monitor {
	return &Monitor{db: db, pub: pub, now: time.Now, targets: map[string]time.Duration{
		KindInspectionReview: inspectionReview,
	}}
}

func (m *Monitor) Migrate() error { return m.db.AutoMigrate(&Event{}) }

// Handle applies one Core event. It is idempotent: a redelivered "opened"
// event hits the unique source_event_id, and closing an already-closed
// deadline is a no-op.
func (m *Monitor) Handle(ctx context.Context, ev CoreEvent) error {
	if ev.EventID == "" || ev.TenantID == "" || ev.CompanyID == "" || ev.ResourceID == "" {
		return ErrMalformed
	}
	switch ev.Type {
	case "inspection.completed":
		return m.open(ctx, KindInspectionReview, ev)
	case "inspection.approved":
		return m.close(ctx, KindInspectionReview, ev)
	}
	return nil // other company events are not deadline-relevant
}

func (m *Monitor) open(ctx context.Context, kind string, ev CoreEvent) error {
	opened := ev.OccurredAt
	if opened.IsZero() {
		opened = m.now()
	}
	row := Event{
		SourceEventID: ev.EventID, TenantID: ev.TenantID, CompanyID: ev.CompanyID, CompanyName: ev.CompanyName,
		Kind: kind, ResourceID: ev.ResourceID, ResourceLabel: ev.ResourceLabel,
		NotifyUserID: ev.ActorUserID, NotifyEmail: ev.ActorEmail,
		OpenedAt: opened, DueAt: opened.Add(m.targets[kind]),
	}
	return m.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&row).Error
}

func (m *Monitor) close(ctx context.Context, kind string, ev CoreEvent) error {
	closed := ev.OccurredAt
	if closed.IsZero() {
		closed = m.now()
	}
	return m.db.WithContext(ctx).Model(&Event{}).
		Where("tenant_id = ? AND company_id = ? AND resource_id = ? AND kind = ? AND closed_at IS NULL", ev.TenantID, ev.CompanyID, ev.ResourceID, kind).
		Update("closed_at", closed).Error
}

// CheckBreaches marks overdue deadlines and publishes one notification per
// breach. Mark and publish share a transaction: if the publish fails the row
// stays unmarked and the next tick retries, so a breach is never silently
// dropped. SKIP LOCKED lets several replicas run the watcher safely.
func (m *Monitor) CheckBreaches(ctx context.Context) (int, error) {
	count := 0
	err := m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var due []Event
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
			Where("closed_at IS NULL AND breached_at IS NULL AND due_at < ?", m.now()).
			Limit(100).Find(&due).Error; err != nil {
			return err
		}
		for _, e := range due {
			now := m.now()
			if err := tx.Model(&Event{}).Where("id = ?", e.ID).Update("breached_at", now).Error; err != nil {
				return err
			}
			if err := m.notifyBreach(ctx, e); err != nil {
				return err
			}
			count++
		}
		return nil
	})
	return count, err
}

func (m *Monitor) notifyBreach(ctx context.Context, e Event) error {
	label := kindLabels[e.Kind]
	target := m.targets[e.Kind]
	payload := map[string]interface{}{
		"kind": "sla.breach", "tenant_id": e.TenantID, "company_id": e.CompanyID,
		"user_id": e.NotifyUserID, "email": e.NotifyEmail,
		"channels": []string{"inapp", "email"}, "priority": "high",
		"title":         fmt.Sprintf("Plazo vencido: %s", label),
		"message":       fmt.Sprintf("%s — %s. Plazo de %s vencido el %s.", e.CompanyName, e.ResourceLabel, human(target), e.DueAt.In(bogota).Format("2006-01-02 15:04")),
		"link":          "/dashboard/inspections",
		"mail_template": "sla_breach",
		"variables": map[string]string{
			"event_label": label, "company_name": e.CompanyName, "target": human(target),
			"opened_at": e.OpenedAt.In(bogota).Format("2006-01-02 15:04"), "resource_id": e.ResourceLabel,
		},
	}
	body, _ := json.Marshal(payload)
	return m.pub.Publish(ctx, "zu.notifications", "notify.sla.breach", body)
}

type Summary struct {
	Open          int64    `json:"open"`
	Breached      int64    `json:"breached_open"`
	ClosedLast30  int64    `json:"closed_last_30d"`
	OnTimeLast30  int64    `json:"closed_on_time_last_30d"`
	CompliancePct *float64 `json:"compliance_pct_last_30d"` // nil when nothing closed yet: no data is not 100%
}

func (m *Monitor) scope(ctx context.Context, tenantID, companyID string) *gorm.DB {
	q := m.db.WithContext(ctx).Model(&Event{}).Where("tenant_id = ?", tenantID)
	if companyID != "" {
		q = q.Where("company_id = ?", companyID)
	}
	return q
}

func (m *Monitor) Summary(ctx context.Context, tenantID, companyID string) (Summary, error) {
	var s Summary
	since := m.now().AddDate(0, 0, -30)
	if err := m.scope(ctx, tenantID, companyID).Where("closed_at IS NULL").Count(&s.Open).Error; err != nil {
		return s, err
	}
	m.scope(ctx, tenantID, companyID).Where("closed_at IS NULL AND breached_at IS NOT NULL").Count(&s.Breached)
	m.scope(ctx, tenantID, companyID).Where("closed_at >= ?", since).Count(&s.ClosedLast30)
	m.scope(ctx, tenantID, companyID).Where("closed_at >= ? AND closed_at <= due_at", since).Count(&s.OnTimeLast30)
	if s.ClosedLast30 > 0 {
		pct := float64(s.OnTimeLast30) / float64(s.ClosedLast30) * 100
		s.CompliancePct = &pct
	}
	return s, nil
}

func (m *Monitor) List(ctx context.Context, tenantID, companyID, status string) ([]Event, error) {
	q := m.scope(ctx, tenantID, companyID)
	switch status {
	case "open":
		q = q.Where("closed_at IS NULL")
	case "breached":
		q = q.Where("breached_at IS NOT NULL")
	case "closed":
		q = q.Where("closed_at IS NOT NULL")
	}
	var out []Event
	return out, q.Order("due_at ASC").Limit(200).Find(&out).Error
}

var bogota = func() *time.Location {
	l, err := time.LoadLocation("America/Bogota")
	if err != nil {
		return time.FixedZone("COT", -5*3600)
	}
	return l
}()

func human(d time.Duration) string {
	if d%(24*time.Hour) == 0 {
		return fmt.Sprintf("%d días", int(d/(24*time.Hour)))
	}
	if d%time.Hour == 0 {
		return fmt.Sprintf("%d horas", int(d/time.Hour))
	}
	return fmt.Sprintf("%d minutos", int(d/time.Minute))
}
