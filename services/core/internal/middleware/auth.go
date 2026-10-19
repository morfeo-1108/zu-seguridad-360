package middleware

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/zu-seguridad/core/config"
)

// JWTProtected accepts only HS256 access tokens with a valid subject, tenant,
// role and expiry. Refresh tokens must never authorize API requests.
func JWTProtected(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			return fiber.NewError(fiber.StatusUnauthorized, "missing or invalid authorization header")
		}

		tokenStr := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
		if tokenStr == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "missing or invalid authorization header")
		}

		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			if t.Method != jwt.SigningMethodHS256 {
				return nil, fiber.NewError(fiber.StatusUnauthorized, "invalid signing method")
			}
			return []byte(cfg.JWTSecret), nil
		}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
		if err != nil || token == nil || !token.Valid {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid or expired token")
		}

		expiresAt, err := claims.GetExpirationTime()
		if err != nil || expiresAt == nil || !expiresAt.Time.After(time.Now()) {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid or expired token")
		}

		userID, userOK := claims["user_id"].(string)
		tenantID, tenantOK := claims["tenant_id"].(string)
		role, roleOK := claims["role"].(string)
		tokenType, typeOK := claims["token_type"].(string)
		subject, subjectOK := claims["sub"].(string)
		if !userOK || !tenantOK || !roleOK || !typeOK || !subjectOK || tokenType != "access" || role == "" || subject != userID {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid token claims")
		}
		if _, err := uuid.Parse(userID); err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid token claims")
		}
		if _, err := uuid.Parse(tenantID); err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid token claims")
		}
		companyID, hasCompanyID := claims["company_id"]
		if hasCompanyID {
			companyIDString, ok := companyID.(string)
			if !ok || companyIDString == "" {
				return fiber.NewError(fiber.StatusUnauthorized, "invalid token claims")
			}
			parsedCompanyID, err := uuid.Parse(companyIDString)
			if err != nil || parsedCompanyID == uuid.Nil {
				return fiber.NewError(fiber.StatusUnauthorized, "invalid token claims")
			}
			c.Locals("company_id", parsedCompanyID.String())
		}

		c.Locals("user_id", userID)
		c.Locals("tenant_id", tenantID)
		c.Locals("role", role)
		if email, ok := claims["email"].(string); ok {
			c.Locals("email", email)
		}
		return c.Next()
	}
}

// TenantExtractor deliberately ignores X-Tenant-ID. JWTProtected supplies the
// tenant from a signed claim; an untrusted request header must not set scope.
func TenantExtractor() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.Next()
	}
}
