package service

import (
	"context"
	"sort"
	"time"

	"github.com/zu-seguridad/attention/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// AttentionItem es un item priorizado en el inbox maestro
type AttentionItem struct {
	ID          string     `json:"id"`
	TenantID    string     `json:"tenant_id"`
	CompanyID   string     `json:"company_id"`
	CompanyName string     `json:"company_name"`
	Type        string     `json:"type"` // incident | medical | document | drone | cert | copasst
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Urgency     string     `json:"urgency"`  // critical | high | medium | low
	Priority    int        `json:"priority"` // score calculado
	ActionURL   string     `json:"action_url"`
	CreatedAt   time.Time  `json:"created_at"`
	SLABreachAt *time.Time `json:"sla_breach_at"`
}

// PriorityWeights para calcular score de prioridad
var PriorityWeights = map[string]int{
	"incident.fatal":         1000,
	"incident.grave":         900,
	"incident.open":          700,
	"medical.pending_review": 600,
	"drone.alert.epp":        500,
	"document.pending_sign":  400,
	"cert.expiring.0d":       800,
	"cert.expiring.7d":       300,
	"exam.expiring":          250,
	"copasst.pending":        200,
}

type InboxService struct{ db *gorm.DB }

func NewInboxService(db *gorm.DB) *InboxService { return &InboxService{db: db} }

func (s *InboxService) GetPrioritizedInbox(ctx context.Context, tenantID string) ([]AttentionItem, error) {
	var items []AttentionItem

	// 1. Incidentes abiertos sin investigar
	s.db.WithContext(ctx).Raw(`
		SELECT
			i.id, i.tenant_id, i.company_id, c.name AS company_name,
			'incident' AS type,
			CASE i.severity
				WHEN 'fatal'   THEN 'critical'
				WHEN 'grave'   THEN 'critical'
				WHEN 'moderado' THEN 'high'
				ELSE 'medium'
			END AS urgency,
			CONCAT('Incidente ', i.severity, ' sin investigar') AS title,
			SUBSTRING(i.description, 1, 120) AS description,
			CONCAT('/incidents/', i.id) AS action_url,
			i.created_at
		FROM incidents i
		JOIN companies c ON c.id = i.company_id
		WHERE i.tenant_id = ? AND i.status = 'open' AND i.investigated_at IS NULL
		  AND NOT EXISTS (SELECT 1 FROM attention_acks a WHERE a.item_id = i.id::text AND a.tenant_id = i.tenant_id::text)
		ORDER BY i.occurred_at DESC
		LIMIT 10
	`, tenantID).Scan(&items)

	// 2. Exámenes médicos pendientes de revisión
	var medItems []AttentionItem
	s.db.WithContext(ctx).Raw(`
		SELECT
			me.id, me.tenant_id, me.company_id, c.name AS company_name,
			'medical' AS type, 'high' AS urgency,
			CONCAT('Aptitud médica pendiente — ', e.first_name, ' ', e.last_name) AS title,
			CONCAT('Examen ', me.type, ' en ', me.ip_s_name) AS description,
			CONCAT('/medical-exams/', me.id, '/review') AS action_url,
			me.created_at
		FROM medical_exams me
		JOIN companies c ON c.id = me.company_id
		JOIN employees e ON e.id = me.employee_id
		WHERE me.tenant_id = ? AND me.aptitude_result = 'pendiente_revision'
		  AND NOT EXISTS (SELECT 1 FROM attention_acks a WHERE a.item_id = me.id::text AND a.tenant_id = me.tenant_id::text)
		ORDER BY me.created_at ASC
		LIMIT 10
	`, tenantID).Scan(&medItems)
	items = append(items, medItems...)

	// 3. Documentos pendientes de firma
	var docItems []AttentionItem
	s.db.WithContext(ctx).Raw(`
		SELECT
			d.id, d.tenant_id, d.company_id, c.name AS company_name,
			'document' AS type, 'medium' AS urgency,
			CONCAT(d.title, ' — pendiente de firma') AS title,
			CONCAT('Categoría: ', d.category) AS description,
			CONCAT('/documents/', d.id, '/sign') AS action_url,
			d.created_at
		FROM documents d
		JOIN companies c ON c.id = d.company_id
		WHERE d.tenant_id = ? AND d.status = 'signing'
		  AND NOT EXISTS (SELECT 1 FROM attention_acks a WHERE a.item_id = d.id::text AND a.tenant_id = d.tenant_id::text)
		ORDER BY d.created_at ASC
		LIMIT 10
	`, tenantID).Scan(&docItems)
	items = append(items, docItems...)

	// Calcular priority score y ordenar
	for i := range items {
		items[i].Priority = s.calculatePriority(items[i])
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].Priority > items[j].Priority
	})

	return items, nil
}

func (s *InboxService) Acknowledge(ctx context.Context, itemID, tenantID string) error {
	ack := model.AttentionAck{ItemID: itemID, TenantID: tenantID, AcknowledgedAt: time.Now().UTC()}
	return s.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "item_id"}, {Name: "tenant_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"acknowledged_at"}),
	}).Create(&ack).Error
}

func (s *InboxService) GetBriefing(ctx context.Context, tenantID string) (map[string]interface{}, error) {
	items, err := s.GetPrioritizedInbox(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	critical := 0
	for _, item := range items {
		if item.Urgency == "critical" {
			critical++
		}
	}

	return map[string]interface{}{
		"date":           time.Now().Format("Monday 2 January 2006"),
		"total_pending":  len(items),
		"critical_count": critical,
		"top_items":      items[:min(5, len(items))],
		"generated_at":   time.Now().Format("15:04"),
	}, nil
}

func (s *InboxService) calculatePriority(item AttentionItem) int {
	base := map[string]int{"critical": 1000, "high": 600, "medium": 300, "low": 100}[item.Urgency]
	// Penalizar por antigüedad (items más viejos, más urgentes)
	hoursSince := int(time.Since(item.CreatedAt).Hours())
	return base + hoursSince*5
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
