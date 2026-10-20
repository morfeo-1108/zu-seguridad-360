package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/zu-seguridad/core/internal/service"
)

type EPPHandler struct{ svc *service.EPPService }

func NewEPPHandler(svc *service.EPPService) *EPPHandler {
	return &EPPHandler{svc: svc}
}

func (h *EPPHandler) CreateItem(c *fiber.Ctx) error {
	tenantID := mustUUID(c.Locals("tenant_id"))
	companyID := mustUUID(c.Locals("company_id"))
	var input service.CreateEPPItemInput
	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "cuerpo inválido")
	}
	item, err := h.svc.CreateItem(c.Context(), input, companyID, tenantID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "no fue posible crear el elemento")
	}
	return c.Status(fiber.StatusCreated).JSON(item)
}

func (h *EPPHandler) ListItems(c *fiber.Ctx) error {
	tenantID := mustUUID(c.Locals("tenant_id"))
	companyID := mustUUID(c.Locals("company_id"))
	items, err := h.svc.ListItems(c.Context(), companyID, tenantID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "no fue posible consultar el catálogo")
	}
	return c.JSON(fiber.Map{"data": items, "total": len(items)})
}

func (h *EPPHandler) CreateDelivery(c *fiber.Ctx) error {
	tenantID := mustUUID(c.Locals("tenant_id"))
	companyID := mustUUID(c.Locals("company_id"))
	userID := mustUUID(c.Locals("user_id"))
	var input service.CreateEPPDeliveryInput
	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "cuerpo inválido")
	}
	delivery, err := h.svc.CreateDelivery(c.Context(), input, companyID, tenantID, userID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.Status(fiber.StatusCreated).JSON(delivery)
}

func (h *EPPHandler) ListByEmployee(c *fiber.Ctx) error {
	tenantID := mustUUID(c.Locals("tenant_id"))
	companyID := mustUUID(c.Locals("company_id"))
	employeeID, err := uuid.Parse(c.Params("employeeId"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "employeeId inválido")
	}
	deliveries, err := h.svc.ListByEmployee(c.Context(), employeeID, companyID, tenantID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "no fue posible consultar las entregas")
	}
	return c.JSON(fiber.Map{"data": deliveries, "total": len(deliveries)})
}

func (h *EPPHandler) ListExpiring(c *fiber.Ctx) error {
	tenantID := mustUUID(c.Locals("tenant_id"))
	companyID := mustUUID(c.Locals("company_id"))
	deliveries, err := h.svc.ListExpiring(c.Context(), companyID, tenantID, 30)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "no fue posible consultar vencimientos")
	}
	return c.JSON(fiber.Map{"data": deliveries, "total": len(deliveries)})
}

func (h *EPPHandler) Return(c *fiber.Ctx) error {
	tenantID := mustUUID(c.Locals("tenant_id"))
	companyID := mustUUID(c.Locals("company_id"))
	deliveryID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "id inválido")
	}
	var body struct {
		Reason string `json:"reason"`
	}
	_ = c.BodyParser(&body)
	delivery, err := h.svc.Return(c.Context(), deliveryID, companyID, tenantID, body.Reason)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "no fue posible registrar la devolución")
	}
	return c.JSON(delivery)
}
