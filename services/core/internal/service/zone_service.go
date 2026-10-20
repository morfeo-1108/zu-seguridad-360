package service

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/zu-seguridad/core/internal/geo"
	"github.com/zu-seguridad/core/internal/model"
	"gorm.io/gorm"
)

type ZoneService struct {
	db *gorm.DB
}

func NewZoneService(db *gorm.DB) *ZoneService {
	return &ZoneService{db: db}
}

// SafetyProtocolCatalogEnabled gates the zone/checklist catalog — pure
// reference data (what EPP a zone type requires per Res. 2400/1979 etc.),
// no AI, no biometrics, no drone action. It does NOT gate vision inference,
// biometric identification, or any drone flight/action — those live in the
// separate AI microservice and remain closed regardless of this flag; some
// of this catalog's own norm citations are still explicitly marked "pendiente
// de validación normativa vigente" in ProtocolTemplates below, by design —
// it is a draft reference, not a certified compliance source.
const SafetyProtocolCatalogEnabled = true

// ProtocolTemplates is DRAFT ONLY. Do not expose or use operationally until
// each requirement and legal citation has been reviewed against current
// Colombian rules by a competent SST professional. The routes are disabled.
var ProtocolTemplates = map[model.ZoneType][]model.ChecklistItem{
	model.ZoneAlturas: {
		{Code: "casco_barbuquejo", Description: "Casco con barbuquejo de 3-4 puntos", Type: model.CheckEPP, DetectionClass: "casco", Mandatory: true, NormRef: "Pendiente de validación normativa vigente", Severity: "critical", Order: 1},
		{Code: "arnes_cuerpo_completo", Description: "Arnés de cuerpo completo certificado", Type: model.CheckEPP, DetectionClass: "arnes", Mandatory: true, NormRef: "Pendiente de validación normativa vigente", Severity: "critical", Order: 2},
		{Code: "eslinga_doble", Description: "Eslinga con absorbedor de energía", Type: model.CheckEPP, DetectionClass: "eslinga", Mandatory: true, NormRef: "Pendiente de validación normativa vigente", Severity: "critical", Order: 3},
		{Code: "linea_vida", Description: "Conexión a línea de vida o punto de anclaje", Type: model.CheckBehavior, DetectionClass: "linea_vida", Mandatory: true, NormRef: "Pendiente de validación normativa vigente", Severity: "critical", Order: 4},
		{Code: "calzado_seguridad", Description: "Calzado de seguridad antideslizante", Type: model.CheckEPP, DetectionClass: "botas", Mandatory: true, NormRef: "Pendiente de validación normativa vigente", Severity: "high", Order: 5},
		{Code: "permiso_alturas", Description: "Permiso de trabajo en alturas vigente", Type: model.CheckPermit, DetectionClass: "", Mandatory: true, NormRef: "Pendiente de validación normativa vigente", Severity: "critical", Order: 6},
	},
	model.ZoneEspacioConfinado: {
		{Code: "monitor_gases", Description: "Monitor de gases activo", Type: model.CheckEnvironment, DetectionClass: "monitor_gases", Mandatory: true, NormRef: "Pendiente de validación normativa vigente", Severity: "critical", Order: 1},
		{Code: "arnes_rescate", Description: "Arnés con punto de rescate dorsal", Type: model.CheckEPP, DetectionClass: "arnes", Mandatory: true, NormRef: "Pendiente de validación normativa vigente", Severity: "critical", Order: 2},
		{Code: "tripode_rescate", Description: "Trípode de rescate en boca de acceso", Type: model.CheckEnvironment, DetectionClass: "tripode", Mandatory: true, NormRef: "Pendiente de validación normativa vigente", Severity: "critical", Order: 3},
		{Code: "vigia_externo", Description: "Vigía externo presente", Type: model.CheckBehavior, DetectionClass: "vigia", Mandatory: true, NormRef: "Pendiente de validación normativa vigente", Severity: "critical", Order: 4},
		{Code: "ventilacion", Description: "Sistema de ventilación forzada", Type: model.CheckEnvironment, DetectionClass: "ventilador", Mandatory: true, NormRef: "Pendiente de validación normativa vigente", Severity: "high", Order: 5},
		{Code: "permiso_confinado", Description: "Permiso de espacio confinado vigente", Type: model.CheckPermit, DetectionClass: "", Mandatory: true, NormRef: "Pendiente de validación normativa vigente", Severity: "critical", Order: 6},
	},
	model.ZoneSoldadura: {
		{Code: "careta_soldar", Description: "Careta de soldar con filtro adecuado", Type: model.CheckEPP, DetectionClass: "careta", Mandatory: true, NormRef: "Res. 2400/1979", Severity: "high", Order: 1},
		{Code: "guantes_carnaza", Description: "Guantes de carnaza largos", Type: model.CheckEPP, DetectionClass: "guantes", Mandatory: true, NormRef: "Res. 2400/1979", Severity: "high", Order: 2},
		{Code: "delantal_cuero", Description: "Delantal de cuero", Type: model.CheckEPP, DetectionClass: "delantal", Mandatory: true, NormRef: "Res. 2400/1979", Severity: "medium", Order: 3},
		{Code: "polainas", Description: "Polainas de cuero", Type: model.CheckEPP, DetectionClass: "polainas", Mandatory: false, NormRef: "Res. 2400/1979", Severity: "medium", Order: 4},
		{Code: "extintor_cercano", Description: "Extintor a menos de 3 metros", Type: model.CheckEnvironment, DetectionClass: "extintor", Mandatory: true, NormRef: "Res. 2400/1979", Severity: "high", Order: 5},
	},
	model.ZoneElectrico: {
		{Code: "guantes_dielectricos", Description: "Guantes dieléctricos clase adecuada", Type: model.CheckEPP, DetectionClass: "guantes_dielectricos", Mandatory: true, NormRef: "RETIE", Severity: "critical", Order: 1},
		{Code: "casco_clase_e", Description: "Casco dieléctrico clase E", Type: model.CheckEPP, DetectionClass: "casco", Mandatory: true, NormRef: "RETIE", Severity: "critical", Order: 2},
		{Code: "calzado_dielectrico", Description: "Calzado dieléctrico", Type: model.CheckEPP, DetectionClass: "botas", Mandatory: true, NormRef: "RETIE", Severity: "high", Order: 3},
		{Code: "bloqueo_loto", Description: "Bloqueo y etiquetado LOTO aplicado", Type: model.CheckBehavior, DetectionClass: "candado_loto", Mandatory: true, NormRef: "RETIE", Severity: "critical", Order: 4},
	},
	model.ZoneGeneral: {
		{Code: "casco", Description: "Casco de seguridad", Type: model.CheckEPP, DetectionClass: "casco", Mandatory: true, NormRef: "Res. 2400/1979", Severity: "high", Order: 1},
		{Code: "chaleco_reflectivo", Description: "Chaleco reflectivo", Type: model.CheckEPP, DetectionClass: "chaleco", Mandatory: true, NormRef: "Res. 2400/1979", Severity: "medium", Order: 2},
		{Code: "botas_punta_acero", Description: "Botas con puntera de acero", Type: model.CheckEPP, DetectionClass: "botas", Mandatory: true, NormRef: "Res. 2400/1979", Severity: "high", Order: 3},
		{Code: "gafas", Description: "Gafas de seguridad", Type: model.CheckEPP, DetectionClass: "gafas", Mandatory: false, NormRef: "Res. 2400/1979", Severity: "low", Order: 4},
	},
}

// CreateProject crea una obra/proyecto del cliente, contenedor de zonas.
func (s *ZoneService) CreateProject(ctx context.Context, tenantID, companyID uuid.UUID, name, address string) (*model.Project, error) {
	if !SafetyProtocolCatalogEnabled {
		return nil, fmt.Errorf("safety zone protocols are disabled pending normative validation")
	}
	project := &model.Project{TenantID: tenantID, CompanyID: companyID, Name: name, Address: address, Status: "active"}
	return project, s.db.WithContext(ctx).Create(project).Error
}

// ListProjects lista los proyectos de una empresa con sus zonas y checklist.
func (s *ZoneService) ListProjects(ctx context.Context, tenantID, companyID uuid.UUID) ([]model.Project, error) {
	if !SafetyProtocolCatalogEnabled {
		return nil, fmt.Errorf("safety zone protocols are disabled pending normative validation")
	}
	var projects []model.Project
	err := s.db.WithContext(ctx).Where("tenant_id = ? AND company_id = ?", tenantID, companyID).
		Preload("Zones.Checklist").Order("created_at DESC").Find(&projects).Error
	return projects, err
}

// CreateZone crea una zona y siembra su checklist desde la plantilla normativa.
func (s *ZoneService) CreateZone(ctx context.Context, tenantID, projectID uuid.UUID, name string, ztype model.ZoneType, geofence model.JSONB, markerID string) (*model.Zone, error) {
	if !SafetyProtocolCatalogEnabled {
		return nil, fmt.Errorf("safety zone protocols are disabled pending normative validation")
	}
	normRef := map[model.ZoneType]string{
		model.ZoneAlturas: "Pendiente de validación normativa vigente", model.ZoneEspacioConfinado: "Pendiente de validación normativa vigente",
		model.ZoneSoldadura: "Res. 2400/1979", model.ZoneElectrico: "RETIE", model.ZoneGeneral: "Res. 2400/1979",
	}[ztype]

	zone := &model.Zone{
		TenantID: tenantID, ProjectID: projectID, Name: name, Type: ztype,
		NormRef: normRef, GeofencePolygon: geofence, VisualMarkerID: markerID,
		RiskClass: s.riskClassFor(ztype),
	}

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// The project must belong to this tenant — otherwise any caller who
		// knows (or guesses) a projectID from another tenant could seed a
		// zone into it.
		var count int64
		if err := tx.Model(&model.Project{}).Where("id = ? AND tenant_id = ?", projectID, tenantID).Count(&count).Error; err != nil {
			return err
		}
		if count != 1 {
			return gorm.ErrRecordNotFound
		}
		if err := tx.Create(zone).Error; err != nil {
			return err
		}
		// Sembrar checklist desde plantilla normativa
		for _, item := range ProtocolTemplates[ztype] {
			item.ZoneID = zone.ID
			if err := tx.Create(&item).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	s.db.WithContext(ctx).Preload("Checklist").First(zone, zone.ID)
	return zone, nil
}

// GetZoneChecklist devuelve el checklist que el dron debe validar en una zona.
func (s *ZoneService) GetZoneChecklist(ctx context.Context, tenantID, companyID, zoneID uuid.UUID) ([]model.ChecklistItem, error) {
	if !SafetyProtocolCatalogEnabled {
		return nil, fmt.Errorf("safety zone protocols are disabled pending normative validation")
	}
	var items []model.ChecklistItem
	err := s.db.WithContext(ctx).
		Joins("JOIN zones ON zones.id = checklist_items.zone_id").
		Joins("JOIN projects ON projects.id = zones.project_id").
		Where("checklist_items.zone_id = ? AND zones.tenant_id = ? AND projects.company_id = ?", zoneID, tenantID, companyID).
		Order("\"order\" ASC").Find(&items).Error
	return items, err
}

// ResolveZoneByLocation determina en qué zona está el dron según su GPS.
// Reconocimiento de zona por geocerca (point-in-polygon).
func (s *ZoneService) ResolveZoneByLocation(ctx context.Context, tenantID, projectID uuid.UUID, lat, lng float64) (*model.Zone, error) {
	if !SafetyProtocolCatalogEnabled {
		return nil, fmt.Errorf("safety zone protocols are disabled pending normative validation")
	}
	var zones []model.Zone
	if err := s.db.WithContext(ctx).Where("project_id = ? AND tenant_id = ?", projectID, tenantID).Preload("Checklist").Find(&zones).Error; err != nil {
		return nil, err
	}
	for i := range zones {
		if polygon, ok := zones[i].GeofencePolygon["polygon"].([]interface{}); ok {
			if geo.PointInPolygon(lat, lng, polygon) {
				return &zones[i], nil
			}
		}
	}
	return nil, fmt.Errorf("ubicación (%.5f, %.5f) no corresponde a ninguna zona definida", lat, lng)
}

// ResolveZoneByMarker determina la zona por marcador visual (ArUco/QR) detectado.
func (s *ZoneService) ResolveZoneByMarker(ctx context.Context, tenantID, projectID uuid.UUID, markerID string) (*model.Zone, error) {
	if !SafetyProtocolCatalogEnabled {
		return nil, fmt.Errorf("safety zone protocols are disabled pending normative validation")
	}
	var zone model.Zone
	err := s.db.WithContext(ctx).
		Where("project_id = ? AND tenant_id = ? AND visual_marker_id = ?", projectID, tenantID, markerID).
		Preload("Checklist").First(&zone).Error
	if err != nil {
		return nil, fmt.Errorf("marcador '%s' no asociado a ninguna zona", markerID)
	}
	return &zone, nil
}

func (s *ZoneService) riskClassFor(t model.ZoneType) string {
	switch t {
	case model.ZoneAlturas, model.ZoneEspacioConfinado, model.ZoneElectrico:
		return "alto"
	case model.ZoneSoldadura, model.ZoneExcavacion, model.ZoneQuimicos:
		return "medio"
	default:
		return "bajo"
	}
}
