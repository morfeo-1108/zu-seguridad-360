package config

import "os"

type Config struct {
	Port, RabbitMQURL, SMTPHost, SMTPPort, SMTPUser, SMTPPass, FromEmail, FromName, SendGridKey string

	VaultEnabled bool
	VaultAddr    string
	VaultToken   string
	VaultPath    string
}

func Load() *Config {
	return &Config{
		Port:        env("PORT", "4005"),
		RabbitMQURL: env("RABBITMQ_URL", "amqp://zurabbit:zurabbit123@rabbitmq:5672/zu_vhost"),
		SMTPHost:    env("SMTP_HOST", "smtp.sendgrid.net"),
		SMTPPort:    env("SMTP_PORT", "587"),
		SMTPUser:    env("SMTP_USER", "apikey"),
		SMTPPass:    env("SMTP_PASS", ""),
		FromEmail:   env("FROM_EMAIL", "noreply@zu360.co"),
		FromName:    env("FROM_NAME", "ZU Seguridad 360"),
		SendGridKey: env("SENDGRID_API_KEY", ""),

		VaultEnabled: env("VAULT_ENABLED", "false") == "true",
		VaultAddr:    env("VAULT_ADDR", "http://localhost:8200"),
		VaultToken:   env("VAULT_TOKEN", "zu-vault-dev-token"),
		VaultPath:    env("VAULT_SECRET_PATH", "secret/data/zu360/mail"),
	}
}
func env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
