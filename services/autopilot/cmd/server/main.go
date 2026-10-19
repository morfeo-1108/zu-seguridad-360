package main

import (
	"log"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/zu-seguridad/autopilot/config"
)

func main() {
	if strings.EqualFold(os.Getenv("AUTOPILOT_ENABLED"), "true") {
		log.Fatal("AUTOPILOT_ENABLED=true rejected: worker triggers and autonomous actions are disabled until identity, authorization, tenant scope, and approval controls are implemented")
	}

	// Closed-pilot build: do not instantiate/start workers and do not expose
	// worker-management routes. Health is the only available endpoint.
	cfg := config.Load()
	app := fiber.New(fiber.Config{AppName: "ZU Autopilot (disabled)", DisableStartupMessage: true})
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "autopilot_enabled": false})
	})
	log.Printf("ZU Autopilot disabled (health-only) on :%s", cfg.Port)
	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
