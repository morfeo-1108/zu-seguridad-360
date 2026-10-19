package handler

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func newTestApp(h *BillingHandler) *fiber.App {
	app := fiber.New()
	app.Post("/v1/webhooks/wompi", h.WompiWebhook)
	return app
}

func TestWompiWebhookRejectsWithoutEventsSecretConfigured(t *testing.T) {
	h := NewBillingHandler(nil, "")
	app := newTestApp(h)

	req := httptest.NewRequest("POST", "/v1/webhooks/wompi", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	if resp.StatusCode != fiber.StatusServiceUnavailable {
		t.Fatalf("expected 503 when WOMPI_EVENTS_SECRET is unset, got %d", resp.StatusCode)
	}
}

func TestWompiWebhookRejectsBadSignature(t *testing.T) {
	h := NewBillingHandler(nil, "real-secret")
	app := newTestApp(h)

	body, _ := json.Marshal(map[string]interface{}{
		"data":      map[string]interface{}{"transaction": map[string]interface{}{"id": "tx_1"}},
		"timestamp": 1700000000,
		"signature": map[string]interface{}{"properties": []string{"transaction.id"}, "checksum": "not-the-real-checksum"},
	})
	req := httptest.NewRequest("POST", "/v1/webhooks/wompi", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("expected 401 for a bad checksum, got %d", resp.StatusCode)
	}
}

func TestWompiWebhookAcceptsValidSignature(t *testing.T) {
	secret := "real-secret"
	h := NewBillingHandler(nil, secret)
	app := newTestApp(h)

	// This payload's top-level "data" keys are flat strings so the handler's
	// naive property lookup (event.Data[prop]) finds them directly.
	concat := "tx_1" + fmt.Sprintf("%d", int64(1700000000)) + secret
	sum := sha256.Sum256([]byte(concat))
	checksum := hex.EncodeToString(sum[:])

	body, _ := json.Marshal(map[string]interface{}{
		"data":      map[string]interface{}{"transaction_id": "tx_1"},
		"timestamp": 1700000000,
		"signature": map[string]interface{}{"properties": []string{"transaction_id"}, "checksum": checksum},
	})
	req := httptest.NewRequest("POST", "/v1/webhooks/wompi", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200 for a valid checksum, got %d", resp.StatusCode)
	}
}
