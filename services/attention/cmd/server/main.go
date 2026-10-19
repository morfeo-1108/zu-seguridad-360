package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/zu-seguridad/attention/config"
	"github.com/zu-seguridad/attention/internal/service"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg := config.Load()
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("DB error: %v", err)
	}

	inbox := service.NewInboxService(db)

	app := fiber.New(fiber.Config{AppName: "ZU Attention v6.0"})
	app.Use(recover.New())
	app.Get("/health", func(c *fiber.Ctx) error { return c.JSON(fiber.Map{"status": "ok", "service": "zu-attention"}) })

	app.Get("/v1/inbox/:tenantId", func(c *fiber.Ctx) error {
		items, err := inbox.GetPrioritizedInbox(c.Context(), c.Params("tenantId"))
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		return c.JSON(fiber.Map{"data": items, "total": len(items)})
	})
	app.Get("/v1/briefing/:tenantId", func(c *fiber.Ctx) error {
		b, err := inbox.GetBriefing(c.Context(), c.Params("tenantId"))
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		return c.JSON(b)
	})
	app.Post("/v1/inbox/:itemId/acknowledge", func(c *fiber.Ctx) error {
		if err := inbox.Acknowledge(c.Context(), c.Params("itemId"), c.Get("X-Tenant-ID")); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
		return c.JSON(fiber.Map{"acknowledged": true})
	})

	go func() {
		if err := app.Listen(":" + cfg.Port); err != nil {
			log.Fatal(err)
		}
	}()
	log.Printf("📥 ZU Attention running on :%s", cfg.Port)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
}
