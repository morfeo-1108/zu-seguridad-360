package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/zu-seguridad/billing/internal/model"
	"gorm.io/gorm"
	"log"
	"strings"
	"time"
)

// PlanInfo is the commercial definition of a plan: price in COP centavos
// and the quotas it sells. FounderPriceMonthly is the discounted price
// offered to the first professionals who join during the pilot (locked in
// for 12 months from signup — see CreateSubscription).
type PlanInfo struct {
	PriceMonthly         int64 // COP en centavos — precio normal
	FounderPriceMonthly  int64 // COP en centavos — precio fundador, primeros 15-20 profesionales, 12 meses
	MaxCompanies         int
	MaxProfessionalUsers int
	MaxWorkers           int
}

var Plans = map[model.PlanCode]PlanInfo{
	model.PlanEsencial:    {PriceMonthly: 12900000, FounderPriceMonthly: 9900000, MaxCompanies: 3, MaxProfessionalUsers: 1, MaxWorkers: 100},
	model.PlanProfesional: {PriceMonthly: 29900000, FounderPriceMonthly: 21900000, MaxCompanies: 10, MaxProfessionalUsers: 3, MaxWorkers: 350},
	model.PlanConsultora:  {PriceMonthly: 59900000, FounderPriceMonthly: 39900000, MaxCompanies: 25, MaxProfessionalUsers: 8, MaxWorkers: 1000},
	model.PlanAgencia:     {PriceMonthly: 99900000, FounderPriceMonthly: 69900000, MaxCompanies: 60, MaxProfessionalUsers: 20, MaxWorkers: 3000},
}

// FounderSlotsRemaining caps how many tenants may ever get IsFounder=true.
// Hardcoded, not configurable at runtime: raising this is a pricing
// decision, not an operational one.
const FounderSlotsTotal = 20

var ErrCompanyQuotaExceeded = errors.New("adding this company would exceed the plan's company quota")
var ErrNoActiveSubscription = errors.New("tenant has no active subscription")

// firstCompanyWithoutPlan is the company quota of a tenant that has not picked a plan yet.
const firstCompanyWithoutPlan = 1

type BillingService struct {
	db *gorm.DB
}

func NewBillingService(db *gorm.DB) *BillingService {
	return &BillingService{db: db}
}

// CreateSubscription provisions one subscription per tenant/professional. A
// tenant with an existing subscription must cancel/upgrade instead of
// creating a second one (TenantID is unique). wantsFounder is a request,
// not a guarantee — it's silently downgraded to the normal price once
// FounderSlotsTotal is reached, so the caller must read IsFounder back off
// the returned subscription rather than assuming it got what it asked for.
func (s *BillingService) CreateSubscription(ctx context.Context, tenantID uuid.UUID, planCode model.PlanCode, wompiToken string, wantsFounder bool) (*model.Subscription, error) {
	plan, ok := Plans[planCode]
	if !ok {
		return nil, fmt.Errorf("plan %s no existe o no es un plan base con cupo de empresas", planCode)
	}

	now := time.Now()
	periodStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	periodEnd := periodStart.AddDate(0, 1, -1)

	price := plan.PriceMonthly
	isFounder := false
	var founderUntil *time.Time
	if wantsFounder {
		var founderCount int64
		if err := s.db.WithContext(ctx).Model(&model.Subscription{}).Where("is_founder = ?", true).Count(&founderCount).Error; err != nil {
			return nil, err
		}
		if founderCount < FounderSlotsTotal {
			price = plan.FounderPriceMonthly
			isFounder = true
			until := now.AddDate(1, 0, 0)
			founderUntil = &until
		}
	}

	sub := &model.Subscription{
		TenantID: tenantID, PlanCode: planCode, PriceMonthly: price,
		MaxCompanies: plan.MaxCompanies, MaxProfessionalUsers: plan.MaxProfessionalUsers, MaxWorkers: plan.MaxWorkers,
		IsFounder: isFounder, FounderPriceUntil: founderUntil,
		Status: model.SubStatusActive, WompiTokenID: wompiToken, BillingDay: 1,
		CurrentPeriodStart: periodStart, CurrentPeriodEnd: periodEnd, GraceDays: 3,
	}

	if err := s.db.WithContext(ctx).Create(sub).Error; err != nil {
		return nil, err
	}
	log.Printf("✅ Suscripción creada: tenant %s, plan %s, founder=%v, %d empresas, $%.0f/mes", tenantID, planCode, isFounder, plan.MaxCompanies, float64(price)/100)
	return sub, nil
}

// CanAddCompany reports whether tenantID's active subscription has room for
// one more client company, counting only non-deleted companies. Core must
// call this before creating a company; it is the enforcement point for the
// "plan buys you N companies" model — nothing in Core's own company_service
// currently checks it.
func (s *BillingService) CanAddCompany(ctx context.Context, tenantID uuid.UUID) (allowed bool, currentCount, maxCompanies int, err error) {
	var count int64
	if err := s.db.WithContext(ctx).Table("companies").
		Where("tenant_id = ? AND deleted_at IS NULL", tenantID).
		Count(&count).Error; err != nil {
		return false, 0, 0, err
	}

	var sub model.Subscription
	if err := s.db.WithContext(ctx).Where("tenant_id = ? AND status = ?", tenantID, model.SubStatusActive).First(&sub).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// The first company never needed a plan (signup used to create it
			// directly); now that it is registered from the dashboard after
			// signup, keep that: one company without a plan, more need one.
			if count == 0 {
				return true, 0, firstCompanyWithoutPlan, nil
			}
			return false, int(count), 0, ErrNoActiveSubscription
		}
		return false, 0, 0, err
	}

	return int(count) < sub.MaxCompanies, int(count), sub.MaxCompanies, nil
}

// GetSubscription returns tenantID's active subscription, if any.
func (s *BillingService) GetSubscription(ctx context.Context, tenantID uuid.UUID) (*model.Subscription, error) {
	var sub model.Subscription
	if err := s.db.WithContext(ctx).Where("tenant_id = ? AND status = ?", tenantID, model.SubStatusActive).First(&sub).Error; err != nil {
		return nil, err
	}
	return &sub, nil
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
	if sub.IsFounder && sub.FounderPriceUntil != nil && time.Now().After(*sub.FounderPriceUntil) {
		// The 12-month founder window closed; nothing recalculates
		// price_monthly back to the plan's normal rate automatically today
		// — flagging it loudly so a human repriced this subscription
		// instead of silently keeping the founder discount forever.
		log.Printf("⚠️  tenant %s: precio fundador venció el %s pero la suscripción sigue facturando a precio fundador — requiere reprecio manual", sub.TenantID, sub.FounderPriceUntil.Format("2006-01-02"))
	}
	iva := ivaOf(sub.PriceMonthly)
	total := sub.PriceMonthly + iva

	invoice := &model.Invoice{
		TenantID:       sub.TenantID,
		SubscriptionID: sub.ID,
		// Internal reference; the DIAN consecutive is assigned when the
		// electronic invoice is issued after payment (Payments.IssueElectronic).
		DIANNumber:  "ZU-" + strings.ToUpper(uuid.NewString()[:8]),
		Status:      model.InvoiceStatusPending,
		Subtotal:    sub.PriceMonthly,
		IVA:         iva,
		Total:       total,
		PeriodStart: sub.CurrentPeriodStart,
		PeriodEnd:   sub.CurrentPeriodEnd,
		DueAt:       time.Now().AddDate(0, 0, 5),
	}

	if err := s.db.WithContext(ctx).Create(invoice).Error; err != nil {
		return nil, err
	}

	// Publicar para cobro automático Wompi + generación PDF DIAN
	log.Printf("📄 Factura %s generada: $%d COP", invoice.DIANNumber, invoice.Total/100)
	return invoice, nil
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

// ── Idempotencia (hallazgo de auditoría #2) ──────────────────

// WasProcessed verifica si una operación ya se ejecutó (evita doble cobro).
func (s *BillingService) WasProcessed(ctx context.Context, key string) bool {
	var count int64
	s.db.WithContext(ctx).Table("idempotency_keys").Where("key = ?", key).Count(&count)
	return count > 0
}

// MarkProcessed registra que una operación se completó.
func (s *BillingService) MarkProcessed(ctx context.Context, key string) error {
	// time.Now() in Go rather than NOW() in SQL — NOW() is Postgres-specific
	// and this package's tests run against sqlite.
	return s.db.WithContext(ctx).Exec(
		"INSERT INTO idempotency_keys (key, processed_at) VALUES (?, ?) ON CONFLICT (key) DO NOTHING",
		key, time.Now().UTC(),
	).Error
}
