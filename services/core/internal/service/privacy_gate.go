package service

import "errors"

var (
	ErrConsentCaptureDisabled          = errors.New("consent capture is disabled pending approved evidence and policy validation")
	ErrConsentAuthorizationUnavailable = errors.New("consent authorization is unavailable until an approved policy is integrated")
	ErrConsentManagementDisabled       = errors.New("consent management is disabled until actor attribution and audit are implemented")
	ErrSensitiveDataProcessingDisabled = errors.New("sensitive worker-data processing is disabled pending privacy and retention approval")
	ErrSensitiveDataExportDisabled     = errors.New("sensitive data export is disabled pending allowlist, authorization, and privacy review")
	ErrSyntheticDataRequired           = errors.New("synthetic pilot accepts only explicitly marked test data")
)
