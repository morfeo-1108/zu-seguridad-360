package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/zu-seguridad/core/internal/service"
)

type TrainingHandler struct{ svc *service.TrainingService }

func NewTrainingHandler(svc *service.TrainingService) *TrainingHandler {
	return &TrainingHandler{svc: svc}
}

func (h *TrainingHandler) List(c *fiber.Ctx) error {
	companyID, _ := uuid.Parse(c.Params("companyId"))
	tenantID := mustUUID(c.Locals("tenant_id"))
	trainings, err := h.svc.List(c.Context(), companyID, tenantID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(fiber.Map{"data": trainings, "total": len(trainings)})
}

func (h *TrainingHandler) Create(c *fiber.Ctx) error {
	companyID, err := uuid.Parse(c.Params("companyId"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "companyId inválido")
	}
	tenantID := mustUUID(c.Locals("tenant_id"))
	var input service.CreateTrainingInput
	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "cuerpo inválido")
	}
	training, err := h.svc.Create(c.Context(), input, companyID, tenantID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.Status(fiber.StatusCreated).JSON(training)
}

func (h *TrainingHandler) Complete(c *fiber.Ctx) error {
	trainingID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "ID inválido")
	}
	companyID, err := uuid.Parse(c.Params("companyId"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "companyId inválido")
	}
	tenantID := mustUUID(c.Locals("tenant_id"))
	var input service.RegisterAttendanceInput
	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "cuerpo inválido")
	}
	certs, err := h.svc.Complete(c.Context(), trainingID, companyID, tenantID, input)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.JSON(fiber.Map{
		"simulation":           true,
		"certificates_created": len(certs),
		"certificates":         certs,
		"notice":               "Constancias sintéticas no válidas; no acreditan capacitación ni vigencia normativa.",
	})
}

func (h *TrainingHandler) VerifyCertificate(c *fiber.Ctx) error {
	folio := c.Params("folio")
	cert, err := h.svc.VerifyCertificate(c.Context(), folio)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	}
	return c.JSON(fiber.Map{"valid": true, "certificate": cert})
}

func (h *TrainingHandler) GetExpiring(c *fiber.Ctx) error {
	tenantID := mustUUID(c.Locals("tenant_id"))
	var companyID *uuid.UUID
	if companyIDText, ok := c.Locals("company_id").(string); ok && companyIDText != "" {
		parsed, err := uuid.Parse(companyIDText)
		if err != nil || parsed == uuid.Nil {
			return fiber.NewError(fiber.StatusForbidden, "company scope required")
		}
		companyID = &parsed
	}
	days := 30
	certs, err := h.svc.GetExpiringCertificates(c.Context(), tenantID, companyID, days)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(fiber.Map{"data": certs, "total": len(certs), "threshold_days": days})
}
