package model

import (
	"github.com/google/uuid"
	"time"
)

// RegulatoryWatch implementa la VIGILANCIA NORMATIVA con humano en el lazo.
//
// PRINCIPIO CRÍTICO: el sistema NO se actualiza solo ante cambios legales. Eso
// sería un riesgo legal grave. Lo que hace es DETECTAR posibles cambios en
// fuentes oficiales y ENCOLARLOS para que un experto los apruebe. Detectar y
// proponer, sí; decidir y aplicar solo, nunca.

// RegulatorySource: fuente oficial monitoreada.
type RegulatorySource struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name        string     `gorm:"size:120" json:"name"` // "Ministerio de Trabajo", "ISO", "Aerocivil"
	URL         string     `gorm:"type:text" json:"url"`
	Framework   string     `gorm:"size:40" json:"framework"` // norma asociada
	LastChecked *time.Time `json:"last_checked"`
	Active      bool       `gorm:"default:true" json:"active"`
}

func (RegulatorySource) TableName() string { return "regulatory_sources" }

// RegulatoryChange: un cambio detectado, en cola de aprobación humana.
type RegulatoryChange struct {
	ID       uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	SourceID uuid.UUID `gorm:"type:uuid;index" json:"source_id"`

	Framework   string    `gorm:"size:40;index" json:"framework"`
	Title       string    `gorm:"type:text" json:"title"`
	Summary     string    `gorm:"type:text" json:"summary"` // resumen detectado
	DetectedURL string    `gorm:"type:text" json:"detected_url"`
	DetectedAt  time.Time `json:"detected_at"`

	// Estado en el lazo de aprobación humana
	Status string `gorm:"size:30;default:'pendiente_revision'" json:"status"`
	// pendiente_revision → en_analisis → aprobado | descartado
	Impact         string `gorm:"size:30" json:"impact"`            // alto|medio|bajo (propuesto por IA)
	ProposedAction string `gorm:"type:text" json:"proposed_action"` // qué sugiere el sistema

	// Aprobación del experto (obligatoria antes de afectar nada)
	ReviewedBy      *uuid.UUID `gorm:"type:uuid" json:"reviewed_by"`
	ReviewedAt      *time.Time `json:"reviewed_at"`
	ReviewNotes     string     `gorm:"type:text" json:"review_notes"`
	AppliedToSystem bool       `gorm:"default:false" json:"applied_to_system"`

	CreatedAt time.Time `json:"created_at"`
}

func (RegulatoryChange) TableName() string { return "regulatory_changes" }
