package middleware

import (
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/zu-seguridad/core/config"
)

func signedTestToken(t *testing.T, secret, tokenType string) string {
	return signedRoleToken(t, secret, tokenType, "professional_sst")
}

func signedRoleToken(t *testing.T, secret, tokenType, role string) string {
	t.Helper()
	userID := uuid.NewString()
	claims := jwt.MapClaims{
		"user_id":    userID,
		"sub":        userID,
		"tenant_id":  uuid.NewString(),
		"role":       role,
		"token_type": tokenType,
		"exp":        time.Now().Add(time.Minute).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("sign test token: %v", err)
	}
	return signed
}

func signedCompanyToken(t *testing.T, secret, role, tenantID, companyID string) string {
	t.Helper()
	userID := uuid.NewString()
	claims := jwt.MapClaims{
		"user_id":    userID,
		"sub":        userID,
		"tenant_id":  tenantID,
		"company_id": companyID,
		"role":       role,
		"token_type": "access",
		"exp":        time.Now().Add(time.Minute).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("sign scoped test token: %v", err)
	}
	return signed
}

func TestJWTProtectedAcceptsAccessToken(t *testing.T) {
	secret := "unit-test-secret"
	app := fiber.New()
	app.Get("/protected", JWTProtected(&config.Config{JWTSecret: secret}), func(c *fiber.Ctx) error {
		return c.SendString(fmt.Sprint(c.Locals("tenant_id")))
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+signedTestToken(t, secret, "access"))
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200 for access token, got %d", resp.StatusCode)
	}
}

func TestJWTProtectedRejectsRefreshToken(t *testing.T) {
	secret := "unit-test-secret"
	app := fiber.New()
	app.Get("/protected", JWTProtected(&config.Config{JWTSecret: secret}), func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+signedTestToken(t, secret, "refresh"))
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("expected 401 for refresh token, got %d", resp.StatusCode)
	}
}

func TestTenantExtractorIgnoresUntrustedHeader(t *testing.T) {
	app := fiber.New()
	app.Get("/tenant", TenantExtractor(), func(c *fiber.Ctx) error {
		if c.Locals("tenant_id") != nil {
			return c.SendStatus(fiber.StatusForbidden)
		}
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/tenant", nil)
	req.Header.Set("X-Tenant-ID", uuid.NewString())
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestJWTProtectedUsesSignedCompanyClaimNotHeader(t *testing.T) {
	secret := "unit-test-secret"
	tenantA, companyA, companyB := uuid.NewString(), uuid.NewString(), uuid.NewString()
	app := fiber.New()
	app.Get("/companies/:id", JWTProtected(&config.Config{JWTSecret: secret}), CompanyScope("id"), func(c *fiber.Ctx) error {
		return c.SendString(c.Locals("company_id").(string))
	})

	req := httptest.NewRequest("GET", "/companies/"+companyA, nil)
	req.Header.Set("Authorization", "Bearer "+signedCompanyToken(t, secret, "employer", tenantA, companyA))
	req.Header.Set("X-Company-ID", companyB)
	req.Header.Set("X-Tenant-ID", uuid.NewString())
	req.URL.RawQuery = "company_id=" + companyB
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected scoped request 200, got %d", resp.StatusCode)
	}
}

func TestCompanyScopeIgnoresHeaderQueryAndBody(t *testing.T) {
	secret := "unit-test-secret"
	tenantA, companyA, companyB := uuid.NewString(), uuid.NewString(), uuid.NewString()
	app := fiber.New()
	app.Post("/companies/:id", JWTProtected(&config.Config{JWTSecret: secret}), CompanyScope("id"), func(c *fiber.Ctx) error {
		return c.SendString(c.Locals("company_id").(string))
	})

	req := httptest.NewRequest("POST", "/companies/"+companyA, strings.NewReader(`{"company_id":"`+companyB+`"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+signedCompanyToken(t, secret, "professional_sst", tenantA, companyA))
	req.Header.Set("X-Company-ID", companyB)
	req.URL.RawQuery = "company_id=" + companyB
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected request scoped to signed company A, got %d", resp.StatusCode)
	}
}

func TestCompanyScopeRejectsCompanyATokenAccessingCompanyB(t *testing.T) {
	secret := "unit-test-secret"
	tenantA, companyA, companyB := uuid.NewString(), uuid.NewString(), uuid.NewString()
	app := fiber.New()
	app.Get("/companies/:id", JWTProtected(&config.Config{JWTSecret: secret}), CompanyScope("id"), func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/companies/"+companyB, nil)
	req.Header.Set("Authorization", "Bearer "+signedCompanyToken(t, secret, "employer", tenantA, companyA))
	req.Header.Set("X-Company-ID", companyB)
	req.URL.RawQuery = "company_id=" + companyB
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != fiber.StatusForbidden {
		t.Fatalf("expected 403 for company A → company B, got %d", resp.StatusCode)
	}
}

func TestCompanyScopeRequiresCompanyClaimForEmployerAndProfessional(t *testing.T) {
	secret := "unit-test-secret"
	for _, role := range []string{"employer", "professional_sst"} {
		app := fiber.New()
		app.Get("/companies/:id", JWTProtected(&config.Config{JWTSecret: secret}), CompanyScope("id"), func(c *fiber.Ctx) error {
			return c.SendStatus(fiber.StatusOK)
		})
		req := httptest.NewRequest("GET", "/companies/"+uuid.NewString(), nil)
		req.Header.Set("Authorization", "Bearer "+signedRoleToken(t, secret, "access", role))
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("app.Test for role %s: %v", role, err)
		}
		resp.Body.Close()
		if resp.StatusCode != fiber.StatusForbidden {
			t.Fatalf("expected 403 for %s token without company claim, got %d", role, resp.StatusCode)
		}
	}
}
