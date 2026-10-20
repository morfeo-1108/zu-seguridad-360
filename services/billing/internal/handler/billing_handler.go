package handler

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/zu-seguridad/billing/internal/model"
	"github.com/zu-seguridad/billing/internal/service"
	"github.com/zu-seguridad/billing/internal/wompi"
)

type BillingHandler struct {
	svc               *service.BillingService
	wompiEventsSecret string
}

func NewBillingHandler(svc *service.BillingService, wompiEventsSecret string) *BillingHandler {
	return &BillingHandler{svc: svc, wompiEventsSecret: wompiEventsSecret}
}

// WompiWebhook receives Wompi's transaction.updated events. It only reacts
// to events whose checksum verifies against WOMPI_EVENTS_SECRET — anything
// else is rejected outright, never processed "just in case". The exact
// property-path resolution Wompi signs (e.g. "data.transaction.id") needs
// confirming against a real payload before this gates real money; see
// internal/wompi/client.go.
func (h *BillingHandler) WompiWebhook(c *fiber.Ctx) error {
	var event struct {
		Data      map[string]interface{} `json:"data"`
		Timestamp int64                  `json:"timestamp"`
		Signature struct {
			Properties []string `json:"properties"`
			Checksum   string   `json:"checksum"`
		} `json:"signature"`
	}
	if err := c.BodyParser(&event); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid webhook payload")
	}
	if h.wompiEventsSecret == "" {
		return fiber.NewError(fiber.StatusServiceUnavailable, "WOMPI_EVENTS_SECRET not configured")
	}
	values := make([]string, 0, len(event.Signature.Properties))
	for _, prop := range event.Signature.Properties {
		v, _ := event.Data[prop].(string)
		values = append(values, v)
	}
	if !wompi.VerifyEventSignature(values, event.Timestamp, event.Signature.Checksum, h.wompiEventsSecret) {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid webhook signature")
	}
	// Signature verified; this does not yet reconcile the event against an
	// invoice (that mapping — which invoice a given Wompi transaction
	// reference belongs to — is handled by ChargeInvoice's synchronous path
	// today, not by this webhook).
	return c.JSON(fiber.Map{"received": true})
}

// CanAddCompany is called by Core before creating a non-synthetic company. It
// never creates or mutates anything; Core enforces the decision.
func (h *BillingHandler) CanAddCompany(c *fiber.Ctx) error {
	tenantID, err := uuid.Parse(c.Params("tenantId"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenantId inválido")
	}
	allowed, current, max, err := h.svc.CanAddCompany(c.Context(), tenantID)
	if err != nil {
		if errors.Is(err, service.ErrNoActiveSubscription) {
			return c.Status(fiber.StatusOK).JSON(fiber.Map{"allowed": false, "reason": "no_active_subscription"})
		}
		return fiber.NewError(fiber.StatusInternalServerError, "no fue posible verificar el cupo del plan")
	}
	return c.JSON(fiber.Map{
		"allowed":       allowed,
		"current_count": current,
		"max_companies": max,
	})
}

// ListPlans is public pricing information — the 4 commercial tiers and
// what each includes. No auth required, same as any pricing page.
func (h *BillingHandler) ListPlans(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"plans": service.Plans})
}

// GetSubscription returns the caller's own tenant subscription.
func (h *BillingHandler) GetSubscription(c *fiber.Ctx) error {
	tenantID, err := uuid.Parse(c.Params("tenantId"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenantId inválido")
	}
	sub, err := h.svc.GetSubscription(c.Context(), tenantID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "sin suscripción activa")
	}
	return c.JSON(sub)
}

// CreateSubscription registers a plan selection for the tenant. It does NOT
// attempt a real Wompi charge — payment collection stays disabled pending a
// sandbox-tested flow (see main.go) — so this is the self-service "pick a
// plan" step the pricing doc calls "autogestión: incluida", not a billing
// event yet.
func (h *BillingHandler) CreateSubscription(c *fiber.Ctx) error {
	tenantID, err := uuid.Parse(c.Params("tenantId"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "tenantId inválido")
	}
	var body struct {
		PlanCode     string `json:"plan_code"`
		WantsFounder bool   `json:"wants_founder"`
	}
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "cuerpo inválido")
	}
	sub, err := h.svc.CreateSubscription(c.Context(), tenantID, model.PlanCode(body.PlanCode), "", body.WantsFounder)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.Status(fiber.StatusCreated).JSON(sub)
}
