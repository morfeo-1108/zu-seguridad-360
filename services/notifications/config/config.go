package config

import "os"

type Config struct {
	Port                  string
	DatabaseURL           string
	RabbitMQURL           string
	RedisAddr             string
	WhatsAppToken         string
	WhatsAppPhoneNumberID string
	PushAPIKey            string
}

func Load() *Config {
	return &Config{
		Port:                  getEnv("PORT", "4006"),
		DatabaseURL:           getEnv("DATABASE_URL", "postgres://zuadmin:zupassword@postgres:5432/zu_core"),
		RabbitMQURL:           getEnv("RABBITMQ_URL", "amqp://zurabbit:zurabbit123@rabbitmq:5672/zu_vhost"),
		RedisAddr:             getEnv("REDIS_ADDR", "redis:6379"),
		WhatsAppToken:         getEnv("WHATSAPP_TOKEN", ""),
		WhatsAppPhoneNumberID: getEnv("WHATSAPP_PHONE_NUMBER_ID", ""),
		PushAPIKey:            getEnv("PUSH_API_KEY", ""),
	}
}

func getEnv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
