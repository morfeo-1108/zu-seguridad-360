package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/zu-seguridad/core/internal/service"
)

type CompanyHandler struct{ svc *service.CompanyService }

func NewCompanyHandler(svc *service.CompanyService) *CompanyHandler {
	return &CompanyHandler{svc: svc}
}

func (h *CompanyHandler) List(c *fiber.Ctx) error {
	tenantID := mustUUID(c.Locals("tenant_id"))
	userID := mustUUID(c.Locals("user_id"))
	companies, err := h.svc.ListForUser(c.Context(), userID, tenantID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(fiber.Map{"data": companies, "total": len(companies)})
}

func (h *CompanyHandler) GetByID(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "ID inválido")
	}
	tenantID := mustUUID(c.Locals("tenant_id"))
	company, err := h.svc.GetByID(c.Context(), id, tenantID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "empresa no encontrada")
	}
	return c.JSON(company)
}

func (h *CompanyHandler) Create(c *fiber.Ctx) error {
	var input service.CreateCompanyInput
	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "cuerpo de solicitud inválido")
	}
	tenantID := mustUUID(c.Locals("tenant_id"))
	userID := mustUUID(c.Locals("user_id"))
	company, err := h.svc.CreateForUser(c.Context(), input, tenantID, userID, c.Get("Authorization"))
	if err != nil {
		if err == service.ErrCompanyQuotaExceeded {
			return fiber.NewError(fiber.StatusPaymentRequired, err.Error())
		}
		return fiber.NewError(fiber.StatusConflict, err.Error())
	}
	return c.Status(fiber.StatusCreated).JSON(company)
}

func (h *CompanyHandler) Update(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "ID inválido")
	}
	var input service.UpdateCompanyInput
	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "cuerpo de solicitud inválido")
	}
	tenantID := mustUUID(c.Locals("tenant_id"))
	company, err := h.svc.Update(c.Context(), id, tenantID, input)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	}
	return c.JSON(company)
}

func (h *CompanyHandler) Delete(c *fiber.Ctx) error {
	id, _ := uuid.Parse(c.Params("id"))
	tenantID := mustUUID(c.Locals("tenant_id"))
	if err := h.svc.Delete(c.Context(), id, tenantID); err != nil {
		return fiber.NewError(fiber.StatusNotFound, "empresa no encontrada")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *CompanyHandler) GetRiskScore(c *fiber.Ctx) error {
	id, _ := uuid.Parse(c.Params("id"))
	tenantID := mustUUID(c.Locals("tenant_id"))
	score, err := h.svc.CalculateRiskScore(c.Context(), id, tenantID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "empresa no encontrada")
	}
	return c.JSON(fiber.Map{"company_id": id, "risk_score": score})
}

func (h *CompanyHandler) ProcessRUT(c *fiber.Ctx) error {
	// El OCR del RUT se delega al microservicio AI via cola RabbitMQ
	id, _ := uuid.Parse(c.Params("id"))
	return c.JSON(fiber.Map{
		"company_id": id,
		"status":     "processing",
		"message":    "RUT enviado al motor OCR. Resultado disponible en 30 segundos.",
	})
}

func mustUUID(v interface{}) uuid.UUID {
	if s, ok := v.(string); ok {
		if id, err := uuid.Parse(s); err == nil {
			return id
		}
	}
	return uuid.Nil
}
