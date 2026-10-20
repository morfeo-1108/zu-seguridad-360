package middleware

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/zu-seguridad/attention/config"
)

// JWTProtected mirrors services/core/internal/middleware/auth.go: Core
// issues the token, Attention only verifies it with the same shared secret.
// Before this, every route here trusted whatever tenantId the caller put in
// the URL (or an X-Tenant-ID header) — anyone could read or acknowledge any
// tenant's inbox by changing a path segment. tenant_id now only ever comes
// from this verified claim.
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

		tenantID, tenantOK := claims["tenant_id"].(string)
		tokenType, typeOK := claims["token_type"].(string)
		if !tenantOK || !typeOK || tokenType != "access" {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid token claims")
		}
		if _, err := uuid.Parse(tenantID); err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid token claims")
		}

		c.Locals("tenant_id", tenantID)
		return c.Next()
	}
}
