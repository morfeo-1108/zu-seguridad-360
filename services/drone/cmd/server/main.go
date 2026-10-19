package main

import (
	"log"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	if strings.EqualFold(os.Getenv("DRONE_ENABLED"), "true") {
		log.Fatal("DRONE_ENABLED=true rejected: mission, stream, and flight APIs are disabled until verified identity, tenant scope, operator authorization, and hardware safety checks are implemented")
	}

	// Closed-pilot build: this binary exposes only health. It does not open a
	// database/MQTT connection or register drone, camera, or flight handlers.
	app := fiber.New(fiber.Config{AppName: "ZU Drone (disabled)", DisableStartupMessage: true})
	app.Use(recover.New())
	app.Get("/v1/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "service": "zu-drone", "mode": "disabled"})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "4012"
	}
	log.Printf("ZU Drone disabled (health-only) on :%s", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatal(err)
	}
}
