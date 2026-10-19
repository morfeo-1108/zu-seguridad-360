package model

import "time"

type EventType string

const (
	EventIncident    EventType = "incident"
	EventDocument    EventType = "document"
	EventMedical     EventType = "medical"
	EventDrone       EventType = "drone"
	EventCertificate EventType = "certificate"
)

// SLA targets en minutos
var SLATargets = map[EventType]int{
	EventIncident:    60,   // 1 hora
	EventDocument:    1440, // 24 horas
	EventMedical:     2880, // 48 horas
	EventDrone:       30,   // 30 minutos
	EventCertificate: 720,  // 12 horas
}

type SLAEvent struct {
	ID             string     `json:"id"`
	TenantID       string     `json:"tenant_id"`
	CompanyID      string     `json:"company_id"`
	Type           EventType  `json:"type"`
	CreatedAt      time.Time  `json:"created_at"`
	ResolvedAt     *time.Time `json:"resolved_at"`
	TargetMinutes  int        `json:"target_minutes"`
	ElapsedMinutes float64    `json:"elapsed_minutes"`
	IsBreached     bool       `json:"is_breached"`
	Status         string     `json:"status"`
}

func (e *SLAEvent) Calculate() {
	target := SLATargets[e.Type]
	e.TargetMinutes = target
	end := time.Now()
	if e.ResolvedAt != nil {
		end = *e.ResolvedAt
	}
	e.ElapsedMinutes = end.Sub(e.CreatedAt).Minutes()
	e.IsBreached = e.ElapsedMinutes > float64(target)
	if e.IsBreached {
		e.Status = "breached"
	} else {
		e.Status = "ok"
	}
}
