package middleware

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestRequireAllowedOriginAcceptsOnlyExactConfiguredOrigin(t *testing.T) {
	app := fiber.New()
	app.Post("/signup", RequireAllowedOrigin("https://app.example.co,https://portal.example.co"), func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNoContent)
	})

	for _, tc := range []struct {
		name   string
		origin string
		want   int
	}{
		{"allowed", "https://app.example.co", fiber.StatusNoContent},
		{"trailing slash normalized", "https://app.example.co/", fiber.StatusNoContent},
		{"unlisted", "https://evil.example", fiber.StatusForbidden},
		{"missing", "", fiber.StatusForbidden},
		{"wildcard-like", "https://app.example.co.evil", fiber.StatusForbidden},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/signup", strings.NewReader("{}"))
			req.Header.Set("Content-Type", "application/json")
			if tc.origin != "" {
				req.Header.Set("Origin", tc.origin)
			}
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("app.Test: %v", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != tc.want {
				t.Fatalf("expected %d, got %d", tc.want, resp.StatusCode)
			}
		})
	}
}

func TestRequestBodyLimitRejectsOversizedBody(t *testing.T) {
	app := fiber.New()
	app.Post("/signup", RequestBodyLimit(8), func(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusNoContent) })
	req := httptest.NewRequest("POST", "/signup", strings.NewReader(strings.Repeat("x", 9)))
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != fiber.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413, got %d", resp.StatusCode)
	}
}
