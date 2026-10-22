//go:build integration

// Full payment chain against a real PostgreSQL (TEST_DATABASE_URL) and a
// local stand-in for Wompi's API (SIMULATED — not Wompi): invoice → checkout
// → signed webhook → reconciliation → paid → outbox → electronic invoice.
package service

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/zu-seguridad/billing/internal/dian"
	"github.com/zu-seguridad/billing/internal/model"
	"github.com/zu-seguridad/billing/internal/wompi"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const eventsSecret = "test_events_SIMULATED"

func TestPaymentChainEndToEnd(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&model.Subscription{}, &model.Invoice{}, &model.IdempotencyKey{}, &model.OutboxEvent{}); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	tenant := uuid.New()

	// SIMULATED Wompi API: answers whatever the "payer" did in checkout.
	var paid struct {
		id, ref string
		amount  int64
		status  string
	}
	fake := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/transactions/"+paid.id {
			w.WriteHeader(404)
			return
		}
		fmt.Fprintf(w, `{"data":{"id":%q,"status":%q,"reference":%q,"amount_in_cents":%d,"currency":"COP"}}`, paid.id, paid.status, paid.ref, paid.amount)
	}))
	defer fake.Close()

	w := wompi.NewClient(wompi.Config{PublicKey: "pub_test_SIMULATED", IntegritySecret: "test_integrity_SIMULATED", EventsSecret: eventsSecret, BaseURL: fake.URL})
	outbox := NewOutboxRelay(db, nil)
	d := DianConfig{
		Resolution: dian.Resolution{Number: "18760000001", Prefix: "SETP", From: 990000000, To: 995000000, ValidFrom: "2019-01-19", ValidTo: "2030-01-19", TechnicalKey: "fc8eac422eba16e22ffd8c6f94b3f40a6e38162c"},
		Issuer:     dian.Issuer{NIT: "900000001", DV: NITCheckDigit("900000001"), Name: "ZU QA (pruebas)", Environment: "2", SoftwareID: "sw", SoftwarePIN: "1"},
	}
	p := NewPayments(db, w, outbox, d, "https://www.zuseguridad360.com")
	if err := p.Migrate(); err != nil {
		t.Fatal(err)
	}
	if _, err := NewBillingService(db).CreateSubscription(ctx, tenant, model.PlanProfesional, "", false); err != nil {
		t.Fatal(err)
	}
	inv, err := p.IssueCurrentInvoice(ctx, tenant)
	if err != nil {
		t.Fatal(err)
	}
	again, _ := p.IssueCurrentInvoice(ctx, tenant)
	if again.ID != inv.ID {
		t.Fatal("one invoice per period")
	}
	if inv.Subtotal != 29900000 || inv.IVA != 5681000 || inv.Total != 35581000 {
		t.Fatalf("amounts: %+v", inv)
	}
	att, link, err := p.StartCheckout(ctx, tenant, inv.ID, "qa@x.co")
	if err != nil {
		t.Fatal(err)
	}
	u, _ := url.Parse(link)
	if u.Query().Get("amount-in-cents") != "35581000" {
		t.Fatalf("checkout amount: %s", link)
	}

	// The payer completes checkout; Wompi sends a signed webhook.
	paid.id, paid.ref, paid.amount, paid.status = "tx-sim-1", att.Reference, inv.Total, "APPROVED"
	ts := time.Now().Unix()
	sum := wompi.EventChecksum([]string{paid.id, paid.status, fmt.Sprint(paid.amount)}, ts, eventsSecret)
	body := fmt.Sprintf(`{"event":"transaction.updated","data":{"transaction":{"id":%q,"status":"APPROVED","reference":%q,"amount_in_cents":%d,"currency":"COP"}},"environment":"test","signature":{"properties":["transaction.id","transaction.status","transaction.amount_in_cents"],"checksum":%q},"timestamp":%d}`,
		paid.id, paid.ref, paid.amount, sum, ts)
	if err := p.HandleEvent(ctx, []byte(body), eventsSecret); err != nil {
		t.Fatal(err)
	}
	if err := p.HandleEvent(ctx, []byte(body), eventsSecret); err != nil { // replayed webhook
		t.Fatal(err)
	}
	var got model.Invoice
	db.First(&got, "id = ?", inv.ID)
	if got.Status != model.InvoiceStatusPaid || got.WompiTxID != "tx-sim-1" {
		t.Fatalf("invoice not paid: %+v", got)
	}
	var events []model.OutboxEvent
	db.Where("tenant_id = ?", tenant).Find(&events)
	if len(events) != 2 {
		t.Fatalf("expected exactly 2 outbox events (paid + notify) despite the replay, got %d", len(events))
	}

	// A tampered webhook claiming a different amount is rejected by signature.
	tampered := strings.Replace(body, fmt.Sprint(`"amount_in_cents":`, paid.amount), `"amount_in_cents":1`, 1)
	if err := p.HandleEvent(ctx, []byte(tampered), eventsSecret); err != wompi.ErrBadSignature {
		t.Fatalf("tampered event: %v", err)
	}

	// Electronic invoice: without fiscal profile it is recorded as such.
	e, err := p.IssueElectronic(ctx, inv.ID)
	if err != nil || e.Status != "sin_perfil_fiscal" {
		t.Fatalf("expected sin_perfil_fiscal, got %+v %v", e, err)
	}
	if err := p.SaveProfile(ctx, &model.BillingProfile{TenantID: tenant, IDType: "13", IDNumber: "1020304050", LegalName: "QA Profesional", Email: "qa@x.co"}); err != nil {
		t.Fatal(err)
	}
	e, err = p.IssueElectronic(ctx, inv.ID)
	if err != nil || e.Status != "pendiente_transmision" || len(e.CUFE) != 96 || !strings.HasPrefix(e.Number, "SETP99") {
		t.Fatalf("electronic invoice: %+v %v", e, err)
	}
	e2, _ := p.IssueElectronic(ctx, inv.ID)
	if e2.Number != e.Number {
		t.Fatal("issuing twice must return the same document")
	}
	t.Logf("factura %s CUFE %s…", e.Number, e.CUFE[:16])
}

func TestNITCheckDigit(t *testing.T) {
	// DIAN's own NIT 800197268-4.
	if NITCheckDigit("800197268") != "4" {
		t.Fatal("DV for 800197268 must be 4")
	}
}
