package handler

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/zu-seguridad/core/internal/service"
)

type IndicatorHandler struct{ svc *service.IndicatorService }

func NewIndicatorHandler(svc *service.IndicatorService) *IndicatorHandler {
	return &IndicatorHandler{svc: svc}
}

func (h *IndicatorHandler) Monthly(c *fiber.Ctx) error {
	tenantID := mustUUID(c.Locals("tenant_id"))
	companyID := mustUUID(c.Locals("company_id"))
	now := time.Now()
	year, _ := strconv.Atoi(c.Query("year", strconv.Itoa(now.Year())))
	month, _ := strconv.Atoi(c.Query("month", strconv.Itoa(int(now.Month()))))
	standardHours, _ := strconv.Atoi(c.Query("standard_hours", "220"))

	result, err := h.svc.MonthlyIndicators(c.Context(), companyID, tenantID, year, month, standardHours)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.JSON(result)
}
