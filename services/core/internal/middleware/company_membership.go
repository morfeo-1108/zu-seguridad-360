package middleware

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// ActiveCompanyMembershipChecker is implemented by the user repository. The
// middleware verifies membership at request time so revocation takes effect
// without waiting for an access token to expire.
type ActiveCompanyMembershipChecker interface {
	HasActiveCompanyMembership(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) (bool, error)
}

// CompanyMembershipScope requires the resource path to match the active,
// verified company_id claim and then confirms the user's persisted membership.
// It never trusts company IDs from headers, query strings, or request bodies.
func CompanyMembershipScope(paramName string, checker ActiveCompanyMembershipChecker) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if checker == nil {
			return fiber.NewError(fiber.StatusForbidden, "company membership unavailable")
		}
		userID, userOK := parseScopeID(c.Locals("user_id"))
		tenantID, tenantOK := parseScopeID(c.Locals("tenant_id"))
		claimID, companyOK := parseScopeID(c.Locals("company_id"))
		if !userOK || !tenantOK || !companyOK {
			return fiber.NewError(fiber.StatusForbidden, "company scope required")
		}

		resourceText := c.Params(paramName)
		if resourceText == "" {
			if c.Method() != fiber.MethodGet {
				return fiber.NewError(fiber.StatusForbidden, "company scope required for mutation")
			}
		} else {
			resourceID, ok := parseScopeID(resourceText)
			if !ok {
				return fiber.NewError(fiber.StatusBadRequest, "invalid company id")
			}
			if resourceID != claimID {
				return fiber.NewError(fiber.StatusForbidden, "company scope mismatch")
			}
		}

		allowed, err := checker.HasActiveCompanyMembership(c.Context(), userID, tenantID, claimID)
		if err != nil || !allowed {
			return fiber.NewError(fiber.StatusForbidden, "company membership required")
		}
		return c.Next()
	}
}

func parseScopeID(value interface{}) (uuid.UUID, bool) {
	text, ok := value.(string)
	if !ok || text == "" {
		return uuid.Nil, false
	}
	id, err := uuid.Parse(text)
	return id, err == nil && id != uuid.Nil
}
