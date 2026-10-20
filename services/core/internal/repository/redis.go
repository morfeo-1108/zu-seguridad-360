package repository

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
	"github.com/zu-seguridad/core/config"
)

func NewRedisClient(cfg *config.Config) *redis.Client {
	// Username vacío (default en dev) autentica contra el usuario "default"
	// de Redis, igual que antes — no rompe nada existente. Con un usuario
	// ACL dedicado (RedisUsername configurado), Core puede tener su propia
	// credencial sin rotar el requirepass compartido que usan los demás
	// servicios (billing/attention/media), evitando romperlos.
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Username: cfg.RedisUsername,
		Password: cfg.RedisPassword,
		DB:       0,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("failed to connect to redis: %v", err)
	}

	log.Println("✅ Redis connected")
	return client
}
