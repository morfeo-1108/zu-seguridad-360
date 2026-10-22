package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/zu-seguridad/onboarding/config"
	"github.com/zu-seguridad/onboarding/internal/auth"
	"github.com/zu-seguridad/onboarding/internal/bus"
	"github.com/zu-seguridad/onboarding/internal/service"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	pub := &bus.Publisher{URL: cfg.RabbitMQURL}
	defer pub.Close()
	tracker := service.NewTracker(db, pub)
	if err := tracker.Migrate(); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	consumer := &bus.Consumer{
		URL: cfg.RabbitMQURL, Queue: "zu.onboarding.companies",
		Bindings: []bus.Binding{{Exchange: "zu.companies", Key: "company.created"}},
		Handler: func(ctx context.Context, d amqp.Delivery) error {
			var ev struct {
				TenantID  string `json:"tenant_id"`
				CompanyID string `json:"company_id"`
			}
			if err := json.Unmarshal(d.Body, &ev); err != nil || ev.TenantID == "" || ev.CompanyID == "" {
				return bus.Permanent(errors.New("malformed company.created event"))
			}
			_, err := tracker.Recompute(ctx, ev.TenantID, ev.CompanyID)
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return bus.Permanent(errors.New("company not found in Core for that tenant"))
			}
			return err
		},
	}
	go consumer.Run(ctx)
	go func() {
		t := time.NewTicker(cfg.RecomputeInterval)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				tracker.RecomputeAll(ctx)
			}
		}
	}()

	app := fiber.New(fiber.Config{AppName: "ZU Onboarding v6.1", DisableStartupMessage: true})
	app.Use(recover.New())
	app.Get("/health", func(c *fiber.Ctx) error {
		sqlDB, _ := db.DB()
		dbOK := sqlDB != nil && sqlDB.PingContext(c.Context()) == nil
		if !consumer.Up() || !dbOK {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"status": "degraded", "queue_consumer": consumer.Up(), "db": dbOK})
		}
		return c.JSON(fiber.Map{"status": "ok", "service": "zu-onboarding", "queue_consumer": true, "db": true})
	})
	v1 := app.Group("/v1/onboarding", auth.JWT(cfg.JWTSecret))
	// Active company of the session (company_id claim).
	v1.Get("/progress", func(c *fiber.Ctx) error {
		id := auth.From(c)
		if id.CompanyID == "" {
			return fiber.NewError(fiber.StatusBadRequest, "no active company in session")
		}
		p, err := tracker.Recompute(c.Context(), id.TenantID, id.CompanyID)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "company not found")
		}
		if err != nil {
			log.Printf("progress: %v", err)
			return fiber.NewError(fiber.StatusInternalServerError, "could not compute onboarding progress")
		}
		return c.JSON(p)
	})
	// Every company of the professional's tenant.
	v1.Get("/portfolio", func(c *fiber.Ctx) error {
		id := auth.From(c)
		ps, err := tracker.Portfolio(c.Context(), id.TenantID)
		if err != nil {
			log.Printf("portfolio: %v", err)
			return fiber.NewError(fiber.StatusInternalServerError, "could not compute onboarding portfolio")
		}
		return c.JSON(fiber.Map{"data": ps, "total": len(ps)})
	})

	go func() {
		if err := app.Listen(":" + cfg.Port); err != nil {
			log.Fatal(err)
		}
	}()
	log.Printf("ZU Onboarding on :%s (%d steps)", cfg.Port, len(service.Steps))
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	cancel()
	if err := app.ShutdownWithTimeout(5 * time.Second); err != nil {
		log.Printf("shutdown: %v", err)
	}
}
