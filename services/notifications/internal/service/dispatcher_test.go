package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWhatsAppWithoutCredentialsFailsHonestly(t *testing.T) {
	d := NewNotificationDispatcher(nil, nil, map[string]string{})
	result := d.sendWhatsApp(context.Background(), NotificationPayload{Recipient: "+573001234567", Title: "x", Message: "y"})
	if result.Success {
		t.Fatal("expected Success=false without credentials — must never fabricate success")
	}
	if result.Error == "" {
		t.Fatal("expected a non-empty error explaining why nothing was sent")
	}
}

func TestWhatsAppMissingRecipientFails(t *testing.T) {
	d := NewNotificationDispatcher(nil, nil, map[string]string{"whatsapp_token": "t", "whatsapp_phone_number_id": "p"})
	result := d.sendWhatsApp(context.Background(), NotificationPayload{Title: "x", Message: "y"})
	if result.Success {
		t.Fatal("expected failure when recipient is missing")
	}
}

func TestWhatsAppSendsRealRequestWhenConfigured(t *testing.T) {
	var gotAuth, gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"messages":[{"id":"wamid.x"}]}`))
	}))
	defer server.Close()

	d := NewNotificationDispatcher(nil, nil, map[string]string{"whatsapp_token": "real-token", "whatsapp_phone_number_id": "123"})
	d.whatsappBaseURL = server.URL

	result := d.sendWhatsApp(context.Background(), NotificationPayload{Recipient: "+573001234567", Title: "t", Message: "m"})
	if !result.Success {
		t.Fatalf("expected success, got error: %s", result.Error)
	}
	if gotAuth != "Bearer real-token" {
		t.Fatalf("expected bearer token forwarded, got %q", gotAuth)
	}
	if gotPath != "/123/messages" {
		t.Fatalf("expected phone_number_id in path, got %q", gotPath)
	}
}

func TestWhatsAppUsesApprovedTemplateWhenConfigured(t *testing.T) {
	var got map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&got)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()
	d := NewNotificationDispatcher(nil, nil, map[string]string{"whatsapp_token": "t", "whatsapp_phone_number_id": "1", "whatsapp_template": "alerta_sst"})
	d.whatsappBaseURL = server.URL
	if r := d.sendWhatsApp(context.Background(), NotificationPayload{Recipient: "+573001234567", Title: "t", Message: "m"}); !r.Success {
		t.Fatal(r.Error)
	}
	if got["type"] != "template" {
		t.Fatalf("expected template message, got %v", got["type"])
	}
}

func TestValidateRejectsMultilineTitle(t *testing.T) {
	p := NotificationPayload{Kind: "k", TenantID: "t", Title: "a\nBcc: x", Channels: []Channel{ChannelEmail}}
	if p.Validate() == nil {
		t.Fatal("expected multi-line title to be rejected")
	}
}
