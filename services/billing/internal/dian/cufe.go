// Package dian builds Colombian electronic invoices (factura electrónica de
// venta, UBL 2.1 per DIAN's Anexo Técnico): consecutive numbering inside an
// authorized resolution range, the CUFE, the QR payload and the XML.
//
// What it does NOT do yet, and why: the XML must be signed (XAdES-EPES) with
// the issuer's digital certificate and transmitted to DIAN's web service as an
// authorized "facturador electrónico" (habilitación with a TestSetId, or
// through a proveedor tecnológico). Those depend on credentials only the
// business can obtain; see Transmitter. Until then a generated invoice stays
// in status "pendiente_transmision" — it is never reported as accepted.
package dian

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"time"
)

type CUFEInput struct {
	Number       string    // prefijo + consecutivo, e.g. SETP990000002
	IssuedAt     time.Time // fecha y hora de expedición (con zona horaria)
	Subtotal     int64     // valor bruto antes de impuestos, en centavos
	IVA          int64     // código 01
	INC          int64     // código 04 (impuesto al consumo)
	ICA          int64     // código 03
	Total        int64     // valor total a pagar, en centavos
	IssuerNIT    string    // NIT del facturador, sin DV
	BuyerID      string    // documento del adquirente
	TechnicalKey string    // clave técnica de la resolución (ClTec)
	Environment  string    // 1 = producción, 2 = pruebas
}

// money renders centavos with exactly two decimals, as the Anexo requires.
func money(c int64) string { return fmt.Sprintf("%d.%02d", c/100, c%100) }

// CUFE = SHA-384(NumFac + FecFac + HorFac + ValFac + 01 + ValImp1 + 04 +
// ValImp2 + 03 + ValImp3 + ValTot + NitOFE + NumAdq + ClTec + TipoAmbiente).
func CUFE(in CUFEInput) string {
	s := in.Number +
		in.IssuedAt.Format("2006-01-02") +
		in.IssuedAt.Format("15:04:05-07:00") +
		money(in.Subtotal) +
		"01" + money(in.IVA) +
		"04" + money(in.INC) +
		"03" + money(in.ICA) +
		money(in.Total) +
		in.IssuerNIT + in.BuyerID + in.TechnicalKey + in.Environment
	sum := sha512.Sum384([]byte(s))
	return hex.EncodeToString(sum[:])
}

// CUFESource exposes the concatenated string (for audit/debug of a CUFE).
func CUFESource(in CUFEInput) string {
	return in.Number + in.IssuedAt.Format("2006-01-02") + in.IssuedAt.Format("15:04:05-07:00") +
		money(in.Subtotal) + "01" + money(in.IVA) + "04" + money(in.INC) + "03" + money(in.ICA) +
		money(in.Total) + in.IssuerNIT + in.BuyerID + in.TechnicalKey + in.Environment
}
