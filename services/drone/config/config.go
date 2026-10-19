package config

import "os"

type Config struct {
	Port        string
	DatabaseURL string
	MQTTBroker  string
	RedisAddr   string
}

func Load() *Config {
	return &Config{
		Port:        getEnv("PORT", "4012"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://zuadmin:zupassword@postgres:5432/zu_core"),
		MQTTBroker:  getEnv("MQTT_BROKER", "tcp://emqx:1883"),
		RedisAddr:   getEnv("REDIS_ADDR", "redis:6379"),
	}
}

func getEnv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
