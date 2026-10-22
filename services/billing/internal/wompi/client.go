// Package wompi integrates Wompi (Bancolombia) payments: Web Checkout links
// signed with the integrity secret, transaction lookups for reconciliation,
// and verification of signed webhook events.
//
// Environment is chosen by the keys alone: pub_test_/prv_test_ keys talk to
// https://sandbox.wompi.co/v1, pub_prod_/prv_prod_ to https://production.wompi.co/v1.
// Going live is therefore only a change of the four secrets in Vault.
package wompi

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var ErrNotConfigured = errors.New("wompi is not configured (public key and integrity secret are required)")

const CheckoutURL = "https://checkout.wompi.co/p/"

type Config struct {
	PublicKey       string
	PrivateKey      string
	EventsSecret    string
	IntegritySecret string
	BaseURL         string // optional override (tests); derived from the keys otherwise
}

type Client struct {
	cfg  Config
	http *http.Client
}

func NewClient(cfg Config) *Client {
	if cfg.BaseURL == "" {
		cfg.BaseURL = BaseURLFor(cfg.PublicKey)
	}
	return &Client{cfg: cfg, http: &http.Client{Timeout: 15 * time.Second}}
}

// BaseURLFor maps a key to its Wompi environment.
func BaseURLFor(key string) string {
	if strings.HasPrefix(key, "pub_prod_") || strings.HasPrefix(key, "prv_prod_") {
		return "https://production.wompi.co/v1"
	}
	return "https://sandbox.wompi.co/v1"
}

func (c *Client) Environment() string {
	switch {
	case c.cfg.PublicKey == "":
		return "unconfigured"
	case strings.HasPrefix(c.cfg.PublicKey, "pub_prod_"):
		return "production"
	default:
		return "sandbox"
	}
}

func (c *Client) Configured() bool { return c.cfg.PublicKey != "" && c.cfg.IntegritySecret != "" }

// IntegritySignature: SHA-256 of reference + amount_in_cents + currency
// (+ expiration time when used) + integrity secret, as Wompi requires on
// every checkout so the amount cannot be altered in the browser.
func IntegritySignature(reference string, amountInCents int64, currency, expiration, secret string) string {
	sum := sha256.Sum256([]byte(reference + strconv.FormatInt(amountInCents, 10) + currency + expiration + secret))
	return hex.EncodeToString(sum[:])
}

type CheckoutInput struct {
	Reference     string
	AmountInCents int64
	CustomerEmail string
	RedirectURL   string
}

// CheckoutLink builds a Web Checkout URL. Card data never touches our servers.
func (c *Client) CheckoutLink(in CheckoutInput) (string, error) {
	if !c.Configured() {
		return "", ErrNotConfigured
	}
	q := url.Values{}
	q.Set("public-key", c.cfg.PublicKey)
	q.Set("currency", "COP")
	q.Set("amount-in-cents", strconv.FormatInt(in.AmountInCents, 10))
	q.Set("reference", in.Reference)
	q.Set("signature:integrity", IntegritySignature(in.Reference, in.AmountInCents, "COP", "", c.cfg.IntegritySecret))
	if in.RedirectURL != "" {
		q.Set("redirect-url", in.RedirectURL)
	}
	if in.CustomerEmail != "" {
		q.Set("customer-data:email", in.CustomerEmail)
	}
	return CheckoutURL + "?" + q.Encode(), nil
}

type Transaction struct {
	ID            string `json:"id"`
	Status        string `json:"status"` // PENDING | APPROVED | DECLINED | VOIDED | ERROR
	Reference     string `json:"reference"`
	AmountInCents int64  `json:"amount_in_cents"`
	Currency      string `json:"currency"`
	PaymentMethod string `json:"payment_method_type"`
	StatusMessage string `json:"status_message"`
	CustomerEmail string `json:"customer_email"`
	FinalizedAt   string `json:"finalized_at"`
}

// GetTransaction asks Wompi for the authoritative state of a transaction.
// Reconciliation never trusts a webhook body alone: amounts and status are
// re-read from Wompi before an invoice is marked paid.
func (c *Client) GetTransaction(ctx context.Context, id string) (*Transaction, error) {
	var out struct {
		Data Transaction `json:"data"`
	}
	if err := c.get(ctx, "/transactions/"+url.PathEscape(id), c.cfg.PublicKey, &out); err != nil {
		return nil, err
	}
	return &out.Data, nil
}

// FindByReference lists transactions for a reference (needs the private key).
func (c *Client) FindByReference(ctx context.Context, reference string) ([]Transaction, error) {
	if c.cfg.PrivateKey == "" {
		return nil, ErrNotConfigured
	}
	var out struct {
		Data []Transaction `json:"data"`
	}
	err := c.get(ctx, "/transactions?reference="+url.QueryEscape(reference), c.cfg.PrivateKey, &out)
	return out.Data, err
}

func (c *Client) get(ctx context.Context, path, bearer string, into interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.cfg.BaseURL+path, nil)
	if err != nil {
		return err
	}
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("wompi unreachable: %w", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode >= 300 {
		return fmt.Errorf("wompi %s returned %d: %s", path, resp.StatusCode, truncate(string(raw), 300))
	}
	return json.Unmarshal(raw, into)
}

// Event is a Wompi webhook event (transaction.updated, ...).
type Event struct {
	Event       string                 `json:"event"`
	Data        map[string]interface{} `json:"data"`
	Environment string                 `json:"environment"`
	Signature   struct {
		Properties []string `json:"properties"`
		Checksum   string   `json:"checksum"`
	} `json:"signature"`
	Timestamp int64  `json:"timestamp"`
	SentAt    string `json:"sent_at"`
}

var ErrBadSignature = errors.New("invalid wompi event signature")

// VerifyEvent checks the event checksum: SHA-256 over the values of
// signature.properties (paths inside data, e.g. "transaction.amount_in_cents"),
// then the timestamp, then the events secret.
func VerifyEvent(raw []byte, eventsSecret string) (*Event, error) {
	if eventsSecret == "" {
		return nil, ErrNotConfigured
	}
	var ev Event
	dec := json.NewDecoder(strings.NewReader(string(raw)))
	dec.UseNumber() // keep amounts exactly as sent ("2500000", not "2.5e+06")
	if err := dec.Decode(&ev); err != nil {
		return nil, fmt.Errorf("invalid event payload: %w", err)
	}
	if len(ev.Signature.Properties) == 0 || ev.Signature.Checksum == "" {
		return nil, ErrBadSignature
	}
	var b strings.Builder
	for _, p := range ev.Signature.Properties {
		v, ok := lookup(ev.Data, p)
		if !ok {
			return nil, ErrBadSignature
		}
		b.WriteString(v)
	}
	b.WriteString(strconv.FormatInt(ev.Timestamp, 10))
	b.WriteString(eventsSecret)
	sum := sha256.Sum256([]byte(b.String()))
	if !strings.EqualFold(hex.EncodeToString(sum[:]), ev.Signature.Checksum) {
		return nil, ErrBadSignature
	}
	return &ev, nil
}

// EventChecksum builds a valid checksum (used by tests and the local simulator).
func EventChecksum(values []string, timestamp int64, secret string) string {
	sum := sha256.Sum256([]byte(strings.Join(values, "") + strconv.FormatInt(timestamp, 10) + secret))
	return strings.ToUpper(hex.EncodeToString(sum[:]))
}

func lookup(m map[string]interface{}, path string) (string, bool) {
	var cur interface{} = m
	for _, part := range strings.Split(path, ".") {
		obj, ok := cur.(map[string]interface{})
		if !ok {
			return "", false
		}
		cur, ok = obj[part]
		if !ok {
			return "", false
		}
	}
	switch v := cur.(type) {
	case string:
		return v, true
	case json.Number:
		return v.String(), true
	case bool:
		return strconv.FormatBool(v), true
	case nil:
		return "", true
	}
	return "", false
}

// TransactionFromEvent extracts data.transaction from a verified event.
func TransactionFromEvent(ev *Event) (*Transaction, error) {
	raw, err := json.Marshal(ev.Data["transaction"])
	if err != nil {
		return nil, err
	}
	var t Transaction
	if err := json.Unmarshal(raw, &t); err != nil || t.ID == "" {
		return nil, errors.New("event has no transaction")
	}
	return &t, nil
}

func truncate(s string, n int) string {
	if len(s) > n {
		return s[:n]
	}
	return s
}
