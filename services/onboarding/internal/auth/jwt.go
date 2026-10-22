// Package auth validates Core's HS256 access tokens with the same rules as
// services/core/internal/middleware/auth.go. Scope (tenant, company, user)
// always comes from the signed claims, never from headers, query or body.
package auth

import (
	"crypto/subtle"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Identity struct {
	UserID, TenantID, CompanyID, Role, Email string
}

func JWT(secret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		h := c.Get("Authorization")
		if !strings.HasPrefix(h, "Bearer ") {
			return fiber.NewError(fiber.StatusUnauthorized, "missing or invalid authorization header")
		}
		id, err := Parse(strings.TrimSpace(strings.TrimPrefix(h, "Bearer ")), secret)
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid or expired token")
		}
		c.Locals("identity", id)
		return c.Next()
	}
}

func Parse(tokenStr, secret string) (*Identity, error) {
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}), jwt.WithExpirationRequired())
	if err != nil || !token.Valid {
		return nil, fiber.ErrUnauthorized
	}
	exp, err := claims.GetExpirationTime()
	if err != nil || exp == nil || !exp.Time.After(time.Now()) {
		return nil, fiber.ErrUnauthorized
	}
	str := func(k string) string { v, _ := claims[k].(string); return v }
	id := &Identity{UserID: str("user_id"), TenantID: str("tenant_id"), Role: str("role"), Email: str("email")}
	if str("token_type") != "access" || id.Role == "" || str("sub") != id.UserID {
		return nil, fiber.ErrUnauthorized
	}
	if _, err := uuid.Parse(id.UserID); err != nil {
		return nil, fiber.ErrUnauthorized
	}
	if _, err := uuid.Parse(id.TenantID); err != nil {
		return nil, fiber.ErrUnauthorized
	}
	if raw, ok := claims["company_id"]; ok {
		s, _ := raw.(string)
		u, err := uuid.Parse(s)
		if err != nil || u == uuid.Nil {
			return nil, fiber.ErrUnauthorized
		}
		id.CompanyID = u.String()
	}
	return id, nil
}

func From(c *fiber.Ctx) *Identity {
	id, _ := c.Locals("identity").(*Identity)
	return id
}

// Internal guards service-to-service endpoints with a shared token. An empty
// configured token disables the endpoint instead of leaving it open.
func Internal(token string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if token == "" {
			return fiber.NewError(fiber.StatusServiceUnavailable, "internal endpoint disabled (INTERNAL_API_TOKEN not configured)")
		}
		if subtle.ConstantTimeCompare([]byte(c.Get("X-Internal-Token")), []byte(token)) != 1 {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid internal token")
		}
		return c.Next()
	}
}
