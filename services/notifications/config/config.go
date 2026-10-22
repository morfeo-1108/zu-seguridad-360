package config

import (
	"fmt"
	"os"
)

type Config struct {
	Env                   string
	Port                  string
	DatabaseURL           string
	RabbitMQURL           string
	JWTSecret             string
	InternalToken         string
	WhatsAppToken         string
	WhatsAppPhoneNumberID string
	WhatsAppTemplate      string

	VaultEnabled bool
	VaultAddr    string
	VaultToken   string
	VaultPath    string
}

func Load() *Config {
	return &Config{
		Env:                   getEnv("ENV", "development"),
		Port:                  getEnv("PORT", "4006"),
		DatabaseURL:           getEnv("DATABASE_URL", "postgres://zuadmin:zupassword@postgres:5432/zu_core"),
		RabbitMQURL:           getEnv("RABBITMQ_URL", "amqp://zurabbit:zurabbit123@rabbitmq:5672/zu_vhost"),
		JWTSecret:             getEnv("JWT_SECRET", ""),
		InternalToken:         getEnv("INTERNAL_API_TOKEN", ""),
		WhatsAppToken:         getEnv("WHATSAPP_TOKEN", ""),
		WhatsAppPhoneNumberID: getEnv("WHATSAPP_PHONE_NUMBER_ID", ""),
		WhatsAppTemplate:      getEnv("WHATSAPP_TEMPLATE", ""),

		VaultEnabled: getEnv("VAULT_ENABLED", "false") == "true",
		VaultAddr:    getEnv("VAULT_ADDR", "http://localhost:8200"),
		VaultToken:   getEnv("VAULT_TOKEN", "zu-vault-dev-token"),
		VaultPath:    getEnv("VAULT_SECRET_PATH", "secret/data/zu360/notifications"),
	}
}

// ValidateProduction: outside development every secret must come from Vault
// and the JWT secret must be a real one (Core enforces the same minimum).
func (c *Config) ValidateProduction() error {
	if len(c.JWTSecret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters")
	}
	if c.Env == "development" {
		return nil
	}
	if !c.VaultEnabled {
		return fmt.Errorf("ENV=%s requires VAULT_ENABLED=true", c.Env)
	}
	return nil
}

func getEnv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
