package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
)

func TestCompanyQuotaIsSkippedInSyntheticPilot(t *testing.T) {
	// Must never call out to billing in synthetic-pilot mode: there is no
	// subscription concept for synthetic tenants.
	svc := NewCompanyService(nil, nil, "http://127.0.0.1:1", true)
	if err := svc.checkCompanyQuota(context.Background(), uuid.New(), "Bearer x"); err != nil {
		t.Fatalf("expected quota check to be skipped in synthetic pilot, got %v", err)
	}
}

func TestCompanyQuotaRejectsWhenBillingDisallows(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "" {
			t.Fatalf("expected the bearer token to be forwarded to billing")
		}
		w.Write([]byte(`{"allowed":false,"current_count":10,"max_companies":10}`))
	}))
	defer server.Close()

	svc := NewCompanyService(nil, nil, server.URL, false)
	err := svc.checkCompanyQuota(context.Background(), uuid.New(), "Bearer real-token")
	if err != ErrCompanyQuotaExceeded {
		t.Fatalf("expected ErrCompanyQuotaExceeded, got %v", err)
	}
}

func TestCompanyQuotaAllowsWhenBillingPermits(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"allowed":true,"current_count":3,"max_companies":10}`))
	}))
	defer server.Close()

	svc := NewCompanyService(nil, nil, server.URL, false)
	if err := svc.checkCompanyQuota(context.Background(), uuid.New(), "Bearer real-token"); err != nil {
		t.Fatalf("expected quota check to pass, got %v", err)
	}
}

func TestCompanyQuotaFailsClosedWhenBillingUnavailable(t *testing.T) {
	svc := NewCompanyService(nil, nil, "http://127.0.0.1:1", false)
	err := svc.checkCompanyQuota(context.Background(), uuid.New(), "Bearer x")
	if err == nil {
		t.Fatal("expected an error when billing is unreachable, got nil (fail-open is not acceptable)")
	}
}
