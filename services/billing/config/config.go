package config

import "os"

type Config struct {
	Port, DatabaseURL, RabbitMQURL, WompiURL, WompiKey, WompiEventsSecret, JWTSecret string
	AutoMigrate                                                                      bool

	VaultEnabled bool
	VaultAddr    string
	VaultToken   string
	VaultPath    string
}

func Load() *Config {
	return &Config{
		Port:              env("PORT", "4010"),
		DatabaseURL:       env("DATABASE_URL", "postgres://zuadmin:zupassword@postgres:5432/zu_core"),
		RabbitMQURL:       env("RABBITMQ_URL", "amqp://zurabbit:zurabbit123@rabbitmq:5672/zu_vhost"),
		WompiURL:          env("WOMPI_URL", "https://production.wompi.co/v1"),
		WompiKey:          env("WOMPI_KEY", ""),
		WompiEventsSecret: env("WOMPI_EVENTS_SECRET", ""),
		JWTSecret:         env("JWT_SECRET", "zu-super-secret-key-change-in-prod"),
		AutoMigrate:       env("AUTO_MIGRATE", "false") == "true",

		VaultEnabled: env("VAULT_ENABLED", "false") == "true",
		VaultAddr:    env("VAULT_ADDR", "http://localhost:8200"),
		VaultToken:   env("VAULT_TOKEN", "zu-vault-dev-token"),
		VaultPath:    env("VAULT_SECRET_PATH", "secret/data/zu360/billing"),
	}
}
func env(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
