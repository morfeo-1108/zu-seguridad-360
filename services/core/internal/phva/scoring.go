// Package phva contiene la lógica pura del ciclo PHVA, testeable sin BD.
package phva

// StandardState representa el estado mínimo de un estándar para el cálculo.
type StandardState struct {
	Weight         int
	Status         string // "cumple" | "no_cumple" | "parcial" | "pendiente"
	HumanValidated bool
}

// ComplianceScore calcula el % de cumplimiento contando SOLO estándares
// validados por humano que están en estado "cumple". Esta es la garantía
// legal del sistema: nada cuenta hasta que un profesional lo aprueba.
func ComplianceScore(standards []StandardState) float64 {
	totalWeight, compliantWeight := 0, 0
	for _, s := range standards {
		totalWeight += s.Weight
		if s.HumanValidated && s.Status == "cumple" {
			compliantWeight += s.Weight
		}
	}
	if totalWeight == 0 {
		return 0
	}
	return float64(compliantWeight) / float64(totalWeight) * 100
}

// NextPhase devuelve la siguiente fase del ciclo PHVA (cierra y reinicia).
func NextPhase(current string) string {
	switch current {
	case "planear":
		return "hacer"
	case "hacer":
		return "verificar"
	case "verificar":
		return "actuar"
	case "actuar":
		return "planear"
	default:
		return "planear"
	}
}
