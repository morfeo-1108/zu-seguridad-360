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
	companyID, _ := uuid.Parse(c.Params("companyId"))
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
	examID, _ := uuid.Parse(c.Params("id"))
	tenantID := mustUUID(c.Locals("tenant_id"))
	reviewedBy := mustUUID(c.Locals("user_id"))
	var input service.ReviewExamInput
	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "cuerpo inválido")
	}
	exam, err := h.svc.Review(c.Context(), examID, tenantID, reviewedBy, input)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.JSON(exam)
}

func (h *MedicalHandler) GetPendingReview(c *fiber.Ctx) error {
	tenantID := mustUUID(c.Locals("tenant_id"))
	exams, err := h.svc.GetPendingReview(c.Context(), tenantID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(fiber.Map{"data": exams, "total": len(exams)})
}

func (h *MedicalHandler) GetExpiring(c *fiber.Ctx) error {
	tenantID := mustUUID(c.Locals("tenant_id"))
	exams, err := h.svc.GetExpiring(c.Context(), tenantID, 30)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(fiber.Map{"data": exams, "total": len(exams)})
}

func (h *MedicalHandler) GetByEmployee(c *fiber.Ctx) error {
	empID, _ := uuid.Parse(c.Params("employeeId"))
	tenantID := mustUUID(c.Locals("tenant_id"))
	exams, err := h.svc.GetByEmployee(c.Context(), empID, tenantID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(fiber.Map{"data": exams, "total": len(exams)})
}
