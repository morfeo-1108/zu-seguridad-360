package handler

import (
	"errors"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/zu-seguridad/billing/internal/model"
	"github.com/zu-seguridad/billing/internal/service"
	"github.com/zu-seguridad/billing/internal/wompi"
)

type BillingHandler struct {
	svc               *service.BillingService
	pay               *service.Payments
	wompiEventsSecret string
}

func NewBillingHandler(svc *service.BillingService, pay *service.Payments, wompiEventsSecret string) *BillingHandler {
	return &BillingHandler{svc: svc, pay: pay, wompiEventsSecret: wompiEventsSecret}
}

// WompiWebhook receives Wompi events. A bad signature is rejected; a valid
// one is reconciled against Wompi's API before anything changes. Wompi
// retries non-2xx responses, so transient failures return 5xx on purpose.
func (h *BillingHandler) WompiWebhook(c *fiber.Ctx) error {
	if h.wompiEventsSecret == "" {
		return fiber.NewError(fiber.StatusServiceUnavailable, "WOMPI_EVENTS_SECRET not configured")
	}
	err := h.pay.HandleEvent(c.Context(), c.Body(), h.wompiEventsSecret)
	switch {
	case err == nil:
		return c.JSON(fiber.Map{"received": true})
	case errors.Is(err, wompi.ErrBadSignature):
		return fiber.NewError(fiber.StatusUnauthorized, "invalid webhook signature")
	case errors.Is(err, service.ErrAmountMismatch):
		log.Printf("wompi webhook: %v", err)
		return c.JSON(fiber.Map{"received": true, "flagged": "amount_mismatch"})
	default:
		log.Printf("wompi webhook processing failed: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "processing failed")
	}
}

func tenantParam(c *fiber.Ctx) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Params("tenantId"))
	if err != nil {
		return uuid.Nil, fiber.NewError(fiber.StatusBadRequest, "tenantId inválido")
	}
	return id, nil
}

// CanAddCompany is called by Core before creating a non-synthetic company. It
// never creates or mutates anything; Core enforces the decision.
func (h *BillingHandler) CanAddCompany(c *fiber.Ctx) error {
	tenantID, err := tenantParam(c)
	if err != nil {
		return err
	}
	allowed, current, max, err := h.svc.CanAddCompany(c.Context(), tenantID)
	if err != nil {
		if errors.Is(err, service.ErrNoActiveSubscription) {
			return c.Status(fiber.StatusOK).JSON(fiber.Map{"allowed": false, "reason": "no_active_subscription"})
		}
		return fiber.NewError(fiber.StatusInternalServerError, "no fue posible verificar el cupo del plan")
	}
	return c.JSON(fiber.Map{"allowed": allowed, "current_count": current, "max_companies": max})
}

// ListPlans is public pricing information.
func (h *BillingHandler) ListPlans(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"plans": service.Plans})
}

func (h *BillingHandler) GetSubscription(c *fiber.Ctx) error {
	tenantID, err := tenantParam(c)
	if err != nil {
		return err
	}
	sub, err := h.svc.GetSubscription(c.Context(), tenantID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "sin suscripción activa")
	}
	return c.JSON(sub)
}

// CreateSubscription records the plan choice; payment happens per invoice.
func (h *BillingHandler) CreateSubscription(c *fiber.Ctx) error {
	tenantID, err := tenantParam(c)
	if err != nil {
		return err
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

func (h *BillingHandler) GetProfile(c *fiber.Ctx) error {
	tenantID, err := tenantParam(c)
	if err != nil {
		return err
	}
	bp, err := h.pay.GetProfile(c.Context(), tenantID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "sin perfil fiscal")
	}
	return c.JSON(bp)
}

func (h *BillingHandler) SaveProfile(c *fiber.Ctx) error {
	tenantID, err := tenantParam(c)
	if err != nil {
		return err
	}
	var bp model.BillingProfile
	if err := c.BodyParser(&bp); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "cuerpo inválido")
	}
	bp.TenantID = tenantID
	if err := h.pay.SaveProfile(c.Context(), &bp); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.JSON(bp)
}

func (h *BillingHandler) ListInvoices(c *fiber.Ctx) error {
	tenantID, err := tenantParam(c)
	if err != nil {
		return err
	}
	inv, err := h.pay.ListInvoices(c.Context(), tenantID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "no fue posible listar facturas")
	}
	return c.JSON(fiber.Map{"data": inv, "total": len(inv)})
}

// IssueInvoice creates (or returns) the current period's invoice.
func (h *BillingHandler) IssueInvoice(c *fiber.Ctx) error {
	tenantID, err := tenantParam(c)
	if err != nil {
		return err
	}
	inv, err := h.pay.IssueCurrentInvoice(c.Context(), tenantID)
	if errors.Is(err, service.ErrNoActiveSubscription) {
		return fiber.NewError(fiber.StatusConflict, "sin suscripción activa: elija un plan primero")
	}
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "no fue posible generar la factura")
	}
	return c.Status(fiber.StatusCreated).JSON(inv)
}

func (h *BillingHandler) Checkout(c *fiber.Ctx) error {
	tenantID, err := tenantParam(c)
	if err != nil {
		return err
	}
	invoiceID, err := uuid.Parse(c.Params("invoiceId"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invoiceId inválido")
	}
	email, _ := c.Locals("email").(string)
	att, link, err := h.pay.StartCheckout(c.Context(), tenantID, invoiceID, email)
	switch {
	case errors.Is(err, service.ErrInvoiceNotFound):
		return fiber.NewError(fiber.StatusNotFound, "factura no encontrada")
	case errors.Is(err, service.ErrInvoiceNotPayable):
		return fiber.NewError(fiber.StatusConflict, "la factura no está pendiente de pago")
	case errors.Is(err, wompi.ErrNotConfigured):
		return fiber.NewError(fiber.StatusServiceUnavailable, "pagos en línea no configurados todavía")
	case err != nil:
		log.Printf("checkout: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "no fue posible iniciar el pago")
	}
	return c.JSON(fiber.Map{"checkout_url": link, "reference": att.Reference, "environment": att.Environment, "amount_in_cents": att.AmountInCents})
}

// RetryElectronic re-attempts a document blocked on missing data (fiscal
// profile, issuer configuration) once the tenant completed it.
func (h *BillingHandler) RetryElectronic(c *fiber.Ctx) error {
	tenantID, err := tenantParam(c)
	if err != nil {
		return err
	}
	invoiceID, err := uuid.Parse(c.Params("invoiceId"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invoiceId inválido")
	}
	if _, err := h.pay.ElectronicFor(c.Context(), tenantID, invoiceID); err != nil {
		return fiber.NewError(fiber.StatusNotFound, "factura electrónica no encontrada")
	}
	e, err := h.pay.IssueElectronic(c.Context(), invoiceID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.JSON(e)
}

func (h *BillingHandler) Electronic(c *fiber.Ctx) error {
	tenantID, err := tenantParam(c)
	if err != nil {
		return err
	}
	invoiceID, err := uuid.Parse(c.Params("invoiceId"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invoiceId inválido")
	}
	e, err := h.pay.ElectronicFor(c.Context(), tenantID, invoiceID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "factura electrónica no generada")
	}
	if c.Query("format") == "xml" && len(e.XML) > 0 {
		c.Set("Content-Type", "application/xml")
		c.Set("Content-Disposition", `attachment; filename="`+e.Number+`.xml"`)
		return c.Send(e.XML)
	}
	return c.JSON(e)
}
