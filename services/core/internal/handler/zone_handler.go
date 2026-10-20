package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/zu-seguridad/core/internal/model"
	"github.com/zu-seguridad/core/internal/service"
	"strconv"
)

type ZoneHandler struct{ svc *service.ZoneService }

func NewZoneHandler(svc *service.ZoneService) *ZoneHandler { return &ZoneHandler{svc: svc} }

func (h *ZoneHandler) CreateProject(c *fiber.Ctx) error {
	companyID := mustUUID(c.Locals("company_id"))
	tenantID := mustUUID(c.Locals("tenant_id"))
	var body struct {
		Name    string `json:"name"`
		Address string `json:"address"`
	}
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "cuerpo inválido")
	}
	project, err := h.svc.CreateProject(c.Context(), tenantID, companyID, body.Name, body.Address)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.Status(fiber.StatusCreated).JSON(project)
}

func (h *ZoneHandler) ListProjects(c *fiber.Ctx) error {
	companyID := mustUUID(c.Locals("company_id"))
	tenantID := mustUUID(c.Locals("tenant_id"))
	projects, err := h.svc.ListProjects(c.Context(), tenantID, companyID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.JSON(fiber.Map{"data": projects, "total": len(projects)})
}

func (h *ZoneHandler) CreateZone(c *fiber.Ctx) error {
	projectID, _ := uuid.Parse(c.Params("projectId"))
	tenantID := mustUUID(c.Locals("tenant_id"))
	var body struct {
		Name     string      `json:"name"`
		Type     string      `json:"type"`
		Geofence model.JSONB `json:"geofence_polygon"`
		MarkerID string      `json:"visual_marker_id"`
	}
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "cuerpo inválido")
	}
	zone, err := h.svc.CreateZone(c.Context(), tenantID, projectID, body.Name, model.ZoneType(body.Type), body.Geofence, body.MarkerID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.Status(fiber.StatusCreated).JSON(zone)
}

func (h *ZoneHandler) GetChecklist(c *fiber.Ctx) error {
	zoneID, _ := uuid.Parse(c.Params("zoneId"))
	companyID := mustUUID(c.Locals("company_id"))
	tenantID := mustUUID(c.Locals("tenant_id"))
	items, err := h.svc.GetZoneChecklist(c.Context(), tenantID, companyID, zoneID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(fiber.Map{"data": items, "total": len(items)})
}

// ResolveByMarker — el dron pregunta "¿en qué zona estoy?" por marcador visual
func (h *ZoneHandler) ResolveByMarker(c *fiber.Ctx) error {
	projectID, _ := uuid.Parse(c.Params("projectId"))
	tenantID := mustUUID(c.Locals("tenant_id"))
	markerID := c.Query("marker_id")
	zone, err := h.svc.ResolveZoneByMarker(c.Context(), tenantID, projectID, markerID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	}
	return c.JSON(zone)
}

// ResolveByLocation — el dron pregunta por GPS
func (h *ZoneHandler) ResolveByLocation(c *fiber.Ctx) error {
	projectID, _ := uuid.Parse(c.Params("projectId"))
	tenantID := mustUUID(c.Locals("tenant_id"))
	lat, _ := strconv.ParseFloat(c.Query("lat"), 64)
	lng, _ := strconv.ParseFloat(c.Query("lng"), 64)
	zone, err := h.svc.ResolveZoneByLocation(c.Context(), tenantID, projectID, lat, lng)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	}
	return c.JSON(zone)
}
