package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BillingProfile is the tenant's fiscal identity (adquirente) for electronic
// invoices. Without it no electronic invoice can be issued.
type BillingProfile struct {
	TenantID  uuid.UUID `gorm:"type:uuid;primaryKey" json:"tenant_id"`
	IDType    string    `gorm:"size:4;not null" json:"id_type"` // 13 cédula, 31 NIT
	IDNumber  string    `gorm:"size:20;not null" json:"id_number"`
	DV        string    `gorm:"size:1" json:"dv"`
	LegalName string    `gorm:"size:255;not null" json:"legal_name"`
	Email     string    `gorm:"size:255;not null" json:"email"`
	Address   string    `gorm:"size:255" json:"address"`
	City      string    `gorm:"size:100" json:"city"`
	IsPerson  bool      `gorm:"not null;default:true" json:"is_person"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (BillingProfile) TableName() string { return "billing_profiles" }

// PaymentAttempt is one Wompi checkout for an invoice. Wompi references are
// unique per attempt, so a declined payment can be retried with a new one.
type PaymentAttempt struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	InvoiceID     uuid.UUID  `gorm:"type:uuid;not null;index" json:"invoice_id"`
	TenantID      uuid.UUID  `gorm:"type:uuid;not null;index" json:"tenant_id"`
	Reference     string     `gorm:"size:64;not null;uniqueIndex" json:"reference"`
	AmountInCents int64      `gorm:"not null" json:"amount_in_cents"`
	Environment   string     `gorm:"size:16;not null" json:"environment"` // sandbox | production
	Status        string     `gorm:"size:16;not null;default:'CREATED'" json:"status"`
	WompiTxID     string     `gorm:"size:100;index" json:"wompi_tx_id"`
	StatusMessage string     `gorm:"size:255" json:"status_message"`
	CheckedAt     *time.Time `json:"checked_at"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

func (PaymentAttempt) TableName() string { return "payment_attempts" }

func (p *PaymentAttempt) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

// ElectronicInvoice is the DIAN document for a paid invoice.
type ElectronicInvoice struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	InvoiceID     uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex" json:"invoice_id"`
	TenantID      uuid.UUID  `gorm:"type:uuid;not null;index" json:"tenant_id"`
	Number        string     `gorm:"size:30;uniqueIndex" json:"number"`
	Environment   string     `gorm:"size:1;not null" json:"environment"` // 1 producción, 2 pruebas
	CUFE          string     `gorm:"size:96" json:"cufe"`
	CUFESource    string     `gorm:"type:text" json:"-"`
	QR            string     `gorm:"type:text" json:"qr"`
	XML           []byte     `json:"-"`
	Status        string     `gorm:"size:32;not null" json:"status"` // pendiente_transmision | sin_perfil_fiscal | ...
	StatusDetail  string     `gorm:"type:text" json:"status_detail"`
	IssuedAt      time.Time  `json:"issued_at"`
	TransmittedAt *time.Time `json:"transmitted_at"`
	CreatedAt     time.Time  `json:"created_at"`
}

func (ElectronicInvoice) TableName() string { return "electronic_invoices" }

func (e *ElectronicInvoice) BeforeCreate(tx *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	return nil
}

// DianCounter holds the next consecutive per resolution prefix. Numbers are
// taken inside the issuing transaction with a row lock: no gaps from races,
// no reuse.
type DianCounter struct {
	Prefix string `gorm:"size:10;primaryKey"`
	Next   int64  `gorm:"not null"`
}

func (DianCounter) TableName() string { return "dian_counters" }
