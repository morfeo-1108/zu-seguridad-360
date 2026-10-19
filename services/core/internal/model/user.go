package model

import (
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"time"
)

type UserRole string

const (
	RoleAdmin      UserRole = "admin"
	RoleProfSST    UserRole = "professional_sst"
	RoleEmployer   UserRole = "employer"
	RoleWorker     UserRole = "worker"
	RoleIPS        UserRole = "ips"
	RoleAuditor    UserRole = "auditor"
	RoleContractor UserRole = "contractor"
)

type User struct {
	ID                    uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantID              uuid.UUID      `gorm:"type:uuid;not null;index" json:"tenant_id"`
	CompanyID             *uuid.UUID     `gorm:"type:uuid;index" json:"company_id"`
	Email                 string         `gorm:"uniqueIndex;not null;size:255" json:"email"`
	PasswordHash          string         `gorm:"not null" json:"-"`
	GoogleSub             *string        `gorm:"size:255;uniqueIndex:users_google_sub_uidx" json:"-"`
	AuthProvider          string         `gorm:"not null;size:24;default:password" json:"auth_provider"`
	Role                  UserRole       `gorm:"not null;size:30" json:"role"`
	FirstName             string         `gorm:"size:100" json:"first_name"`
	LastName              string         `gorm:"size:100" json:"last_name"`
	Phone                 string         `gorm:"size:20" json:"phone"`
	AvatarURL             string         `gorm:"size:500" json:"avatar_url"`
	IsActive              bool           `gorm:"default:true" json:"is_active"`
	IsSynthetic           bool           `gorm:"not null;default:false;index" json:"is_synthetic"`
	TermsAcceptedAt       *time.Time     `json:"-"`
	PrivacyAcknowledgedAt *time.Time     `json:"-"`
	TermsVersion          string         `gorm:"size:32" json:"-"`
	PrivacyNoticeVersion  string         `gorm:"size:32" json:"-"`
	LastLogin             *time.Time     `json:"last_login"`
	Preferences           JSONB          `gorm:"type:jsonb;default:'{}'" json:"preferences"`
	CreatedAt             time.Time      `json:"created_at"`
	UpdatedAt             time.Time      `json:"updated_at"`
	DeletedAt             gorm.DeletedAt `gorm:"index" json:"-"`
}

func (User) TableName() string { return "users" }

func (u *User) SetPassword(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hash)
	return nil
}

func (u *User) CheckPassword(password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) == nil
}
