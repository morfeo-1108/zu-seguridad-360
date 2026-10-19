package service

import "testing"

func TestValidateSyntheticEmployeeInput(t *testing.T) {
	valid := CreateEmployeeInput{
		FirstName: "Persona", LastName: "Sintético A1", DocumentType: "SYN",
		DocumentNumber: "SYN-000001", Email: "persona1@zu360.test",
		Position: "Demostración SST", Department: "Sintético",
	}
	if err := ValidateSyntheticEmployeeInput(valid); err != nil {
		t.Fatalf("valid synthetic employee rejected: %v", err)
	}

	cases := []struct {
		name string
		edit func(*CreateEmployeeInput)
	}{
		{"real-name-like", func(in *CreateEmployeeInput) { in.FirstName = "Carlos" }},
		{"real-document", func(in *CreateEmployeeInput) { in.DocumentNumber = "1020304050" }},
		{"real-email-domain", func(in *CreateEmployeeInput) { in.Email = "person@example.com" }},
		{"phone", func(in *CreateEmployeeInput) { in.Phone = "+573001234567" }},
		{"unmarked-position", func(in *CreateEmployeeInput) { in.Position = "Operario" }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			input := valid
			tc.edit(&input)
			if err := ValidateSyntheticEmployeeInput(input); err == nil {
				t.Fatal("non-synthetic input was accepted")
			}
		})
	}
}
