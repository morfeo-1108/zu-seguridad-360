package handler

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"time"
)

type HealthHandler struct {
	db    *gorm.DB
	redis *redis.Client
}

func NewHealthHandler(db *gorm.DB, redis *redis.Client) *HealthHandler {
	return &HealthHandler{db: db, redis: redis}
}

func (h *HealthHandler) Check(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 3*time.Second)
	defer cancel()

	status := fiber.Map{
		"service":   "zu-core",
		"version":   "6.0.0",
		"timestamp": time.Now().UTC(),
		"checks":    fiber.Map{},
	}

	checks := status["checks"].(fiber.Map)

	// PostgreSQL check
	sqlDB, _ := h.db.DB()
	if err := sqlDB.PingContext(ctx); err != nil {
		checks["postgres"] = fiber.Map{"status": "down", "error": err.Error()}
		return c.Status(fiber.StatusServiceUnavailable).JSON(status)
	}
	checks["postgres"] = fiber.Map{"status": "up"}

	// Redis check
	if err := h.redis.Ping(ctx).Err(); err != nil {
		checks["redis"] = fiber.Map{"status": "down", "error": err.Error()}
	} else {
		checks["redis"] = fiber.Map{"status": "up"}
	}

	status["status"] = "ok"
	return c.JSON(status)
}
