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
	"github.com/zu-seguridad/notifications/config"
	"github.com/zu-seguridad/notifications/internal/auth"
	"github.com/zu-seguridad/notifications/internal/bus"
	"github.com/zu-seguridad/notifications/internal/service"
	"github.com/zu-seguridad/notifications/internal/store"
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
	st := store.New(db)
	if err := st.Migrate(); err != nil {
		log.Fatalf("migrate: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pub := &bus.Publisher{URL: cfg.RabbitMQURL}
	defer pub.Close()
	dispatcher := service.NewNotificationDispatcher(st, pub, map[string]string{
		"whatsapp_token": cfg.WhatsAppToken, "whatsapp_phone_number_id": cfg.WhatsAppPhoneNumberID,
		"whatsapp_template": cfg.WhatsAppTemplate,
	})

	consumer := &bus.Consumer{
		URL: cfg.RabbitMQURL, Queue: "zu.notifications.dispatch",
		Bindings: []bus.Binding{{Exchange: "zu.notifications", Key: "notify.#"}},
		Handler: func(ctx context.Context, d amqp.Delivery) error {
			var p service.NotificationPayload
			if err := json.Unmarshal(d.Body, &p); err != nil {
				return bus.Permanent(err)
			}
			if err := p.Validate(); err != nil {
				return bus.Permanent(err)
			}
			// Channels fail independently and each attempt is audited; one
			// channel failing must not re-send the others on retry.
			dispatcher.Dispatch(ctx, p)
			return nil
		},
	}
	go consumer.Run(ctx)

	app := fiber.New(fiber.Config{AppName: "ZU Notifications v6.1", DisableStartupMessage: true})
	app.Use(recover.New())
	app.Get("/health", func(c *fiber.Ctx) error {
		sqlDB, _ := db.DB()
		dbOK := sqlDB != nil && sqlDB.PingContext(c.Context()) == nil
		if !consumer.Up() || !dbOK {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"status": "degraded", "queue_consumer": consumer.Up(), "db": dbOK})
		}
		return c.JSON(fiber.Map{"status": "ok", "service": "zu-notifications", "queue_consumer": true, "db": true})
	})

	app.Post("/v1/dispatch", auth.Internal(cfg.InternalToken), func(c *fiber.Ctx) error {
		var p service.NotificationPayload
		if err := c.BodyParser(&p); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid body")
		}
		if err := p.Validate(); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
		return c.JSON(fiber.Map{"results": dispatcher.Dispatch(c.Context(), p)})
	})

	// In-app inbox for the signed-in user (bell in the dashboard).
	me := app.Group("/v1/me/notifications", auth.JWT(cfg.JWTSecret))
	me.Get("/", func(c *fiber.Ctx) error {
		id := auth.From(c)
		items, unread, err := st.List(c.Context(), id.TenantID, id.UserID, c.QueryBool("unread"), c.QueryInt("limit", 30))
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "could not load notifications")
		}
		return c.JSON(fiber.Map{"items": items, "unread": unread})
	})
	me.Post("/read-all", func(c *fiber.Ctx) error {
		id := auth.From(c)
		n, err := st.MarkAllRead(c.Context(), id.TenantID, id.UserID)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "could not update notifications")
		}
		return c.JSON(fiber.Map{"updated": n})
	})
	me.Post("/:id/read", func(c *fiber.Ctx) error {
		id := auth.From(c)
		err := st.MarkRead(c.Context(), id.TenantID, id.UserID, c.Params("id"))
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "notification not found")
		}
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid notification id")
		}
		return c.JSON(fiber.Map{"read": true})
	})

	go func() {
		if err := app.Listen(":" + cfg.Port); err != nil {
			log.Fatal(err)
		}
	}()
	log.Printf("ZU Notifications on :%s", cfg.Port)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	cancel()
	app.ShutdownWithTimeout(5 * time.Second)
}
