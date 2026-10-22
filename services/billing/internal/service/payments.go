package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/zu-seguridad/billing/internal/dian"
	"github.com/zu-seguridad/billing/internal/model"
	"github.com/zu-seguridad/billing/internal/wompi"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrInvoiceNotFound   = errors.New("invoice not found")
	ErrInvoiceNotPayable = errors.New("invoice is not pending")
	ErrAmountMismatch    = errors.New("wompi transaction does not match the invoice amount/currency")
)

// Exchange and keys of billing's own domain events (published via outbox).
const (
	BillingExchange = "zu.billing"
	KeyInvoicePaid  = "invoice.paid"
	NotifyExchange  = "zu.notifications"
	KeyNotifyPaid   = "notify.billing.payment"
)

// ivaOf computes 19% IVA in integer centavos, rounding half up — never float.
func ivaOf(subtotal int64) int64 { return (subtotal*19 + 50) / 100 }

type DianConfig struct {
	Resolution dian.Resolution
	Issuer     dian.Issuer
}

// Configured reports whether the issuer identity needed for a DIAN document exists.
func (c DianConfig) Configured() bool {
	return c.Issuer.NIT != "" && c.Resolution.Prefix != "" && c.Resolution.TechnicalKey != ""
}

type Payments struct {
	db     *gorm.DB
	wompi  *wompi.Client
	outbox *OutboxRelay
	dian   DianConfig
	appURL string
	now    func() time.Time
}

func NewPayments(db *gorm.DB, w *wompi.Client, outbox *OutboxRelay, d DianConfig, appURL string) *Payments {
	return &Payments{db: db, wompi: w, outbox: outbox, dian: d, appURL: strings.TrimRight(appURL, "/"), now: time.Now}
}

func (p *Payments) Migrate() error {
	return p.db.AutoMigrate(&model.BillingProfile{}, &model.PaymentAttempt{}, &model.ElectronicInvoice{}, &model.DianCounter{})
}

// IssueCurrentInvoice creates the invoice for the subscription's current
// period, or returns the existing one (one invoice per subscription/period).
func (p *Payments) IssueCurrentInvoice(ctx context.Context, tenantID uuid.UUID) (*model.Invoice, error) {
	var sub model.Subscription
	if err := p.db.WithContext(ctx).Where("tenant_id = ? AND status = ?", tenantID, model.SubStatusActive).First(&sub).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNoActiveSubscription
		}
		return nil, err
	}
	var existing model.Invoice
	err := p.db.WithContext(ctx).Where("subscription_id = ? AND period_start = ?", sub.ID, sub.CurrentPeriodStart).First(&existing).Error
	if err == nil {
		return &existing, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	iva := ivaOf(sub.PriceMonthly)
	inv := &model.Invoice{
		TenantID: tenantID, SubscriptionID: sub.ID, Status: model.InvoiceStatusPending,
		// Internal reference until a DIAN number is assigned on payment.
		DIANNumber: "ZU-" + strings.ToUpper(uuid.NewString()[:8]),
		Subtotal:   sub.PriceMonthly, IVA: iva, Total: sub.PriceMonthly + iva,
		PeriodStart: sub.CurrentPeriodStart, PeriodEnd: sub.CurrentPeriodEnd, DueAt: p.now().AddDate(0, 0, 5),
		Metadata: map[string]interface{}{"plan_code": string(sub.PlanCode), "is_founder": sub.IsFounder},
	}
	return inv, p.db.WithContext(ctx).Create(inv).Error
}

func (p *Payments) ListInvoices(ctx context.Context, tenantID uuid.UUID) ([]model.Invoice, error) {
	var out []model.Invoice
	return out, p.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Order("created_at DESC").Limit(50).Find(&out).Error
}

func (p *Payments) invoice(ctx context.Context, tenantID, invoiceID uuid.UUID) (*model.Invoice, error) {
	var inv model.Invoice
	err := p.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", invoiceID, tenantID).First(&inv).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrInvoiceNotFound
	}
	return &inv, err
}

// StartCheckout opens a new Wompi Web Checkout attempt for a pending invoice.
func (p *Payments) StartCheckout(ctx context.Context, tenantID, invoiceID uuid.UUID, email string) (*model.PaymentAttempt, string, error) {
	inv, err := p.invoice(ctx, tenantID, invoiceID)
	if err != nil {
		return nil, "", err
	}
	if inv.Status != model.InvoiceStatusPending && inv.Status != model.InvoiceStatusOverdue {
		return nil, "", ErrInvoiceNotPayable
	}
	if !p.wompi.Configured() {
		return nil, "", wompi.ErrNotConfigured
	}
	att := &model.PaymentAttempt{
		InvoiceID: inv.ID, TenantID: tenantID, AmountInCents: inv.Total, Environment: p.wompi.Environment(),
		Reference: fmt.Sprintf("ZU-%s-%d", strings.ToUpper(inv.ID.String()[:8]), p.now().UnixNano()%1_000_000_000),
	}
	if err := p.db.WithContext(ctx).Create(att).Error; err != nil {
		return nil, "", err
	}
	link, err := p.wompi.CheckoutLink(wompi.CheckoutInput{
		Reference: att.Reference, AmountInCents: att.AmountInCents, CustomerEmail: email,
		RedirectURL: p.appURL + "/dashboard/billing?ref=" + att.Reference,
	})
	return att, link, err
}

// HandleEvent processes a Wompi webhook body: verify the signature, then
// reconcile against Wompi's API (the body alone is never trusted).
func (p *Payments) HandleEvent(ctx context.Context, raw []byte, eventsSecret string) error {
	ev, err := wompi.VerifyEvent(raw, eventsSecret)
	if err != nil {
		return err
	}
	if ev.Event != "transaction.updated" {
		return nil
	}
	tx, err := wompi.TransactionFromEvent(ev)
	if err != nil {
		return err
	}
	_, err = p.Reconcile(ctx, tx.ID)
	return err
}

// Reconcile reads a transaction from Wompi and applies it to its attempt and
// invoice. Idempotent: replays and late duplicates change nothing.
func (p *Payments) Reconcile(ctx context.Context, wompiTxID string) (*model.PaymentAttempt, error) {
	tx, err := p.wompi.GetTransaction(ctx, wompiTxID)
	if err != nil {
		return nil, err
	}
	return p.apply(ctx, tx)
}

func (p *Payments) apply(ctx context.Context, tx *wompi.Transaction) (*model.PaymentAttempt, error) {
	var att model.PaymentAttempt
	err := p.db.WithContext(ctx).Transaction(func(db *gorm.DB) error {
		if err := db.Clauses(clause.Locking{Strength: "UPDATE"}).Where("reference = ?", tx.Reference).First(&att).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("unknown payment reference %q", tx.Reference)
			}
			return err
		}
		if tx.AmountInCents != att.AmountInCents || tx.Currency != "COP" {
			now := p.now()
			db.Model(&att).Updates(map[string]interface{}{"status": "MISMATCH", "wompi_tx_id": tx.ID, "checked_at": now,
				"status_message": fmt.Sprintf("Wompi %d %s ≠ esperado %d COP", tx.AmountInCents, tx.Currency, att.AmountInCents)})
			return ErrAmountMismatch
		}
		now := p.now()
		if err := db.Model(&att).Updates(map[string]interface{}{
			"status": tx.Status, "wompi_tx_id": tx.ID, "status_message": tx.StatusMessage, "checked_at": now,
		}).Error; err != nil {
			return err
		}
		if tx.Status != "APPROVED" {
			return nil
		}
		// Only the first approval of a pending invoice marks it paid.
		res := db.Model(&model.Invoice{}).
			Where("id = ? AND status IN ?", att.InvoiceID, []model.InvoiceStatus{model.InvoiceStatusPending, model.InvoiceStatusOverdue}).
			Updates(map[string]interface{}{"status": model.InvoiceStatusPaid, "paid_at": now, "wompi_tx_id": tx.ID})
		if res.Error != nil || res.RowsAffected == 0 {
			return res.Error
		}
		var inv model.Invoice
		if err := db.Where("id = ?", att.InvoiceID).First(&inv).Error; err != nil {
			return err
		}
		evt := map[string]interface{}{"invoice_id": inv.ID.String(), "tenant_id": inv.TenantID.String(), "total": inv.Total,
			"wompi_tx_id": tx.ID, "reference": att.Reference, "environment": att.Environment, "paid_at": now}
		if err := p.outbox.Enqueue(db, inv.TenantID, BillingExchange, KeyInvoicePaid, evt); err != nil {
			return err
		}
		return p.outbox.Enqueue(db, inv.TenantID, NotifyExchange, KeyNotifyPaid, map[string]interface{}{
			"kind": "billing.payment", "tenant_id": inv.TenantID.String(), "channels": []string{"inapp"}, "priority": "normal",
			"title":   "Pago recibido",
			"message": fmt.Sprintf("Recibimos su pago de $%s COP (referencia %s).", thousands(inv.Total/100), att.Reference),
			"link":    "/dashboard/billing",
		})
	})
	return &att, err
}

// ReconcilePending re-checks attempts Wompi has not resolved yet (lost or
// delayed webhooks). Needs the private key to search by reference.
func (p *Payments) ReconcilePending(ctx context.Context) (int, error) {
	var atts []model.PaymentAttempt
	cutoff := p.now().Add(-2 * time.Minute)
	if err := p.db.WithContext(ctx).Where("status IN ? AND created_at < ? AND created_at > ?",
		[]string{"CREATED", "PENDING"}, cutoff, p.now().Add(-72*time.Hour)).Limit(50).Find(&atts).Error; err != nil {
		return 0, err
	}
	n := 0
	for _, a := range atts {
		var txs []wompi.Transaction
		if a.WompiTxID != "" {
			tx, err := p.wompi.GetTransaction(ctx, a.WompiTxID)
			if err != nil {
				continue
			}
			txs = []wompi.Transaction{*tx}
		} else {
			found, err := p.wompi.FindByReference(ctx, a.Reference)
			if err != nil {
				if errors.Is(err, wompi.ErrNotConfigured) {
					return n, nil
				}
				continue
			}
			txs = found
		}
		for i := range txs {
			if _, err := p.apply(ctx, &txs[i]); err == nil {
				n++
			}
		}
	}
	return n, nil
}

// --- Billing profile -------------------------------------------------------

func (p *Payments) GetProfile(ctx context.Context, tenantID uuid.UUID) (*model.BillingProfile, error) {
	var bp model.BillingProfile
	if err := p.db.WithContext(ctx).Where("tenant_id = ?", tenantID).First(&bp).Error; err != nil {
		return nil, err
	}
	return &bp, nil
}

func (p *Payments) SaveProfile(ctx context.Context, bp *model.BillingProfile) error {
	bp.IDNumber = strings.TrimSpace(bp.IDNumber)
	if (bp.IDType != "13" && bp.IDType != "31") || bp.IDNumber == "" || strings.TrimSpace(bp.LegalName) == "" || !strings.Contains(bp.Email, "@") {
		return fmt.Errorf("perfil fiscal incompleto: tipo (13 cédula / 31 NIT), número, razón social y correo son obligatorios")
	}
	for _, r := range bp.IDNumber {
		if r < '0' || r > '9' {
			return fmt.Errorf("el número de identificación solo admite dígitos (sin DV)")
		}
	}
	if bp.IDType == "31" {
		bp.DV = NITCheckDigit(bp.IDNumber)
		bp.IsPerson = false
	} else {
		bp.DV = ""
		bp.IsPerson = true
	}
	return p.db.WithContext(ctx).Save(bp).Error
}

// NITCheckDigit computes the DIAN verification digit (módulo 11).
func NITCheckDigit(nit string) string {
	weights := []int{3, 7, 13, 17, 19, 23, 29, 37, 41, 43, 47, 53, 59, 67, 71}
	sum := 0
	for i := 0; i < len(nit) && i < len(weights); i++ {
		sum += int(nit[len(nit)-1-i]-'0') * weights[i]
	}
	r := sum % 11
	if r > 1 {
		r = 11 - r
	}
	return fmt.Sprint(r)
}

// --- Electronic invoice ----------------------------------------------------

// IssueElectronic builds the DIAN document of a paid invoice. Idempotent per
// invoice. It assigns the next consecutive of the resolution, computes the
// CUFE and stores the unsigned UBL; status stays "pendiente_transmision"
// until signing and transmission exist (see package dian).
func (p *Payments) IssueElectronic(ctx context.Context, invoiceID uuid.UUID) (*model.ElectronicInvoice, error) {
	var out model.ElectronicInvoice
	err := p.db.WithContext(ctx).Transaction(func(db *gorm.DB) error {
		if err := db.Where("invoice_id = ?", invoiceID).First(&out).Error; err == nil {
			if out.Status != "sin_perfil_fiscal" && out.Status != "sin_configuracion_emisor" {
				return nil
			}
			// Blocked on missing data: retry now that it may exist.
			if err := db.Delete(&out).Error; err != nil {
				return err
			}
			out = model.ElectronicInvoice{}
		}
		var inv model.Invoice
		if err := db.Where("id = ? AND status = ?", invoiceID, model.InvoiceStatusPaid).First(&inv).Error; err != nil {
			return fmt.Errorf("paid invoice %s not found: %w", invoiceID, err)
		}
		out = model.ElectronicInvoice{InvoiceID: inv.ID, TenantID: inv.TenantID, Environment: p.dian.Issuer.Environment, IssuedAt: p.now()}
		var bp model.BillingProfile
		if err := db.Where("tenant_id = ?", inv.TenantID).First(&bp).Error; err != nil {
			out.Status, out.StatusDetail = "sin_perfil_fiscal", "El cliente no ha registrado su perfil fiscal (NIT/cédula, razón social)."
			return db.Create(&out).Error
		}
		if !p.dian.Configured() {
			out.Status, out.StatusDetail = "sin_configuracion_emisor", "Falta NIT del emisor, resolución de facturación o clave técnica."
			return db.Create(&out).Error
		}
		r := p.dian.Resolution
		var ctr model.DianCounter
		if err := db.Clauses(clause.Locking{Strength: "UPDATE"}).Where("prefix = ?", r.Prefix).First(&ctr).Error; err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			ctr = model.DianCounter{Prefix: r.Prefix, Next: r.From}
			if err := db.Create(&ctr).Error; err != nil {
				return err
			}
		}
		if ctr.Next > r.To {
			return fmt.Errorf("rango de numeración %s agotado (%d–%d)", r.Prefix, r.From, r.To)
		}
		number := fmt.Sprintf("%s%d", r.Prefix, ctr.Next)
		if err := db.Model(&ctr).Where("prefix = ?", r.Prefix).Update("next", ctr.Next+1).Error; err != nil {
			return err
		}
		plan, _ := inv.Metadata["plan_code"].(string)
		res, err := dian.Build(dian.Invoice{
			Number: number, IssuedAt: out.IssuedAt.In(bogota), DueAt: inv.DueAt.In(bogota), Resolution: r, Issuer: p.dian.Issuer,
			Buyer: dian.Buyer{IDType: bp.IDType, ID: bp.IDNumber, DV: bp.DV, Name: bp.LegalName, Email: bp.Email, Address: bp.Address, City: bp.City, IsPerson: bp.IsPerson},
			Lines: []dian.Line{{Description: fmt.Sprintf("Suscripción ZU Seguridad 360 — plan %s — %s", plan, inv.PeriodStart.Format("2006-01")),
				Quantity: 1, UnitPrice: inv.Subtotal, Subtotal: inv.Subtotal, IVA: inv.IVA}},
			Subtotal: inv.Subtotal, IVA: inv.IVA, Total: inv.Total,
			Note: "Pago Wompi " + inv.WompiTxID,
		})
		if err != nil {
			return err
		}
		out.Number, out.CUFE, out.CUFESource, out.QR, out.XML = number, res.CUFE, res.Source, res.QR, res.XML
		out.Status = "pendiente_transmision"
		out.StatusDetail = "XML UBL 2.1 y CUFE generados. Falta firma digital (certificado) y envío a la DIAN o al proveedor tecnológico."
		if out.Environment != "1" {
			out.StatusDetail += " Ambiente de PRUEBAS: no es una factura con validez fiscal."
		}
		if err := db.Create(&out).Error; err != nil {
			return err
		}
		return db.Model(&model.Invoice{}).Where("id = ?", inv.ID).Update("dian_number", number).Error
	})
	if err != nil {
		return nil, err
	}
	log.Printf("factura electrónica %s para invoice %s: %s", out.Number, invoiceID, out.Status)
	return &out, nil
}

func (p *Payments) ElectronicFor(ctx context.Context, tenantID, invoiceID uuid.UUID) (*model.ElectronicInvoice, error) {
	var e model.ElectronicInvoice
	err := p.db.WithContext(ctx).Where("invoice_id = ? AND tenant_id = ?", invoiceID, tenantID).First(&e).Error
	return &e, err
}

var bogota = func() *time.Location {
	l, err := time.LoadLocation("America/Bogota")
	if err != nil {
		return time.FixedZone("COT", -5*3600)
	}
	return l
}()

func thousands(n int64) string {
	s := fmt.Sprint(n)
	var b strings.Builder
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			b.WriteByte('.')
		}
		b.WriteRune(c)
	}
	return b.String()
}
