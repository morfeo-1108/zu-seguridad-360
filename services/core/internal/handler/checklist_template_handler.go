package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/zu-seguridad/core/internal/service"
	"gorm.io/gorm"
)

type ChecklistTemplateHandler struct{ svc *service.ChecklistTemplateService }

func NewChecklistTemplateHandler(svc *service.ChecklistTemplateService) *ChecklistTemplateHandler {
	return &ChecklistTemplateHandler{svc: svc}
}

func (h *ChecklistTemplateHandler) List(c *fiber.Ctx) error {
	companyID := mustUUID(c.Locals("company_id"))
	tenantID := mustUUID(c.Locals("tenant_id"))
	templates, err := h.svc.List(c.Context(), companyID, tenantID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(fiber.Map{"data": templates, "total": len(templates)})
}

func (h *ChecklistTemplateHandler) GetByID(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "ID inválido")
	}
	companyID := mustUUID(c.Locals("company_id"))
	tenantID := mustUUID(c.Locals("tenant_id"))
	template, err := h.svc.GetByID(c.Context(), id, companyID, tenantID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "plantilla no encontrada")
	}
	return c.JSON(template)
}

func (h *ChecklistTemplateHandler) Create(c *fiber.Ctx) error {
	companyID := mustUUID(c.Locals("company_id"))
	tenantID := mustUUID(c.Locals("tenant_id"))
	userID := mustUUID(c.Locals("user_id"))
	var input service.CreateChecklistTemplateInput
	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "cuerpo inválido")
	}
	template, err := h.svc.Create(c.Context(), input, companyID, tenantID, userID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.Status(fiber.StatusCreated).JSON(template)
}

func (h *ChecklistTemplateHandler) Delete(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "ID inválido")
	}
	companyID := mustUUID(c.Locals("company_id"))
	tenantID := mustUUID(c.Locals("tenant_id"))
	if err := h.svc.Delete(c.Context(), id, companyID, tenantID); err != nil {
		if err == gorm.ErrRecordNotFound {
			return fiber.NewError(fiber.StatusNotFound, "plantilla no encontrada")
		}
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.SendStatus(fiber.StatusNoContent)
}
