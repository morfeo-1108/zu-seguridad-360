package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/zu-seguridad/core/internal/service"
)

type AlertHandler struct{ svc *service.AlertService }

func NewAlertHandler(svc *service.AlertService) *AlertHandler {
	return &AlertHandler{svc: svc}
}

func (h *AlertHandler) List(c *fiber.Ctx) error {
	tenantID := mustUUID(c.Locals("tenant_id"))
	companyID := mustUUID(c.Locals("company_id"))
	alerts, err := h.svc.GetAll(c.Context(), companyID, tenantID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "no fue posible consultar alertas")
	}
	return c.JSON(fiber.Map{"data": alerts, "total": len(alerts)})
}
