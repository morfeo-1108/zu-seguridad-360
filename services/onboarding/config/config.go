package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Env               string
	Port              string
	DatabaseURL       string
	RabbitMQURL       string
	JWTSecret         string
	RecomputeInterval time.Duration

	VaultEnabled bool
	VaultAddr    string
	VaultToken   string
	VaultPath    string
}

func Load() *Config {
	return &Config{
		Env:               getEnv("ENV", "development"),
		Port:              getEnv("PORT", "4009"),
		DatabaseURL:       getEnv("DATABASE_URL", "postgres://zuadmin:zupassword@postgres:5432/zu_core"),
		RabbitMQURL:       getEnv("RABBITMQ_URL", "amqp://zurabbit:zurabbit123@rabbitmq:5672/zu_vhost"),
		JWTSecret:         getEnv("JWT_SECRET", ""),
		RecomputeInterval: time.Duration(getInt("ONBOARDING_RECOMPUTE_SECONDS", 600)) * time.Second,

		VaultEnabled: getEnv("VAULT_ENABLED", "false") == "true",
		VaultAddr:    getEnv("VAULT_ADDR", "http://localhost:8200"),
		VaultToken:   getEnv("VAULT_TOKEN", "zu-vault-dev-token"),
		VaultPath:    getEnv("VAULT_SECRET_PATH", "secret/data/zu360/onboarding"),
	}
}

func (c *Config) ValidateProduction() error {
	if len(c.JWTSecret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters")
	}
	if c.Env != "development" && !c.VaultEnabled {
		return fmt.Errorf("ENV=%s requires VAULT_ENABLED=true", c.Env)
	}
	return nil
}

func getInt(k string, d int) int {
	if n, err := strconv.Atoi(os.Getenv(k)); err == nil && n > 0 {
		return n
	}
	return d
}

func getEnv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
