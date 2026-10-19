package config

import "os"

type Config struct {
	Port        string
	DatabaseURL string
	RedisAddr   string
	RabbitMQURL string
	CoreURL     string
	AIURL       string
}

func Load() *Config {
	return &Config{
		Port:        getEnv("PORT", "4007"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://zuadmin:zupassword@localhost:5432/zu_core"),
		RedisAddr:   getEnv("REDIS_ADDR", "localhost:6379"),
		RabbitMQURL: getEnv("RABBITMQ_URL", "amqp://zurabbit:zurabbit123@localhost:5672/zu_vhost"),
		CoreURL:     getEnv("CORE_URL", "http://core:4001"),
		AIURL:       getEnv("AI_URL", "http://ai:4002"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
