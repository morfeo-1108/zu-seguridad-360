package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/zu-seguridad/core/config"
)

func TestSensitiveDataGateIsClosedByDefault(t *testing.T) {
	app := fiber.New()
	app.Post("/sensitive", SensitiveDataProcessingDisabled(), func(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusOK) })
	resp, err := app.Test(httptest.NewRequest("POST", "/sensitive", nil))
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != fiber.StatusNotFound {
		t.Fatalf("default must stay closed: got %d", resp.StatusCode)
	}
}

func TestSyntheticDataGateRequiresServerConfiguration(t *testing.T) {
	for _, tc := range []struct {
		name    string
		enabled bool
		want    int
	}{
		{"disabled", false, fiber.StatusNotFound},
		{"synthetic enabled", true, fiber.StatusCreated},
	} {
		t.Run(tc.name, func(t *testing.T) {
			app := fiber.New()
			app.Post("/sensitive", SensitiveDataProcessingGate(tc.enabled), func(c *fiber.Ctx) error {
				if tc.enabled && c.Locals("synthetic_data_mode") != true {
					t.Fatal("server-approved synthetic mode was not set in request locals")
				}
				return c.SendStatus(fiber.StatusCreated)
			})
			req := httptest.NewRequest("POST", "/sensitive?synthetic=true", nil)
			req.Header.Set("X-Pilot-Synthetic", "true")
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

func TestSensitiveDataGatePreservesCrossCompany403(t *testing.T) {
	const secret = "unit-test-secret"
	tenantID, companyA, companyB := uuid.NewString(), uuid.NewString(), uuid.NewString()
	app := fiber.New()
	app.Get("/companies/:companyId/employees", JWTProtected(&config.Config{JWTSecret: secret}), CompanyScope("companyId"), SensitiveDataProcessingGate(true), func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})
	request := func(companyID string) int {
		req := httptest.NewRequest("GET", "/companies/"+companyID+"/employees", nil)
		req.Header.Set("Authorization", "Bearer "+signedCompanyToken(t, secret, "professional_sst", tenantID, companyA))
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("app.Test: %v", err)
		}
		defer resp.Body.Close()
		return resp.StatusCode
	}
	if got := request(companyA); got != fiber.StatusOK {
		t.Fatalf("authorized company should reach synthetic handler; expected %d, got %d", fiber.StatusOK, got)
	}
	if got := request(companyB); got != fiber.StatusForbidden {
		t.Fatalf("cross-company request must be rejected before the synthetic gate; expected 403, got %d", got)
	}
}
