package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/zu-seguridad/core/internal/model"
	"gorm.io/gorm"
)

type EPPService struct{ db *gorm.DB }

func NewEPPService(db *gorm.DB) *EPPService { return &EPPService{db: db} }

type CreateEPPItemInput struct {
	Name         string `json:"name" validate:"required"`
	Category     string `json:"category" validate:"required"`
	LifespanDays int    `json:"lifespan_days" validate:"required,min=1"`
}

func (s *EPPService) CreateItem(ctx context.Context, input CreateEPPItemInput, companyID, tenantID uuid.UUID) (*model.EPPItem, error) {
	item := &model.EPPItem{TenantID: tenantID, CompanyID: companyID, Name: input.Name, Category: input.Category, LifespanDays: input.LifespanDays}
	return item, s.db.WithContext(ctx).Create(item).Error
}

func (s *EPPService) ListItems(ctx context.Context, companyID, tenantID uuid.UUID) ([]model.EPPItem, error) {
	var items []model.EPPItem
	err := s.db.WithContext(ctx).Where("company_id = ? AND tenant_id = ?", companyID, tenantID).Order("name ASC").Find(&items).Error
	return items, err
}

type CreateEPPDeliveryInput struct {
	EmployeeID string `json:"employee_id" validate:"required"`
	ItemID     string `json:"item_id" validate:"required"`
	Quantity   int    `json:"quantity"`
	Condition  string `json:"condition"`
}

// CreateDelivery requires the employee and the item to both belong to the
// caller's active company — same ownership check pattern used by evidence
// and medical exams earlier this session.
func (s *EPPService) CreateDelivery(ctx context.Context, input CreateEPPDeliveryInput, companyID, tenantID, recordedBy uuid.UUID) (*model.EPPDelivery, error) {
	employeeID, err := uuid.Parse(input.EmployeeID)
	if err != nil {
		return nil, fmt.Errorf("employee_id inválido")
	}
	itemID, err := uuid.Parse(input.ItemID)
	if err != nil {
		return nil, fmt.Errorf("item_id inválido")
	}
	quantity := input.Quantity
	if quantity < 1 {
		quantity = 1
	}
	condition := input.Condition
	if condition == "" {
		condition = "nuevo"
	}

	var item model.EPPItem
	now := time.Now().UTC()
	delivery := &model.EPPDelivery{
		TenantID: tenantID, CompanyID: companyID, EmployeeID: employeeID, ItemID: itemID,
		Quantity: quantity, Condition: condition, DeliveredAt: now,
		ConfirmedBy: &recordedBy, ConfirmedAt: &now,
	}

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Table("employees").Where("id = ? AND tenant_id = ? AND company_id = ?", employeeID, tenantID, companyID).Count(&count).Error; err != nil {
			return err
		}
		if count != 1 {
			return gorm.ErrRecordNotFound
		}
		if err := tx.Where("id = ? AND tenant_id = ? AND company_id = ?", itemID, tenantID, companyID).First(&item).Error; err != nil {
			return err
		}
		if item.LifespanDays > 0 {
			expires := now.AddDate(0, 0, item.LifespanDays)
			delivery.ExpiresAt = &expires
		}
		return tx.Create(delivery).Error
	})
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("el trabajador o el elemento no pertenecen a la empresa activa")
		}
		return nil, err
	}
	delivery.Item = item
	return delivery, nil
}

func (s *EPPService) ListByEmployee(ctx context.Context, employeeID, companyID, tenantID uuid.UUID) ([]model.EPPDelivery, error) {
	var deliveries []model.EPPDelivery
	err := s.db.WithContext(ctx).Preload("Item").
		Where("employee_id = ? AND company_id = ? AND tenant_id = ?", employeeID, companyID, tenantID).
		Order("delivered_at DESC").Find(&deliveries).Error
	return deliveries, err
}

func (s *EPPService) ListExpiring(ctx context.Context, companyID, tenantID uuid.UUID, days int) ([]model.EPPDelivery, error) {
	cutoff := time.Now().UTC().AddDate(0, 0, days)
	var deliveries []model.EPPDelivery
	err := s.db.WithContext(ctx).Preload("Item").
		Where("company_id = ? AND tenant_id = ? AND returned_at IS NULL AND expires_at IS NOT NULL AND expires_at <= ?", companyID, tenantID, cutoff).
		Order("expires_at ASC").Find(&deliveries).Error
	return deliveries, err
}

func (s *EPPService) Return(ctx context.Context, deliveryID, companyID, tenantID uuid.UUID, reason string) (*model.EPPDelivery, error) {
	now := time.Now().UTC()
	result := s.db.WithContext(ctx).Model(&model.EPPDelivery{}).
		Where("id = ? AND company_id = ? AND tenant_id = ?", deliveryID, companyID, tenantID).
		Updates(map[string]interface{}{"returned_at": now, "return_reason": reason})
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected != 1 {
		return nil, gorm.ErrRecordNotFound
	}
	var delivery model.EPPDelivery
	if err := s.db.WithContext(ctx).Preload("Item").Where("id = ? AND company_id = ? AND tenant_id = ?", deliveryID, companyID, tenantID).First(&delivery).Error; err != nil {
		return nil, err
	}
	return &delivery, nil
}
