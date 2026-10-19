//go:build integration
// +build integration

package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/zu-seguridad/billing/internal/model"
	"github.com/zu-seguridad/billing/internal/wompi"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupInvoiceTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("could not open test DB: %v", err)
	}
	if err := db.AutoMigrate(&model.Invoice{}); err != nil {
		t.Fatalf("migration failed: %v", err)
	}
	return db
}

func TestChargeInvoiceMarksPaidOnApproved(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"data":{"id":"tx_approved","status":"APPROVED","reference":"x","amount_in_cents":119000}}`))
	}))
	defer server.Close()

	db := setupInvoiceTestDB(t)
	invoice := &model.Invoice{TenantID: uuid.New(), Status: model.InvoiceStatusPending, Total: 119000}
	if err := db.Create(invoice).Error; err != nil {
		t.Fatalf("seed invoice failed: %v", err)
	}

	svc := NewBillingService(db, wompi.NewClient(server.URL, "prv_test"))
	if err := svc.ChargeInvoice(context.Background(), invoice.ID, nil, "tok"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var updated model.Invoice
	db.First(&updated, "id = ?", invoice.ID)
	if updated.Status != model.InvoiceStatusPaid {
		t.Fatalf("expected status paid, got %s", updated.Status)
	}
	if updated.WompiTxID != "tx_approved" {
		t.Fatalf("expected wompi tx id recorded, got %q", updated.WompiTxID)
	}
	if updated.PaidAt == nil {
		t.Fatal("expected paid_at to be set")
	}
}

func TestChargeInvoiceMarksOverdueOnDeclined(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"data":{"id":"tx_declined","status":"DECLINED","reference":"x","amount_in_cents":119000}}`))
	}))
	defer server.Close()

	db := setupInvoiceTestDB(t)
	invoice := &model.Invoice{TenantID: uuid.New(), Status: model.InvoiceStatusPending, Total: 119000}
	db.Create(invoice)

	svc := NewBillingService(db, wompi.NewClient(server.URL, "prv_test"))
	if err := svc.ChargeInvoice(context.Background(), invoice.ID, nil, "tok"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var updated model.Invoice
	db.First(&updated, "id = ?", invoice.ID)
	if updated.Status != model.InvoiceStatusOverdue {
		t.Fatalf("expected status overdue on decline, got %s", updated.Status)
	}
	if updated.PaidAt != nil {
		t.Fatal("a declined charge must never set paid_at")
	}
}

func TestChargeInvoiceFailsWithoutWompiCredentials(t *testing.T) {
	db := setupInvoiceTestDB(t)
	invoice := &model.Invoice{TenantID: uuid.New(), Status: model.InvoiceStatusPending, Total: 119000}
	db.Create(invoice)

	svc := NewBillingService(db, wompi.NewClient("https://example.test", ""))
	err := svc.ChargeInvoice(context.Background(), invoice.ID, nil, "tok")
	if err == nil {
		t.Fatal("expected an error when wompi has no private key configured")
	}

	var updated model.Invoice
	db.First(&updated, "id = ?", invoice.ID)
	if updated.Status != model.InvoiceStatusPending {
		t.Fatalf("invoice status must stay pending when the charge was never attempted, got %s", updated.Status)
	}
}

func TestChargeInvoiceRefusesNonPendingInvoice(t *testing.T) {
	db := setupInvoiceTestDB(t)
	invoice := &model.Invoice{TenantID: uuid.New(), Status: model.InvoiceStatusPaid, Total: 119000}
	db.Create(invoice)

	svc := NewBillingService(db, wompi.NewClient("https://example.test", "prv_test"))
	if err := svc.ChargeInvoice(context.Background(), invoice.ID, nil, "tok"); err == nil {
		t.Fatal("expected an error when charging an already-paid invoice (prevents double charge)")
	}
}
