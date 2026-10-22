package main

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"log"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/zu-seguridad/mail/config"
	"github.com/zu-seguridad/mail/internal/service"
)

const (
	queueOutbound = "zu.mail.outbound"
	queueRetry    = "zu.mail.retry"
	queueDead     = "zu.mail.dead"
	maxAttempts   = 5
	retryDelayMS  = 60000
)

var consumerUp atomic.Bool

func main() {
	cfg := config.Load()
	if err := cfg.LoadVaultSecrets(); err != nil {
		log.Fatalf("vault secret loading failed: %v", err)
	}
	if err := cfg.ValidateProduction(); err != nil {
		log.Fatalf("config rejected: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mailer := service.NewMailerService(service.SMTPConfig{
		Host: cfg.SMTPHost, Port: cfg.SMTPPort, User: cfg.SMTPUser, Pass: cfg.SMTPPass,
		TLSMode: cfg.SMTPTLS, FromEmail: cfg.FromEmail, FromName: cfg.FromName,
	})
	go consumeForever(ctx, cfg.RabbitMQURL, mailer)

	app := fiber.New(fiber.Config{AppName: "ZU Mail v6.1", DisableStartupMessage: true})
	app.Use(recover.New())
	app.Get("/health", func(c *fiber.Ctx) error {
		if !consumerUp.Load() {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"status": "degraded", "service": "zu-mail", "queue_consumer": false})
		}
		return c.JSON(fiber.Map{"status": "ok", "service": "zu-mail", "queue_consumer": true})
	})
	// Internal endpoint for operators and service-to-service sends. Without a
	// configured token it is disabled: an unauthenticated send endpoint is an
	// open relay under our domain.
	app.Post("/v1/send", func(c *fiber.Ctx) error {
		if cfg.InternalToken == "" {
			return fiber.NewError(fiber.StatusServiceUnavailable, "send endpoint disabled (INTERNAL_API_TOKEN not configured)")
		}
		if subtle.ConstantTimeCompare([]byte(c.Get("X-Internal-Token")), []byte(cfg.InternalToken)) != 1 {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid internal token")
		}
		var msg service.MailMessage
		if err := c.BodyParser(&msg); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid body")
		}
		res, err := mailer.Send(msg)
		if errors.Is(err, service.ErrPermanent) {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
		if err != nil {
			log.Printf("send failed: %v", err)
			return fiber.NewError(fiber.StatusBadGateway, "smtp delivery failed")
		}
		return c.JSON(fiber.Map{"sent": true, "message_id": res.MessageID, "sent_at": res.SentAt})
	})

	go func() {
		if err := app.Listen(":" + cfg.Port); err != nil {
			log.Fatal(err)
		}
	}()
	log.Printf("ZU Mail on :%s (smtp %s:%s, tls=%s)", cfg.Port, cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPTLS)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	cancel()
	app.ShutdownWithTimeout(5 * time.Second)
}

// consumeForever keeps a consumer attached to zu.mail.outbound, reconnecting
// with backoff whenever RabbitMQ is unavailable or drops the connection.
func consumeForever(ctx context.Context, url string, mailer *service.MailerService) {
	backoff := time.Second
	for ctx.Err() == nil {
		err := consumeOnce(ctx, url, mailer)
		consumerUp.Store(false)
		if ctx.Err() != nil {
			return
		}
		log.Printf("mail consumer disconnected: %v (retrying in %s)", err, backoff)
		select {
		case <-ctx.Done():
			return
		case <-time.After(backoff):
		}
		if backoff < 30*time.Second {
			backoff *= 2
		}
	}
}

func consumeOnce(ctx context.Context, url string, mailer *service.MailerService) error {
	conn, err := amqp.Dial(url)
	if err != nil {
		return err
	}
	defer conn.Close()
	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	if err := declareTopology(ch); err != nil {
		return err
	}
	if err := ch.Qos(5, 0, false); err != nil {
		return err
	}
	deliveries, err := ch.Consume(queueOutbound, "zu-mail", false, false, false, false, nil)
	if err != nil {
		return err
	}
	consumerUp.Store(true)
	log.Printf("mail consumer attached to %s", queueOutbound)
	closed := conn.NotifyClose(make(chan *amqp.Error, 1))
	for {
		select {
		case <-ctx.Done():
			return nil
		case e := <-closed:
			return e
		case d, ok := <-deliveries:
			if !ok {
				return errors.New("delivery channel closed")
			}
			handle(ctx, ch, d, mailer)
		}
	}
}

func declareTopology(ch *amqp.Channel) error {
	if _, err := ch.QueueDeclare(queueOutbound, true, false, false, false, amqp.Table{
		"x-dead-letter-exchange": "", "x-dead-letter-routing-key": queueRetry,
	}); err != nil {
		return err
	}
	if _, err := ch.QueueDeclare(queueRetry, true, false, false, false, amqp.Table{
		"x-message-ttl": int32(retryDelayMS), "x-dead-letter-exchange": "", "x-dead-letter-routing-key": queueOutbound,
	}); err != nil {
		return err
	}
	_, err := ch.QueueDeclare(queueDead, true, false, false, false, nil)
	return err
}

func handle(ctx context.Context, ch *amqp.Channel, d amqp.Delivery, mailer *service.MailerService) {
	var msg service.MailMessage
	if err := json.Unmarshal(d.Body, &msg); err != nil {
		deadLetter(ctx, ch, d, "invalid json: "+err.Error())
		return
	}
	res, err := mailer.Send(msg)
	switch {
	case err == nil:
		log.Printf("mail sent %s to %d recipient(s) [tenant=%s]", res.MessageID, len(msg.To)+len(msg.CC), msg.TenantID)
		d.Ack(false)
	case errors.Is(err, service.ErrPermanent):
		deadLetter(ctx, ch, d, err.Error())
	case attempts(d) >= maxAttempts-1:
		deadLetter(ctx, ch, d, "max attempts reached: "+err.Error())
	default:
		log.Printf("mail send failed (attempt %d/%d, retry in %ds): %v", attempts(d)+1, maxAttempts, retryDelayMS/1000, err)
		d.Nack(false, false) // dead-letters into zu.mail.retry, back after the TTL
	}
}

func attempts(d amqp.Delivery) int {
	deaths, _ := d.Headers["x-death"].([]interface{})
	for _, x := range deaths {
		if t, ok := x.(amqp.Table); ok && t["queue"] == queueOutbound {
			if n, ok := t["count"].(int64); ok {
				return int(n)
			}
		}
	}
	return 0
}

func deadLetter(ctx context.Context, ch *amqp.Channel, d amqp.Delivery, reason string) {
	log.Printf("mail dead-lettered: %s", reason)
	err := ch.PublishWithContext(ctx, "", queueDead, false, false, amqp.Publishing{
		ContentType: "application/json", DeliveryMode: amqp.Persistent, Body: d.Body,
		Headers: amqp.Table{"x-zu-reason": reason},
	})
	if err != nil {
		d.Nack(false, true) // keep it rather than lose it
		return
	}
	d.Ack(false)
}
