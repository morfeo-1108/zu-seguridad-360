package model

import (
	"time"

	"github.com/google/uuid"
)

// CompanyMembership grants one user access to one company within the user's
// tenant. Company access is never inferred from a client-supplied company ID.
type CompanyMembership struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantID    uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_company_membership_user_company,priority:1" json:"tenant_id"`
	UserID      uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_company_membership_user_company,priority:2" json:"user_id"`
	CompanyID   uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_company_membership_user_company,priority:3" json:"company_id"`
	Role        string    `gorm:"size:40;not null;default:'manager'" json:"role"`
	IsActive    bool      `gorm:"not null;default:true;index" json:"is_active"`
	IsSynthetic bool      `gorm:"not null;default:false;index" json:"is_synthetic"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (CompanyMembership) TableName() string { return "company_memberships" }
