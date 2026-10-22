package config

import (
	"fmt"
	"os"
)

type Config struct {
	Env, Port, RabbitMQURL                 string
	SMTPHost, SMTPPort, SMTPUser, SMTPPass string
	SMTPTLS                                string // starttls | tls | none
	FromEmail, FromName                    string
	InternalToken                          string // guards POST /v1/send; empty disables the endpoint

	VaultEnabled bool
	VaultAddr    string
	VaultToken   string
	VaultPath    string
}

func Load() *Config {
	return &Config{
		Env:           env("ENV", "development"),
		Port:          env("PORT", "4005"),
		RabbitMQURL:   env("RABBITMQ_URL", "amqp://zurabbit:zurabbit123@rabbitmq:5672/zu_vhost"),
		SMTPHost:      env("SMTP_HOST", "smtp.sendgrid.net"),
		SMTPPort:      env("SMTP_PORT", "587"),
		SMTPUser:      env("SMTP_USER", ""),
		SMTPPass:      env("SMTP_PASS", ""),
		SMTPTLS:       env("SMTP_TLS", "starttls"),
		FromEmail:     env("FROM_EMAIL", "noreply@zuseguridad360.com"),
		FromName:      env("FROM_NAME", "ZU Seguridad 360"),
		InternalToken: env("INTERNAL_API_TOKEN", ""),

		VaultEnabled: env("VAULT_ENABLED", "false") == "true",
		VaultAddr:    env("VAULT_ADDR", "http://localhost:8200"),
		VaultToken:   env("VAULT_TOKEN", "zu-vault-dev-token"),
		VaultPath:    env("VAULT_SECRET_PATH", "secret/data/zu360/mail"),
	}
}

// ValidateProduction refuses to start outside development with settings that
// would leak mail in clear text or leave the send endpoint guarded by a
// trivially guessable token.
func (c *Config) ValidateProduction() error {
	if c.Env == "development" {
		return nil
	}
	if c.SMTPTLS != "starttls" && c.SMTPTLS != "tls" && c.SMTPHost != "mailpit" {
		return fmt.Errorf("SMTP_TLS=%q is only allowed against the local mailpit capture server outside development", c.SMTPTLS)
	}
	if c.InternalToken != "" && len(c.InternalToken) < 32 {
		return fmt.Errorf("INTERNAL_API_TOKEN must be at least 32 characters")
	}
	if !c.VaultEnabled {
		return fmt.Errorf("ENV=%s requires VAULT_ENABLED=true", c.Env)
	}
	return nil
}

func env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
