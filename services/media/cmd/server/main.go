package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/zu-seguridad/media/config"
	"github.com/zu-seguridad/media/internal/handler"
	"github.com/zu-seguridad/media/internal/middleware"
)

func main() {
	cfg := config.Load()
	if err := cfg.LoadVaultSecrets(); err != nil {
		log.Fatalf("vault secret loading failed: %v", err)
	}
	uploadH := handler.NewUploadHandler(cfg.S3Bucket, cfg.AWSRegion, cfg.S3Endpoint, cfg.S3PublicEndpoint, cfg.AWSAccessKeyID, cfg.AWSSecretAccessKey).
		WithLinkSecret(cfg.JWTSecret)

	app := fiber.New(fiber.Config{AppName: "ZU Media v6.0", BodyLimit: 50 * 1024 * 1024})
	app.Use(recover.New())
	app.Get("/health", func(c *fiber.Ctx) error { return c.JSON(fiber.Map{"status": "ok", "service": "zu-media"}) })
	app.Post("/v1/upload", middleware.JWTProtected(cfg), uploadH.Upload)
	app.Post("/v1/files/link", middleware.JWTProtected(cfg), uploadH.Link)
	app.Get("/v1/files/*", uploadH.Download) // authorized by the signed link itself

	go func() {
		if err := app.Listen(":" + cfg.Port); err != nil {
			log.Fatal(err)
		}
	}()
	log.Printf("📁 ZU Media running on :%s (Sophos + S3 WORM)", cfg.Port)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
}
