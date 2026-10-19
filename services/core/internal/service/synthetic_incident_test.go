package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestCloseSyntheticRemainsDisabledByDefault(t *testing.T) {
	service := NewIncidentService(nil, nil)
	_, err := service.CloseSynthetic(context.Background(), uuid.New(), uuid.New(), uuid.New(), uuid.New())
	if !errors.Is(err, ErrSensitiveDataProcessingDisabled) {
		t.Fatalf("expected synthetic incident close to remain disabled by default, got %v", err)
	}
}
