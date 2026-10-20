package model

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type SubscriptionStatus string
type InvoiceStatus string
type PlanCode string

const (
	SubStatusActive    SubscriptionStatus = "active"
	SubStatusSuspended SubscriptionStatus = "suspended"
	SubStatusCancelled SubscriptionStatus = "cancelled"
	SubStatusPastDue   SubscriptionStatus = "past_due"

	InvoiceStatusPending InvoiceStatus = "pending"
	InvoiceStatusPaid    InvoiceStatus = "paid"
	InvoiceStatusOverdue InvoiceStatus = "overdue"
	InvoiceStatusVoided  InvoiceStatus = "voided"

	PlanEsencial    PlanCode = "esencial"
	PlanProfesional PlanCode = "profesional"
	PlanConsultora  PlanCode = "consultora"
	PlanAgencia     PlanCode = "agencia"
)

// Subscription is one per professional/tenant, not per company: the plan
// sells a quota of companies, professional users, and workers managed —
// see service.Plans for the current commercial definition of each tier.
// MaxCompanies is actively enforced (BillingService.CanAddCompany, called
// by Core before creating a company). MaxProfessionalUsers and MaxWorkers
// are stored and shown to the customer but NOT enforced yet: Core has no
// "invite a second professional user under the same tenant" flow, and no
// per-tenant worker-count cap — both would be new Core features, not just
// a billing check.
type Subscription struct {
	ID                   uuid.UUID          `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantID             uuid.UUID          `gorm:"type:uuid;not null;uniqueIndex" json:"tenant_id"`
	PlanCode             PlanCode           `gorm:"size:30;not null" json:"plan_code"`
	PriceMonthly         int64              `gorm:"not null" json:"price_monthly"` // COP en centavos; precio vigente (normal o fundador) al momento de facturar
	MaxCompanies         int                `gorm:"not null" json:"max_companies"`
	MaxProfessionalUsers int                `gorm:"not null;default:1" json:"max_professional_users"`
	MaxWorkers           int                `gorm:"not null;default:0" json:"max_workers"`
	IsFounder            bool               `gorm:"default:false" json:"is_founder"`
	FounderPriceUntil    *time.Time         `json:"founder_price_until"` // tras esta fecha, la próxima factura debe recalcularse al precio normal del plan
	Status               SubscriptionStatus `gorm:"size:20;default:'active'" json:"status"`
	WompiTokenID         string             `gorm:"size:255" json:"wompi_token_id"` // tokenización tarjeta
	BillingDay           int                `gorm:"default:1" json:"billing_day"`
	TrialEndsAt          *time.Time         `json:"trial_ends_at"`
	CurrentPeriodStart   time.Time          `json:"current_period_start"`
	CurrentPeriodEnd     time.Time          `json:"current_period_end"`
	CancelledAt          *time.Time         `json:"cancelled_at"`
	SuspendedAt          *time.Time         `json:"suspended_at"`
	GraceDays            int                `gorm:"default:3" json:"grace_days"`
	CreatedAt            time.Time          `json:"created_at"`
	UpdatedAt            time.Time          `json:"updated_at"`
}

// Invoice bills the professional/tenant (ZU360's actual customer), not an
// individual client company in their portfolio.
type Invoice struct {
	ID             uuid.UUID              `gorm:"type:uuid;primaryKey" json:"id"`
	TenantID       uuid.UUID              `gorm:"type:uuid;not null;index" json:"tenant_id"`
	SubscriptionID uuid.UUID              `gorm:"type:uuid;not null" json:"subscription_id"`
	DIANNumber     string                 `gorm:"uniqueIndex;size:50" json:"dian_number"`
	Status         InvoiceStatus          `gorm:"size:20;default:'pending'" json:"status"`
	Subtotal       int64                  `json:"subtotal"` // COP centavos
	IVA            int64                  `json:"iva"`      // 19%
	Total          int64                  `json:"total"`
	PeriodStart    time.Time              `json:"period_start"`
	PeriodEnd      time.Time              `json:"period_end"`
	DueAt          time.Time              `gorm:"index" json:"due_at"`
	PaidAt         *time.Time             `json:"paid_at"`
	WompiTxID      string                 `gorm:"size:100" json:"wompi_tx_id"`
	PDFURL         string                 `gorm:"size:500" json:"pdf_url"`
	SentAt         *time.Time             `json:"sent_at"`
	Metadata       map[string]interface{} `gorm:"serializer:json" json:"metadata"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

func (Subscription) TableName() string { return "subscriptions" }
func (Invoice) TableName() string      { return "invoices" }

// BeforeCreate generates the primary key in Go rather than relying on a
// Postgres-only gen_random_uuid() column default — portable across dialects
// (and required for the sqlite-backed integration tests in this package).
func (i *Invoice) BeforeCreate(tx *gorm.DB) error {
	if i.ID == uuid.Nil {
		i.ID = uuid.New()
	}
	return nil
}
