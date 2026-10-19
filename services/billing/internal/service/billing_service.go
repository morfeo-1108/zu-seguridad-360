package service

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/zu-seguridad/billing/internal/model"
	"gorm.io/gorm"
	"log"
	"time"
)

// Precios en COP (en centavos)
var PlanPrices = map[model.PlanCode]int64{
	model.PlanBasico:   100000000, // $1,000,000 COP
	model.PlanEstandar: 150000000, // $1,500,000 COP
	model.PlanPremium:  215000000, // $2,150,000 COP
	model.PlanDronAdd:  65000000,  // $650,000 COP
}

const IVARate = 0.19

type BillingService struct{ db *gorm.DB }

func NewBillingService(db *gorm.DB) *BillingService { return &BillingService{db: db} }

func (s *BillingService) CreateSubscription(ctx context.Context, companyID, tenantID uuid.UUID, planCode model.PlanCode, wompiToken string) (*model.Subscription, error) {
	price, ok := PlanPrices[planCode]
	if !ok {
		return nil, fmt.Errorf("plan %s no existe", planCode)
	}

	now := time.Now()
	periodStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	periodEnd := periodStart.AddDate(0, 1, -1)

	sub := &model.Subscription{
		TenantID:           tenantID,
		CompanyID:          companyID,
		PlanCode:           planCode,
		PriceMonthly:       price,
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
	log.Printf("✅ Suscripción creada: empresa %s, plan %s, $%.0f/mes", companyID, planCode, float64(price)/100)
	return sub, nil
}

func (s *BillingService) GenerateMonthlyInvoices(ctx context.Context) (int, error) {
	var subs []model.Subscription
	if err := s.db.WithContext(ctx).Where("status = 'active' AND billing_day = ?", time.Now().Day()).Find(&subs).Error; err != nil {
		return 0, err
	}

	count := 0
	for _, sub := range subs {
		if _, err := s.generateInvoice(ctx, sub); err != nil {
			log.Printf("Error generando factura para empresa %s: %v", sub.CompanyID, err)
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
		CompanyID:      sub.CompanyID,
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
