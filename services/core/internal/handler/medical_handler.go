package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/zu-seguridad/core/internal/service"
)

type MedicalHandler struct{ svc *service.MedicalExamService }

func NewMedicalHandler(svc *service.MedicalExamService) *MedicalHandler {
	return &MedicalHandler{svc: svc}
}

func (h *MedicalHandler) Schedule(c *fiber.Ctx) error {
	companyID := mustUUID(c.Locals("company_id"))
	tenantID := mustUUID(c.Locals("tenant_id"))
	var input service.ScheduleExamInput
	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "cuerpo inválido")
	}
	exam, err := h.svc.Schedule(c.Context(), input, companyID, tenantID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.Status(fiber.StatusCreated).JSON(exam)
}

func (h *MedicalHandler) Review(c *fiber.Ctx) error {
	examID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "id inválido")
	}
	tenantID := mustUUID(c.Locals("tenant_id"))
	companyID := mustUUID(c.Locals("company_id"))
	reviewedBy := mustUUID(c.Locals("user_id"))
	var input service.ReviewExamInput
	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "cuerpo inválido")
	}
	exam, err := h.svc.Review(c.Context(), examID, tenantID, companyID, reviewedBy, input)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.JSON(exam)
}

func (h *MedicalHandler) GetPendingReview(c *fiber.Ctx) error {
	tenantID := mustUUID(c.Locals("tenant_id"))
	companyID := mustUUID(c.Locals("company_id"))
	exams, err := h.svc.GetPendingReview(c.Context(), tenantID, companyID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(fiber.Map{"data": exams, "total": len(exams)})
}

func (h *MedicalHandler) GetExpiring(c *fiber.Ctx) error {
	tenantID := mustUUID(c.Locals("tenant_id"))
	companyID := mustUUID(c.Locals("company_id"))
	exams, err := h.svc.GetExpiring(c.Context(), tenantID, companyID, 30)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(fiber.Map{"data": exams, "total": len(exams)})
}

func (h *MedicalHandler) GetByEmployee(c *fiber.Ctx) error {
	empID, err := uuid.Parse(c.Params("employeeId"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "employeeId inválido")
	}
	tenantID := mustUUID(c.Locals("tenant_id"))
	companyID := mustUUID(c.Locals("company_id"))
	exams, err := h.svc.GetByEmployee(c.Context(), empID, tenantID, companyID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(fiber.Map{"data": exams, "total": len(exams)})
}
