package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/zu-seguridad/core/internal/service"
)

type PortfolioHandler struct {
	svc *service.PortfolioService
}

func NewPortfolioHandler(svc *service.PortfolioService) *PortfolioHandler {
	return &PortfolioHandler{svc: svc}
}

func (h *PortfolioHandler) Summary(c *fiber.Ctx) error {
	tenantID := mustUUID(c.Locals("tenant_id"))
	userID := mustUUID(c.Locals("user_id"))

	summaries, err := h.svc.Summary(c.Context(), userID, tenantID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "no se pudo construir el panel de portafolio")
	}
	return c.JSON(fiber.Map{"data": summaries})
}
