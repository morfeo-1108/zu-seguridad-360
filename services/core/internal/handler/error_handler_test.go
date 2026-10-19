package handler

import (
	"io"
	"log"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestGlobalErrorHandlerRedactsInternalErrorMessages(t *testing.T) {
	const sensitive = "PII-SECRET-MARKER-8841"
	previousWriter := log.Writer()
	var logs strings.Builder
	log.SetOutput(&logs)
	defer log.SetOutput(previousWriter)

	app := fiber.New(fiber.Config{ErrorHandler: GlobalErrorHandler})
	app.Get("/failure", func(c *fiber.Ctx) error {
		return fiber.NewError(fiber.StatusInternalServerError, sensitive)
	})

	response, err := app.Test(httptest.NewRequest("GET", "/failure", nil))
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read response: %v", err)
	}
	if response.StatusCode != fiber.StatusInternalServerError {
		t.Fatalf("expected HTTP 500, got %d", response.StatusCode)
	}
	if strings.Contains(string(body), sensitive) {
		t.Fatal("response exposed internal error text")
	}
	if strings.Contains(logs.String(), sensitive) {
		t.Fatal("log exposed internal error text")
	}
	if !strings.Contains(string(body), "internal server error") {
		t.Fatalf("expected a generic error response, got %s", body)
	}
}
