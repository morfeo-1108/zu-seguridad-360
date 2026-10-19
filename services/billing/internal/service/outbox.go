package service

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/zu-seguridad/billing/internal/model"
	"gorm.io/gorm"
)

// OutboxRelay is not started by the service entrypoint until publisher confirms,
// transactional enqueue coverage and downstream idempotency are tested.
type OutboxRelay struct {
	db *gorm.DB
	ch *amqp.Channel
}

func NewOutboxRelay(db *gorm.DB, ch *amqp.Channel) *OutboxRelay {
	return &OutboxRelay{db: db, ch: ch}
}

func (r *OutboxRelay) Enqueue(tx *gorm.DB, tenantID uuid.UUID, exchange, routingKey string, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	var object model.JSONB
	if err := json.Unmarshal(data, &object); err != nil || object == nil {
		return errInvalidOutboxPayload
	}
	evt := &model.OutboxEvent{
		TenantID: tenantID, Exchange: exchange, RoutingKey: routingKey,
		Payload: object, Status: "pending",
	}
	return tx.Create(evt).Error
}

func (r *OutboxRelay) Run(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.publishPending(ctx)
		}
	}
}

func (r *OutboxRelay) publishPending(ctx context.Context) {
	if r.ch == nil {
		return
	}
	var events []model.OutboxEvent
	if err := r.db.WithContext(ctx).
		Where("status = ? AND attempts < ?", "pending", 5).
		Order("created_at ASC").Limit(100).Find(&events).Error; err != nil {
		log.Printf("outbox query failed: %v", err)
		return
	}

	for _, evt := range events {
		body, err := json.Marshal(evt.Payload)
		if err == nil {
			err = r.ch.Publish(evt.Exchange, evt.RoutingKey, false, false, amqp.Publishing{
				ContentType: "application/json", Body: body,
				DeliveryMode: amqp.Persistent, MessageId: evt.ID.String(),
			})
		}
		now := time.Now()
		if err != nil {
			attempts := evt.Attempts + 1
			status := "pending"
			if attempts >= 5 {
				status = "failed"
			}
			r.db.WithContext(ctx).Model(&evt).Updates(map[string]interface{}{
				"attempts": attempts, "status": status,
			})
			log.Printf("outbox publish failed (attempt %d): %v", attempts, err)
			continue
		}
		r.db.WithContext(ctx).Model(&evt).Updates(map[string]interface{}{
			"status": "published", "published_at": &now,
		})
	}
}

var errInvalidOutboxPayload = &outboxPayloadError{}

type outboxPayloadError struct{}

func (*outboxPayloadError) Error() string { return "outbox payload must be a JSON object" }
