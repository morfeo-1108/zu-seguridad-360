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
	"github.com/zu-seguridad/sla/config"
	"github.com/zu-seguridad/sla/internal/auth"
	"github.com/zu-seguridad/sla/internal/bus"
	"github.com/zu-seguridad/sla/internal/service"
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
	monitor := service.NewMonitor(db, pub, cfg.InspectionReview)
	if err := monitor.Migrate(); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	consumer := &bus.Consumer{
		URL: cfg.RabbitMQURL, Queue: "zu.sla.events",
		Bindings: []bus.Binding{
			{Exchange: "zu.companies", Key: "inspection.completed"},
			{Exchange: "zu.companies", Key: "inspection.approved"},
		},
		Handler: func(ctx context.Context, d amqp.Delivery) error {
			var ev service.CoreEvent
			if err := json.Unmarshal(d.Body, &ev); err != nil {
				return bus.Permanent(err)
			}
			if ev.Type == "" {
				ev.Type = d.RoutingKey
			}
			err := monitor.Handle(ctx, ev)
			if errors.Is(err, service.ErrMalformed) {
				return bus.Permanent(err)
			}
			return err
		},
	}
	go consumer.Run(ctx)

	go func() {
		t := time.NewTicker(cfg.CheckInterval)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				n, err := monitor.CheckBreaches(ctx)
				if err != nil {
					log.Printf("breach check failed (will retry): %v", err)
				} else if n > 0 {
					log.Printf("%d SLA breach(es) notified", n)
				}
			}
		}
	}()

	app := fiber.New(fiber.Config{AppName: "ZU SLA v6.1", DisableStartupMessage: true})
	app.Use(recover.New())
	app.Get("/health", func(c *fiber.Ctx) error {
		sqlDB, _ := db.DB()
		dbOK := sqlDB != nil && sqlDB.PingContext(c.Context()) == nil
		if !consumer.Up() || !dbOK {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"status": "degraded", "queue_consumer": consumer.Up(), "db": dbOK})
		}
		return c.JSON(fiber.Map{"status": "ok", "service": "zu-sla", "queue_consumer": true, "db": true})
	})

	// Scope comes from the token: tenant always, company when the session is
	// switched into one. A professional without an active company sees the
	// whole portfolio of their own tenant.
	v1 := app.Group("/v1/sla", auth.JWT(cfg.JWTSecret))
	v1.Get("/summary", func(c *fiber.Ctx) error {
		id := auth.From(c)
		s, err := monitor.Summary(c.Context(), id.TenantID, id.CompanyID)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "could not compute SLA summary")
		}
		return c.JSON(s)
	})
	v1.Get("/events", func(c *fiber.Ctx) error {
		id := auth.From(c)
		events, err := monitor.List(c.Context(), id.TenantID, id.CompanyID, c.Query("status"))
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "could not list SLA events")
		}
		return c.JSON(fiber.Map{"data": events, "total": len(events)})
	})

	go func() {
		if err := app.Listen(":" + cfg.Port); err != nil {
			log.Fatal(err)
		}
	}()
	log.Printf("ZU SLA on :%s (inspection review target %s)", cfg.Port, cfg.InspectionReview)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	cancel()
	app.ShutdownWithTimeout(5 * time.Second)
}
