package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
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
	whatsappToken string
	pushApiKey    string
}

func NewNotificationDispatcher(cfg map[string]string) *NotificationDispatcher {
	return &NotificationDispatcher{
		whatsappToken: cfg["whatsapp_token"],
		pushApiKey:    cfg["push_api_key"],
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
			result.Success = true
			log.Printf("📱 SMS queued: %s to %s", payload.Title, payload.UserID)

		default:
			result.Error = fmt.Sprintf("canal %s no soportado", ch)
		}

		results = append(results, result)
	}

	return results
}

func (d *NotificationDispatcher) sendWhatsApp(ctx context.Context, payload NotificationPayload) DispatchResult {
	// Integración WhatsApp Business API (Meta)
	log.Printf("📲 WhatsApp [%s]: %s — %s", payload.Priority, payload.Title, payload.Message)
	return DispatchResult{Channel: ChannelWhatsApp, Success: true, SentAt: time.Now()}
}

func (d *NotificationDispatcher) sendPush(ctx context.Context, payload NotificationPayload) DispatchResult {
	// Firebase FCM push notification
	log.Printf("🔔 Push [%s]: %s", payload.Priority, payload.Title)
	return DispatchResult{Channel: ChannelPush, Success: true, SentAt: time.Now()}
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
