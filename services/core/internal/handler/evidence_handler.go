package handler

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/zu-seguridad/core/internal/model"
	"github.com/zu-seguridad/core/internal/service"
	"gorm.io/gorm"
)

type EvidenceHandler struct{ svc *service.EvidenceService }

func NewEvidenceHandler(svc *service.EvidenceService) *EvidenceHandler {
	return &EvidenceHandler{svc: svc}
}

func parseEvidenceID(value, name string) (uuid.UUID, error) {
	id, err := uuid.Parse(value)
	if err != nil {
		return uuid.Nil, fiber.NewError(fiber.StatusBadRequest, name+" inválido")
	}
	return id, nil
}

func (h *EvidenceHandler) Record(c *fiber.Ctx) error {
	var ev model.EvidenceRecord
	if err := c.BodyParser(&ev); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "cuerpo inválido")
	}
	ev.TenantID = mustUUID(c.Locals("tenant_id"))
	// company_id must come from the token's active company, never the client body —
	// otherwise a user could record evidence under any company in their tenant,
	// bypassing the switch-company membership check.
	ev.CompanyID = mustUUID(c.Locals("company_id"))
	// Client-supplied detections, compliance decisions and hashes are not trusted.
	ev.ID = uuid.Nil
	ev.Compliant = nil
	ev.DetectionData = nil
	ev.Severity = ""
	ev.PhotoChecksum = ""
	ev.LegalHold = false
	ev.Source = "manual"
	if err := h.svc.Record(c.Context(), &ev); err != nil {
		if err == gorm.ErrRecordNotFound {
			return fiber.NewError(fiber.StatusNotFound, "recurso relacionado no disponible")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "no fue posible registrar la evidencia")
	}
	return c.Status(fiber.StatusCreated).JSON(ev)
}

func (h *EvidenceHandler) EmployeeTimeline(c *fiber.Ctx) error {
	tenantID := mustUUID(c.Locals("tenant_id"))
	companyID := mustUUID(c.Locals("company_id"))
	empID, err := parseEvidenceID(c.Params("employeeId"), "employeeId")
	if err != nil {
		return err
	}
	limit, err := strconv.Atoi(c.Query("limit", "50"))
	if err != nil || limit < 1 || limit > 200 {
		return fiber.NewError(fiber.StatusBadRequest, "limit debe estar entre 1 y 200")
	}
	records, err := h.svc.EmployeeTimeline(c.Context(), tenantID, companyID, empID, limit)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "no fue posible consultar evidencias")
	}
	stats, err := h.svc.Stats(c.Context(), tenantID, companyID, empID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "no fue posible calcular estadísticas")
	}
	return c.JSON(fiber.Map{"data": records, "total": len(records), "stats": stats})
}

func (h *EvidenceHandler) ComplianceProof(c *fiber.Ctx) error {
	tenantID := mustUUID(c.Locals("tenant_id"))
	companyID := mustUUID(c.Locals("company_id"))
	empID, err := parseEvidenceID(c.Params("employeeId"), "employeeId")
	if err != nil {
		return err
	}
	date, err := time.Parse("2006-01-02", c.Query("date"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "date debe tener formato YYYY-MM-DD")
	}
	code := c.Query("code")
	if code == "" || len(code) > 50 {
		return fiber.NewError(fiber.StatusBadRequest, "code requerido")
	}
	proof, err := h.svc.ComplianceProof(c.Context(), tenantID, companyID, empID, code, date)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"found": false, "message": "No hay prueba disponible para ese ítem y fecha"})
		}
		return fiber.NewError(fiber.StatusInternalServerError, "no fue posible consultar la prueba")
	}
	return c.JSON(fiber.Map{"found": true, "proof": proof})
}

func (h *EvidenceHandler) ProjectEvidence(c *fiber.Ctx) error {
	tenantID := mustUUID(c.Locals("tenant_id"))
	companyID := mustUUID(c.Locals("company_id"))
	projectID, err := parseEvidenceID(c.Params("projectId"), "projectId")
	if err != nil {
		return err
	}
	from, err := time.Parse("2006-01-02", c.Query("from", "2020-01-01"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "from debe tener formato YYYY-MM-DD")
	}
	to, err := time.Parse("2006-01-02", c.Query("to", time.Now().Format("2006-01-02")))
	if err != nil || to.Before(from) {
		return fiber.NewError(fiber.StatusBadRequest, "rango de fechas inválido")
	}
	records, err := h.svc.ProjectEvidence(c.Context(), tenantID, companyID, projectID, from, to.Add(24*time.Hour))
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "no fue posible consultar evidencias")
	}
	return c.JSON(fiber.Map{"data": records, "total": len(records)})
}
