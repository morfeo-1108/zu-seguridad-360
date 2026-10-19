package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DataExportService remains closed until a reviewed, field-level allowlist and
// an authenticated, audited authorization workflow exist.
type DataExportService struct {
	db *gorm.DB
}

func NewDataExportService(db *gorm.DB) *DataExportService {
	return &DataExportService{db: db}
}

// CompanyExport is retained as a versioned contract placeholder only. It must
// not be populated from arbitrary table maps before privacy review.
type CompanyExport struct {
	SchemaVersion     string                   `json:"schema_version"`
	ExportedAt        time.Time                `json:"exported_at"`
	CompanyID         uuid.UUID                `json:"company_id"`
	Company           map[string]interface{}   `json:"company"`
	Employees         []map[string]interface{} `json:"employees"`
	Incidents         []map[string]interface{} `json:"incidents"`
	Certifications    []map[string]interface{} `json:"certifications"`
	MedicalExams      []map[string]interface{} `json:"medical_exams"`
	Documents         []map[string]interface{} `json:"documents"`
	Risks             []map[string]interface{} `json:"risks"`
	ManagementSystems []map[string]interface{} `json:"management_systems"`
	Note              string                   `json:"note"`
}

func (s *DataExportService) ExportCompany(context.Context, uuid.UUID, uuid.UUID) (*CompanyExport, error) {
	return nil, ErrSensitiveDataExportDisabled
}

func (s *DataExportService) ExportCompanyJSON(context.Context, uuid.UUID, uuid.UUID) ([]byte, error) {
	return nil, ErrSensitiveDataExportDisabled
}
