package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/zu-seguridad/core/internal/service"
	"gorm.io/gorm"
)

type DocumentHandler struct{ svc *service.DocumentService }

func NewDocumentHandler(svc *service.DocumentService) *DocumentHandler {
	return &DocumentHandler{svc: svc}
}

func (h *DocumentHandler) List(c *fiber.Ctx) error {
	companyID := mustUUID(c.Locals("company_id"))
	tenantID := mustUUID(c.Locals("tenant_id"))
	docs, err := h.svc.List(c.Context(), companyID, tenantID, c.Query("category"))
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(fiber.Map{"data": docs, "total": len(docs)})
}

func (h *DocumentHandler) Create(c *fiber.Ctx) error {
	companyID := mustUUID(c.Locals("company_id"))
	tenantID := mustUUID(c.Locals("tenant_id"))
	userID := mustUUID(c.Locals("user_id"))
	var input service.CreateDocumentInput
	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "cuerpo inválido")
	}
	doc, err := h.svc.Create(c.Context(), input, companyID, tenantID, userID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.Status(fiber.StatusCreated).JSON(doc)
}

func (h *DocumentHandler) Delete(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "ID inválido")
	}
	companyID := mustUUID(c.Locals("company_id"))
	tenantID := mustUUID(c.Locals("tenant_id"))
	if err := h.svc.Delete(c.Context(), id, companyID, tenantID); err != nil {
		if err == gorm.ErrRecordNotFound {
			return fiber.NewError(fiber.StatusNotFound, "documento no encontrado")
		}
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.SendStatus(fiber.StatusNoContent)
}
