package service

import (
	"bufio"
	"errors"
	"io"
	"net"
	"strings"
	"testing"
	"time"
)

func TestRejectsHeaderInjection(t *testing.T) {
	s := NewMailerService(SMTPConfig{Host: "x", Port: "25", TLSMode: "none", FromEmail: "a@b.co"})
	cases := []MailMessage{
		{To: []string{"victim@x.co\r\nBcc: evil@x.co"}, Subject: "s", BodyText: "b"},
		{To: []string{"ok@x.co"}, Subject: "s\r\nBcc: evil@x.co", BodyText: "b"},
		{To: []string{"Name <ok@x.co>"}, Subject: "s", BodyText: "b"},
		{To: nil, Subject: "s", BodyText: "b"},
		{To: []string{"ok@x.co"}, TemplateKey: "nope"},
	}
	for i, m := range cases {
		if _, err := s.Send(m); !errors.Is(err, ErrPermanent) {
			t.Fatalf("case %d: expected permanent error, got %v", i, err)
		}
	}
}

func TestTemplateVariablesAreEscaped(t *testing.T) {
	s := NewMailerService(SMTPConfig{})
	m, err := s.prepare(MailMessage{
		To: []string{"ok@x.co"}, TemplateKey: "incident_reported",
		Variables: map[string]string{"company_name": "<script>x</script>\r\nBcc: e@x.co", "description": "a & b"},
		Items:     []string{"<b>no</b>"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(m.BodyHTML, "<script>") || !strings.Contains(m.BodyHTML, "a &amp; b") {
		t.Fatalf("body not escaped: %s", m.BodyHTML)
	}
	if strings.ContainsAny(m.Subject, "\r\n") {
		t.Fatalf("subject kept line breaks: %q", m.Subject)
	}
}

// fakeSMTP is a minimal plaintext SMTP server that records the DATA payload.
func fakeSMTP(t *testing.T, offerStartTLS bool) (addr string, got chan string) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	got = make(chan string, 1)
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		r := bufio.NewReader(conn)
		w := func(s string) { io.WriteString(conn, s+"\r\n") }
		w("220 fake")
		for {
			line, err := r.ReadString('\n')
			if err != nil {
				return
			}
			cmd := strings.ToUpper(strings.TrimSpace(line))
			switch {
			case strings.HasPrefix(cmd, "EHLO"):
				if offerStartTLS {
					w("250-fake")
					w("250 STARTTLS")
				} else {
					w("250 fake")
				}
			case cmd == "DATA":
				w("354 go")
				var b strings.Builder
				for {
					l, _ := r.ReadString('\n')
					if l == ".\r\n" {
						break
					}
					b.WriteString(l)
				}
				got <- b.String()
				w("250 ok")
			case cmd == "QUIT":
				w("221 bye")
				return
			default:
				w("250 ok")
			}
		}
	}()
	t.Cleanup(func() { ln.Close() })
	return ln.Addr().String(), got
}

func TestDeliversMultipartOverPlainSMTP(t *testing.T) {
	addr, got := fakeSMTP(t, false)
	host, port, _ := net.SplitHostPort(addr)
	s := NewMailerService(SMTPConfig{Host: host, Port: port, TLSMode: "none", FromEmail: "noreply@zuseguridad360.com", FromName: "ZU"})
	res, err := s.Send(MailMessage{To: []string{"sst@example.co"}, TemplateKey: "sla_breach", Variables: map[string]string{"event_label": "Investigación", "company_name": "ACME"}})
	if err != nil {
		t.Fatal(err)
	}
	select {
	case data := <-got:
		for _, want := range []string{"multipart/alternative", "text/plain", "text/html", res.MessageID, "ACME"} {
			if !strings.Contains(data, want) {
				t.Fatalf("payload missing %q:\n%s", want, data)
			}
		}
	case <-time.After(3 * time.Second):
		t.Fatal("no DATA received")
	}
}

func TestStartTLSRequiredIsNeverDowngraded(t *testing.T) {
	addr, _ := fakeSMTP(t, false)
	host, port, _ := net.SplitHostPort(addr)
	s := NewMailerService(SMTPConfig{Host: host, Port: port, TLSMode: "starttls", User: "u", Pass: "p", FromEmail: "a@b.co"})
	_, err := s.Send(MailMessage{To: []string{"x@y.co"}, Subject: "s", BodyText: "b"})
	if err == nil || !strings.Contains(err.Error(), "STARTTLS") {
		t.Fatalf("expected refusal without STARTTLS, got %v", err)
	}
}

func TestCredentialsNeverSentInClear(t *testing.T) {
	addr, _ := fakeSMTP(t, false)
	host, port, _ := net.SplitHostPort(addr)
	s := NewMailerService(SMTPConfig{Host: host, Port: port, TLSMode: "none", User: "u", Pass: "p", FromEmail: "a@b.co"})
	if _, err := s.Send(MailMessage{To: []string{"x@y.co"}, Subject: "s", BodyText: "b"}); err == nil {
		t.Fatal("expected refusal to authenticate over plaintext")
	}
}
