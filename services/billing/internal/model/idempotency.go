package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
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
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
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

// BeforeCreate generates the primary key in Go rather than relying on a
// Postgres-only gen_random_uuid() column default (AutoMigrate with that
// default in the DDL fails outright on sqlite, which the tests in this
// package use).
func (e *OutboxEvent) BeforeCreate(tx *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	return nil
}

// JSONB tipo auxiliar para columnas jsonb de PostgreSQL.
type JSONB map[string]interface{}

// Scan implements sql.Scanner so GORM can populate JSONB columns read back from Postgres.
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("JSONB: unsupported Scan source, expected []byte")
	}
	if len(bytes) == 0 {
		*j = nil
		return nil
	}
	return json.Unmarshal(bytes, j)
}

// Value implements driver.Valuer so GORM can persist JSONB columns to Postgres.
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}
