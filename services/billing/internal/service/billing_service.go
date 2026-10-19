package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/zu-seguridad/billing/internal/model"
	"gorm.io/gorm"
	"log"
	"time"
)

// PlanInfo is the commercial definition of a plan: price and how many client
// companies a professional may manage under it. dron_addon is not a
// standalone plan (it carries no company quota of its own) — it is meant to
// be combined with a base plan; CreateSubscription rejects it alone.
type PlanInfo struct {
	PriceMonthly int64 // COP en centavos
	MaxCompanies int
}

var Plans = map[model.PlanCode]PlanInfo{
	model.PlanBasico:   {PriceMonthly: 100000000, MaxCompanies: 10}, // $1,000,000 COP
	model.PlanEstandar: {PriceMonthly: 150000000, MaxCompanies: 20}, // $1,500,000 COP
	model.PlanPremium:  {PriceMonthly: 215000000, MaxCompanies: 50}, // $2,150,000 COP
}

const IVARate = 0.19

var ErrCompanyQuotaExceeded = errors.New("adding this company would exceed the plan's company quota")
var ErrNoActiveSubscription = errors.New("tenant has no active subscription")

type BillingService struct{ db *gorm.DB }

func NewBillingService(db *gorm.DB) *BillingService { return &BillingService{db: db} }

// CreateSubscription provisions one subscription per tenant/professional. A
// tenant with an existing subscription must cancel/upgrade instead of
// creating a second one (TenantID is unique).
func (s *BillingService) CreateSubscription(ctx context.Context, tenantID uuid.UUID, planCode model.PlanCode, wompiToken string) (*model.Subscription, error) {
	plan, ok := Plans[planCode]
	if !ok {
		return nil, fmt.Errorf("plan %s no existe o no es un plan base con cupo de empresas", planCode)
	}

	now := time.Now()
	periodStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	periodEnd := periodStart.AddDate(0, 1, -1)

	sub := &model.Subscription{
		TenantID:           tenantID,
		PlanCode:           planCode,
		PriceMonthly:       plan.PriceMonthly,
		MaxCompanies:       plan.MaxCompanies,
		Status:             model.SubStatusActive,
		WompiTokenID:       wompiToken,
		BillingDay:         1,
		CurrentPeriodStart: periodStart,
		CurrentPeriodEnd:   periodEnd,
		GraceDays:          3,
	}

	if err := s.db.WithContext(ctx).Create(sub).Error; err != nil {
		return nil, err
	}
	log.Printf("✅ Suscripción creada: tenant %s, plan %s, %d empresas, $%.0f/mes", tenantID, planCode, plan.MaxCompanies, float64(plan.PriceMonthly)/100)
	return sub, nil
}

// CanAddCompany reports whether tenantID's active subscription has room for
// one more client company, counting only non-deleted companies. Core must
// call this before creating a company; it is the enforcement point for the
// "plan buys you N companies" model — nothing in Core's own company_service
// currently checks it.
func (s *BillingService) CanAddCompany(ctx context.Context, tenantID uuid.UUID) (allowed bool, currentCount, maxCompanies int, err error) {
	var sub model.Subscription
	if err := s.db.WithContext(ctx).Where("tenant_id = ? AND status = ?", tenantID, model.SubStatusActive).First(&sub).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, 0, 0, ErrNoActiveSubscription
		}
		return false, 0, 0, err
	}

	var count int64
	if err := s.db.WithContext(ctx).Table("companies").
		Where("tenant_id = ? AND deleted_at IS NULL", tenantID).
		Count(&count).Error; err != nil {
		return false, 0, sub.MaxCompanies, err
	}

	return int(count) < sub.MaxCompanies, int(count), sub.MaxCompanies, nil
}

func (s *BillingService) GenerateMonthlyInvoices(ctx context.Context) (int, error) {
	var subs []model.Subscription
	if err := s.db.WithContext(ctx).Where("status = 'active' AND billing_day = ?", time.Now().Day()).Find(&subs).Error; err != nil {
		return 0, err
	}

	count := 0
	for _, sub := range subs {
		if _, err := s.generateInvoice(ctx, sub); err != nil {
			log.Printf("Error generando factura para tenant %s: %v", sub.TenantID, err)
			continue
		}
		count++
	}
	return count, nil
}

func (s *BillingService) generateInvoice(ctx context.Context, sub model.Subscription) (*model.Invoice, error) {
	iva := int64(float64(sub.PriceMonthly) * IVARate)
	total := sub.PriceMonthly + iva

	invoice := &model.Invoice{
		TenantID:       sub.TenantID,
		SubscriptionID: sub.ID,
		DIANNumber:     s.generateDIANNumber(),
		Status:         model.InvoiceStatusPending,
		Subtotal:       sub.PriceMonthly,
		IVA:            iva,
		Total:          total,
		PeriodStart:    sub.CurrentPeriodStart,
		PeriodEnd:      sub.CurrentPeriodEnd,
		DueAt:          time.Now().AddDate(0, 0, 5),
	}

	if err := s.db.WithContext(ctx).Create(invoice).Error; err != nil {
		return nil, err
	}

	// Publicar para cobro automático Wompi + generación PDF DIAN
	log.Printf("📄 Factura %s generada: $%d COP", invoice.DIANNumber, invoice.Total/100)
	return invoice, nil
}

func (s *BillingService) ProcessPayment(ctx context.Context, invoiceID uuid.UUID, wompiTxID string) error {
	now := time.Now()
	return s.db.WithContext(ctx).Model(&model.Invoice{}).
		Where("id = ?", invoiceID).
		Updates(map[string]interface{}{
			"status":      model.InvoiceStatusPaid,
			"paid_at":     now,
			"wompi_tx_id": wompiTxID,
		}).Error
}

func (s *BillingService) SuspendOverdue(ctx context.Context) (int, error) {
	result := s.db.WithContext(ctx).
		Model(&model.Subscription{}).
		Where("status = 'active' AND id IN (SELECT subscription_id FROM invoices WHERE status = 'overdue' AND due_at < NOW() - INTERVAL '3 days')").
		Update("status", model.SubStatusSuspended)
	return int(result.RowsAffected), result.Error
}

func (s *BillingService) GetMRR(ctx context.Context, tenantID uuid.UUID) (map[string]interface{}, error) {
	var subs []model.Subscription
	s.db.WithContext(ctx).Where("tenant_id = ? AND status = 'active'", tenantID).Find(&subs)

	var mrr int64
	for _, sub := range subs {
		mrr += sub.PriceMonthly
	}

	return map[string]interface{}{
		"mrr_cop":     mrr / 100,
		"mrr_display": fmt.Sprintf("$%.1fM COP", float64(mrr)/100/1000000),
		"active_subs": len(subs),
		"arr_cop":     mrr * 12 / 100,
	}, nil
}

func (s *BillingService) generateDIANNumber() string {
	return fmt.Sprintf("FE-%d-%d", time.Now().Year(), time.Now().UnixNano()%1000000)
}

// ── Idempotencia (hallazgo de auditoría #2) ──────────────────

// WasProcessed verifica si una operación ya se ejecutó (evita doble cobro).
func (s *BillingService) WasProcessed(ctx context.Context, key string) bool {
	var count int64
	s.db.WithContext(ctx).Table("idempotency_keys").Where("key = ?", key).Count(&count)
	return count > 0
}

// MarkProcessed registra que una operación se completó.
func (s *BillingService) MarkProcessed(ctx context.Context, key string) error {
	return s.db.WithContext(ctx).Exec(
		"INSERT INTO idempotency_keys (key, processed_at) VALUES (?, NOW()) ON CONFLICT (key) DO NOTHING",
		key,
	).Error
}
