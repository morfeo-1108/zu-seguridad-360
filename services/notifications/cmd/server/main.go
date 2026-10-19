package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/zu-seguridad/notifications/config"
	"github.com/zu-seguridad/notifications/internal/service"
)

func main() {
	cfg := config.Load()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dispatcher := service.NewNotificationDispatcher(map[string]string{
		"whatsapp_token": cfg.WhatsAppToken, "whatsapp_phone_number_id": cfg.WhatsAppPhoneNumberID,
		"push_api_key": cfg.PushAPIKey,
	})

	if conn, err := amqp.Dial(cfg.RabbitMQURL); err == nil {
		ch, _ := conn.Channel()
		q, _ := ch.QueueDeclare("zu.notifications.dispatch", true, false, false, false, nil)
		deliveries, _ := ch.Consume(q.Name, "", false, false, false, false, nil)
		raw := make(chan []byte, 64)
		go func() {
			for d := range deliveries {
				raw <- d.Body
				d.Ack(false)
			}
		}()
		go dispatcher.ConsumeQueue(ctx, raw)
		defer conn.Close()
		log.Println("📬 Notifications consumer activo: zu.notifications.dispatch")
	} else {
		log.Printf("⚠️  RabbitMQ no disponible (dev): %v", err)
	}

	app := fiber.New(fiber.Config{AppName: "ZU Notifications v6.0"})
	app.Use(recover.New())
	app.Get("/health", func(c *fiber.Ctx) error { return c.JSON(fiber.Map{"status": "ok", "service": "zu-notifications"}) })
	app.Post("/v1/dispatch", func(c *fiber.Ctx) error {
		var p service.NotificationPayload
		if err := c.BodyParser(&p); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid body")
		}
		return c.JSON(fiber.Map{"results": dispatcher.Dispatch(c.Context(), p)})
	})

	go func() {
		if err := app.Listen(":" + cfg.Port); err != nil {
			log.Fatal(err)
		}
	}()
	log.Printf("🔔 ZU Notifications running on :%s", cfg.Port)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	cancel()
}
