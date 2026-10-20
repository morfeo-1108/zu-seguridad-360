package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/zu-seguridad/core/internal/service"
)

type InspectionHandler struct{ svc *service.InspectionService }

func NewInspectionHandler(svc *service.InspectionService) *InspectionHandler {
	return &InspectionHandler{svc: svc}
}

func (h *InspectionHandler) Create(c *fiber.Ctx) error {
	tenantID := mustUUID(c.Locals("tenant_id"))
	companyID := mustUUID(c.Locals("company_id"))
	userID := mustUUID(c.Locals("user_id"))
	var input service.CreateInspectionInput
	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "cuerpo inválido")
	}
	inspection, err := h.svc.Create(c.Context(), input, companyID, tenantID, userID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.Status(fiber.StatusCreated).JSON(inspection)
}

func (h *InspectionHandler) List(c *fiber.Ctx) error {
	tenantID := mustUUID(c.Locals("tenant_id"))
	companyID := mustUUID(c.Locals("company_id"))
	inspections, err := h.svc.List(c.Context(), companyID, tenantID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "no fue posible consultar inspecciones")
	}
	return c.JSON(fiber.Map{"data": inspections, "total": len(inspections)})
}

func (h *InspectionHandler) GetByID(c *fiber.Ctx) error {
	tenantID := mustUUID(c.Locals("tenant_id"))
	companyID := mustUUID(c.Locals("company_id"))
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "id inválido")
	}
	inspection, err := h.svc.GetByID(c.Context(), id, companyID, tenantID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "inspección no encontrada")
	}
	return c.JSON(inspection)
}

func (h *InspectionHandler) Approve(c *fiber.Ctx) error {
	tenantID := mustUUID(c.Locals("tenant_id"))
	companyID := mustUUID(c.Locals("company_id"))
	userID := mustUUID(c.Locals("user_id"))
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "id inválido")
	}
	inspection, err := h.svc.Approve(c.Context(), id, companyID, tenantID, userID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "no fue posible aprobar la inspección")
	}
	return c.JSON(inspection)
}
