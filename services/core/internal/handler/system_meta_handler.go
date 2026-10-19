package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/zu-seguridad/core/internal/model"
)

// SystemMetaHandler expone metadata verificable del sistema: los roles RBAC,
// la estrategia multi-tenant y las capacidades. Responde a las críticas de
// auditoría de tipo "no verificable" haciendo la información consultable.
type SystemMetaHandler struct{}

func NewSystemMetaHandler() *SystemMetaHandler { return &SystemMetaHandler{} }

// GET /v1/system/roles — los 7 roles RBAC con su descripción (verificable)
func (h *SystemMetaHandler) Roles(c *fiber.Ctx) error {
	roles := []fiber.Map{
		{"role": model.RoleAdmin, "name": "Administrador", "scope": "plataforma", "status": "enabled_global_privilege", "description": "Gestión global; cuenta altamente privilegiada"},
		{"role": model.RoleProfSST, "name": "Profesional SST", "scope": "portafolio", "status": "enabled; tenant/object coverage still under validation", "description": "Opera el portafolio asociado al tenant"},
		{"role": model.RoleEmployer, "name": "Empleador", "scope": "empresa", "status": "disabled_pending_company_membership", "description": "Sin acceso hasta implementar y probar pertenencia a empresa"},
		{"role": model.RoleWorker, "name": "Trabajador", "scope": "individual", "status": "disabled_pending_company_membership", "description": "Sin acceso hasta implementar y probar pertenencia individual/empresa"},
		{"role": model.RoleIPS, "name": "IPS", "scope": "médico", "status": "disabled_pending_contract_and_membership", "description": "Sin acceso a información médica desde las rutas actuales"},
		{"role": model.RoleAuditor, "name": "Auditor", "scope": "lectura", "status": "disabled_pending_scoped_access", "description": "Sin acceso hasta delimitar alcance y empresa"},
		{"role": model.RoleContractor, "name": "Contratista", "scope": "empresa", "status": "disabled_pending_company_membership", "description": "Sin acceso hasta implementar pertenencia y alcance documental"},
	}
	return c.JSON(fiber.Map{"roles": roles, "total": len(roles), "rbac_engine": "Casbin"})
}

// GET /v1/system/tenancy — la estrategia multi-tenant, explícita
func (h *SystemMetaHandler) Tenancy(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"model":             "shared-schema",
		"engine":            "PostgreSQL",
		"isolation":         "filtros por tenant_id en aplicación; la cobertura debe verificarse por endpoint y objeto",
		"tenant_extraction": "claim tenant_id de JWT firmado; se ignoran headers tenant no confiables",
		"database_rls":      "not_enabled_or_not_verified",
		"status":            "partial; no representa una garantía de aislamiento",
	})
}

// GET /v1/system/capabilities — capacidades y su estado de madurez (honesto)
func (h *SystemMetaHandler) Capabilities(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status": "partial; capabilities require independent build, authorization and operational verification",
		"modules": fiber.Map{
			"core_sst":                        "in_development",
			"employee_data_api":               "disabled_pending_privacy_consent_and_retention_approval",
			"incident_training_api":           "disabled_pending_privacy_and_retention_review",
			"billing":                         "health_only; financial_routes_and_cron_disabled",
			"ai_and_vision":                   "disabled_until_reviewed",
			"biometrics":                      "disabled_until_reviewed",
			"drones":                          "disabled_until_hardware_and_safety_validation",
			"documents_generation_api":        "disabled_pending_auth_scope_and_privacy_review",
			"safety_protocols":                "disabled_until_normative_validation",
			"public_certificate_verification": "disabled_pending_privacy_review",
			"data_export":                     "disabled_until_sensitive_data_review",
			"medical_exam_endpoints":          "disabled_pending_consent_and_company_authorization",
			"consent_capture":                 "disabled_until_evidence_and_policy_review",
			"retention_automation":            "disabled_no_approved_category_schedule",
			"employee_data_erasure":           "disabled_pending_retention_and_schema_review",
		},
	})
}
