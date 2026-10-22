package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/robfig/cron/v3"
	"github.com/zu-seguridad/autopilot/config"
	"github.com/zu-seguridad/autopilot/internal/auth"
	"github.com/zu-seguridad/autopilot/internal/briefing"
	"github.com/zu-seguridad/autopilot/internal/bus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Autopilot runs scheduled, read-only jobs over Core's records (today: the
// morning briefing). It never acts on a record by itself. AUTOPILOT_ENABLED
// stays reserved for autonomous actions, which remain disabled until
// identity, authorization, tenant scope and human-approval controls exist.
func main() {
	if strings.EqualFold(os.Getenv("AUTOPILOT_ENABLED"), "true") {
		log.Fatal("AUTOPILOT_ENABLED=true rejected: autonomous actions are disabled until identity, authorization, tenant scope, and approval controls are implemented")
	}
	cfg := config.Load()
	if err := cfg.LoadVaultSecrets(); err != nil {
		log.Fatalf("vault secret loading failed: %v", err)
	}
	if err := cfg.ValidateProduction(); err != nil {
		log.Fatalf("config rejected: %v", err)
	}
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{Logger: logger.Default.LogMode(logger.Warn)})
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	pub := &bus.Publisher{URL: cfg.RabbitMQURL}
	defer pub.Close()
	svc := briefing.New(db, pub)
	svc.OnlyTenant = os.Getenv("BRIEFING_ONLY_TENANT")
	if err := svc.Migrate(); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	sched := cron.New(cron.WithLocation(briefing.Bogota))
	if _, err := sched.AddFunc(cfg.BriefingCron, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()
		n, err := svc.RunDaily(ctx)
		log.Printf("daily briefing: %d sent, err=%v", n, err)
	}); err != nil {
		log.Fatalf("invalid BRIEFING_CRON %q: %v", cfg.BriefingCron, err)
	}
	sched.Start()
	defer sched.Stop()

	app := fiber.New(fiber.Config{AppName: "ZU Autopilot v6.1", DisableStartupMessage: true})
	app.Use(recover.New())
	app.Get("/health", func(c *fiber.Ctx) error {
		sqlDB, _ := db.DB()
		if sqlDB == nil || sqlDB.PingContext(c.Context()) != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"status": "degraded", "db": false})
		}
		next := sched.Entries()[0].Next
		return c.JSON(fiber.Map{"status": "ok", "service": "zu-autopilot", "autonomous_actions": false, "next_briefing": next})
	})
	v1 := app.Group("/v1/briefing", auth.JWT(cfg.JWTSecret))
	v1.Get("/today", func(c *fiber.Ctx) error {
		id := auth.From(c)
		b, err := svc.Today(c.Context(), id.TenantID, id.UserID)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "no briefing yet")
		}
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "could not load briefing")
		}
		return c.JSON(b)
	})
	// Recompute now for the caller (dashboard "actualizar"); does not notify.
	v1.Post("/refresh", func(c *fiber.Ctx) error {
		id := auth.From(c)
		b, _, err := svc.Build(c.Context(), id.TenantID, id.UserID)
		if err != nil {
			log.Printf("refresh: %v", err)
			return fiber.NewError(fiber.StatusInternalServerError, "could not build briefing")
		}
		return c.JSON(b)
	})

	go func() {
		if err := app.Listen(":" + cfg.Port); err != nil {
			log.Fatal(err)
		}
	}()
	log.Printf("ZU Autopilot on :%s (briefing cron %q America/Bogota)", cfg.Port, cfg.BriefingCron)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	if err := app.ShutdownWithTimeout(5 * time.Second); err != nil {
		log.Printf("shutdown: %v", err)
	}
}
