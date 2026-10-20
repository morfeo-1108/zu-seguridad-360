package middleware

import (
	"fmt"

	"github.com/casbin/casbin/v2"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func NewCasbinEnforcer(db *gorm.DB) (*casbin.Enforcer, error) {
	if db == nil {
		return nil, fmt.Errorf("casbin database is required")
	}
	adapter, err := gormadapter.NewAdapterByDB(db)
	if err != nil {
		return nil, fmt.Errorf("initialize casbin adapter: %w", err)
	}
	enforcer, err := casbin.NewEnforcer("config/rbac_model.conf", adapter)
	if err != nil {
		return nil, fmt.Errorf("initialize casbin enforcer: %w", err)
	}
	if err := LoadDefaultPolicies(enforcer); err != nil {
		return nil, err
	}
	return enforcer, nil
}

func LoadDefaultPolicies(e *casbin.Enforcer) error {
	if e == nil {
		return fmt.Errorf("casbin enforcer is required")
	}
	add := func(sub, obj, act string) error {
		if _, err := e.AddPolicy(sub, obj, act); err != nil {
			return fmt.Errorf("load policy %s %s %s: %w", sub, obj, act, err)
		}
		return nil
	}
	// These roles currently lack enforced company-level membership checks.
	// Remove legacy persisted grants on every boot; restore them only with
	// object-scope middleware and cross-company tests in place.
	for _, role := range []string{"employer", "worker", "ips", "auditor", "contractor"} {
		if _, err := e.RemoveFilteredPolicy(0, role); err != nil {
			return fmt.Errorf("remove unsafe legacy policies for %s: %w", role, err)
		}
	}
	// admin puede todo
	if err := add("admin", "*", "*"); err != nil {
		return err
	}
	// professional_sst
	for _, p := range [][3]string{
		{"professional_sst", "companies", "read"}, {"professional_sst", "companies", "write"},
		{"professional_sst", "employees", "read"}, {"professional_sst", "employees", "write"},
		{"professional_sst", "incidents", "read"}, {"professional_sst", "incidents", "write"},
		{"professional_sst", "documents", "read"}, {"professional_sst", "documents", "write"},
		{"professional_sst", "company_export", "read"}, {"professional_sst", "system", "read"},
	} {
		if err := add(p[0], p[1], p[2]); err != nil {
			return err
		}
	}
	for _, resource := range []string{"medical_exams", "incidents", "trainings", "projects", "evidence", "management", "regulatory_read", "risks", "epp", "inspections"} {
		if err := add("professional_sst", resource, "read"); err != nil {
			return err
		}
		if err := add("professional_sst", resource, "write"); err != nil {
			return err
		}
	}
	for _, p := range [][3]string{
		{"admin", "company_export", "read"}, {"admin", "system", "read"},
		{"admin", "regulatory_review", "write"},
	} {
		if err := add(p[0], p[1], p[2]); err != nil {
			return err
		}
	}
	if err := e.SavePolicy(); err != nil {
		return fmt.Errorf("persist casbin policies: %w", err)
	}
	return nil
}

func Authorize(enforcer *casbin.Enforcer, resource string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		role, ok := c.Locals("role").(string)
		if !ok || role == "" {
			return fiber.NewError(fiber.StatusForbidden, "role not found")
		}

		action := "read"
		if c.Method() != "GET" {
			action = "write"
		}

		allowed, err := enforcer.Enforce(role, resource, action)
		if err != nil || !allowed {
			return fiber.NewError(fiber.StatusForbidden, "insufficient permissions")
		}

		return c.Next()
	}
}
