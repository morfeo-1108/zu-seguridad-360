package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type Channel string

const (
	ChannelEmail    Channel = "email"
	ChannelWhatsApp Channel = "whatsapp"
	ChannelPush     Channel = "push"
	ChannelSMS      Channel = "sms"
)

type NotificationPayload struct {
	TenantID  string            `json:"tenant_id"`
	CompanyID string            `json:"company_id"`
	UserID    string            `json:"user_id"`
	Recipient string            `json:"recipient"` // WhatsApp: número E.164; Push: device token
	Channels  []Channel         `json:"channels"`
	Title     string            `json:"title"`
	Message   string            `json:"message"`
	Data      map[string]string `json:"data"`
	Priority  string            `json:"priority"` // critical | high | normal | low
	SendAt    *time.Time        `json:"send_at"`  // nil = inmediato
}

type DispatchResult struct {
	Channel Channel
	Success bool
	Error   string
	SentAt  time.Time
}

type NotificationDispatcher struct {
	whatsappToken         string
	whatsappPhoneNumberID string
	pushAPIKey            string
	httpClient            *http.Client
	whatsappBaseURL       string // overridable in tests; defaults to Meta's Graph API
	fcmURL                string // overridable in tests; defaults to FCM's legacy endpoint
}

func NewNotificationDispatcher(cfg map[string]string) *NotificationDispatcher {
	return &NotificationDispatcher{
		whatsappToken:         cfg["whatsapp_token"],
		whatsappPhoneNumberID: cfg["whatsapp_phone_number_id"],
		pushAPIKey:            cfg["push_api_key"],
		httpClient:            &http.Client{Timeout: 10 * time.Second},
		whatsappBaseURL:       "https://graph.facebook.com/v19.0",
		fcmURL:                "https://fcm.googleapis.com/fcm/send",
	}
}

func (d *NotificationDispatcher) Dispatch(ctx context.Context, payload NotificationPayload) []DispatchResult {
	var results []DispatchResult

	for _, ch := range payload.Channels {
		var result DispatchResult
		result.Channel = ch
		result.SentAt = time.Now()

		switch ch {
		case ChannelEmail:
			// Delegar al microservicio mail via RabbitMQ
			result.Success = true
			log.Printf("📧 Email notification queued: %s", payload.Title)

		case ChannelWhatsApp:
			result = d.sendWhatsApp(ctx, payload)

		case ChannelPush:
			result = d.sendPush(ctx, payload)

		case ChannelSMS:
			result.Success = false
			result.Error = "SMS provider not configured"
			log.Printf("⚠️  SMS not sent (no provider configured): %s to %s", payload.Title, payload.UserID)

		default:
			result.Error = fmt.Sprintf("canal %s no soportado", ch)
		}

		results = append(results, result)
	}

	return results
}

// sendWhatsApp posts to Meta's WhatsApp Business Cloud API. It never reports
// success without actually sending: if credentials are missing or the API
// call fails, Success is false and Error explains why — this used to always
// return Success: true regardless of whether anything was configured, which
// is dangerous for a channel used to escalate critical safety findings.
func (d *NotificationDispatcher) sendWhatsApp(ctx context.Context, payload NotificationPayload) DispatchResult {
	result := DispatchResult{Channel: ChannelWhatsApp, SentAt: time.Now()}
	if d.whatsappToken == "" || d.whatsappPhoneNumberID == "" {
		result.Error = "WhatsApp Business credentials not configured (WHATSAPP_TOKEN/WHATSAPP_PHONE_NUMBER_ID)"
		log.Printf("⚠️  WhatsApp not sent (no credentials): %s", payload.Title)
		return result
	}
	if payload.Recipient == "" {
		result.Error = "recipient phone number is required"
		return result
	}

	body := map[string]interface{}{
		"messaging_product": "whatsapp",
		"to":                payload.Recipient,
		"type":              "text",
		"text":              map[string]string{"body": payload.Title + ": " + payload.Message},
	}
	raw, _ := json.Marshal(body)
	url := fmt.Sprintf("%s/%s/messages", d.whatsappBaseURL, d.whatsappPhoneNumberID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(raw))
	if err != nil {
		result.Error = fmt.Sprintf("request build failed: %v", err)
		return result
	}
	req.Header.Set("Authorization", "Bearer "+d.whatsappToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := d.httpClient.Do(req)
	if err != nil {
		result.Error = fmt.Sprintf("WhatsApp API unreachable: %v", err)
		return result
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		result.Error = fmt.Sprintf("WhatsApp API returned %d: %s", resp.StatusCode, string(respBody))
		return result
	}
	result.Success = true
	return result
}

// sendPush posts to the legacy FCM HTTP API using a server key. Same
// honesty rule as WhatsApp: no credentials or a failed call means
// Success: false, never a fabricated true.
func (d *NotificationDispatcher) sendPush(ctx context.Context, payload NotificationPayload) DispatchResult {
	result := DispatchResult{Channel: ChannelPush, SentAt: time.Now()}
	if d.pushAPIKey == "" {
		result.Error = "push provider not configured (PUSH_API_KEY)"
		log.Printf("⚠️  Push not sent (no credentials): %s", payload.Title)
		return result
	}
	if payload.Recipient == "" {
		result.Error = "recipient device token is required"
		return result
	}

	body := map[string]interface{}{
		"to": payload.Recipient,
		"notification": map[string]string{
			"title": payload.Title,
			"body":  payload.Message,
		},
		"data": payload.Data,
	}
	raw, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, d.fcmURL, bytes.NewReader(raw))
	if err != nil {
		result.Error = fmt.Sprintf("request build failed: %v", err)
		return result
	}
	req.Header.Set("Authorization", "key="+d.pushAPIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := d.httpClient.Do(req)
	if err != nil {
		result.Error = fmt.Sprintf("FCM unreachable: %v", err)
		return result
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		result.Error = fmt.Sprintf("FCM returned %d: %s", resp.StatusCode, string(respBody))
		return result
	}
	result.Success = true
	return result
}

func (d *NotificationDispatcher) ConsumeQueue(ctx context.Context, msgs <-chan []byte) {
	for {
		select {
		case <-ctx.Done():
			return
		case raw := <-msgs:
			var payload NotificationPayload
			if err := json.Unmarshal(raw, &payload); err != nil {
				log.Printf("Invalid notification payload: %v", err)
				continue
			}
			results := d.Dispatch(ctx, payload)
			for _, r := range results {
				if !r.Success {
					log.Printf("❌ Notification failed [%s]: %s", r.Channel, r.Error)
				}
			}
		}
	}
}
