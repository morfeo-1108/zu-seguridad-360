package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CommitteeType string

const (
	CommitteeCOPASST CommitteeType = "copasst" // >= 10 trabajadores
	CommitteeVigia    CommitteeType = "vigia"   // < 10 trabajadores
)

// Committee is a company's COPASST/Vigía Ocupacional for one period
// (standard term in Colombia is 2 years). Members are kept as a JSON array
// ({employee_id, role, is_alternate}) rather than a separate table — the
// membership roster only changes at election time, unlike meetings which
// accumulate indefinitely.
type Committee struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	TenantID    uuid.UUID      `gorm:"type:uuid;not null;index" json:"tenant_id"`
	CompanyID   uuid.UUID      `gorm:"type:uuid;not null;index" json:"company_id"`
	Type        CommitteeType  `gorm:"size:20;not null" json:"type"`
	PeriodStart time.Time      `json:"period_start"`
	PeriodEnd   time.Time      `json:"period_end"`
	Members     JSONBArray     `gorm:"type:jsonb;not null;default:'[]'" json:"members"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Committee) TableName() string { return "committees" }

func (c *Committee) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

// IsActive reports whether now falls within the committee's elected period.
func (c *Committee) IsActive(now time.Time) bool {
	return !now.Before(c.PeriodStart) && !now.After(c.PeriodEnd)
}

// CommitteeMeeting is one minuted session (acta) of a committee, with
// attendees and the decisions/votes recorded in that session.
type CommitteeMeeting struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	TenantID    uuid.UUID      `gorm:"type:uuid;not null;index" json:"tenant_id"`
	CompanyID   uuid.UUID      `gorm:"type:uuid;not null;index" json:"company_id"`
	CommitteeID uuid.UUID      `gorm:"type:uuid;not null;index" json:"committee_id"`
	HeldAt      time.Time      `json:"held_at"`
	Agenda      string         `gorm:"type:text" json:"agenda"`
	Minutes     string         `gorm:"type:text" json:"minutes"` // acta
	Attendees   JSONBArray     `gorm:"type:jsonb;default:'[]'" json:"attendees"`
	Decisions   JSONBArray     `gorm:"type:jsonb;default:'[]'" json:"decisions"` // [{topic, vote_for, vote_against, result}]
	RecordedBy  uuid.UUID      `gorm:"type:uuid;not null" json:"recorded_by"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (CommitteeMeeting) TableName() string { return "committee_meetings" }

func (m *CommitteeMeeting) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}
