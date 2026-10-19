package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/zu-seguridad/core/internal/model"
	"gorm.io/gorm"
)

var ErrDataErasureDisabled = errors.New("data erasure is disabled until retention and schema rules are independently reviewed")

type ConsentService struct {
	db *gorm.DB
}

func NewConsentService(db *gorm.DB) *ConsentService {
	return &ConsentService{db: db}
}

// GrantConsent is deliberately closed: arbitrary caller strings and a URL are
// not sufficient proof of informed, purpose-specific consent.
func (s *ConsentService) GrantConsent(context.Context, uuid.UUID, uuid.UUID, string, string, string, string) (*model.ConsentRecord, error) {
	return nil, ErrConsentCaptureDisabled
}

// HasValidConsent never treats legacy Granted=true rows as authorization. A
// versioned policy and verified evidence flow must be implemented first.
func (s *ConsentService) HasValidConsent(context.Context, uuid.UUID, uuid.UUID, string) (bool, error) {
	return false, ErrConsentAuthorizationUnavailable
}

// RevokeConsent is closed until actor identity and append-only audit are wired.
// No consent capture or sensitive processing is available in the pilot build.
func (s *ConsentService) RevokeConsent(context.Context, uuid.UUID, uuid.UUID) error {
	return ErrConsentManagementDisabled
}

// RequestDataErasure is deliberately fail-closed until legal retention and
// schema-dependent deletion/anonymous-retention behavior have been reviewed.
func (s *ConsentService) RequestDataErasure(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) (map[string]interface{}, error) {
	return nil, ErrDataErasureDisabled
}
