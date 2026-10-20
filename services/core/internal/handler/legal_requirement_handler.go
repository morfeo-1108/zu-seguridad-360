package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/zu-seguridad/core/internal/service"
)

type LegalRequirementHandler struct{ svc *service.LegalRequirementService }

func NewLegalRequirementHandler(svc *service.LegalRequirementService) *LegalRequirementHandler {
	return &LegalRequirementHandler{svc: svc}
}

func (h *LegalRequirementHandler) Create(c *fiber.Ctx) error {
	tenantID := mustUUID(c.Locals("tenant_id"))
	companyID := mustUUID(c.Locals("company_id"))
	var input service.CreateLegalRequirementInput
	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "cuerpo inválido")
	}
	req, err := h.svc.Create(c.Context(), input, companyID, tenantID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "no fue posible registrar el requisito legal")
	}
	return c.Status(fiber.StatusCreated).JSON(req)
}

func (h *LegalRequirementHandler) List(c *fiber.Ctx) error {
	tenantID := mustUUID(c.Locals("tenant_id"))
	companyID := mustUUID(c.Locals("company_id"))
	requirements, err := h.svc.List(c.Context(), companyID, tenantID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "no fue posible consultar la matriz legal")
	}
	return c.JSON(fiber.Map{"data": requirements, "total": len(requirements)})
}

func (h *LegalRequirementHandler) UpdateCompliance(c *fiber.Ctx) error {
	tenantID := mustUUID(c.Locals("tenant_id"))
	companyID := mustUUID(c.Locals("company_id"))
	userID := mustUUID(c.Locals("user_id"))
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "id inválido")
	}
	var input service.UpdateComplianceInput
	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "cuerpo inválido")
	}
	req, err := h.svc.UpdateCompliance(c.Context(), id, companyID, tenantID, userID, input)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "no fue posible actualizar el cumplimiento")
	}
	return c.JSON(req)
}
