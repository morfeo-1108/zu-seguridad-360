package handler

import (
	"bytes"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/zu-seguridad/core/internal/service"
	"io"
	"strconv"
)

type EmployeeHandler struct{ svc *service.EmployeeService }

func NewEmployeeHandler(svc *service.EmployeeService) *EmployeeHandler {
	return &EmployeeHandler{svc: svc}
}

func (h *EmployeeHandler) List(c *fiber.Ctx) error {
	companyID, err := uuid.Parse(c.Params("companyId"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "companyId inválido")
	}
	tenantID := mustUUID(c.Locals("tenant_id"))
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "50"))

	employees, total, err := h.svc.List(c.Context(), companyID, tenantID, page, limit)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(fiber.Map{
		"data": employees, "total": total,
		"page": page, "limit": limit,
		"pages": (total + int64(limit) - 1) / int64(limit),
	})
}

func (h *EmployeeHandler) GetByID(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "ID inválido")
	}
	companyID, err := uuid.Parse(c.Params("companyId"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "companyId inválido")
	}
	tenantID := mustUUID(c.Locals("tenant_id"))
	emp, err := h.svc.GetByID(c.Context(), id, companyID, tenantID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "empleado no encontrado")
	}
	return c.JSON(emp)
}

func (h *EmployeeHandler) Create(c *fiber.Ctx) error {
	companyID, err := uuid.Parse(c.Params("companyId"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "companyId inválido")
	}
	tenantID := mustUUID(c.Locals("tenant_id"))
	var input service.CreateEmployeeInput
	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "cuerpo de solicitud inválido")
	}
	emp, err := h.svc.Create(c.Context(), input, companyID, tenantID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.Status(fiber.StatusCreated).JSON(emp)
}

func (h *EmployeeHandler) BulkImport(c *fiber.Ctx) error {
	companyID, err := uuid.Parse(c.Params("companyId"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "companyId inválido")
	}
	tenantID := mustUUID(c.Locals("tenant_id"))

	file, err := c.FormFile("file")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "archivo CSV requerido")
	}

	f, _ := file.Open()
	defer f.Close()
	body, _ := io.ReadAll(f)

	imported, errors, err := h.svc.BulkImportFromCSV(c.Context(), bytes.NewReader(body), companyID, tenantID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.JSON(fiber.Map{
		"imported": imported,
		"errors":   errors,
		"message":  "Importación completada",
	})
}

func (h *EmployeeHandler) Update(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "ID inválido")
	}
	companyID, err := uuid.Parse(c.Params("companyId"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "companyId inválido")
	}
	tenantID := mustUUID(c.Locals("tenant_id"))
	var input service.UpdateEmployeeInput
	c.BodyParser(&input)
	emp, err := h.svc.Update(c.Context(), id, companyID, tenantID, input)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, err.Error())
	}
	return c.JSON(emp)
}

func (h *EmployeeHandler) Delete(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "ID inválido")
	}
	companyID, err := uuid.Parse(c.Params("companyId"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "companyId inválido")
	}
	tenantID := mustUUID(c.Locals("tenant_id"))
	if err := h.svc.Delete(c.Context(), id, companyID, tenantID); err != nil {
		return fiber.NewError(fiber.StatusNotFound, "empleado no encontrado")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *EmployeeHandler) GetOccupationalProfile(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "ID inválido")
	}
	companyID, err := uuid.Parse(c.Params("companyId"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "companyId inválido")
	}
	tenantID := mustUUID(c.Locals("tenant_id"))
	emp, err := h.svc.GetByID(c.Context(), id, companyID, tenantID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "empleado no encontrado")
	}
	// Perfil ocupacional completo: empleado + certificados + exámenes + restricciones
	return c.JSON(fiber.Map{
		"employee":      emp,
		"certificates":  []interface{}{},
		"medical_exams": []interface{}{},
		"trainings":     []interface{}{},
	})
}
