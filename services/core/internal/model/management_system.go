package model

import (
	"github.com/google/uuid"
	"time"
)

// ── DOMINIO DEL SISTEMA DE GESTIÓN (columna vertebral PHVA) ──
//
// El SG-SST no es un conjunto de tareas sueltas: es un sistema de gestión que
// vive en el ciclo Planear-Hacer-Verificar-Actuar (PHVA) y se somete a normas
// con estructura de alto nivel (HLS) compartida. Este dominio es el CONTENEDOR
// que organiza todos los módulos existentes como ciclo de mejora continua.
//
// Normas modeladas (versiones vigentes verificadas):
//   ISO 45001:2018 (SST) · ISO 9001:2015 (calidad) · ISO 14001:2015 (ambiental)
//   ISO 31000:2018 (gestión de riesgos) · Res. 0312/2019 (estándares mínimos Colombia)
//
// Principio de diseño: el sistema PROPONE, ORGANIZA y GENERA; el profesional SST
// APRUEBA y FIRMA. Nada legalmente vinculante se cierra sin validación humana.

// PHVAPhase: las cuatro fases del ciclo.
type PHVAPhase string

const (
	PhasePlanear   PHVAPhase = "planear"   // política, objetivos, plan anual, requisitos legales
	PhaseHacer     PHVAPhase = "hacer"     // ejecución: capacitaciones, EPP, controles
	PhaseVerificar PHVAPhase = "verificar" // auditorías, indicadores vs metas, revisión por dirección
	PhaseActuar    PHVAPhase = "actuar"    // acciones correctivas, gestión del cambio, mejora
)

// ManagementSystem es el SG-SST de UNA empresa para UN ciclo anual.
// Es el agregado raíz que ata todos los módulos al ciclo PHVA.
type ManagementSystem struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantID  uuid.UUID `gorm:"type:uuid;index" json:"tenant_id"`
	CompanyID uuid.UUID `gorm:"type:uuid;index" json:"company_id"`

	Year         int       `json:"year"` // ciclo anual del SG-SST
	CurrentPhase PHVAPhase `gorm:"size:20" json:"current_phase"`
	Status       string    `gorm:"size:30;default:'active'" json:"status"`

	// Estándares (cumplimiento del marco normativo) y elementos del ciclo
	Standards  []ComplianceStandard `gorm:"foreignKey:SystemID" json:"standards,omitempty"`
	Objectives []SSTObjective       `gorm:"foreignKey:SystemID" json:"objectives,omitempty"`

	// Cierre del lazo: lecciones del ciclo anterior incorporadas a este
	PreviousSystemID *uuid.UUID `gorm:"type:uuid" json:"previous_system_id"`
	CarriedLessons   JSONB      `gorm:"type:jsonb" json:"carried_lessons"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (ManagementSystem) TableName() string { return "management_systems" }

// NormFramework: marco normativo al que pertenece un requisito.
type NormFramework string

const (
	FwISO45001 NormFramework = "iso_45001_2018"
	FwISO9001  NormFramework = "iso_9001_2015"
	FwISO14001 NormFramework = "iso_14001_2015"
	FwISO31000 NormFramework = "iso_31000_2018"
	FwRes0312  NormFramework = "res_0312_2019"
)

// ComplianceStandard: un requisito/estándar de una norma, ligado al módulo que
// lo satisface y a su evidencia. Es el modelo de cumplimiento VIVO.
type ComplianceStandard struct {
	ID       uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	SystemID uuid.UUID `gorm:"type:uuid;index" json:"system_id"`
	TenantID uuid.UUID `gorm:"type:uuid;index" json:"tenant_id"`

	Framework NormFramework `gorm:"size:30;index" json:"framework"`
	Clause    string        `gorm:"size:30" json:"clause"` // ej: "6.1.2", "Estándar 1.1.1"
	Title     string        `gorm:"type:text" json:"title"`
	Phase     PHVAPhase     `gorm:"size:20" json:"phase"` // a qué fase del ciclo pertenece

	// Trazabilidad: qué módulo lo satisface y con qué evidencia
	LinkedModule string `gorm:"size:50" json:"linked_module"` // "risks", "training", "incidents"...
	EvidenceRef  string `gorm:"type:text" json:"evidence_ref"`

	// Estado de cumplimiento (lo propone el sistema, lo valida el profesional)
	Status           string     `gorm:"size:30;default:'pendiente'" json:"status"` // pendiente|cumple|no_cumple|parcial
	SystemAssessment string     `gorm:"type:text" json:"system_assessment"`        // evaluación automática
	HumanValidated   bool       `gorm:"default:false" json:"human_validated"`      // ← aprobación humana obligatoria
	ValidatedBy      *uuid.UUID `gorm:"type:uuid" json:"validated_by"`
	ValidatedAt      *time.Time `json:"validated_at"`

	Weight    int       `json:"weight"` // ponderación para el score
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (ComplianceStandard) TableName() string { return "compliance_standards" }

// SSTObjective: objetivo medible del SG-SST (fase Planear).
type SSTObjective struct {
	ID       uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	SystemID uuid.UUID `gorm:"type:uuid;index" json:"system_id"`
	TenantID uuid.UUID `gorm:"type:uuid;index" json:"tenant_id"`

	Description string     `gorm:"type:text" json:"description"`
	Indicator   string     `gorm:"size:120" json:"indicator"` // IF, IS, cobertura, etc.
	Target      float64    `json:"target"`                    // meta planeada
	Current     float64    `json:"current"`                   // valor actual (de los módulos)
	Unit        string     `gorm:"size:30" json:"unit"`
	Deadline    *time.Time `json:"deadline"`
	Status      string     `gorm:"size:30;default:'en_progreso'" json:"status"`

	CreatedAt time.Time `json:"created_at"`
}

func (SSTObjective) TableName() string { return "sst_objectives" }

// CorrectiveAction: el cierre del lazo PHVA (fase Actuar).
// Un hallazgo (auditoría, incidente, inspección) genera una acción correctiva;
// su cierre retroalimenta el plan del siguiente ciclo.
type CorrectiveAction struct {
	ID       uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	SystemID uuid.UUID `gorm:"type:uuid;index" json:"system_id"`
	TenantID uuid.UUID `gorm:"type:uuid;index" json:"tenant_id"`

	// Origen del hallazgo (cierra el lazo desde Verificar hacia Actuar)
	SourceType string     `gorm:"size:40" json:"source_type"` // auditoria|incidente|inspeccion|no_conformidad
	SourceID   *uuid.UUID `gorm:"type:uuid" json:"source_id"`
	Finding    string     `gorm:"type:text" json:"finding"`

	// Análisis de causa raíz (metodología, reusa el motor de incidentes)
	RootCause     string     `gorm:"type:text" json:"root_cause"`
	Action        string     `gorm:"type:text" json:"action"`
	Responsible   string     `gorm:"size:120" json:"responsible"`
	DueDate       *time.Time `json:"due_date"`
	Status        string     `gorm:"size:30;default:'abierta'" json:"status"` // abierta|en_progreso|cerrada|verificada
	Effectiveness string     `gorm:"type:text" json:"effectiveness"`          // verificación de eficacia

	// Retroalimentación al ciclo siguiente
	FeedsNextCycle bool `gorm:"default:false" json:"feeds_next_cycle"`

	ClosedAt  *time.Time `json:"closed_at"`
	CreatedAt time.Time  `json:"created_at"`
}

func (CorrectiveAction) TableName() string { return "corrective_actions" }
