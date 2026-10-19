package service

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/zu-seguridad/core/internal/model"
	"gorm.io/gorm"
	"time"
)

// ManagementSystemService da vida al ciclo PHVA y al cumplimiento normativo.
//
// El sistema PROPONE evaluaciones de cumplimiento automáticas a partir de los
// datos de los módulos; el profesional SST las VALIDA. El score solo cuenta
// estándares validados por humano: nada legalmente relevante se cierra solo.
type ManagementSystemService struct {
	db *gorm.DB
}

func NewManagementSystemService(db *gorm.DB) *ManagementSystemService {
	return &ManagementSystemService{db: db}
}

// ── Catálogo de estándares por marco normativo ───────────────
// Cada entrada mapea cláusula → fase PHVA → módulo que la satisface.
// Esto es la "matriz de trazabilidad" que pide la auditoría del Ministerio.

type stdTemplate struct {
	Clause, Title, Module string
	Phase                 model.PHVAPhase
	Weight                int
}

var FrameworkStandards = map[model.NormFramework][]stdTemplate{
	// ── ISO 45001:2018 (estructura HLS, cláusulas 4–10) ──
	model.FwISO45001: {
		{"4.1", "Comprensión de la organización y su contexto", "company", model.PhasePlanear, 2},
		{"4.2", "Necesidades y expectativas de los trabajadores y partes interesadas", "company", model.PhasePlanear, 2},
		{"5.1", "Liderazgo y compromiso de la alta dirección", "documents", model.PhasePlanear, 3},
		{"5.4", "Consulta y participación de los trabajadores", "training", model.PhasePlanear, 3},
		{"6.1.2", "Identificación de peligros y evaluación de riesgos (GTC 45)", "risks", model.PhasePlanear, 3},
		{"6.1.3", "Determinación de requisitos legales aplicables", "compliance", model.PhasePlanear, 3},
		{"6.2", "Objetivos de SST y planificación para lograrlos", "management", model.PhasePlanear, 3},
		{"7.2", "Competencia (capacitación)", "training", model.PhaseHacer, 2},
		{"7.5", "Información documentada controlada", "documents", model.PhaseHacer, 2},
		{"8.1", "Planificación y control operacional", "risks", model.PhaseHacer, 3},
		{"8.2", "Preparación y respuesta ante emergencias", "documents", model.PhaseHacer, 3},
		{"9.1", "Seguimiento, medición, análisis y evaluación del desempeño", "management", model.PhaseVerificar, 3},
		{"9.2", "Auditoría interna", "management", model.PhaseVerificar, 3},
		{"9.3", "Revisión por la dirección", "management", model.PhaseVerificar, 3},
		{"10.2", "Incidentes, no conformidades y acciones correctivas", "incidents", model.PhaseActuar, 3},
		{"10.3", "Mejora continua", "management", model.PhaseActuar, 3},
	},
	// ── ISO 9001:2015 (calidad — cláusulas relevantes al SG) ──
	model.FwISO9001: {
		{"4.4", "Sistema de gestión de calidad y sus procesos", "management", model.PhasePlanear, 2},
		{"6.1", "Acciones para abordar riesgos y oportunidades", "risks", model.PhasePlanear, 2},
		{"7.5", "Información documentada", "documents", model.PhaseHacer, 2},
		{"8.5", "Producción y provisión del servicio", "management", model.PhaseHacer, 2},
		{"9.2", "Auditoría interna", "management", model.PhaseVerificar, 2},
		{"10.2", "No conformidad y acción correctiva", "incidents", model.PhaseActuar, 2},
	},
	// ── ISO 14001:2015 (ambiental — para empresas que lo integran) ──
	model.FwISO14001: {
		{"6.1.2", "Aspectos ambientales", "risks", model.PhasePlanear, 2},
		{"6.1.3", "Requisitos legales ambientales", "compliance", model.PhasePlanear, 2},
		{"8.1", "Planificación y control operacional ambiental", "risks", model.PhaseHacer, 2},
		{"8.2", "Preparación y respuesta ante emergencias ambientales", "documents", model.PhaseHacer, 2},
		{"9.1", "Seguimiento y medición ambiental", "management", model.PhaseVerificar, 2},
	},
	// ── ISO 31000:2018 (gestión de riesgos — principios y proceso) ──
	model.FwISO31000: {
		{"5.2", "Liderazgo y compromiso en la gestión del riesgo", "documents", model.PhasePlanear, 2},
		{"6.3", "Alcance, contexto y criterios del riesgo", "risks", model.PhasePlanear, 2},
		{"6.4", "Proceso de evaluación del riesgo", "risks", model.PhaseHacer, 3},
		{"6.5", "Tratamiento del riesgo", "risks", model.PhaseHacer, 3},
		{"6.6", "Seguimiento y revisión del riesgo", "management", model.PhaseVerificar, 2},
	},
	// ── Res. 0312/2019 (estándares mínimos Colombia — núcleo legal) ──
	model.FwRes0312: {
		{"1.1.1", "Asignación de responsable del SG-SST", "company", model.PhasePlanear, 2},
		{"1.1.4", "Afiliación al Sistema de Seguridad Social Integral", "company", model.PhasePlanear, 2},
		{"1.2.1", "Programa de capacitación anual", "training", model.PhasePlanear, 2},
		{"2.1.1", "Política del SG-SST firmada y divulgada", "documents", model.PhasePlanear, 3},
		{"2.2.1", "Objetivos del SG-SST", "management", model.PhasePlanear, 2},
		{"3.1.1", "Evaluación médica ocupacional", "medical", model.PhaseHacer, 3},
		{"4.1.1", "Identificación de peligros y valoración de riesgos", "risks", model.PhaseHacer, 3},
		{"4.2.1", "Medidas de prevención y control", "risks", model.PhaseHacer, 3},
		{"5.1.1", "Plan de prevención, preparación y respuesta ante emergencias", "documents", model.PhaseHacer, 3},
		{"6.1.1", "Indicadores del SG-SST (frecuencia, severidad, etc.)", "management", model.PhaseVerificar, 3},
		{"6.1.2", "Auditoría anual del SG-SST", "management", model.PhaseVerificar, 3},
		{"6.1.3", "Revisión por la alta dirección", "management", model.PhaseVerificar, 2},
		{"7.1.1", "Acciones correctivas y preventivas", "incidents", model.PhaseActuar, 3},
	},
}

// InitializeSystem crea el SG-SST de una empresa para un año, sembrando todos
// los estándares de los marcos seleccionados.
func (s *ManagementSystemService) InitializeSystem(ctx context.Context, tenantID, companyID uuid.UUID, year int, frameworks []model.NormFramework) (*model.ManagementSystem, error) {
	if tenantID == uuid.Nil || companyID == uuid.Nil || year < 2000 || year > 2200 {
		return nil, fmt.Errorf("invalid management system scope")
	}
	if len(frameworks) == 0 {
		return nil, fmt.Errorf("at least one framework is required")
	}
	for _, fw := range frameworks {
		if _, ok := FrameworkStandards[fw]; !ok {
			return nil, fmt.Errorf("unsupported framework")
		}
	}
	sys := &model.ManagementSystem{
		TenantID: tenantID, CompanyID: companyID, Year: year,
		CurrentPhase: model.PhasePlanear, Status: "active",
	}

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Table("companies").Where("id = ? AND tenant_id = ?", companyID, tenantID).Count(&count).Error; err != nil {
			return err
		}
		if count != 1 {
			return gorm.ErrRecordNotFound
		}
		if err := tx.Create(sys).Error; err != nil {
			return err
		}
		for _, fw := range frameworks {
			for _, t := range FrameworkStandards[fw] {
				std := model.ComplianceStandard{
					SystemID: sys.ID, TenantID: tenantID, Framework: fw,
					Clause: t.Clause, Title: t.Title, Phase: t.Phase,
					LinkedModule: t.Module, Status: "pendiente", Weight: t.Weight,
				}
				if err := tx.Create(&std).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return sys, nil
}

// AssessCompliance: el sistema PROPONE el estado de cada estándar a partir de
// los datos reales de los módulos. NO marca cumplimiento definitivo: deja la
// evaluación lista para validación humana.

func (s *ManagementSystemService) AssessCompliance(ctx context.Context, tenantID, companyID, systemID uuid.UUID) (map[string]interface{}, error) {
	var system model.ManagementSystem
	if err := s.db.WithContext(ctx).Where("id = ? AND tenant_id = ? AND company_id = ?", systemID, tenantID, companyID).First(&system).Error; err != nil {
		return nil, err
	}
	var standards []model.ComplianceStandard
	if err := s.db.WithContext(ctx).Where("system_id = ? AND tenant_id = ?", systemID, tenantID).Find(&standards).Error; err != nil {
		return nil, err
	}

	assessed := 0
	for i := range standards {
		if standards[i].HumanValidated {
			continue
		}
		// Do not infer legal compliance from record counts or module presence.
		assessment, proposedStatus := s.evaluateStandard(&standards[i])
		if err := s.db.WithContext(ctx).Model(&standards[i]).Where("id = ? AND tenant_id = ? AND system_id = ?", standards[i].ID, tenantID, systemID).Updates(map[string]interface{}{
			"system_assessment": assessment,
			"status":            proposedStatus,
		}).Error; err != nil {
			return nil, err
		}
		assessed++
	}

	return map[string]interface{}{
		"system_id":          systemID,
		"standards_assessed": assessed,
		"note":               "No se emite dictamen automático. Todo estándar queda pendiente de revisión de aplicabilidad, vigencia y evidencia por el profesional SST.",
	}, nil
}

func (s *ManagementSystemService) evaluateStandard(std *model.ComplianceStandard) (string, string) {
	return fmt.Sprintf("Módulo %q: revisar aplicabilidad, fuente normativa vigente y evidencia suficiente; el sistema no emite dictamen automático", std.LinkedModule), "pendiente"
}

// ValidateStandard: el profesional SST APRUEBA (o corrige) la evaluación.
// Solo lo validado por humano cuenta para el score de cumplimiento legal.
func (s *ManagementSystemService) ValidateStandard(ctx context.Context, tenantID, companyID, standardID, validatedBy uuid.UUID, status string) error {
	if status != "cumple" && status != "no_cumple" && status != "parcial" {
		return fmt.Errorf("invalid standard status")
	}
	var count int64
	if err := s.db.WithContext(ctx).Table("compliance_standards AS cs").
		Joins("JOIN management_systems AS ms ON ms.id = cs.system_id AND ms.tenant_id = cs.tenant_id").
		Where("cs.id = ? AND cs.tenant_id = ? AND ms.company_id = ?", standardID, tenantID, companyID).
		Count(&count).Error; err != nil {
		return err
	}
	if count != 1 {
		return gorm.ErrRecordNotFound
	}
	now := time.Now()
	result := s.db.WithContext(ctx).Model(&model.ComplianceStandard{}).
		Where("id = ? AND tenant_id = ?", standardID, tenantID).
		Updates(map[string]interface{}{
			"status":          status,
			"human_validated": true,
			"validated_by":    validatedBy,
			"validated_at":    now,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// ComplianceScore: solo cuenta estándares VALIDADOS por humano.

func (s *ManagementSystemService) ComplianceScore(ctx context.Context, tenantID, companyID, systemID uuid.UUID) (map[string]interface{}, error) {
	var system model.ManagementSystem
	if err := s.db.WithContext(ctx).Where("id = ? AND tenant_id = ? AND company_id = ?", systemID, tenantID, companyID).First(&system).Error; err != nil {
		return nil, err
	}
	var standards []model.ComplianceStandard
	if err := s.db.WithContext(ctx).Where("system_id = ? AND tenant_id = ?", systemID, tenantID).Find(&standards).Error; err != nil {
		return nil, err
	}

	byFramework := map[model.NormFramework]map[string]int{}
	totalWeight, validatedWeight, compliantWeight, validated := 0, 0, 0, 0

	for _, std := range standards {
		if byFramework[std.Framework] == nil {
			byFramework[std.Framework] = map[string]int{"total": 0, "cumple": 0, "validados": 0}
		}
		byFramework[std.Framework]["total"]++
		totalWeight += std.Weight
		if std.HumanValidated {
			validated++
			validatedWeight += std.Weight
			byFramework[std.Framework]["validados"]++
			if std.Status == "cumple" {
				compliantWeight += std.Weight
				byFramework[std.Framework]["cumple"]++
			}
		}
	}

	var score interface{}
	if validatedWeight > 0 {
		score = float64(compliantWeight) / float64(validatedWeight) * 100
	}
	coverage := 0.0
	if totalWeight > 0 {
		coverage = float64(validatedWeight) / float64(totalWeight) * 100
	}

	return map[string]interface{}{
		"system_id":           systemID,
		"compliance_score":    score,
		"validation_coverage": coverage,
		"total_standards":     len(standards),
		"human_validated":     validated,
		"by_framework":        byFramework,
		"note":                "El score cubre únicamente el peso de estándares validados; la cobertura se informa por separado y no equivale a certificación legal.",
	}, nil
}

// AdvancePhase: avanza el ciclo PHVA. Al cerrar 'Actuar', las acciones
// correctivas marcadas alimentan el plan del siguiente ciclo (cierre del lazo).
func (s *ManagementSystemService) AdvancePhase(ctx context.Context, tenantID, companyID, systemID uuid.UUID) (model.PHVAPhase, error) {
	var sys model.ManagementSystem
	if err := s.db.WithContext(ctx).Where("id = ? AND tenant_id = ? AND company_id = ?", systemID, tenantID, companyID).First(&sys).Error; err != nil {
		return "", err
	}
	next := map[model.PHVAPhase]model.PHVAPhase{
		model.PhasePlanear:   model.PhaseHacer,
		model.PhaseHacer:     model.PhaseVerificar,
		model.PhaseVerificar: model.PhaseActuar,
		model.PhaseActuar:    model.PhasePlanear, // cierra el ciclo y reinicia
	}[sys.CurrentPhase]
	if next == "" {
		return "", fmt.Errorf("invalid management phase")
	}

	result := s.db.WithContext(ctx).Model(&sys).Where("id = ? AND tenant_id = ?", systemID, tenantID).Update("current_phase", next)
	if result.Error != nil {
		return "", result.Error
	}
	if result.RowsAffected != 1 {
		return "", gorm.ErrRecordNotFound
	}
	return next, nil
}

// CloseLoop: crea el SG-SST del siguiente año heredando las lecciones de las
// acciones correctivas del ciclo actual (mejora continua real).
func (s *ManagementSystemService) CloseLoop(ctx context.Context, tenantID, companyID, systemID uuid.UUID) (*model.ManagementSystem, error) {
	var current model.ManagementSystem
	if err := s.db.WithContext(ctx).Where("id = ? AND tenant_id = ? AND company_id = ?", systemID, tenantID, companyID).First(&current).Error; err != nil {
		return nil, err
	}

	// Recoger lecciones de las acciones correctivas marcadas
	var actions []model.CorrectiveAction
	if err := s.db.WithContext(ctx).Where("tenant_id = ? AND system_id = ? AND feeds_next_cycle = true", tenantID, systemID).Find(&actions).Error; err != nil {
		return nil, err
	}
	lessons := model.JSONB{"count": len(actions), "from_year": current.Year}

	frameworks, err := s.frameworksOf(ctx, tenantID, systemID)
	if err != nil {
		return nil, err
	}
	next, err := s.InitializeSystem(ctx, current.TenantID, current.CompanyID, current.Year+1, frameworks)
	if err != nil {
		return nil, err
	}
	result := s.db.WithContext(ctx).Model(next).Where("id = ? AND tenant_id = ?", next.ID, tenantID).Updates(map[string]interface{}{
		"previous_system_id": current.ID,
		"carried_lessons":    lessons,
	})
	if result.Error != nil {
		return nil, result.Error
	}
	return next, nil
}

func (s *ManagementSystemService) frameworksOf(ctx context.Context, tenantID, systemID uuid.UUID) ([]model.NormFramework, error) {
	var rows []model.ComplianceStandard
	if err := s.db.WithContext(ctx).Where("tenant_id = ? AND system_id = ?", tenantID, systemID).Distinct("framework").Find(&rows).Error; err != nil {
		return nil, err
	}
	seen := map[model.NormFramework]bool{}
	var out []model.NormFramework
	for _, r := range rows {
		if !seen[r.Framework] {
			seen[r.Framework] = true
			out = append(out, r.Framework)
		}
	}
	return out, nil
}
