package middleware

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/zu-seguridad/billing/config"
)

// JWTProtected validates Core's own HS256 access token (same shared secret
// and claim shape) so billing endpoints are not reachable without a real
// authenticated session, even when the caller is Core forwarding a request
// on a professional's behalf.
func JWTProtected(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			return fiber.NewError(fiber.StatusUnauthorized, "missing or invalid authorization header")
		}
		tokenStr := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
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
		tenantID, ok := claims["tenant_id"].(string)
		if !ok || tenantID == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid token claims")
		}
		c.Locals("tenant_id", tenantID)
		return c.Next()
	}
}

// TenantScope rejects any request whose path tenantId does not match the
// token's own tenant_id claim — a professional may only query their own
// subscription/quota, never another tenant's.
func TenantScope(paramName string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claim, _ := c.Locals("tenant_id").(string)
		if claim == "" || claim != c.Params(paramName) {
			return fiber.NewError(fiber.StatusForbidden, "tenant scope mismatch")
		}
		return c.Next()
	}
}
