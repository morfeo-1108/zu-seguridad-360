package dian

import (
	"testing"
	"time"
)

// Worked example of the CUFE in DIAN's Anexo Técnico de Factura Electrónica.
func TestCUFEMatchesAnexoTecnicoExample(t *testing.T) {
	at, _ := time.Parse("2006-01-02 15:04:05-07:00", "2019-01-16 10:53:10-05:00")
	in := CUFEInput{
		Number: "323200000129", IssuedAt: at,
		Subtotal: 150000000, IVA: 28500000, Total: 178500000,
		IssuerNIT: "700085371", BuyerID: "800199436",
		TechnicalKey: "693ff6f2a553c3646a063436fd4dd9ded0311471", Environment: "1",
	}
	want := "8bb918b19ba22a694f1da11c643b5e9de39adf60311cf179179e9b33381030bcd4c3c3f156c506ed5908f9276f5bd9b4"
	if got := CUFE(in); got != want {
		t.Fatalf("CUFE mismatch\n src  %s\n got  %s\n want %s", CUFESource(in), got, want)
	}
}
