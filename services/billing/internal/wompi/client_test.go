package wompi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// Example from Wompi's integrity-signature documentation.
func TestIntegritySignatureMatchesWompiExample(t *testing.T) {
	got := IntegritySignature("sk8-438k4-xmxm392-sn2m", 2490000, "COP", "", "prod_integrity_Z5mMke9x0k8gpErbDqwrJXMqsI6SFli6")
	want := "37c8407747e595535433ef8f6a811d853cd943046624a0ec04662b17bbf33bf5"
	if got != want {
		t.Fatalf("integrity signature mismatch:\n got %s\nwant %s", got, want)
	}
}

func TestBaseURLFollowsTheKeys(t *testing.T) {
	if BaseURLFor("pub_test_x") != "https://sandbox.wompi.co/v1" || BaseURLFor("pub_prod_x") != "https://production.wompi.co/v1" {
		t.Fatal("environment must be derived from the key prefix")
	}
}

func TestCheckoutLinkIsSignedAndUnconfiguredRefuses(t *testing.T) {
	if _, err := NewClient(Config{}).CheckoutLink(CheckoutInput{Reference: "r", AmountInCents: 100}); err != ErrNotConfigured {
		t.Fatalf("expected ErrNotConfigured, got %v", err)
	}
	c := NewClient(Config{PublicKey: "pub_test_abc", IntegritySecret: "test_integrity_s"})
	link, err := c.CheckoutLink(CheckoutInput{Reference: "INV-1", AmountInCents: 11781000, RedirectURL: "https://www.zuseguridad360.com/dashboard/billing"})
	if err != nil {
		t.Fatal(err)
	}
	u, _ := url.Parse(link)
	q := u.Query()
	if q.Get("public-key") != "pub_test_abc" || q.Get("amount-in-cents") != "11781000" || q.Get("currency") != "COP" {
		t.Fatalf("bad checkout params: %v", q)
	}
	if q.Get("signature:integrity") != IntegritySignature("INV-1", 11781000, "COP", "", "test_integrity_s") {
		t.Fatal("checkout must carry the integrity signature")
	}
}

const event = `{"event":"transaction.updated","data":{"transaction":{"id":"1234-1610641025-49201","amount_in_cents":4490000,"reference":"MZQ3X2DE2SMX","customer_email":"juan.perez@gmail.com","currency":"COP","payment_method_type":"NEQUI","status":"APPROVED"}},"environment":"test","signature":{"properties":["transaction.id","transaction.status","transaction.amount_in_cents"],"checksum":"%s"},"timestamp":1530291411,"sent_at":"2018-07-20T16:45:05.000Z"}`

func TestVerifyEventWithNestedNumericProperties(t *testing.T) {
	sum := EventChecksum([]string{"1234-1610641025-49201", "APPROVED", "4490000"}, 1530291411, "test_events_secret")
	raw := []byte(strings.Replace(event, "%s", sum, 1))
	ev, err := VerifyEvent(raw, "test_events_secret")
	if err != nil {
		t.Fatalf("valid event rejected: %v", err)
	}
	tx, err := TransactionFromEvent(ev)
	if err != nil || tx.AmountInCents != 4490000 || tx.Reference != "MZQ3X2DE2SMX" {
		t.Fatalf("bad transaction: %+v %v", tx, err)
	}
	// Lower-case checksum is the same digest.
	if _, err := VerifyEvent([]byte(strings.Replace(event, "%s", strings.ToLower(sum), 1)), "test_events_secret"); err != nil {
		t.Fatal("checksum comparison must be case-insensitive")
	}
}

func TestVerifyEventRejectsTampering(t *testing.T) {
	sum := EventChecksum([]string{"1234-1610641025-49201", "APPROVED", "4490000"}, 1530291411, "test_events_secret")
	tampered := strings.Replace(strings.Replace(event, "%s", sum, 1), `"amount_in_cents":4490000`, `"amount_in_cents":100`, 1)
	if _, err := VerifyEvent([]byte(tampered), "test_events_secret"); err != ErrBadSignature {
		t.Fatalf("tampered amount must fail, got %v", err)
	}
	if _, err := VerifyEvent([]byte(strings.Replace(event, "%s", sum, 1)), "other"); err != ErrBadSignature {
		t.Fatal("wrong secret must fail")
	}
	if _, err := VerifyEvent([]byte(strings.Replace(event, "%s", sum, 1)), ""); err != ErrNotConfigured {
		t.Fatal("missing secret must refuse, never accept")
	}
}

func TestGetTransactionReadsAuthoritativeState(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/transactions/tx-1" {
			w.WriteHeader(404)
			return
		}
		_, _ = w.Write([]byte(`{"data":{"id":"tx-1","status":"APPROVED","reference":"INV-1","amount_in_cents":500,"currency":"COP"}}`))
	}))
	defer srv.Close()
	c := NewClient(Config{PublicKey: "pub_test_x", BaseURL: srv.URL})
	tx, err := c.GetTransaction(context.Background(), "tx-1")
	if err != nil || tx.Status != "APPROVED" || tx.AmountInCents != 500 {
		t.Fatalf("got %+v %v", tx, err)
	}
}
