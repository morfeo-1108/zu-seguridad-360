package model

import (
	"github.com/google/uuid"
	"time"
)

// EvidenceRecord es una prueba fotográfica inmutable ligada a un trabajador.
//
// Cada validación del dron, cada inspección, cada incidente, cada entrega de EPP
// genera un registro con foto georreferenciada y timestamp. El perfil del trabajador
// acumula su línea de tiempo de evidencias para validación exacta ante cualquier
// auditoría, protocolo o imprevisto (ej: ARL pide prueba de que usaba arnés).
//
// Inmutabilidad: append-only. La foto vive en S3 (WORM opcional). El registro nunca
// se edita ni borra mientras esté bajo legal-hold.
type EvidenceType string

const (
	EvidenceEPPValidation EvidenceType = "epp_validation"   // validación de dron por zona
	EvidenceIncident      EvidenceType = "incident"         // foto de incidente
	EvidenceInspection    EvidenceType = "inspection"       // inspección de seguridad
	EvidenceEPPDelivery   EvidenceType = "epp_delivery"     // entrega de EPP firmada
	EvidenceBiometric     EvidenceType = "biometric_enroll" // enrolamiento facial
	EvidenceTraining      EvidenceType = "training"         // asistencia a capacitación
	EvidencePermit        EvidenceType = "permit"           // permiso de trabajo
)

type EvidenceRecord struct {
	ID         uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantID   uuid.UUID  `gorm:"type:uuid;index" json:"tenant_id"`
	ProjectID  *uuid.UUID `gorm:"type:uuid;index" json:"project_id"`
	ZoneID     *uuid.UUID `gorm:"type:uuid;index" json:"zone_id"`
	EmployeeID uuid.UUID  `gorm:"type:uuid;index:idx_evidence_employee" json:"employee_id"`
	CompanyID  uuid.UUID  `gorm:"type:uuid;index" json:"company_id"`

	Type         EvidenceType `gorm:"size:30;index" json:"type"`
	PhotoURL     string       `gorm:"type:text" json:"photo_url"` // S3
	ThumbnailURL string       `gorm:"type:text" json:"thumbnail_url"`
	Source       string       `gorm:"size:20" json:"source"` // drone | mobile | manual

	// Georreferenciación y temporalidad (clave para validación exacta)
	GPSLat     float64   `json:"gps_lat"`
	GPSLng     float64   `json:"gps_lng"`
	CapturedAt time.Time `gorm:"index" json:"captured_at"`

	// Resultado de la validación que generó esta evidencia
	ChecklistCode string `gorm:"size:50" json:"checklist_code"`    // ej: "arnes", "casco"
	DetectionData JSONB  `gorm:"type:jsonb" json:"detection_data"` // clases detectadas, confianza
	Compliant     *bool  `json:"compliant"`                        // cumplió o no
	Severity      string `gorm:"size:20" json:"severity"`

	// Integridad: hash de la foto para verificar que no se alteró
	PhotoChecksum string `gorm:"size:64" json:"photo_checksum"`
	LegalHold     bool   `gorm:"default:false" json:"legal_hold"`

	CreatedAt time.Time `json:"created_at"`
}

func (EvidenceRecord) TableName() string { return "evidence_records" }

// AgentEvent es el registro append-only de actividad de los agentes.
// Alimenta el feed EN VIVO y queda como bitácora auditable de qué decidió/hizo
// cada agente, cuándo y sobre qué recurso.
type AgentEvent struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantID  uuid.UUID  `gorm:"type:uuid;index:idx_agentevent_tenant" json:"tenant_id"`
	ProjectID *uuid.UUID `gorm:"type:uuid;index" json:"project_id"`

	AgentKey   string `gorm:"size:40;index" json:"agent_key"` // ej: "vision", "billing"
	AgentClass string `gorm:"size:20" json:"agent_class"`     // reasoning | deterministic
	Action     string `gorm:"type:text" json:"action"`        // descripción legible
	Severity   string `gorm:"size:20" json:"severity"`        // info | warning | critical

	ResourceType string     `gorm:"size:40" json:"resource_type"` // employee | incident | zone | invoice
	ResourceID   *uuid.UUID `gorm:"type:uuid" json:"resource_id"`
	EmployeeID   *uuid.UUID `gorm:"type:uuid;index" json:"employee_id"`

	Metadata  JSONB     `gorm:"type:jsonb" json:"metadata"`
	CreatedAt time.Time `gorm:"index;autoCreateTime" json:"created_at"`
}

func (AgentEvent) TableName() string { return "agent_events" }
