package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/smtp"
	"strings"
	"time"
)

type MailProvider string

const (
	ProviderSendGrid MailProvider = "sendgrid"
	ProviderMailgun  MailProvider = "mailgun"
	ProviderSES      MailProvider = "ses"
	ProviderSMTP     MailProvider = "smtp"
)

type MailMessage struct {
	To          []string          `json:"to"`
	CC          []string          `json:"cc"`
	Subject     string            `json:"subject"`
	BodyHTML    string            `json:"body_html"`
	BodyText    string            `json:"body_text"`
	TemplateKey string            `json:"template_key"`
	Variables   map[string]string `json:"variables"`
	AttachURLs  []string          `json:"attach_urls"`
	TenantID    string            `json:"tenant_id"`
	CompanyID   string            `json:"company_id"`
	Priority    int               `json:"priority"` // 1=high, 2=normal, 3=low
}

type SendResult struct {
	MessageID string
	Provider  MailProvider
	SentAt    time.Time
	Error     error
}

type MailerService struct {
	providers []MailProvider
	smtpHost  string
	smtpPort  string
	smtpUser  string
	smtpPass  string
	fromEmail string
	fromName  string
}

func NewMailerService(cfg map[string]string) *MailerService {
	return &MailerService{
		providers: []MailProvider{ProviderSendGrid, ProviderSMTP},
		smtpHost:  cfg["smtp_host"],
		smtpPort:  cfg["smtp_port"],
		smtpUser:  cfg["smtp_user"],
		smtpPass:  cfg["smtp_pass"],
		fromEmail: cfg["from_email"],
		fromName:  cfg["from_name"],
	}
}

func (s *MailerService) Send(ctx context.Context, msg MailMessage) SendResult {
	// Aplicar template si se especificó
	if msg.TemplateKey != "" {
		msg = s.applyTemplate(msg)
	}

	// Intentar proveedores en orden (fallback automático)
	for _, provider := range s.providers {
		result := s.sendViaProvider(ctx, msg, provider)
		if result.Error == nil {
			log.Printf("📧 Email enviado via %s a %v: %s", provider, msg.To, msg.Subject)
			return result
		}
		log.Printf("⚠️ Provider %s failed, trying next: %v", provider, result.Error)
	}

	return SendResult{Error: fmt.Errorf("all mail providers failed")}
}

func (s *MailerService) sendViaProvider(ctx context.Context, msg MailMessage, provider MailProvider) SendResult {
	switch provider {
	case ProviderSMTP:
		return s.sendSMTP(msg)
	default:
		return SendResult{Error: fmt.Errorf("provider %s not configured", provider)}
	}
}

func (s *MailerService) sendSMTP(msg MailMessage) SendResult {
	auth := smtp.PlainAuth("", s.smtpUser, s.smtpPass, s.smtpHost)

	header := fmt.Sprintf(
		"From: %s <%s>\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n",
		s.fromName, s.fromEmail, strings.Join(msg.To, ", "), msg.Subject,
	)
	body := header + msg.BodyHTML

	addr := fmt.Sprintf("%s:%s", s.smtpHost, s.smtpPort)
	err := smtp.SendMail(addr, auth, s.fromEmail, msg.To, []byte(body))
	if err != nil {
		return SendResult{Error: err}
	}
	return SendResult{Provider: ProviderSMTP, SentAt: time.Now()}
}

// Templates SST predefinidos
var emailTemplates = map[string]struct{ Subject, Body string }{
	"welcome": {
		Subject: "Bienvenido a ZU Seguridad Integral 360 — {{company_name}}",
		Body:    `<h2>Bienvenido, {{contact_name}}</h2><p>Su empresa <strong>{{company_name}}</strong> ha sido registrada exitosamente en ZU Seguridad Integral 360.</p><p>Su profesional SST asignada es <strong>{{sst_name}}</strong>.</p>`,
	},
	"cert_expiring": {
		Subject: "⏰ Certificado próximo a vencer — {{employee_name}}",
		Body:    `<p>El certificado de <strong>{{cert_type}}</strong> del empleado <strong>{{employee_name}}</strong> vence el <strong>{{expiry_date}}</strong>.</p><p>Por favor gestione la renovación a la brevedad.</p>`,
	},
	"incident_reported": {
		Subject: "🚨 Nuevo incidente reportado — {{company_name}}",
		Body:    `<p>Se ha reportado un nuevo incidente en <strong>{{company_name}}</strong>.</p><p><strong>Tipo:</strong> {{incident_type}}<br><strong>Descripción:</strong> {{description}}<br><strong>Fecha:</strong> {{date}}</p>`,
	},
	"monthly_report": {
		Subject: "📊 Informe SST mensual — {{company_name}} — {{month}}",
		Body:    `<h2>Informe Mensual SST</h2><p>Adjunto encontrará el informe de indicadores SST correspondiente a {{month}}.</p><p><strong>IF:</strong> {{if_value}} | <strong>IS:</strong> {{is_value}} | <strong>Cumplimiento:</strong> {{compliance}}%</p>`,
	},
	"furat_sent": {
		Subject: "📋 FURAT enviado a ARL {{arl_name}} — {{company_name}}",
		Body:    `<p>El Formulario Único de Reporte de Accidente de Trabajo (FURAT) ha sido enviado exitosamente a <strong>{{arl_name}}</strong>.</p><p>Número de radicado: <strong>{{furat_number}}</strong></p>`,
	},
}

func (s *MailerService) applyTemplate(msg MailMessage) MailMessage {
	tpl, ok := emailTemplates[msg.TemplateKey]
	if !ok {
		return msg
	}

	subject := tpl.Subject
	body := tpl.Body
	for k, v := range msg.Variables {
		subject = strings.ReplaceAll(subject, "{{"+k+"}}", v)
		body = strings.ReplaceAll(body, "{{"+k+"}}", v)
	}
	msg.Subject = subject
	msg.BodyHTML = body
	return msg
}

// ConsumeQueue procesa mensajes de la cola RabbitMQ
func (s *MailerService) ConsumeQueue(ctx context.Context, msgs <-chan []byte) {
	for {
		select {
		case <-ctx.Done():
			return
		case raw := <-msgs:
			var msg MailMessage
			if err := json.Unmarshal(raw, &msg); err != nil {
				log.Printf("Invalid mail message: %v", err)
				continue
			}
			s.Send(ctx, msg)
		}
	}
}
