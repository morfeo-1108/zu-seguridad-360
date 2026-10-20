package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/zu-seguridad/onboarding/config"
	"github.com/zu-seguridad/onboarding/internal/service"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type amqpPublisher struct{ ch *amqp.Channel }

func (a *amqpPublisher) Publish(exchange, key string, body []byte) error {
	if a.ch == nil {
		return nil
	}
	return a.ch.Publish(exchange, key, false, false, amqp.Publishing{ContentType: "application/json", Body: body})
}

func main() {
	cfg := config.Load()
	if err := cfg.LoadVaultSecrets(); err != nil {
		log.Fatalf("vault secret loading failed: %v", err)
	}
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("DB error: %v", err)
	}

	pub := &amqpPublisher{}
	if conn, err := amqp.Dial(cfg.RabbitMQURL); err == nil {
		pub.ch, _ = conn.Channel()
		defer conn.Close()
	} else {
		log.Printf("⚠️  RabbitMQ no disponible (dev): %v", err)
	}

	svc := service.NewOnboardingService(db, pub)

	app := fiber.New(fiber.Config{AppName: "ZU Onboarding v6.0"})
	app.Use(recover.New())
	app.Get("/health", func(c *fiber.Ctx) error { return c.JSON(fiber.Map{"status": "ok", "service": "zu-onboarding"}) })
	app.Get("/v1/steps", func(c *fiber.Ctx) error { return c.JSON(fiber.Map{"steps": svc.GetAllSteps()}) })
	app.Post("/v1/companies/:companyId/start", func(c *fiber.Ctx) error {
		state, err := svc.Start(c.Context(), c.Params("companyId"), c.Get("X-Tenant-ID"))
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
		return c.Status(fiber.StatusCreated).JSON(state)
	})
	app.Get("/v1/companies/:companyId/state", func(c *fiber.Ctx) error {
		state, err := svc.GetState(c.Context(), c.Params("companyId"))
		if err != nil {
			return fiber.NewError(fiber.StatusNotFound, err.Error())
		}
		return c.JSON(state)
	})
	app.Post("/v1/companies/:companyId/steps/:stepKey/complete", func(c *fiber.Ctx) error {
		state, err := svc.CompleteStep(c.Context(), c.Params("companyId"), c.Params("stepKey"))
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
		return c.JSON(state)
	})

	go func() {
		if err := app.Listen(":" + cfg.Port); err != nil {
			log.Fatal(err)
		}
	}()
	log.Printf("🚀 ZU Onboarding running on :%s — 9 pasos autónomos", cfg.Port)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
}
