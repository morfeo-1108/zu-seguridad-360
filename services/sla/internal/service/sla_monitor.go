package service

import (
	"context"
	"fmt"
	"gorm.io/gorm"
	"log"
	"time"
)

// SLA targets por tipo de evento (en minutos)
var SLATargets = map[string]int{
	"incident.reported":     60,   // 1 hora para apertura de investigación
	"incident.grave":        30,   // 30 min para incidentes graves
	"incident.fatal":        15,   // 15 min para fatales
	"exam.pending_review":   480,  // 8 horas para revisión de aptitud
	"document.pending_sign": 1440, // 24 horas para firma
	"drone.alert.epp":       30,   // 30 min para validar alerta de dron
	"cert.expiring.7d":      4320, // 3 días para iniciar renovación
}

type SLAEvent struct {
	ID          string     `json:"id"`
	TenantID    string     `json:"tenant_id"`
	CompanyID   string     `json:"company_id"`
	EventType   string     `json:"event_type"`
	ResourceID  string     `json:"resource_id"`
	OpenedAt    time.Time  `json:"opened_at"`
	ClosedAt    *time.Time `json:"closed_at"`
	TargetMins  int        `json:"target_minutes"`
	ElapsedMins int        `json:"elapsed_minutes"`
	Breached    bool       `json:"breached"`
	BreachAt    time.Time  `json:"breach_at"`
}

type SLAMonitor struct{ db *gorm.DB }

func NewSLAMonitor(db *gorm.DB) *SLAMonitor { return &SLAMonitor{db: db} }

func (m *SLAMonitor) OpenEvent(ctx context.Context, tenantID, companyID, resourceID, eventType string) (*SLAEvent, error) {
	target, ok := SLATargets[eventType]
	if !ok {
		target = 1440 // default 24h
	}

	now := time.Now()
	event := &SLAEvent{
		TenantID:   tenantID,
		CompanyID:  companyID,
		EventType:  eventType,
		ResourceID: resourceID,
		OpenedAt:   now,
		TargetMins: target,
		BreachAt:   now.Add(time.Duration(target) * time.Minute),
	}

	if err := m.db.WithContext(ctx).Table("sla_events").Create(event).Error; err != nil {
		return nil, err
	}

	log.Printf("⏱️  SLA abierto: %s | Target: %d min | Breach: %s", eventType, target, event.BreachAt.Format("15:04"))
	return event, nil
}

func (m *SLAMonitor) CloseEvent(ctx context.Context, resourceID, eventType string) error {
	now := time.Now()
	return m.db.WithContext(ctx).Table("sla_events").
		Where("resource_id = ? AND event_type = ? AND closed_at IS NULL", resourceID, eventType).
		Updates(map[string]interface{}{"closed_at": now}).Error
}

func (m *SLAMonitor) CheckBreaches(ctx context.Context) ([]SLAEvent, error) {
	var breached []SLAEvent
	err := m.db.WithContext(ctx).Table("sla_events").
		Where("closed_at IS NULL AND breach_at < NOW() AND breached = false").
		Find(&breached).Error
	if err != nil {
		return nil, err
	}

	// Marcar como incumplidos y alertar
	for _, event := range breached {
		m.db.WithContext(ctx).Table("sla_events").
			Where("id = ?", event.ID).
			Updates(map[string]interface{}{"breached": true})

		log.Printf("🚨 SLA BREACH: %s | Empresa: %s | Recurso: %s", event.EventType, event.CompanyID, event.ResourceID)
	}

	return breached, nil
}

func (m *SLAMonitor) GetDashboard(ctx context.Context, tenantID string) (map[string]interface{}, error) {
	var open, breached int64
	m.db.WithContext(ctx).Table("sla_events").Where("tenant_id = ? AND closed_at IS NULL", tenantID).Count(&open)
	m.db.WithContext(ctx).Table("sla_events").Where("tenant_id = ? AND breached = true AND closed_at IS NULL", tenantID).Count(&breached)

	compliance := 100.0
	if open > 0 {
		compliance = float64(open-breached) / float64(open) * 100
	}

	return map[string]interface{}{
		"open_events":     open,
		"breached_events": breached,
		"compliance_pct":  fmt.Sprintf("%.1f%%", compliance),
		"sla_targets":     SLATargets,
	}, nil
}
