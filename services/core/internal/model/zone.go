package model

import (
	"github.com/google/uuid"
	"time"
)

// ── Modelo de Zonas y Protocolos ─────────────────────────────
//
// Modelo de dominio propuesto. No acredita que el producto identifique personas,
// valide EPP o aplique protocolos operativos; drones/biometría están deshabilitados.
//
// Las categorías y sus requisitos normativos deben ser revisados por un profesional
// SST contra fuentes vigentes antes de exponerse o utilizarse operativamente.

// Project es una obra/proyecto del cliente. Contiene múltiples zonas.
type Project struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantID  uuid.UUID `gorm:"type:uuid;index" json:"tenant_id"`
	CompanyID uuid.UUID `gorm:"type:uuid;index" json:"company_id"`
	Name      string    `gorm:"size:200" json:"name"`
	Address   string    `gorm:"type:text" json:"address"`
	Status    string    `gorm:"size:30;default:'active'" json:"status"`
	Zones     []Zone    `gorm:"foreignKey:ProjectID" json:"zones,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Project) TableName() string { return "projects" }

// ZoneType define categorías de zona; no implica que exista una norma verificada.
type ZoneType string

const (
	ZoneAlturas          ZoneType = "alturas"
	ZoneEspacioConfinado ZoneType = "espacio_confinado"
	ZoneSoldadura        ZoneType = "soldadura"
	ZoneElectrico        ZoneType = "electrico"
	ZoneExcavacion       ZoneType = "excavacion"
	ZoneQuimicos         ZoneType = "quimicos"
	ZoneGeneral          ZoneType = "general"
)

// Zone es un entorno físico dentro de un proyecto con un protocolo propio.
type Zone struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantID  uuid.UUID `gorm:"type:uuid;index" json:"tenant_id"`
	ProjectID uuid.UUID `gorm:"type:uuid;index" json:"project_id"`
	Name      string    `gorm:"size:150" json:"name"`
	Type      ZoneType  `gorm:"size:40" json:"type"`
	NormRef   string    `gorm:"size:60" json:"norm_ref"`

	// Reconocimiento de zona: geocerca (polígono GPS) + marcadores visuales
	GeofencePolygon JSONB  `gorm:"type:jsonb" json:"geofence_polygon"` // [[lat,lng],...]
	VisualMarkerID  string `gorm:"size:60" json:"visual_marker_id"`    // ArUco/QR físico en la zona
	RiskClass       string `gorm:"size:10" json:"risk_class"`          // alto/medio/bajo

	// Protocolo: el checklist que el dron debe validar en esta zona
	Checklist []ChecklistItem `gorm:"foreignKey:ZoneID" json:"checklist,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

func (Zone) TableName() string { return "zones" }

// ChecklistItemType: qué valida el dron por cada ítem.
type ChecklistItemType string

const (
	CheckEPP         ChecklistItemType = "epp"      // presencia de EPP específico
	CheckBehavior    ChecklistItemType = "conducta" // postura, no aglomeración
	CheckEnvironment ChecklistItemType = "entorno"  // señalización, demarcación
	CheckPermit      ChecklistItemType = "permiso"  // permiso de trabajo vigente
)

// ChecklistItem es un requisito que el trabajador debe cumplir en la zona.
// El dron valida cada uno mediante visión IA.
type ChecklistItem struct {
	ID             uuid.UUID         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ZoneID         uuid.UUID         `gorm:"type:uuid;index" json:"zone_id"`
	Code           string            `gorm:"size:50" json:"code"` // ej: "casco", "arnes", "linea_vida"
	Description    string            `gorm:"type:text" json:"description"`
	Type           ChecklistItemType `gorm:"size:20" json:"type"`
	Mandatory      bool              `gorm:"default:true" json:"mandatory"`
	DetectionClass string            `gorm:"size:50" json:"detection_class"` // clase que el modelo de visión debe detectar
	NormRef        string            `gorm:"size:60" json:"norm_ref"`
	Severity       string            `gorm:"size:20" json:"severity"` // crítico si su ausencia puede ser fatal
	Order          int               `json:"order"`
}

func (ChecklistItem) TableName() string { return "checklist_items" }
