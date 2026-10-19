package service

import (
	"context"
	"github.com/google/uuid"
	"github.com/zu-seguridad/core/internal/model"
	"gorm.io/gorm"
	"time"
)

// RegulatoryWatchService: vigilancia normativa con humano en el lazo.
type RegulatoryWatchService struct {
	db *gorm.DB
}

func NewRegulatoryWatchService(db *gorm.DB) *RegulatoryWatchService {
	return &RegulatoryWatchService{db: db}
}

// SeedSources registra las fuentes oficiales a monitorear.
func (s *RegulatoryWatchService) SeedSources(ctx context.Context) error {
	sources := []model.RegulatorySource{
		{Name: "Ministerio de Trabajo (Colombia)", URL: "https://www.mintrabajo.gov.co/normatividad", Framework: "res_0312_2019", Active: true},
		{Name: "ISO — ISO 45001", URL: "https://www.iso.org/standard/63787.html", Framework: "iso_45001_2018", Active: true},
		{Name: "ISO — ISO 9001", URL: "https://www.iso.org/standard/62085.html", Framework: "iso_9001_2015", Active: true},
		{Name: "ISO — ISO 14001", URL: "https://www.iso.org/standard/60857.html", Framework: "iso_14001_2015", Active: true},
		{Name: "ISO — ISO 31000", URL: "https://www.iso.org/standard/65694.html", Framework: "iso_31000_2018", Active: true},
		{Name: "Aerocivil — RAC", URL: "https://www.aerocivil.gov.co/normatividad", Framework: "rac_91_100", Active: true},
	}
	for i := range sources {
		s.db.WithContext(ctx).Where("name = ?", sources[i].Name).FirstOrCreate(&sources[i])
	}
	return nil
}

// QueueDetectedChange: registra un cambio detectado en cola de revisión humana.
// Lo llamaría un worker que monitorea las fuentes (scraping/diff/feeds).
func (s *RegulatoryWatchService) QueueDetectedChange(ctx context.Context, sourceID uuid.UUID, framework, title, summary, url, impact, proposedAction string) (*model.RegulatoryChange, error) {
	change := &model.RegulatoryChange{
		SourceID: sourceID, Framework: framework, Title: title, Summary: summary,
		DetectedURL: url, DetectedAt: time.Now(), Status: "pendiente_revision",
		Impact: impact, ProposedAction: proposedAction,
	}
	if err := s.db.WithContext(ctx).Create(change).Error; err != nil {
		return nil, err
	}
	return change, nil
}

// PendingChanges: cola de cambios esperando aprobación del experto.
func (s *RegulatoryWatchService) PendingChanges(ctx context.Context) ([]model.RegulatoryChange, error) {
	var changes []model.RegulatoryChange
	err := s.db.WithContext(ctx).
		Where("status = ?", "pendiente_revision").
		Order("detected_at DESC").Find(&changes).Error
	return changes, err
}

// ReviewChange: el experto APRUEBA o DESCARTA un cambio. Solo lo aprobado puede
// afectar el sistema, y siempre tras decisión humana explícita.
func (s *RegulatoryWatchService) ReviewChange(ctx context.Context, changeID, reviewedBy uuid.UUID, decision, notes string) error {
	now := time.Now()
	status := "descartado"
	if decision == "aprobar" {
		status = "aprobado"
	}
	return s.db.WithContext(ctx).Model(&model.RegulatoryChange{}).
		Where("id = ?", changeID).
		Updates(map[string]interface{}{
			"status":       status,
			"reviewed_by":  reviewedBy,
			"reviewed_at":  now,
			"review_notes": notes,
		}).Error
}
