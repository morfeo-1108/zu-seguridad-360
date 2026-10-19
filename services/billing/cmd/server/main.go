package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/zu-seguridad/billing/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg := config.Load()
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("Billing database connection failed: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Billing database handle failed: %v", err)
	}
	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("Billing database ping failed: %v", err)
	}

	// Billing routes, cron, outbox relay, and startup migrations remain disabled
	// until authentication, sandbox payment flow, idempotency, and schema changes
	// have dedicated tests and an explicit release gate.
	app := fiber.New(fiber.Config{AppName: "ZU Billing — test harness"})
	app.Use(recover.New())
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "service": "zu-billing", "mode": "health-only"})
	})

	go func() {
		if err := app.Listen(":" + cfg.Port); err != nil {
			log.Printf("Billing HTTP server stopped: %v", err)
		}
	}()
	log.Printf("ZU Billing health-only mode on :%s; financial endpoints are disabled", cfg.Port)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	_ = app.Shutdown()
	_ = sqlDB.Close()
}
