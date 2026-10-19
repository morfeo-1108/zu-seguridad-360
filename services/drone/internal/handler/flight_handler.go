package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/zu-seguridad/drone/internal/flight"
	"github.com/zu-seguridad/drone/internal/service"
)

type FlightHandler struct{ svc *service.FlightService }

func NewFlightHandler(svc *service.FlightService) *FlightHandler {
	return &FlightHandler{svc: svc}
}

// PlanMission — genera plan + pre-vuelo + clima. NO vuela; el PIC revisa primero.
func (h *FlightHandler) PlanMission(c *fiber.Ctx) error {
	var req service.PlanRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "cuerpo inválido")
	}
	result := h.svc.PlanMission(req)
	return c.JSON(result)
}

// Authorize — el PIC autoriza y la misión arranca de forma autónoma.
func (h *FlightHandler) Authorize(c *fiber.Ctx) error {
	missionID := c.Params("missionId")
	var body struct {
		Preflight  flight.PreflightResult `json:"preflight"`
		PICName    string                 `json:"pic_name"`
		PICLicense string                 `json:"pic_license"`
	}
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "cuerpo inválido")
	}
	// Requisito legal: debe identificarse el piloto al mando
	if body.PICName == "" || body.PICLicense == "" {
		return fiber.NewError(fiber.StatusForbidden, "se requiere piloto al mando (PIC) identificado con licencia")
	}
	if err := h.svc.AuthorizeAndStart(c.Context(), missionID, body.Preflight); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.JSON(fiber.Map{"mission_id": missionID, "status": "autonomous_flight_started", "pic": body.PICName})
}

// TakeControl — el PIC interrumpe la autonomía y toma control manual.
func (h *FlightHandler) TakeControl(c *fiber.Ctx) error {
	missionID := c.Params("missionId")
	if err := h.svc.TakeControl(missionID); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.JSON(fiber.Map{"mission_id": missionID, "status": "manual_control_taken"})
}

// Status — fase actual + bitácora de decisiones del piloto (caja negra).
func (h *FlightHandler) Status(c *fiber.Ctx) error {
	missionID := c.Params("missionId")
	status, err := h.svc.Status(missionID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	}
	return c.JSON(status)
}
