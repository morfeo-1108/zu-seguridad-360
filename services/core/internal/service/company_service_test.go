package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func TestSyntheticCompanyRequiresReservedRiskMetadata(t *testing.T) {
	service := NewCompanyService(nil, nil, "", true)
	base := CreateCompanyInput{
		Name:          "ZU360 PILOTO SINTÉTICO - EMPRESA DE PRUEBA",
		NIT:           "TEST-NIT-9001",
		CIIU:          "TEST",
		ARL:           "ARL SINTÉTICA",
		LegalRep:      "Persona sintética representante",
		LegalRepEmail: "representante@zu360.test",
	}

	tests := []struct {
		name   string
		mutate func(*CreateCompanyInput)
	}{
		{name: "real CIIU classification", mutate: func(input *CreateCompanyInput) { input.CIIU = "0111" }},
		{name: "real ARL name", mutate: func(input *CreateCompanyInput) { input.ARL = "ARL externa" }},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input := base
			test.mutate(&input)
			_, err := service.CreateForUser(context.Background(), input, uuid.New(), uuid.New(), "")
			if err != ErrSyntheticDataRequired {
				t.Fatalf("expected ErrSyntheticDataRequired, got %v", err)
			}
		})
	}
}

func TestSyntheticPilotRiskScoreIsDisabled(t *testing.T) {
	service := NewCompanyService(nil, nil, "", true)
	value, err := service.CalculateRiskScore(context.Background(), uuid.New(), uuid.New())
	if err != ErrSensitiveDataProcessingDisabled {
		t.Fatalf("expected ErrSensitiveDataProcessingDisabled, got %v", err)
	}
	if value != 0 {
		t.Fatalf("expected no risk score, got %v", value)
	}
}

func TestSyntheticARLMarker(t *testing.T) {
	if !isSyntheticARL("ARL SINTÉTICA") || !isSyntheticARL("Demostración ARL") {
		t.Fatal("expected reserved synthetic ARL labels to pass")
	}
	if isSyntheticARL("ARL externa") {
		t.Fatal("expected real provider names to be rejected")
	}
}
