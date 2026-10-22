package service

import (
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"errors"
	"fmt"
	"html"
	"mime"
	"net"
	"net/mail"
	"net/smtp"
	"strings"
	"time"
)

// MailMessage is the contract shared by the HTTP endpoint and the
// zu.mail.outbound queue. Recipients and subject are validated before any
// SMTP traffic: a CR/LF in either would let a caller inject headers.
type MailMessage struct {
	To          []string          `json:"to"`
	CC          []string          `json:"cc"`
	Subject     string            `json:"subject"`
	BodyHTML    string            `json:"body_html"`
	BodyText    string            `json:"body_text"`
	TemplateKey string            `json:"template_key"`
	Variables   map[string]string `json:"variables"`
	Items       []string          `json:"items"` // rendered, escaped, as a <ul> in place of {{items}}
	TenantID    string            `json:"tenant_id"`
	CompanyID   string            `json:"company_id"`
}

type SendResult struct {
	MessageID string
	SentAt    time.Time
}

// ErrPermanent marks failures that retrying cannot fix (bad address, unknown
// template); the queue consumer dead-letters these immediately.
var ErrPermanent = errors.New("permanent mail failure")

type SMTPConfig struct {
	Host, Port, User, Pass string
	TLSMode                string // starttls | tls | none
	FromEmail, FromName    string
}

type MailerService struct {
	cfg  SMTPConfig
	dial func(addr string) (smtpClient, error) // overridable in tests
}

func NewMailerService(cfg SMTPConfig) *MailerService {
	s := &MailerService{cfg: cfg}
	s.dial = s.defaultDial
	return s
}

func (s *MailerService) Send(msg MailMessage) (SendResult, error) {
	msg, err := s.prepare(msg)
	if err != nil {
		return SendResult{}, err
	}
	id := newMessageID(s.cfg.FromEmail)
	raw := s.render(msg, id)
	rcpts := append(append([]string{}, msg.To...), msg.CC...)
	if err := s.deliver(rcpts, raw); err != nil {
		return SendResult{}, err
	}
	return SendResult{MessageID: id, SentAt: time.Now()}, nil
}

func (s *MailerService) prepare(msg MailMessage) (MailMessage, error) {
	if msg.TemplateKey != "" {
		tpl, ok := emailTemplates[msg.TemplateKey]
		if !ok {
			return msg, fmt.Errorf("%w: unknown template %q", ErrPermanent, msg.TemplateKey)
		}
		msg.Subject = fill(tpl.Subject, msg.Variables, false)
		msg.BodyHTML = strings.ReplaceAll(fill(tpl.Body, msg.Variables, true), "{{items}}", itemList(msg.Items))
		msg.BodyText = ""
	}
	if len(msg.To) == 0 {
		return msg, fmt.Errorf("%w: no recipients", ErrPermanent)
	}
	for _, list := range [][]string{msg.To, msg.CC} {
		for _, a := range list {
			if err := validAddress(a); err != nil {
				return msg, err
			}
		}
	}
	if strings.TrimSpace(msg.Subject) == "" || strings.ContainsAny(msg.Subject, "\r\n") {
		return msg, fmt.Errorf("%w: subject is empty or contains line breaks", ErrPermanent)
	}
	if msg.BodyHTML == "" && msg.BodyText == "" {
		return msg, fmt.Errorf("%w: empty body", ErrPermanent)
	}
	return msg, nil
}

func validAddress(a string) error {
	if strings.ContainsAny(a, "\r\n,;<>") {
		return fmt.Errorf("%w: invalid recipient %q", ErrPermanent, a)
	}
	parsed, err := mail.ParseAddress(a)
	if err != nil || parsed.Address != a {
		return fmt.Errorf("%w: invalid recipient %q", ErrPermanent, a)
	}
	return nil
}

func (s *MailerService) render(msg MailMessage, id string) []byte {
	boundary := "zu-" + randHex(12)
	from := mail.Address{Name: s.cfg.FromName, Address: s.cfg.FromEmail}
	var b strings.Builder
	h := func(k, v string) { b.WriteString(k + ": " + v + "\r\n") }
	h("From", from.String())
	h("To", strings.Join(msg.To, ", "))
	if len(msg.CC) > 0 {
		h("Cc", strings.Join(msg.CC, ", "))
	}
	h("Subject", mime.QEncoding.Encode("utf-8", msg.Subject))
	h("Date", time.Now().Format(time.RFC1123Z))
	h("Message-ID", id)
	h("MIME-Version", "1.0")
	h("Content-Type", `multipart/alternative; boundary="`+boundary+`"`)
	b.WriteString("\r\n")

	text := msg.BodyText
	if text == "" {
		text = htmlToText(msg.BodyHTML)
	}
	part := func(ctype, body string) {
		b.WriteString("--" + boundary + "\r\n")
		b.WriteString("Content-Type: " + ctype + "; charset=UTF-8\r\n")
		b.WriteString("Content-Transfer-Encoding: 8bit\r\n\r\n")
		b.WriteString(strings.ReplaceAll(body, "\n", "\r\n") + "\r\n")
	}
	part("text/plain", text)
	if msg.BodyHTML != "" {
		part("text/html", msg.BodyHTML)
	}
	b.WriteString("--" + boundary + "--\r\n")
	return []byte(b.String())
}

// smtpClient is the subset of *smtp.Client used, so tests can fake a server.
type smtpClient interface {
	Hello(string) error
	Extension(string) (bool, string)
	StartTLS(*tls.Config) error
	Auth(smtp.Auth) error
	Mail(string) error
	Rcpt(string) error
	Data() (interface {
		Write([]byte) (int, error)
		Close() error
	}, error)
	Quit() error
	Close() error
}

type stdClient struct{ *smtp.Client }

func (c stdClient) Data() (interface {
	Write([]byte) (int, error)
	Close() error
}, error) {
	return c.Client.Data()
}

func (s *MailerService) defaultDial(addr string) (smtpClient, error) {
	if s.cfg.TLSMode == "tls" {
		conn, err := tls.DialWithDialer(&net.Dialer{Timeout: 15 * time.Second}, "tcp", addr, &tls.Config{ServerName: s.cfg.Host})
		if err != nil {
			return nil, err
		}
		c, err := smtp.NewClient(conn, s.cfg.Host)
		if err != nil {
			return nil, err
		}
		return stdClient{c}, nil
	}
	conn, err := net.DialTimeout("tcp", addr, 15*time.Second)
	if err != nil {
		return nil, err
	}
	c, err := smtp.NewClient(conn, s.cfg.Host)
	if err != nil {
		return nil, err
	}
	return stdClient{c}, nil
}

func (s *MailerService) deliver(rcpts []string, raw []byte) error {
	c, err := s.dial(net.JoinHostPort(s.cfg.Host, s.cfg.Port))
	if err != nil {
		return fmt.Errorf("smtp connect: %w", err)
	}
	defer c.Close()
	if err := c.Hello("zuseguridad360.com"); err != nil {
		return fmt.Errorf("smtp hello: %w", err)
	}
	if s.cfg.TLSMode == "starttls" {
		// Never downgrade: a server that does not offer STARTTLS gets no mail
		// and, above all, no credentials.
		if ok, _ := c.Extension("STARTTLS"); !ok {
			return fmt.Errorf("smtp server %s does not offer STARTTLS", s.cfg.Host)
		}
		if err := c.StartTLS(&tls.Config{ServerName: s.cfg.Host}); err != nil {
			return fmt.Errorf("smtp starttls: %w", err)
		}
	}
	if s.cfg.User != "" {
		if s.cfg.TLSMode == "none" {
			return fmt.Errorf("refusing to send SMTP credentials over an unencrypted connection")
		}
		if err := c.Auth(smtp.PlainAuth("", s.cfg.User, s.cfg.Pass, s.cfg.Host)); err != nil {
			return fmt.Errorf("smtp auth: %w", err)
		}
	}
	if err := c.Mail(s.cfg.FromEmail); err != nil {
		return fmt.Errorf("smtp mail from: %w", err)
	}
	for _, r := range rcpts {
		if err := c.Rcpt(r); err != nil {
			return fmt.Errorf("smtp rcpt %s: %w", r, err)
		}
	}
	w, err := c.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}
	if _, err := w.Write(raw); err != nil {
		return fmt.Errorf("smtp write: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("smtp data close: %w", err)
	}
	return c.Quit()
}

// fill substitutes {{var}} placeholders. Values going into HTML are escaped:
// they come from tenant data (company names, incident descriptions) and must
// not be able to inject markup into a mail the recipient trusts.
func fill(tpl string, vars map[string]string, escape bool) string {
	for k, v := range vars {
		if escape {
			v = html.EscapeString(v)
		} else {
			v = strings.NewReplacer("\r", " ", "\n", " ").Replace(v)
		}
		tpl = strings.ReplaceAll(tpl, "{{"+k+"}}", v)
	}
	return tpl
}

func itemList(items []string) string {
	if len(items) == 0 {
		return "<p>Sin novedades.</p>"
	}
	var b strings.Builder
	b.WriteString("<ul>")
	for _, it := range items {
		b.WriteString("<li>" + html.EscapeString(it) + "</li>")
	}
	b.WriteString("</ul>")
	return b.String()
}

func htmlToText(h string) string {
	r := strings.NewReplacer("<br>", "\n", "<br/>", "\n", "</p>", "\n\n", "</h2>", "\n\n", "</li>", "\n")
	s := r.Replace(h)
	var b strings.Builder
	in := false
	for _, ch := range s {
		switch {
		case ch == '<':
			in = true
		case ch == '>':
			in = false
		case !in:
			b.WriteRune(ch)
		}
	}
	return strings.TrimSpace(html.UnescapeString(b.String()))
}

func newMessageID(from string) string {
	domain := "zuseguridad360.com"
	if i := strings.LastIndex(from, "@"); i >= 0 {
		domain = from[i+1:]
	}
	return "<" + randHex(16) + "@" + domain + ">"
}

func randHex(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)
}
