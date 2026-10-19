package phva

import "testing"

func TestComplianceScore_OnlyCountsHumanValidated(t *testing.T) {
	standards := []StandardState{
		{Weight: 3, Status: "cumple", HumanValidated: true},    // cuenta: 3
		{Weight: 2, Status: "cumple", HumanValidated: false},   // NO cuenta (sin validar)
		{Weight: 3, Status: "no_cumple", HumanValidated: true}, // NO cuenta (no cumple)
		{Weight: 2, Status: "cumple", HumanValidated: true},    // cuenta: 2
	}
	// Esperado: (3+2) / (3+2+3+2) = 5/10 = 50%
	got := ComplianceScore(standards)
	if got != 50.0 {
		t.Errorf("ComplianceScore = %.1f, esperaba 50.0 (solo validados que cumplen)", got)
	}
}

func TestComplianceScore_SystemProposalDoesNotCount(t *testing.T) {
	// Aunque el sistema proponga "cumple", sin validación humana NO suma.
	standards := []StandardState{
		{Weight: 5, Status: "cumple", HumanValidated: false},
	}
	if got := ComplianceScore(standards); got != 0 {
		t.Errorf("propuesta sin validar no debe contar: got %.1f, esperaba 0", got)
	}
}

func TestComplianceScore_Empty(t *testing.T) {
	if got := ComplianceScore(nil); got != 0 {
		t.Errorf("sin estándares = 0, got %.1f", got)
	}
}

func TestNextPhase_CyclesCorrectly(t *testing.T) {
	seq := []struct{ from, to string }{
		{"planear", "hacer"},
		{"hacer", "verificar"},
		{"verificar", "actuar"},
		{"actuar", "planear"}, // cierra el lazo
	}
	for _, s := range seq {
		if got := NextPhase(s.from); got != s.to {
			t.Errorf("NextPhase(%q) = %q, esperaba %q", s.from, got, s.to)
		}
	}
}
