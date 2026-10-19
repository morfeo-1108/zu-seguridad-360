package handler

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/zu-seguridad/core/internal/service"
	"gorm.io/gorm"
)

type DataExportHandler struct {
	svc *service.DataExportService
}

func NewDataExportHandler(svc *service.DataExportService) *DataExportHandler {
	return &DataExportHandler{svc: svc}
}

// GET /v1/companies/:companyId/export — descarga el paquete portable de la empresa.
// Continuidad del negocio: los datos son de la empresa y salen en formato abierto.
func (h *DataExportHandler) ExportCompany(c *fiber.Ctx) error {
	tenantID := mustUUID(c.Locals("tenant_id"))
	companyID, err := uuid.Parse(c.Params("companyId"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "companyId inválido")
	}
	data, err := h.svc.ExportCompanyJSON(c.Context(), tenantID, companyID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fiber.NewError(fiber.StatusNotFound, "empresa no disponible")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "no fue posible generar la exportación")
	}
	c.Set("Content-Type", "application/json")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=export-%s.json", companyID))
	return c.Send(data)
}
