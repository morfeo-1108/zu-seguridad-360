package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/zu-seguridad/core/internal/service"
)

type CommitteeHandler struct{ svc *service.CommitteeService }

func NewCommitteeHandler(svc *service.CommitteeService) *CommitteeHandler {
	return &CommitteeHandler{svc: svc}
}

func (h *CommitteeHandler) Create(c *fiber.Ctx) error {
	tenantID := mustUUID(c.Locals("tenant_id"))
	companyID := mustUUID(c.Locals("company_id"))
	var input service.CreateCommitteeInput
	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "cuerpo inválido")
	}
	committee, err := h.svc.CreateCommittee(c.Context(), input, companyID, tenantID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.Status(fiber.StatusCreated).JSON(committee)
}

func (h *CommitteeHandler) List(c *fiber.Ctx) error {
	tenantID := mustUUID(c.Locals("tenant_id"))
	companyID := mustUUID(c.Locals("company_id"))
	committees, err := h.svc.List(c.Context(), companyID, tenantID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "no fue posible consultar comités")
	}
	return c.JSON(fiber.Map{"data": committees, "total": len(committees)})
}

func (h *CommitteeHandler) CreateMeeting(c *fiber.Ctx) error {
	tenantID := mustUUID(c.Locals("tenant_id"))
	companyID := mustUUID(c.Locals("company_id"))
	userID := mustUUID(c.Locals("user_id"))
	committeeID, err := uuid.Parse(c.Params("committeeId"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "committeeId inválido")
	}
	var input service.CreateMeetingInput
	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "cuerpo inválido")
	}
	meeting, err := h.svc.CreateMeeting(c.Context(), committeeID, input, companyID, tenantID, userID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.Status(fiber.StatusCreated).JSON(meeting)
}

func (h *CommitteeHandler) ListMeetings(c *fiber.Ctx) error {
	tenantID := mustUUID(c.Locals("tenant_id"))
	companyID := mustUUID(c.Locals("company_id"))
	committeeID, err := uuid.Parse(c.Params("committeeId"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "committeeId inválido")
	}
	meetings, err := h.svc.ListMeetings(c.Context(), committeeID, companyID, tenantID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "no fue posible consultar actas")
	}
	return c.JSON(fiber.Map{"data": meetings, "total": len(meetings)})
}
