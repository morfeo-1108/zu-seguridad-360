package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/zu-seguridad/notifications/internal/store"
)

type Channel string

const (
	ChannelInApp    Channel = "inapp"
	ChannelEmail    Channel = "email"
	ChannelWhatsApp Channel = "whatsapp"
	ChannelPush     Channel = "push"
	ChannelSMS      Channel = "sms"
)

// NotificationPayload is what producers publish to the zu.notifications
// exchange (routing key notify.<kind>) or POST to /v1/dispatch.
type NotificationPayload struct {
	Kind      string    `json:"kind"` // e.g. sla.breach, briefing.morning
	TenantID  string    `json:"tenant_id"`
	CompanyID string    `json:"company_id"`
	UserID    string    `json:"user_id"`   // in-app addressee; empty = whole tenant
	Email     string    `json:"email"`     // email channel addressee
	Recipient string    `json:"recipient"` // WhatsApp: E.164 number; push: device token
	Channels  []Channel `json:"channels"`
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	Link      string    `json:"link"`
	Priority  string    `json:"priority"` // critical | high | normal | low
	// Optional mail template (services/mail) and its variables/items; without
	// it the email uses the generic "notification" template (title + message).
	MailTemplate string            `json:"mail_template"`
	Variables    map[string]string `json:"variables"`
	Items        []string          `json:"items"`
}

type DispatchResult struct {
	Channel Channel `json:"channel"`
	Success bool    `json:"success"`
	Error   string  `json:"error,omitempty"`
}

type Publisher interface {
	Publish(ctx context.Context, exchange, key string, body []byte) error
}

type NotificationDispatcher struct {
	store                 *store.Store
	pub                   Publisher
	whatsappToken         string
	whatsappPhoneNumberID string
	whatsappTemplate      string
	httpClient            *http.Client
	whatsappBaseURL       string // overridable in tests; defaults to Meta's Graph API
}

func NewNotificationDispatcher(st *store.Store, pub Publisher, cfg map[string]string) *NotificationDispatcher {
	return &NotificationDispatcher{
		store:                 st,
		pub:                   pub,
		whatsappToken:         cfg["whatsapp_token"],
		whatsappPhoneNumberID: cfg["whatsapp_phone_number_id"],
		whatsappTemplate:      cfg["whatsapp_template"],
		httpClient:            &http.Client{Timeout: 10 * time.Second},
		whatsappBaseURL:       "https://graph.facebook.com/v19.0",
	}
}

func (p NotificationPayload) Validate() error {
	if p.TenantID == "" || p.Title == "" || p.Kind == "" || len(p.Channels) == 0 {
		return fmt.Errorf("kind, tenant_id, title and channels are required")
	}
	if strings.ContainsAny(p.Title, "\r\n") {
		return fmt.Errorf("title must be a single line")
	}
	return nil
}

// Dispatch delivers on every requested channel independently and records an
// audit row per attempt. A channel never reports success it did not achieve.
func (d *NotificationDispatcher) Dispatch(ctx context.Context, p NotificationPayload) []DispatchResult {
	if p.Priority == "" {
		p.Priority = "normal"
	}
	var notifID *string
	var results []DispatchResult
	for _, ch := range p.Channels {
		var r DispatchResult
		switch ch {
		case ChannelInApp:
			r = d.saveInApp(ctx, p, &notifID)
		case ChannelEmail:
			r = d.queueEmail(ctx, p)
		case ChannelWhatsApp:
			r = d.sendWhatsApp(ctx, p)
		case ChannelPush:
			r = DispatchResult{Channel: ChannelPush, Error: "push not available: FCM's legacy API was retired in 2024 and the HTTP v1 client (service account) is not implemented"}
		case ChannelSMS:
			r = DispatchResult{Channel: ChannelSMS, Error: "SMS provider not configured"}
		default:
			r = DispatchResult{Channel: ch, Error: fmt.Sprintf("unsupported channel %q", ch)}
		}
		if err := d.store.RecordDelivery(ctx, &store.Delivery{
			NotificationID: notifID, TenantID: p.TenantID, Channel: string(r.Channel), Kind: p.Kind, Success: r.Success, Error: r.Error,
		}); err != nil {
			log.Printf("delivery audit write failed: %v", err)
		}
		if !r.Success {
			log.Printf("notification %s [%s] not delivered: %s", p.Kind, r.Channel, r.Error)
		}
		results = append(results, r)
	}
	return results
}

func (d *NotificationDispatcher) saveInApp(ctx context.Context, p NotificationPayload, id **string) DispatchResult {
	n := &store.Notification{TenantID: p.TenantID, Kind: p.Kind, Title: p.Title, Message: p.Message, Priority: p.Priority, Link: p.Link}
	if p.UserID != "" {
		n.UserID = &p.UserID
	}
	if p.CompanyID != "" {
		n.CompanyID = &p.CompanyID
	}
	if err := d.store.Create(ctx, n); err != nil {
		return DispatchResult{Channel: ChannelInApp, Error: "in-app store failed: " + err.Error()}
	}
	*id = &n.ID
	return DispatchResult{Channel: ChannelInApp, Success: true}
}

// queueEmail hands the message to services/mail through its durable queue.
// Success here means "accepted by the mail queue"; SMTP delivery, retries and
// dead-lettering are mail's responsibility.
func (d *NotificationDispatcher) queueEmail(ctx context.Context, p NotificationPayload) DispatchResult {
	if p.Email == "" {
		return DispatchResult{Channel: ChannelEmail, Error: "no email address for this notification"}
	}
	tpl, vars := p.MailTemplate, p.Variables
	if tpl == "" {
		tpl, vars = "notification", map[string]string{"title": p.Title, "message": p.Message}
	}
	body, _ := json.Marshal(map[string]interface{}{
		"to": []string{p.Email}, "template_key": tpl, "variables": vars, "items": p.Items,
		"tenant_id": p.TenantID, "company_id": p.CompanyID,
	})
	if err := d.pub.Publish(ctx, "", "zu.mail.outbound", body); err != nil {
		return DispatchResult{Channel: ChannelEmail, Error: "mail queue unavailable: " + err.Error()}
	}
	return DispatchResult{Channel: ChannelEmail, Success: true}
}

// sendWhatsApp posts to Meta's WhatsApp Business Cloud API. It never reports
// success without actually sending: if credentials are missing or the API
// call fails, Success is false and Error explains why — this used to always
// return Success: true regardless of whether anything was configured, which
// is dangerous for a channel used to escalate critical safety findings.
func (d *NotificationDispatcher) sendWhatsApp(ctx context.Context, payload NotificationPayload) DispatchResult {
	result := DispatchResult{Channel: ChannelWhatsApp}
	if d.whatsappToken == "" || d.whatsappPhoneNumberID == "" {
		result.Error = "WhatsApp Business credentials not configured (WHATSAPP_TOKEN/WHATSAPP_PHONE_NUMBER_ID)"
		log.Printf("⚠️  WhatsApp not sent (no credentials): %s", payload.Title)
		return result
	}
	if payload.Recipient == "" {
		result.Error = "recipient phone number is required"
		return result
	}

	// Business-initiated messages outside a 24h customer window must use a
	// template approved by Meta; free text only works inside that window.
	body := map[string]interface{}{"messaging_product": "whatsapp", "to": payload.Recipient}
	if d.whatsappTemplate != "" {
		body["type"] = "template"
		body["template"] = map[string]interface{}{
			"name": d.whatsappTemplate, "language": map[string]string{"code": "es"},
			"components": []map[string]interface{}{{"type": "body", "parameters": []map[string]string{
				{"type": "text", "text": payload.Title}, {"type": "text", "text": payload.Message},
			}}},
		}
	} else {
		body["type"] = "text"
		body["text"] = map[string]string{"body": payload.Title + ": " + payload.Message}
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
