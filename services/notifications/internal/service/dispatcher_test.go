package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWhatsAppWithoutCredentialsFailsHonestly(t *testing.T) {
	d := NewNotificationDispatcher(map[string]string{})
	result := d.sendWhatsApp(context.Background(), NotificationPayload{Recipient: "+573001234567", Title: "x", Message: "y"})
	if result.Success {
		t.Fatal("expected Success=false without credentials — must never fabricate success")
	}
	if result.Error == "" {
		t.Fatal("expected a non-empty error explaining why nothing was sent")
	}
}

func TestPushWithoutCredentialsFailsHonestly(t *testing.T) {
	d := NewNotificationDispatcher(map[string]string{})
	result := d.sendPush(context.Background(), NotificationPayload{Recipient: "device-token", Title: "x", Message: "y"})
	if result.Success {
		t.Fatal("expected Success=false without credentials — must never fabricate success")
	}
}

func TestWhatsAppMissingRecipientFails(t *testing.T) {
	d := NewNotificationDispatcher(map[string]string{"whatsapp_token": "t", "whatsapp_phone_number_id": "p"})
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

	d := NewNotificationDispatcher(map[string]string{"whatsapp_token": "real-token", "whatsapp_phone_number_id": "123"})
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

func TestPushSendsRealRequestWhenConfigured(t *testing.T) {
	var gotAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success":1}`))
	}))
	defer server.Close()

	d := NewNotificationDispatcher(map[string]string{"push_api_key": "server-key"})
	d.fcmURL = server.URL

	result := d.sendPush(context.Background(), NotificationPayload{Recipient: "device-token", Title: "t", Message: "m"})
	if !result.Success {
		t.Fatalf("expected success, got error: %s", result.Error)
	}
	if gotAuth != "key=server-key" {
		t.Fatalf("expected FCM server-key auth header, got %q", gotAuth)
	}
}
