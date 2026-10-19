package wompi

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateTransactionRefusesWithoutPrivateKey(t *testing.T) {
	c := NewClient("https://example.test", "")
	_, err := c.CreateTransaction(context.Background(), CreateTransactionInput{})
	if err != ErrNotConfigured {
		t.Fatalf("expected ErrNotConfigured, got %v", err)
	}
}

func TestCreateTransactionSendsRealRequestWhenConfigured(t *testing.T) {
	var gotAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"data":{"id":"tx_1","status":"PENDING","reference":"inv-1","amount_in_cents":100000}}`))
	}))
	defer server.Close()

	c := NewClient(server.URL, "prv_test_key")
	resp, err := c.CreateTransaction(context.Background(), CreateTransactionInput{AmountInCents: 100000, Currency: "COP", Reference: "inv-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data.Status != "PENDING" {
		t.Fatalf("expected PENDING status, got %q", resp.Data.Status)
	}
	if gotAuth != "Bearer prv_test_key" {
		t.Fatalf("expected private key forwarded as bearer token, got %q", gotAuth)
	}
}

func TestVerifyEventSignatureMatchesKnownVector(t *testing.T) {
	values := []string{"tx_1", "APPROVED", "100000"}
	timestamp := int64(1700000000)
	secret := "events-secret"
	concat := ""
	for _, v := range values {
		concat += v
	}
	concat += fmt.Sprintf("%d", timestamp)
	concat += secret
	sum := sha256.Sum256([]byte(concat))
	checksum := hex.EncodeToString(sum[:])

	if !VerifyEventSignature(values, timestamp, checksum, secret) {
		t.Fatal("expected signature to verify against its own checksum computation")
	}
	if VerifyEventSignature(values, timestamp, "wrong", secret) {
		t.Fatal("expected mismatched checksum to fail verification")
	}
	if VerifyEventSignature(values, timestamp, checksum, "") {
		t.Fatal("expected empty events secret to fail verification, not pass trivially")
	}
}
