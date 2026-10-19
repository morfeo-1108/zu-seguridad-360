package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// CompanyScope compares a company resource path parameter with the company_id
// claim populated by JWTProtected. It never reads company scope from headers,
// query parameters, or request bodies. Roles without an explicit company claim
// are denied unless they are tenant-wide professional/admin roles; those roles
// still rely on tenant-scoped queries in every handler/service.
func CompanyScope(paramName string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		role, _ := c.Locals("role").(string)
		claim, hasCompany := c.Locals("company_id").(string)
		if !hasCompany || claim == "" {
			if requiresCompanyClaim(role) {
				return fiber.NewError(fiber.StatusForbidden, "company scope required")
			}
			return c.Next()
		}

		claimID, err := uuid.Parse(claim)
		if err != nil || claimID == uuid.Nil {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid company scope")
		}

		resourceIDText := c.Params(paramName)
		if resourceIDText == "" {
			// Collection handlers must filter their result by this signed claim;
			// an unscoped mutation (for example POST /companies) is not allowed.
			if c.Method() != fiber.MethodGet {
				return fiber.NewError(fiber.StatusForbidden, "company scope required for mutation")
			}
			return c.Next()
		}
		resourceID, err := uuid.Parse(resourceIDText)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid company id")
		}
		if resourceID != claimID {
			return fiber.NewError(fiber.StatusForbidden, "company scope mismatch")
		}
		return c.Next()
	}
}

func requiresCompanyClaim(role string) bool {
	switch role {
	case "professional_sst", "employer", "worker", "ips", "auditor", "contractor":
		return true
	default:
		// Only tenant administrators may use tenant-wide access. All other
		// roles need an assigned company claim until membership is implemented.
		return false
	}
}
