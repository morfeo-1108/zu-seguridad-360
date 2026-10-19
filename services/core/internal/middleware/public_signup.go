package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

// RequireAllowedOrigin is an additional CSRF guard for credential-bearing
// browser endpoints. It rejects absent, wildcard, malformed and unlisted origins.
func RequireAllowedOrigin(allowlist string) fiber.Handler {
	allowed := make(map[string]struct{})
	for _, item := range strings.Split(allowlist, ",") {
		origin := strings.TrimRight(strings.TrimSpace(item), "/")
		if origin != "" && origin != "*" {
			allowed[origin] = struct{}{}
		}
	}
	return func(c *fiber.Ctx) error {
		origin := strings.TrimRight(strings.TrimSpace(c.Get("Origin")), "/")
		if origin == "" {
			return fiber.NewError(fiber.StatusForbidden, "request origin is not allowed")
		}
		if _, ok := allowed[origin]; !ok {
			return fiber.NewError(fiber.StatusForbidden, "request origin is not allowed")
		}
		return c.Next()
	}
}

// RequestBodyLimit applies a smaller cap to public authentication payloads than
// the platform-wide upload limit. Fiber's BodyLimit still provides the hard cap.
func RequestBodyLimit(limit int) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if limit <= 0 || c.Request().Header.ContentLength() > limit || len(c.Body()) > limit {
			return fiber.NewError(fiber.StatusRequestEntityTooLarge, "request body too large")
		}
		return c.Next()
	}
}
