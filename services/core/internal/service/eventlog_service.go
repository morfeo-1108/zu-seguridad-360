package service

import (
	"context"
	"github.com/google/uuid"
	"github.com/zu-seguridad/core/internal/model"
	"gorm.io/gorm"
	"sync"
)

// EventLogService: bitácora append-only de agentes + bus de eventos en vivo.
//
// Dos responsabilidades:
//  1. REGISTRO: persistir cada evento de agente (auditable, consultable).
//  2. EN VIVO: publicar a suscriptores (SSE) para el feed del dashboard sin polling.
type EventLogService struct {
	db          *gorm.DB
	mu          sync.RWMutex
	subscribers map[string]map[chan model.AgentEvent]struct{} // tenantID -> set de canales
}

func NewEventLogService(db *gorm.DB) *EventLogService {
	return &EventLogService{
		db:          db,
		subscribers: make(map[string]map[chan model.AgentEvent]struct{}),
	}
}

// Append registra un evento y lo publica en vivo a los suscriptores del tenant.
func (s *EventLogService) Append(ctx context.Context, ev *model.AgentEvent) error {
	if err := s.db.WithContext(ctx).Create(ev).Error; err != nil {
		return err
	}
	s.publish(ev)
	return nil
}

// Recent devuelve los últimos N eventos (registro), filtrable por proyecto.
func (s *EventLogService) Recent(ctx context.Context, tenantID uuid.UUID, projectID *uuid.UUID, limit int) ([]model.AgentEvent, error) {
	var events []model.AgentEvent
	q := s.db.WithContext(ctx).Where("tenant_id = ?", tenantID)
	if projectID != nil {
		q = q.Where("project_id = ?", *projectID)
	}
	if limit <= 0 {
		limit = 50
	}
	err := q.Order("created_at DESC").Limit(limit).Find(&events).Error
	return events, err
}

// EmployeeEvents: eventos de agentes relacionados a un trabajador específico.
func (s *EventLogService) EmployeeEvents(ctx context.Context, employeeID uuid.UUID, limit int) ([]model.AgentEvent, error) {
	var events []model.AgentEvent
	if limit <= 0 {
		limit = 30
	}
	err := s.db.WithContext(ctx).
		Where("employee_id = ?", employeeID).
		Order("created_at DESC").Limit(limit).Find(&events).Error
	return events, err
}

// ── Pub/Sub en vivo (SSE) ─────────────────────────────────────
func (s *EventLogService) Subscribe(tenantID string) chan model.AgentEvent {
	s.mu.Lock()
	defer s.mu.Unlock()
	ch := make(chan model.AgentEvent, 32)
	if s.subscribers[tenantID] == nil {
		s.subscribers[tenantID] = make(map[chan model.AgentEvent]struct{})
	}
	s.subscribers[tenantID][ch] = struct{}{}
	return ch
}

func (s *EventLogService) Unsubscribe(tenantID string, ch chan model.AgentEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if subs, ok := s.subscribers[tenantID]; ok {
		delete(subs, ch)
		close(ch)
	}
}

func (s *EventLogService) publish(ev *model.AgentEvent) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	tenantID := ev.TenantID.String()
	for ch := range s.subscribers[tenantID] {
		select {
		case ch <- *ev:
		default: // suscriptor lento: no bloquear el bus
		}
	}
}
