package repository

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/zu-seguridad/billing/internal/model"
	"gorm.io/gorm"
	"time"
)

type SubscriptionRepository struct{ db *gorm.DB }
type InvoiceRepository struct{ db *gorm.DB }

func NewSubscriptionRepository(db *gorm.DB) *SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}
func NewInvoiceRepository(db *gorm.DB) *InvoiceRepository { return &InvoiceRepository{db: db} }

func (r *SubscriptionRepository) FindDueToday(ctx context.Context) ([]model.Subscription, error) {
	var subs []model.Subscription
	today := time.Now().Truncate(24 * time.Hour)
	err := r.db.WithContext(ctx).Where("status = 'active' AND next_billing_date <= ?", today.Add(24*time.Hour)).Find(&subs).Error
	return subs, err
}

func (r *SubscriptionRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status model.SubscriptionStatus) error {
	return r.db.WithContext(ctx).Model(&model.Subscription{}).Where("id = ?", id).Update("status", status).Error
}

func (r *SubscriptionRepository) UpdateNextBilling(ctx context.Context, id uuid.UUID, next time.Time) error {
	return r.db.WithContext(ctx).Model(&model.Subscription{}).Where("id = ?", id).Update("next_billing_date", next).Error
}

func (r *InvoiceRepository) Create(ctx context.Context, inv *model.Invoice) error {
	return r.db.WithContext(ctx).Create(inv).Error
}

func (r *InvoiceRepository) FindByTenant(ctx context.Context, tenantID uuid.UUID) ([]model.Invoice, error) {
	var invoices []model.Invoice
	err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Order("created_at DESC").Find(&invoices).Error
	return invoices, err
}

type RabbitMQ struct{}

func (r *RabbitMQ) Publish(exchange, key string, body []byte) error {
	return errors.New("RabbitMQ publisher is not configured")
}
