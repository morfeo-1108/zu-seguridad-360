package handler

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/zu-seguridad/billing/internal/service"
	"github.com/zu-seguridad/billing/internal/wompi"
)

func app(secret string) *fiber.App {
	pay := service.NewPayments(nil, wompi.NewClient(wompi.Config{}), nil, service.DianConfig{}, "")
	h := NewBillingHandler(nil, pay, secret)
	a := fiber.New()
	a.Post("/webhook", h.WompiWebhook)
	return a
}

func post(t *testing.T, a *fiber.App, body string) int {
	resp, err := a.Test(httptest.NewRequest("POST", "/webhook", strings.NewReader(body)))
	if err != nil {
		t.Fatal(err)
	}
	return resp.StatusCode
}

func TestWompiWebhookRejectsWithoutEventsSecretConfigured(t *testing.T) {
	if code := post(t, app(""), `{}`); code != 503 {
		t.Fatalf("expected 503, got %d", code)
	}
}

func TestWompiWebhookRejectsBadSignature(t *testing.T) {
	body := `{"event":"transaction.updated","data":{"transaction":{"id":"t","status":"APPROVED","amount_in_cents":1}},"signature":{"properties":["transaction.id"],"checksum":"bad"},"timestamp":1}`
	if code := post(t, app("test_events_x"), body); code != 401 {
		t.Fatalf("expected 401, got %d", code)
	}
}
