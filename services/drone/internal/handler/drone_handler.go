package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/zu-seguridad/drone/internal/service"
)

type DroneHandler struct{ fleet *service.FleetManager }

func NewDroneHandler(fleet *service.FleetManager) *DroneHandler {
	return &DroneHandler{fleet: fleet}
}

func (h *DroneHandler) GetFleet(c *fiber.Ctx) error {
	tenantID, _ := uuid.Parse(c.Locals("tenant_id").(string))
	drones, err := h.fleet.GetFleetStatus(c.Context(), tenantID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(fiber.Map{"data": drones, "total": len(drones)})
}

func (h *DroneHandler) ScheduleMission(c *fiber.Ctx) error {
	var body struct {
		CompanyID   string               `json:"company_id"`
		MissionType string               `json:"mission_type"`
		Waypoints   []map[string]float64 `json:"waypoints"`
	}
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "cuerpo inválido")
	}
	companyID, err := uuid.Parse(body.CompanyID)
	if err != nil {
		if p, perr := uuid.Parse(c.Params("companyId")); perr == nil {
			companyID = p
		} else {
			return fiber.NewError(fiber.StatusBadRequest, "company_id requerido (body o ruta)")
		}
	}
	missionID, err := h.fleet.ScheduleMission(c.Context(), companyID, body.MissionType, body.Waypoints)
	if err != nil {
		return fiber.NewError(fiber.StatusServiceUnavailable, err.Error())
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"mission_id": missionID, "status": "dispatched",
		"message": "Misión enviada al dron. Telemetría disponible via WebSocket.",
	})
}

func (h *DroneHandler) Assign(c *fiber.Ctx) error {
	droneID, _ := uuid.Parse(c.Params("id"))
	companyID, _ := uuid.Parse(c.Params("companyId"))
	if err := h.fleet.AssignDroneToCompany(c.Context(), droneID, companyID); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.JSON(fiber.Map{"assigned": true})
}

func (h *DroneHandler) GetLiveStatus(c *fiber.Ctx) error {
	// En producción: retornar estado en tiempo real desde cache Redis
	return c.JSON(fiber.Map{
		"status": "live_telemetry_via_websocket",
		"ws_url": "wss://api.zu360.co/v1/drones/live",
		"note":   "Conectar via WebSocket para telemetría en tiempo real",
	})
}

// ListDrones — lista la flota (alias semántico de GetFleet para la ruta /drones).
func (h *DroneHandler) ListDrones(c *fiber.Ctx) error {
	return h.GetFleet(c)
}

// GetDrone — detalle de un dron por ID.
func (h *DroneHandler) GetDrone(c *fiber.Ctx) error {
	droneID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "id de dron inválido")
	}
	drone, err := h.fleet.GetDroneByID(c.Context(), droneID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "dron no encontrado")
	}
	return c.JSON(drone)
}

// GetLiveStream — información de conexión al stream en vivo del dron.
func (h *DroneHandler) GetLiveStream(c *fiber.Ctx) error {
	return h.GetLiveStatus(c)
}

// GetInspectionReports — reportes de inspección de una empresa.
func (h *DroneHandler) GetInspectionReports(c *fiber.Ctx) error {
	companyID, err := uuid.Parse(c.Params("companyId"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "companyId inválido")
	}
	reports, err := h.fleet.GetInspectionReports(c.Context(), companyID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(fiber.Map{"data": reports, "total": len(reports)})
}
