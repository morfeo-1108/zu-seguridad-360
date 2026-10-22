package dian

import (
	"bytes"
	"encoding/xml"
	"io"
	"strings"
	"testing"
	"time"
)

func sample() Invoice {
	at, _ := time.Parse(time.RFC3339, "2026-10-22T09:30:00-05:00")
	return Invoice{
		Number: "SETP990000002", IssuedAt: at, DueAt: at.AddDate(0, 0, 5),
		Resolution: Resolution{Number: "18760000001", Prefix: "SETP", From: 990000000, To: 995000000, ValidFrom: "2019-01-19", ValidTo: "2030-01-19", TechnicalKey: "fc8eac422eba16e22ffd8c6f94b3f40a6e38162c"},
		Issuer:     Issuer{NIT: "900000001", DV: "0", Name: "ZU Seguridad 360 S.A.S. (pruebas)", Address: "Cra 1 # 2-3", City: "Bogotá", CityCode: "11001", Department: "Bogotá", DepartmentCode: "11", Email: "facturacion@zuseguridad360.com", TaxLevel: "R-99-PN", SoftwareID: "sw-test", SoftwarePIN: "12345", ProviderNIT: "900000001", ProviderDV: "0", Environment: "2"},
		Buyer:      Buyer{IDType: "13", ID: "1020304050", Name: "Profesional <SST> & Cía", Email: "qa.sst@zuseguridad360.com", IsPerson: true},
		Lines:      []Line{{Description: "Plan Profesional — octubre 2026", Quantity: 1, UnitPrice: 29900000, Subtotal: 29900000, IVA: 5681000}},
		Subtotal:   29900000, IVA: 5681000, Total: 35581000, Note: "Suscripción mensual",
	}
}

func TestBuildProducesWellFormedUBLWithCUFE(t *testing.T) {
	res, err := Build(sample())
	if err != nil {
		t.Fatal(err)
	}
	d := xml.NewDecoder(bytes.NewReader(res.XML))
	var uuid string
	inUUID := false
	for {
		tok, err := d.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("XML not well-formed: %v", err)
		}
		switch v := tok.(type) {
		case xml.StartElement:
			inUUID = v.Name.Local == "UUID"
		case xml.CharData:
			if inUUID {
				uuid += string(v)
			}
		case xml.EndElement:
			inUUID = false
		}
	}
	if uuid != res.CUFE || len(res.CUFE) != 96 {
		t.Fatalf("UUID %q != CUFE %q", uuid, res.CUFE)
	}
	if !strings.Contains(string(res.XML), "Profesional &lt;SST&gt; &amp; Cía") {
		t.Fatal("buyer name must be XML-escaped")
	}
	if !strings.Contains(res.QR, "catalogo-vpfe-hab.dian.gov.co") {
		t.Fatal("test environment QR must point to habilitación catalog")
	}
}

func TestBuildRejectsInconsistentTotals(t *testing.T) {
	inv := sample()
	inv.Total++
	if _, err := Build(inv); err == nil {
		t.Fatal("totals that do not add up must be rejected")
	}
}
