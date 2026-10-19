package middleware

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/zu-seguridad/core/config"
)

type fakeMembershipChecker struct {
	allowedCompany uuid.UUID
	active         bool
}

func (f fakeMembershipChecker) HasActiveCompanyMembership(_ context.Context, _, _, companyID uuid.UUID) (bool, error) {
	return f.active && companyID == f.allowedCompany, nil
}

func TestCompanyMembershipScopeRequiresPersistedMembershipAndSignedCompany(t *testing.T) {
	const secret = "unit-test-secret"
	tenantID, companyA, companyB := uuid.New(), uuid.New(), uuid.New()
	checker := fakeMembershipChecker{allowedCompany: companyA, active: true}
	app := fiber.New()
	app.Get("/companies/:companyId/resources", JWTProtected(&config.Config{JWTSecret: secret}), CompanyMembershipScope("companyId", checker), func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})
	request := func(companyID string) int {
		req := httptest.NewRequest("GET", "/companies/"+companyID+"/resources", nil)
		req.Header.Set("Authorization", "Bearer "+signedCompanyToken(t, secret, "professional_sst", tenantID.String(), companyA.String()))
		req.Header.Set("X-Company-ID", companyB.String())
		req.URL.RawQuery = "company_id=" + companyB.String()
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("app.Test: %v", err)
		}
		defer resp.Body.Close()
		return resp.StatusCode
	}
	if got := request(companyA.String()); got != fiber.StatusOK {
		t.Fatalf("active signed membership should allow company A, got %d", got)
	}
	if got := request(companyB.String()); got != fiber.StatusForbidden {
		t.Fatalf("company A token must not access company B even with injected header/query, got %d", got)
	}

	checker.active = false
	appRevoked := fiber.New()
	appRevoked.Get("/companies/:companyId/resources", JWTProtected(&config.Config{JWTSecret: secret}), CompanyMembershipScope("companyId", checker), func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})
	req := httptest.NewRequest("GET", "/companies/"+companyA.String()+"/resources", nil)
	req.Header.Set("Authorization", "Bearer "+signedCompanyToken(t, secret, "professional_sst", tenantID.String(), companyA.String()))
	resp, err := appRevoked.Test(req)
	if err != nil {
		t.Fatalf("app.Test revoked: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != fiber.StatusForbidden {
		t.Fatalf("revoked membership must deny the next request, got %d", resp.StatusCode)
	}
}
