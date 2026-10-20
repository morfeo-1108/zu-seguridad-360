package config

import "os"

type Config struct {
	Port        string
	DatabaseURL string
	RabbitMQURL string
	RedisAddr   string
	JWTSecret   string

	VaultEnabled bool
	VaultAddr    string
	VaultToken   string
	VaultPath    string
}

func Load() *Config {
	return &Config{
		Port:        getEnv("PORT", "4008"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://zuadmin:zupassword@postgres:5432/zu_core"),
		RabbitMQURL: getEnv("RABBITMQ_URL", "amqp://zurabbit:zurabbit123@rabbitmq:5672/zu_vhost"),
		RedisAddr:   getEnv("REDIS_ADDR", "redis:6379"),
		JWTSecret:   getEnv("JWT_SECRET", "zu-super-secret-key-change-in-prod"),

		VaultEnabled: getEnv("VAULT_ENABLED", "false") == "true",
		VaultAddr:    getEnv("VAULT_ADDR", "http://localhost:8200"),
		VaultToken:   getEnv("VAULT_TOKEN", "zu-vault-dev-token"),
		VaultPath:    getEnv("VAULT_SECRET_PATH", "secret/data/zu360/attention"),
	}
}

func getEnv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
