package handler

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/zu-seguridad/billing/internal/service"
)

type BillingHandler struct{ svc *service.BillingService }

func NewBillingHandler(svc *service.BillingService) *BillingHandler {
	return &BillingHandler{svc: svc}
}

// CanAddCompany is called by Core before creating a non-synthetic company. It
// never creates or mutates anything; Core enforces the decision.
func (h *BillingHandler) CanAddCompany(c *fiber.Ctx) error {
	tenantID, err := uuid.Parse(c.Params("tenantId"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenantId inválido")
	}
	allowed, current, max, err := h.svc.CanAddCompany(c.Context(), tenantID)
	if err != nil {
		if errors.Is(err, service.ErrNoActiveSubscription) {
			return c.Status(fiber.StatusOK).JSON(fiber.Map{"allowed": false, "reason": "no_active_subscription"})
		}
		return fiber.NewError(fiber.StatusInternalServerError, "no fue posible verificar el cupo del plan")
	}
	return c.JSON(fiber.Map{
		"allowed":        allowed,
		"current_count":  current,
		"max_companies":  max,
	})
}
