package service

import "testing"

func TestValidateSyntheticTrainingInput(t *testing.T) {
	valid := CreateTrainingInput{
		Title: "Demostración de inducción", Category: "Prueba", DurationHours: 1,
		ValidityMonths: 3, EvaluationMin: 0, InstructorName: "Sintético Facilitador",
	}
	if err := ValidateSyntheticTrainingInput(valid); err != nil {
		t.Fatalf("valid synthetic training rejected: %v", err)
	}
	invalid := []CreateTrainingInput{
		{Title: "Inducción SST", Category: "Legal", DurationHours: 8, ValidityMonths: 12, EvaluationMin: 70, NormRef: "Res. pendiente"},
		{Title: "Demostración", Category: "Prueba", DurationHours: 1, ValidityMonths: 0, EvaluationMin: 0},
		{Title: "Demostración", Category: "Prueba", DurationHours: 1, ValidityMonths: 0, EvaluationMin: 70},
	}
	for index, input := range invalid {
		if err := ValidateSyntheticTrainingInput(input); err == nil {
			t.Fatalf("unsafe training input %d was accepted", index)
		}
	}
}
