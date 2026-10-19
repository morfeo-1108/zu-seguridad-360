package model

import (
	"time"

	"github.com/google/uuid"
)

type ConsentRecord struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantID      uuid.UUID  `gorm:"type:uuid;not null;index" json:"tenant_id"`
	EmployeeID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"employee_id"`
	DataCategory  string     `gorm:"size:60;not null" json:"data_category"`
	Purpose       string     `gorm:"type:text;not null" json:"purpose"`
	Granted       bool       `json:"granted"`
	GrantedAt     time.Time  `gorm:"not null" json:"granted_at"`
	RevokedAt     *time.Time `json:"revoked_at"`
	Method        string     `gorm:"size:40" json:"method"`
	EvidenceURL   string     `gorm:"type:text" json:"evidence_url"`
	IPAddress     string     `gorm:"size:45" json:"ip_address"`
	Evidence      string     `gorm:"type:text" json:"evidence"`
	PolicyVersion string     `gorm:"size:16" json:"policy_version"`
	CreatedAt     time.Time  `json:"created_at"`
}

func (ConsentRecord) TableName() string { return "consent_records" }
