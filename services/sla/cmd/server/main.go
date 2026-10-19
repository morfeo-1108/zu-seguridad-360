package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/zu-seguridad/sla/config"
	"github.com/zu-seguridad/sla/internal/service"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg := config.Load()
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("DB error: %v", err)
	}

	monitor := service.NewSLAMonitor(db)
	go runBreachWatcher(monitor)

	app := fiber.New(fiber.Config{AppName: "ZU SLA v6.0"})
	app.Use(recover.New())
	app.Get("/health", func(c *fiber.Ctx) error { return c.JSON(fiber.Map{"status": "ok", "service": "zu-sla"}) })

	app.Post("/v1/events/open", func(c *fiber.Ctx) error {
		var b struct{ TenantID, CompanyID, ResourceID, EventType string }
		if err := c.BodyParser(&b); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
		ev, err := monitor.OpenEvent(c.Context(), b.TenantID, b.CompanyID, b.ResourceID, b.EventType)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		return c.JSON(ev)
	})
	app.Post("/v1/events/close", func(c *fiber.Ctx) error {
		var b struct{ ResourceID, EventType string }
		c.BodyParser(&b)
		if err := monitor.CloseEvent(c.Context(), b.ResourceID, b.EventType); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
		return c.JSON(fiber.Map{"closed": true})
	})
	app.Get("/v1/dashboard/:tenantId", func(c *fiber.Ctx) error {
		d, err := monitor.GetDashboard(c.Context(), c.Params("tenantId"))
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		return c.JSON(d)
	})

	go func() {
		if err := app.Listen(":" + cfg.Port); err != nil {
			log.Fatal(err)
		}
	}()
	log.Printf("⏱️  ZU SLA running on :%s", cfg.Port)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
}

func runBreachWatcher(monitor *service.SLAMonitor) {
	ticker := time.NewTicker(1 * time.Minute)
	for range ticker.C {
		if breached, err := monitor.CheckBreaches(context.Background()); err == nil && len(breached) > 0 {
			log.Printf("🚨 %d nuevos SLA breaches", len(breached))
		}
	}
}
