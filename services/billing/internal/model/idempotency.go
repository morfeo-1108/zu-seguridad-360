package model

import (
	"time"

	"github.com/google/uuid"
)

// IdempotencyKey evita procesar dos veces la misma operación de cobro.
type IdempotencyKey struct {
	Key         string    `gorm:"primaryKey;size:255" json:"key"`
	ProcessedAt time.Time `json:"processed_at"`
	Result      JSONB     `gorm:"type:jsonb" json:"result"`
}

func (IdempotencyKey) TableName() string { return "idempotency_keys" }

// OutboxEvent is the persistence contract for a future authenticated publisher.
type OutboxEvent struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"tenant_id"`
	Exchange    string     `gorm:"size:120" json:"exchange"`
	RoutingKey  string     `gorm:"size:120" json:"routing_key"`
	Payload     JSONB      `gorm:"type:jsonb;not null" json:"payload"`
	Status      string     `gorm:"size:20;not null;default:'pending';index" json:"status"`
	Attempts    int        `gorm:"not null;default:0" json:"attempts"`
	PublishedAt *time.Time `json:"published_at"`
	CreatedAt   time.Time  `json:"created_at"`
}

func (OutboxEvent) TableName() string { return "outbox_events" }

// JSONB tipo auxiliar para columnas jsonb de PostgreSQL.
type JSONB map[string]interface{}
