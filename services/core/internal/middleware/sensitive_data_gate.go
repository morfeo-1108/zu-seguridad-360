package middleware

import "github.com/gofiber/fiber/v2"

// SensitiveDataProcessingGate is enabled only by validated server-side synthetic
// pilot configuration. It never reads a client header, query parameter, or body.
func SensitiveDataProcessingGate(syntheticPilotEnabled bool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if !syntheticPilotEnabled {
			return fiber.NewError(fiber.StatusNotFound, "resource unavailable")
		}
		c.Locals("synthetic_data_mode", true)
		return c.Next()
	}
}

// SensitiveDataProcessingDisabled preserves a fail-closed default for routes
// that do not explicitly opt into the validated synthetic pilot configuration.
func SensitiveDataProcessingDisabled() fiber.Handler {
	return SensitiveDataProcessingGate(false)
}
