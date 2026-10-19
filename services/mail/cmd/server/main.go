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
	"github.com/zu-seguridad/mail/config"
	"github.com/zu-seguridad/mail/internal/service"
)

func main() {
	cfg := config.Load()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mailer := service.NewMailerService(map[string]string{
		"smtp_host": cfg.SMTPHost, "smtp_port": cfg.SMTPPort,
		"smtp_user": cfg.SMTPUser, "smtp_pass": cfg.SMTPPass,
		"from_email": cfg.FromEmail, "from_name": cfg.FromName,
	})

	// Consumer RabbitMQ: cola zu.mail.outbound
	if conn, err := amqp.Dial(cfg.RabbitMQURL); err == nil {
		ch, _ := conn.Channel()
		q, _ := ch.QueueDeclare("zu.mail.outbound", true, false, false, false, nil)
		deliveries, _ := ch.Consume(q.Name, "", false, false, false, false, nil)
		raw := make(chan []byte, 64)
		go func() {
			for d := range deliveries {
				raw <- d.Body
				d.Ack(false)
			}
		}()
		go mailer.ConsumeQueue(ctx, raw)
		defer conn.Close()
		log.Println("📬 Mail queue consumer activo: zu.mail.outbound")
	} else {
		log.Printf("⚠️  RabbitMQ no disponible (dev): %v", err)
	}

	app := fiber.New(fiber.Config{AppName: "ZU Mail v6.0"})
	app.Use(recover.New())
	app.Get("/health", func(c *fiber.Ctx) error { return c.JSON(fiber.Map{"status": "ok", "service": "zu-mail"}) })
	app.Post("/v1/send", func(c *fiber.Ctx) error {
		var msg service.MailMessage
		if err := c.BodyParser(&msg); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid body")
		}
		res := mailer.Send(c.Context(), msg)
		if res.Error != nil {
			return fiber.NewError(fiber.StatusBadGateway, res.Error.Error())
		}
		return c.JSON(fiber.Map{"sent": true, "provider": res.Provider, "sent_at": res.SentAt})
	})

	go func() {
		if err := app.Listen(":" + cfg.Port); err != nil {
			log.Fatal(err)
		}
	}()
	log.Printf("📧 ZU Mail running on :%s", cfg.Port)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	cancel()
	app.ShutdownWithContext(ctx)
}
