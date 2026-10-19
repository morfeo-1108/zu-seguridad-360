package model

import (
	"github.com/google/uuid"
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

	PlanBasico   PlanCode = "basico"
	PlanEstandar PlanCode = "estandar"
	PlanPremium  PlanCode = "premium"
	PlanDronAdd  PlanCode = "dron_addon"
)

type Subscription struct {
	ID                 uuid.UUID          `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantID           uuid.UUID          `gorm:"type:uuid;not null;index" json:"tenant_id"`
	CompanyID          uuid.UUID          `gorm:"type:uuid;not null;uniqueIndex" json:"company_id"`
	PlanCode           PlanCode           `gorm:"size:30;not null" json:"plan_code"`
	PriceMonthly       int64              `gorm:"not null" json:"price_monthly"` // COP en centavos
	Status             SubscriptionStatus `gorm:"size:20;default:'active'" json:"status"`
	WompiTokenID       string             `gorm:"size:255" json:"wompi_token_id"` // tokenización tarjeta
	BillingDay         int                `gorm:"default:1" json:"billing_day"`
	TrialEndsAt        *time.Time         `json:"trial_ends_at"`
	CurrentPeriodStart time.Time          `json:"current_period_start"`
	CurrentPeriodEnd   time.Time          `json:"current_period_end"`
	CancelledAt        *time.Time         `json:"cancelled_at"`
	SuspendedAt        *time.Time         `json:"suspended_at"`
	GraceDays          int                `gorm:"default:3" json:"grace_days"`
	CreatedAt          time.Time          `json:"created_at"`
	UpdatedAt          time.Time          `json:"updated_at"`
}

type Invoice struct {
	ID             uuid.UUID              `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantID       uuid.UUID              `gorm:"type:uuid;not null;index" json:"tenant_id"`
	CompanyID      uuid.UUID              `gorm:"type:uuid;not null;index" json:"company_id"`
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
