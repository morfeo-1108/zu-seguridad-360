package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/zu-seguridad/core/internal/service"
)

type IncidentHandler struct{ svc *service.IncidentService }

func NewIncidentHandler(svc *service.IncidentService) *IncidentHandler {
	return &IncidentHandler{svc: svc}
}

func (h *IncidentHandler) List(c *fiber.Ctx) error {
	companyID, _ := uuid.Parse(c.Params("companyId"))
	tenantID := mustUUID(c.Locals("tenant_id"))
	incidents, err := h.svc.GetByCompany(c.Context(), companyID, tenantID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(fiber.Map{"data": incidents, "total": len(incidents)})
}

func (h *IncidentHandler) Create(c *fiber.Ctx) error {
	companyID, err := uuid.Parse(c.Params("companyId"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "companyId inválido")
	}
	tenantID := mustUUID(c.Locals("tenant_id"))
	reportedBy := mustUUID(c.Locals("user_id"))

	var input service.CreateIncidentInput
	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "cuerpo inválido")
	}
	incident, err := h.svc.Create(c.Context(), input, companyID, tenantID, reportedBy)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.Status(fiber.StatusCreated).JSON(incident)
}

func (h *IncidentHandler) GetByID(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "ID inválido")
	}
	companyID, err := uuid.Parse(c.Params("companyId"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "companyId inválido")
	}
	tenantID := mustUUID(c.Locals("tenant_id"))
	incident, err := h.svc.GetByID(c.Context(), id, companyID, tenantID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "incidente no encontrado")
	}
	return c.JSON(incident)
}

func (h *IncidentHandler) Investigate(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "ID inválido")
	}
	companyID, err := uuid.Parse(c.Params("companyId"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "companyId inválido")
	}
	tenantID := mustUUID(c.Locals("tenant_id"))
	investigatedBy := mustUUID(c.Locals("user_id"))

	var input service.UpdateInvestigationInput
	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "cuerpo inválido")
	}
	incident, err := h.svc.UpdateInvestigation(c.Context(), id, companyID, tenantID, investigatedBy, input)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	}
	return c.JSON(incident)
}

func (h *IncidentHandler) GenerateFURAT(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "ID inválido")
	}
	companyID, err := uuid.Parse(c.Params("companyId"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "companyId inválido")
	}
	tenantID := mustUUID(c.Locals("tenant_id"))
	furat, err := h.svc.GenerateFURAT(c.Context(), id, companyID, tenantID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	}
	return c.JSON(furat)
}

func (h *IncidentHandler) Close(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "ID inválido")
	}
	companyID, err := uuid.Parse(c.Params("companyId"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "companyId inválido")
	}
	tenantID := mustUUID(c.Locals("tenant_id"))
	closedBy := mustUUID(c.Locals("user_id"))
	incident, err := h.svc.CloseSynthetic(c.Context(), id, companyID, tenantID, closedBy)
	if err != nil {
		return fiber.NewError(fiber.StatusConflict, "No fue posible cerrar el incidente sintético.")
	}
	return c.JSON(incident)
}

func (h *IncidentHandler) GetOpenByTenant(c *fiber.Ctx) error {
	tenantID := mustUUID(c.Locals("tenant_id"))
	var companyID *uuid.UUID
	if companyIDText, ok := c.Locals("company_id").(string); ok && companyIDText != "" {
		parsed, err := uuid.Parse(companyIDText)
		if err != nil || parsed == uuid.Nil {
			return fiber.NewError(fiber.StatusForbidden, "company scope required")
		}
		companyID = &parsed
	}
	incidents, err := h.svc.GetOpenByTenant(c.Context(), tenantID, companyID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(fiber.Map{"data": incidents, "total": len(incidents)})
}
