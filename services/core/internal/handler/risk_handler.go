package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/zu-seguridad/core/internal/service"
)

type RiskHandler struct{ svc *service.RiskService }

func NewRiskHandler(svc *service.RiskService) *RiskHandler {
	return &RiskHandler{svc: svc}
}

func (h *RiskHandler) Create(c *fiber.Ctx) error {
	tenantID := mustUUID(c.Locals("tenant_id"))
	companyID := mustUUID(c.Locals("company_id"))
	var input service.CreateRiskInput
	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "cuerpo inválido")
	}
	entry, err := h.svc.CreateEntry(c.Context(), input, companyID, tenantID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "no fue posible registrar el riesgo")
	}
	return c.Status(fiber.StatusCreated).JSON(entry)
}

func (h *RiskHandler) List(c *fiber.Ctx) error {
	tenantID := mustUUID(c.Locals("tenant_id"))
	companyID := mustUUID(c.Locals("company_id"))
	entries, err := h.svc.GetMatrix(c.Context(), companyID, tenantID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "no fue posible consultar la matriz")
	}
	return c.JSON(fiber.Map{"data": entries, "total": len(entries)})
}

func (h *RiskHandler) NeedingReview(c *fiber.Ctx) error {
	tenantID := mustUUID(c.Locals("tenant_id"))
	companyID := mustUUID(c.Locals("company_id"))
	entries, err := h.svc.ListNeedingReview(c.Context(), companyID, tenantID, 30)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "no fue posible consultar revisiones pendientes")
	}
	return c.JSON(fiber.Map{"data": entries, "total": len(entries)})
}

func (h *RiskHandler) HeatMap(c *fiber.Ctx) error {
	tenantID := mustUUID(c.Locals("tenant_id"))
	companyID := mustUUID(c.Locals("company_id"))
	heatmap, err := h.svc.GetHeatMap(c.Context(), companyID, tenantID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "no fue posible calcular el mapa de calor")
	}
	return c.JSON(heatmap)
}
