// Package wompi is a thin client for Wompi's Payments API
// (https://docs.wompi.co). It is scaffolding: shaped after Wompi's publicly
// documented transaction-creation contract, but has never been exercised
// against Wompi's sandbox or production endpoints (no credentials exist in
// this environment to do so). Before taking a real charge, verify the
// request/response fields here against Wompi's current API reference —
// field names and required values do change between API versions.
package wompi

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

var ErrNotConfigured = errors.New("wompi is not configured (WOMPI_KEY is empty) — refusing to attempt a real charge")

type Client struct {
	baseURL    string
	privateKey string
	httpClient *http.Client
}

func NewClient(baseURL, privateKey string) *Client {
	return &Client{baseURL: baseURL, privateKey: privateKey, httpClient: &http.Client{Timeout: 15 * time.Second}}
}

type CreateTransactionInput struct {
	AmountInCents   int64                  `json:"amount_in_cents"`
	Currency        string                 `json:"currency"` // "COP"
	CustomerEmail   string                 `json:"customer_email"`
	Reference       string                 `json:"reference"` // must be unique per attempt — use the invoice ID
	PaymentMethod   map[string]interface{} `json:"payment_method"`
	AcceptanceToken string                 `json:"acceptance_token"`
}

type TransactionResponse struct {
	Data struct {
		ID        string `json:"id"`
		Status    string `json:"status"` // PENDING | APPROVED | DECLINED | ERROR
		Reference string `json:"reference"`
		AmountIn  int64  `json:"amount_in_cents"`
	} `json:"data"`
}

// CreateTransaction attempts a real charge. It never fabricates a response —
// a misconfigured client returns ErrNotConfigured instead of pretending to
// have charged anything.
func (c *Client) CreateTransaction(ctx context.Context, input CreateTransactionInput) (*TransactionResponse, error) {
	if c.privateKey == "" {
		return nil, ErrNotConfigured
	}
	body, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("wompi request encode failed: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/transactions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("wompi request build failed: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.privateKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("wompi API unreachable: %w", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("wompi API returned %d: %s", resp.StatusCode, string(raw))
	}
	var out TransactionResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("wompi response decode failed: %w", err)
	}
	return &out, nil
}

// VerifyEventSignature checks a Wompi webhook's checksum. Per Wompi's
// documented scheme: concatenate the values of the properties listed in
// event.signature.properties (in order), append the timestamp, append the
// events secret, then SHA-256 the result and compare hex digests.
// Unverified against a live webhook — the exact property-path resolution
// (e.g. "transaction.id") needs confirmation against a real payload before
// this gates any real money movement.
func VerifyEventSignature(propertyValues []string, timestamp int64, checksum, eventsSecret string) bool {
	if eventsSecret == "" {
		return false
	}
	concat := ""
	for _, v := range propertyValues {
		concat += v
	}
	concat += fmt.Sprintf("%d", timestamp)
	concat += eventsSecret
	sum := sha256.Sum256([]byte(concat))
	return hex.EncodeToString(sum[:]) == checksum
}
