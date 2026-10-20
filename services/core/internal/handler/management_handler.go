package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/zu-seguridad/core/internal/model"
	"github.com/zu-seguridad/core/internal/service"
)

type ManagementHandler struct {
	svc   *service.ManagementSystemService
	watch *service.RegulatoryWatchService
}

func NewManagementHandler(svc *service.ManagementSystemService, watch *service.RegulatoryWatchService) *ManagementHandler {
	return &ManagementHandler{svc: svc, watch: watch}
}

func managementPathID(c *fiber.Ctx) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return uuid.Nil, fiber.NewError(fiber.StatusBadRequest, "id inválido")
	}
	return id, nil
}

func (h *ManagementHandler) InitSystem(c *fiber.Ctx) error {
	tenantID := mustUUID(c.Locals("tenant_id"))
	companyID := mustUUID(c.Locals("company_id"))
	var body struct {
		Year       int                   `json:"year"`
		Frameworks []model.NormFramework `json:"frameworks"`
	}
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "cuerpo inválido")
	}
	if len(body.Frameworks) == 0 {
		body.Frameworks = []model.NormFramework{model.FwRes0312, model.FwISO45001}
	}
	sys, err := h.svc.InitializeSystem(c.Context(), tenantID, companyID, body.Year, body.Frameworks)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "no fue posible inicializar el sistema para esa empresa")
	}
	return c.Status(fiber.StatusCreated).JSON(sys)
}

func (h *ManagementHandler) GetActive(c *fiber.Ctx) error {
	sys, err := h.svc.GetActiveForCompany(c.Context(), mustUUID(c.Locals("tenant_id")), mustUUID(c.Locals("company_id")))
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "sin sistema de gestión activo para esta empresa")
	}
	return c.JSON(sys)
}

func (h *ManagementHandler) Assess(c *fiber.Ctx) error {
	systemID, err := managementPathID(c)
	if err != nil {
		return err
	}
	result, err := h.svc.AssessCompliance(c.Context(), mustUUID(c.Locals("tenant_id")), mustUUID(c.Locals("company_id")), systemID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "sistema no disponible")
	}
	return c.JSON(result)
}

func (h *ManagementHandler) ValidateStandard(c *fiber.Ctx) error {
	standardID, err := managementPathID(c)
	if err != nil {
		return err
	}
	var body struct {
		Status string `json:"status"`
	}
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "cuerpo inválido")
	}
	err = h.svc.ValidateStandard(c.Context(), mustUUID(c.Locals("tenant_id")), mustUUID(c.Locals("company_id")), standardID, mustUUID(c.Locals("user_id")), body.Status)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "no fue posible validar el estándar")
	}
	return c.JSON(fiber.Map{"validated": true, "standard_id": standardID})
}

func (h *ManagementHandler) Score(c *fiber.Ctx) error {
	systemID, err := managementPathID(c)
	if err != nil {
		return err
	}
	result, err := h.svc.ComplianceScore(c.Context(), mustUUID(c.Locals("tenant_id")), mustUUID(c.Locals("company_id")), systemID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "sistema no disponible")
	}
	return c.JSON(result)
}

func (h *ManagementHandler) AdvancePhase(c *fiber.Ctx) error {
	systemID, err := managementPathID(c)
	if err != nil {
		return err
	}
	next, err := h.svc.AdvancePhase(c.Context(), mustUUID(c.Locals("tenant_id")), mustUUID(c.Locals("company_id")), systemID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "sistema no disponible")
	}
	return c.JSON(fiber.Map{"system_id": systemID, "current_phase": next})
}

func (h *ManagementHandler) CloseLoop(c *fiber.Ctx) error {
	systemID, err := managementPathID(c)
	if err != nil {
		return err
	}
	next, err := h.svc.CloseLoop(c.Context(), mustUUID(c.Locals("tenant_id")), mustUUID(c.Locals("company_id")), systemID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "sistema no disponible")
	}
	return c.Status(fiber.StatusCreated).JSON(next)
}

func (h *ManagementHandler) PendingChanges(c *fiber.Ctx) error {
	changes, err := h.watch.PendingChanges(c.Context())
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "no fue posible consultar cambios normativos")
	}
	return c.JSON(fiber.Map{"pending": changes, "count": len(changes)})
}

func (h *ManagementHandler) ReviewChange(c *fiber.Ctx) error {
	changeID, err := managementPathID(c)
	if err != nil {
		return err
	}
	var body struct {
		Decision string `json:"decision"`
		Notes    string `json:"notes"`
	}
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "cuerpo inválido")
	}
	if body.Decision != "aprobar" && body.Decision != "descartar" {
		return fiber.NewError(fiber.StatusBadRequest, "decision inválida")
	}
	if err := h.watch.ReviewChange(c.Context(), changeID, mustUUID(c.Locals("user_id")), body.Decision, body.Notes); err != nil {
		return fiber.NewError(fiber.StatusNotFound, "cambio normativo no disponible")
	}
	return c.JSON(fiber.Map{"reviewed": true, "decision": body.Decision})
}
