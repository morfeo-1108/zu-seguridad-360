package dian

import (
	"bytes"
	"crypto/sha512"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"strings"
	"text/template"
	"time"
)

// Resolution is the DIAN numbering authorization (resolución de facturación).
type Resolution struct {
	Number       string // número de resolución
	Prefix       string
	From, To     int64
	ValidFrom    string // YYYY-MM-DD
	ValidTo      string
	TechnicalKey string // ClTec
}

// Issuer is ZU360's own legal identity as invoicing party.
type Issuer struct {
	NIT, DV, Name, Address, City, CityCode, Department, DepartmentCode, Email string
	TaxLevel                                                                  string // responsabilidades fiscales, e.g. "O-13" or "R-99-PN"
	SoftwareID, SoftwarePIN, ProviderNIT, ProviderDV                          string
	Environment                                                               string // 1 producción, 2 pruebas
}

// Buyer is the customer (adquirente): the SST professional or their firm.
type Buyer struct {
	IDType   string // 31 = NIT, 13 = cédula de ciudadanía
	ID, DV   string
	Name     string
	Email    string
	Address  string
	City     string
	IsPerson bool // persona natural (2) vs jurídica (1)
}

type Line struct {
	Description string
	Quantity    int64
	UnitPrice   int64 // centavos
	Subtotal    int64 // centavos
	IVA         int64 // centavos (19%)
}

type Invoice struct {
	Number     string
	IssuedAt   time.Time
	DueAt      time.Time
	Resolution Resolution
	Issuer     Issuer
	Buyer      Buyer
	Lines      []Line
	Subtotal   int64
	IVA        int64
	Total      int64
	Note       string
}

type Result struct {
	CUFE   string
	QR     string
	XML    []byte
	Source string // CUFE source string, kept for audit
}

// SoftwareSecurityCode = SHA-384(SoftwareID + PIN + NumFac).
func SoftwareSecurityCode(softwareID, pin, number string) string {
	sum := sha512.Sum384([]byte(softwareID + pin + number))
	return hex.EncodeToString(sum[:])
}

func QRURL(env, cufe string) string {
	host := "catalogo-vpfe.dian.gov.co"
	if env != "1" {
		host = "catalogo-vpfe-hab.dian.gov.co"
	}
	return "https://" + host + "/document/searchqr?documentkey=" + cufe
}

// Build computes CUFE and QR and renders the unsigned UBL 2.1 XML.
func Build(inv Invoice) (*Result, error) {
	if inv.Number == "" || inv.Issuer.NIT == "" || inv.Buyer.ID == "" || len(inv.Lines) == 0 {
		return nil, fmt.Errorf("invoice incomplete: number, issuer NIT, buyer ID and lines are required")
	}
	var sub, iva int64
	for _, l := range inv.Lines {
		sub += l.Subtotal
		iva += l.IVA
	}
	if sub != inv.Subtotal || iva != inv.IVA || inv.Subtotal+inv.IVA != inv.Total {
		return nil, fmt.Errorf("totals do not add up (lines %d+%d, invoice %d+%d=%d)", sub, iva, inv.Subtotal, inv.IVA, inv.Total)
	}
	cin := CUFEInput{Number: inv.Number, IssuedAt: inv.IssuedAt, Subtotal: inv.Subtotal, IVA: inv.IVA, Total: inv.Total,
		IssuerNIT: inv.Issuer.NIT, BuyerID: inv.Buyer.ID, TechnicalKey: inv.Resolution.TechnicalKey, Environment: inv.Issuer.Environment}
	cufe := CUFE(cin)
	qr := strings.Join([]string{
		"NumFac: " + inv.Number,
		"FecFac: " + inv.IssuedAt.Format("2006-01-02"),
		"HorFac: " + inv.IssuedAt.Format("15:04:05-07:00"),
		"NitFac: " + inv.Issuer.NIT,
		"DocAdq: " + inv.Buyer.ID,
		"ValFac: " + money(inv.Subtotal),
		"ValIva: " + money(inv.IVA),
		"ValOtroIm: 0.00",
		"ValTolFac: " + money(inv.Total),
		"CUFE: " + cufe,
		QRURL(inv.Issuer.Environment, cufe),
	}, "\n")
	var buf bytes.Buffer
	err := ublTemplate.Execute(&buf, map[string]interface{}{
		"I": inv, "CUFE": cufe, "QR": qr,
		"SSC":  SoftwareSecurityCode(inv.Issuer.SoftwareID, inv.Issuer.SoftwarePIN, inv.Number),
		"Date": inv.IssuedAt.Format("2006-01-02"), "Time": inv.IssuedAt.Format("15:04:05-07:00"),
		"Due": inv.DueAt.Format("2006-01-02"),
	})
	if err != nil {
		return nil, err
	}
	return &Result{CUFE: cufe, QR: qr, XML: buf.Bytes(), Source: CUFESource(cin)}, nil
}

func esc(s string) string {
	var b bytes.Buffer
	_ = xml.EscapeText(&b, []byte(s)) // writes to a bytes.Buffer cannot fail
	return b.String()
}

func orgCode(isPerson bool) string {
	if isPerson {
		return "2"
	}
	return "1"
}

var ublTemplate = template.Must(template.New("ubl").Funcs(template.FuncMap{"x": esc, "m": money, "org": orgCode, "inc": func(i int) int { return i + 1 }}).Parse(`<?xml version="1.0" encoding="UTF-8" standalone="no"?>
<Invoice xmlns="urn:oasis:names:specification:ubl:schema:xsd:Invoice-2"
 xmlns:cac="urn:oasis:names:specification:ubl:schema:xsd:CommonAggregateComponents-2"
 xmlns:cbc="urn:oasis:names:specification:ubl:schema:xsd:CommonBasicComponents-2"
 xmlns:ext="urn:oasis:names:specification:ubl:schema:xsd:CommonExtensionComponents-2"
 xmlns:sts="dian:gov:co:facturaelectronica:Structures-2-1"
 xmlns:ds="http://www.w3.org/2000/09/xmldsig#">
 <ext:UBLExtensions>
  <ext:UBLExtension><ext:ExtensionContent><sts:DianExtensions>
   <sts:InvoiceControl>
    <sts:InvoiceAuthorization>{{x .I.Resolution.Number}}</sts:InvoiceAuthorization>
    <sts:AuthorizationPeriod><cbc:StartDate>{{x .I.Resolution.ValidFrom}}</cbc:StartDate><cbc:EndDate>{{x .I.Resolution.ValidTo}}</cbc:EndDate></sts:AuthorizationPeriod>
    <sts:AuthorizedInvoices><sts:Prefix>{{x .I.Resolution.Prefix}}</sts:Prefix><sts:From>{{.I.Resolution.From}}</sts:From><sts:To>{{.I.Resolution.To}}</sts:To></sts:AuthorizedInvoices>
   </sts:InvoiceControl>
   <sts:InvoiceSource><cbc:IdentificationCode listAgencyID="6" listAgencyName="United Nations Economic Commission for Europe" listSchemeURI="urn:oasis:names:specification:ubl:codelist:gc:CountryIdentificationCode-2.1">CO</cbc:IdentificationCode></sts:InvoiceSource>
   <sts:SoftwareProvider>
    <sts:ProviderID schemeAgencyID="195" schemeAgencyName="CO, DIAN (Dirección de Impuestos y Aduanas Nacionales)" schemeID="{{x .I.Issuer.ProviderDV}}" schemeName="31">{{x .I.Issuer.ProviderNIT}}</sts:ProviderID>
    <sts:SoftwareID schemeAgencyID="195" schemeAgencyName="CO, DIAN (Dirección de Impuestos y Aduanas Nacionales)">{{x .I.Issuer.SoftwareID}}</sts:SoftwareID>
   </sts:SoftwareProvider>
   <sts:SoftwareSecurityCode schemeAgencyID="195" schemeAgencyName="CO, DIAN (Dirección de Impuestos y Aduanas Nacionales)">{{.SSC}}</sts:SoftwareSecurityCode>
   <sts:AuthorizationProvider><sts:AuthorizationProviderID schemeAgencyID="195" schemeAgencyName="CO, DIAN (Dirección de Impuestos y Aduanas Nacionales)" schemeID="4" schemeName="31">800197268</sts:AuthorizationProviderID></sts:AuthorizationProvider>
   <sts:QRCode>{{x .QR}}</sts:QRCode>
  </sts:DianExtensions></ext:ExtensionContent></ext:UBLExtension>
  <ext:UBLExtension><ext:ExtensionContent><!-- Firma XAdES-EPES: pendiente (requiere certificado digital del facturador) --></ext:ExtensionContent></ext:UBLExtension>
 </ext:UBLExtensions>
 <cbc:UBLVersionID>UBL 2.1</cbc:UBLVersionID>
 <cbc:CustomizationID>10</cbc:CustomizationID>
 <cbc:ProfileID>DIAN 2.1: Factura Electrónica de Venta</cbc:ProfileID>
 <cbc:ProfileExecutionID>{{x .I.Issuer.Environment}}</cbc:ProfileExecutionID>
 <cbc:ID>{{x .I.Number}}</cbc:ID>
 <cbc:UUID schemeID="{{x .I.Issuer.Environment}}" schemeName="CUFE-SHA384">{{.CUFE}}</cbc:UUID>
 <cbc:IssueDate>{{.Date}}</cbc:IssueDate>
 <cbc:IssueTime>{{.Time}}</cbc:IssueTime>
 <cbc:DueDate>{{.Due}}</cbc:DueDate>
 <cbc:InvoiceTypeCode>01</cbc:InvoiceTypeCode>
 <cbc:Note>{{x .I.Note}}</cbc:Note>
 <cbc:DocumentCurrencyCode>COP</cbc:DocumentCurrencyCode>
 <cbc:LineCountNumeric>{{len .I.Lines}}</cbc:LineCountNumeric>
 <cac:AccountingSupplierParty>
  <cbc:AdditionalAccountID>1</cbc:AdditionalAccountID>
  <cac:Party>
   <cac:PartyName><cbc:Name>{{x .I.Issuer.Name}}</cbc:Name></cac:PartyName>
   <cac:PhysicalLocation><cac:Address><cbc:ID>{{x .I.Issuer.CityCode}}</cbc:ID><cbc:CityName>{{x .I.Issuer.City}}</cbc:CityName><cbc:CountrySubentity>{{x .I.Issuer.Department}}</cbc:CountrySubentity><cbc:CountrySubentityCode>{{x .I.Issuer.DepartmentCode}}</cbc:CountrySubentityCode><cac:AddressLine><cbc:Line>{{x .I.Issuer.Address}}</cbc:Line></cac:AddressLine><cac:Country><cbc:IdentificationCode>CO</cbc:IdentificationCode><cbc:Name languageID="es">Colombia</cbc:Name></cac:Country></cac:Address></cac:PhysicalLocation>
   <cac:PartyTaxScheme>
    <cbc:RegistrationName>{{x .I.Issuer.Name}}</cbc:RegistrationName>
    <cbc:CompanyID schemeAgencyID="195" schemeAgencyName="CO, DIAN (Dirección de Impuestos y Aduanas Nacionales)" schemeID="{{x .I.Issuer.DV}}" schemeName="31">{{x .I.Issuer.NIT}}</cbc:CompanyID>
    <cbc:TaxLevelCode listName="48">{{x .I.Issuer.TaxLevel}}</cbc:TaxLevelCode>
    <cac:TaxScheme><cbc:ID>01</cbc:ID><cbc:Name>IVA</cbc:Name></cac:TaxScheme>
   </cac:PartyTaxScheme>
   <cac:PartyLegalEntity><cbc:RegistrationName>{{x .I.Issuer.Name}}</cbc:RegistrationName><cbc:CompanyID schemeAgencyID="195" schemeID="{{x .I.Issuer.DV}}" schemeName="31">{{x .I.Issuer.NIT}}</cbc:CompanyID><cac:CorporateRegistrationScheme><cbc:ID>{{x .I.Resolution.Prefix}}</cbc:ID></cac:CorporateRegistrationScheme></cac:PartyLegalEntity>
   <cac:Contact><cbc:ElectronicMail>{{x .I.Issuer.Email}}</cbc:ElectronicMail></cac:Contact>
  </cac:Party>
 </cac:AccountingSupplierParty>
 <cac:AccountingCustomerParty>
  <cbc:AdditionalAccountID>{{org .I.Buyer.IsPerson}}</cbc:AdditionalAccountID>
  <cac:Party>
   <cac:PartyName><cbc:Name>{{x .I.Buyer.Name}}</cbc:Name></cac:PartyName>
   <cac:PartyTaxScheme>
    <cbc:RegistrationName>{{x .I.Buyer.Name}}</cbc:RegistrationName>
    <cbc:CompanyID schemeAgencyID="195" schemeAgencyName="CO, DIAN (Dirección de Impuestos y Aduanas Nacionales)"{{if .I.Buyer.DV}} schemeID="{{x .I.Buyer.DV}}"{{end}} schemeName="{{x .I.Buyer.IDType}}">{{x .I.Buyer.ID}}</cbc:CompanyID>
    <cbc:TaxLevelCode listName="48">R-99-PN</cbc:TaxLevelCode>
    <cac:TaxScheme><cbc:ID>ZZ</cbc:ID><cbc:Name>No aplica</cbc:Name></cac:TaxScheme>
   </cac:PartyTaxScheme>
   <cac:PartyLegalEntity><cbc:RegistrationName>{{x .I.Buyer.Name}}</cbc:RegistrationName><cbc:CompanyID schemeAgencyID="195"{{if .I.Buyer.DV}} schemeID="{{x .I.Buyer.DV}}"{{end}} schemeName="{{x .I.Buyer.IDType}}">{{x .I.Buyer.ID}}</cbc:CompanyID></cac:PartyLegalEntity>
   <cac:Contact><cbc:ElectronicMail>{{x .I.Buyer.Email}}</cbc:ElectronicMail></cac:Contact>
  </cac:Party>
 </cac:AccountingCustomerParty>
 <cac:PaymentMeans><cbc:ID>1</cbc:ID><cbc:PaymentMeansCode>48</cbc:PaymentMeansCode><cbc:PaymentDueDate>{{.Due}}</cbc:PaymentDueDate></cac:PaymentMeans>
 <cac:TaxTotal>
  <cbc:TaxAmount currencyID="COP">{{m .I.IVA}}</cbc:TaxAmount>
  <cac:TaxSubtotal><cbc:TaxableAmount currencyID="COP">{{m .I.Subtotal}}</cbc:TaxableAmount><cbc:TaxAmount currencyID="COP">{{m .I.IVA}}</cbc:TaxAmount><cac:TaxCategory><cbc:Percent>19.00</cbc:Percent><cac:TaxScheme><cbc:ID>01</cbc:ID><cbc:Name>IVA</cbc:Name></cac:TaxScheme></cac:TaxCategory></cac:TaxSubtotal>
 </cac:TaxTotal>
 <cac:LegalMonetaryTotal>
  <cbc:LineExtensionAmount currencyID="COP">{{m .I.Subtotal}}</cbc:LineExtensionAmount>
  <cbc:TaxExclusiveAmount currencyID="COP">{{m .I.Subtotal}}</cbc:TaxExclusiveAmount>
  <cbc:TaxInclusiveAmount currencyID="COP">{{m .I.Total}}</cbc:TaxInclusiveAmount>
  <cbc:PayableAmount currencyID="COP">{{m .I.Total}}</cbc:PayableAmount>
 </cac:LegalMonetaryTotal>
{{range $i, $l := .I.Lines}} <cac:InvoiceLine>
  <cbc:ID>{{inc $i}}</cbc:ID>
  <cbc:InvoicedQuantity unitCode="94">{{$l.Quantity}}</cbc:InvoicedQuantity>
  <cbc:LineExtensionAmount currencyID="COP">{{m $l.Subtotal}}</cbc:LineExtensionAmount>
  <cac:TaxTotal><cbc:TaxAmount currencyID="COP">{{m $l.IVA}}</cbc:TaxAmount><cac:TaxSubtotal><cbc:TaxableAmount currencyID="COP">{{m $l.Subtotal}}</cbc:TaxableAmount><cbc:TaxAmount currencyID="COP">{{m $l.IVA}}</cbc:TaxAmount><cac:TaxCategory><cbc:Percent>19.00</cbc:Percent><cac:TaxScheme><cbc:ID>01</cbc:ID><cbc:Name>IVA</cbc:Name></cac:TaxScheme></cac:TaxCategory></cac:TaxSubtotal></cac:TaxTotal>
  <cac:Item><cbc:Description>{{x $l.Description}}</cbc:Description><cac:StandardItemIdentification><cbc:ID schemeID="999">ZU360-SUSCRIPCION</cbc:ID></cac:StandardItemIdentification></cac:Item>
  <cac:Price><cbc:PriceAmount currencyID="COP">{{m $l.UnitPrice}}</cbc:PriceAmount><cbc:BaseQuantity unitCode="94">1</cbc:BaseQuantity></cac:Price>
 </cac:InvoiceLine>
{{end}}</Invoice>
`))
