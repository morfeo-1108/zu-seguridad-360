package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/zu-seguridad/core/internal/model"
	"gorm.io/gorm"
)

var ErrPrivacyErasureDisabled = ErrDataErasureDisabled

// PrivacyService is a partial implementation, not a compliance guarantee.
type PrivacyService struct{ db *gorm.DB }

func NewPrivacyService(db *gorm.DB) *PrivacyService { return &PrivacyService{db: db} }

// RegisterConsent must not create a consent record from arbitrary request data.
func (s *PrivacyService) RegisterConsent(context.Context, uuid.UUID, uuid.UUID, string, string) (*model.ConsentRecord, error) {
	return nil, ErrConsentCaptureDisabled
}

// HasValidConsent denies processing until a purpose-specific policy and verified
// evidence authorization flow are implemented.
func (s *PrivacyService) HasValidConsent(context.Context, uuid.UUID, uuid.UUID) (bool, error) {
	return false, ErrConsentAuthorizationUnavailable
}

// RevokeConsent is unavailable until an authenticated actor and audit event are
// mandatory inputs; the current signature cannot provide that attribution.
func (s *PrivacyService) RevokeConsent(context.Context, uuid.UUID, uuid.UUID) error {
	return ErrConsentManagementDisabled
}

// Erasure is intentionally disabled: retention periods, evidence preservation,
// and the current Employee schema require legal and engineering review first.
func (s *PrivacyService) EraseEmployeeData(context.Context, uuid.UUID, uuid.UUID, string) error {
	return ErrPrivacyErasureDisabled
}
