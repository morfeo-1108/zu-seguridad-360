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
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/zu-seguridad/billing/config"
	"github.com/zu-seguridad/billing/internal/bus"
	"github.com/zu-seguridad/billing/internal/dian"
	"github.com/zu-seguridad/billing/internal/handler"
	"github.com/zu-seguridad/billing/internal/middleware"
	"github.com/zu-seguridad/billing/internal/model"
	"github.com/zu-seguridad/billing/internal/service"
	"github.com/zu-seguridad/billing/internal/wompi"
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
		log.Fatalf("Billing database connection failed: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil || sqlDB.Ping() != nil {
		log.Fatalf("Billing database unavailable: %v", err)
	}
	if cfg.AutoMigrate {
		if err := db.AutoMigrate(&model.Subscription{}, &model.Invoice{}, &model.IdempotencyKey{}, &model.OutboxEvent{}); err != nil {
			log.Fatalf("Billing migration failed: %v", err)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pub := &bus.Publisher{URL: cfg.RabbitMQURL}
	defer pub.Close()
	if err := declareExchange(cfg.RabbitMQURL, service.BillingExchange); err != nil {
		log.Fatalf("rabbitmq: %v", err)
	}
	outbox := service.NewOutboxRelay(db, pub)
	go outbox.Run(ctx)

	wompiClient := wompi.NewClient(wompi.Config{
		PublicKey: cfg.WompiPublicKey, PrivateKey: cfg.WompiPrivateKey,
		EventsSecret: cfg.WompiEventsSecret, IntegritySecret: cfg.WompiIntegritySecret,
	})
	dianCfg := service.DianConfig{
		Resolution: dian.Resolution{Number: cfg.DianResolution, Prefix: cfg.DianPrefix, From: cfg.DianFrom, To: cfg.DianTo,
			ValidFrom: cfg.DianValidFrom, ValidTo: cfg.DianValidTo, TechnicalKey: cfg.DianTechnicalKey},
		Issuer: dian.Issuer{NIT: cfg.DianIssuerNIT, DV: service.NITCheckDigit(cfg.DianIssuerNIT), Name: cfg.DianIssuerName,
			Address: cfg.DianIssuerAddress, City: "Bogotá", CityCode: "11001", Department: "Bogotá", DepartmentCode: "11",
			Email: cfg.DianIssuerEmail, TaxLevel: cfg.DianTaxLevel, SoftwareID: cfg.DianSoftwareID, SoftwarePIN: cfg.DianSoftwarePIN,
			ProviderNIT: cfg.DianIssuerNIT, ProviderDV: service.NITCheckDigit(cfg.DianIssuerNIT), Environment: cfg.DianEnv},
	}
	payments := service.NewPayments(db, wompiClient, outbox, dianCfg, cfg.PublicAppURL)
	if err := payments.Migrate(); err != nil {
		log.Fatalf("payments migration failed: %v", err)
	}

	// invoice.paid → electronic invoice (own queue: retried, dead-lettered, idempotent).
	dianConsumer := &bus.Consumer{
		URL: cfg.RabbitMQURL, Queue: "zu.billing.dian",
		Bindings: []bus.Binding{{Exchange: service.BillingExchange, Key: service.KeyInvoicePaid}},
		Handler: func(ctx context.Context, d amqp.Delivery) error {
			var ev struct {
				InvoiceID string `json:"invoice_id"`
			}
			if err := json.Unmarshal(d.Body, &ev); err != nil {
				return bus.Permanent(err)
			}
			id, err := uuid.Parse(ev.InvoiceID)
			if err != nil {
				return bus.Permanent(err)
			}
			_, err = payments.IssueElectronic(ctx, id)
			return err
		},
	}
	go dianConsumer.Run(ctx)

	go func() { // lost/late webhooks: ask Wompi directly
		t := time.NewTicker(3 * time.Minute)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				if n, err := payments.ReconcilePending(ctx); err != nil || n > 0 {
					log.Printf("reconcile pending: %d updated, err=%v", n, err)
				}
			}
		}
	}()

	billingSvc := service.NewBillingService(db)
	billingH := handler.NewBillingHandler(billingSvc, payments, cfg.WompiEventsSecret)
	jwt := middleware.JWTProtected(cfg)
	scoped := func(h fiber.Handler) []fiber.Handler {
		return []fiber.Handler{jwt, middleware.TenantScope("tenantId"), h}
	}

	app := fiber.New(fiber.Config{AppName: "ZU Billing", DisableStartupMessage: true})
	app.Use(recover.New())
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "service": "zu-billing", "wompi": wompiClient.Environment(),
			"dian_env": cfg.DianEnv, "dian_issuer_configured": dianCfg.Configured(), "dian_queue": dianConsumer.Up()})
	})
	app.Get("/v1/plans", billingH.ListPlans)
	app.Get("/v1/tenants/:tenantId/can-add-company", scoped(billingH.CanAddCompany)...)
	app.Get("/v1/tenants/:tenantId/subscription", scoped(billingH.GetSubscription)...)
	app.Post("/v1/tenants/:tenantId/subscription", scoped(billingH.CreateSubscription)...)
	app.Get("/v1/tenants/:tenantId/billing-profile", scoped(billingH.GetProfile)...)
	app.Put("/v1/tenants/:tenantId/billing-profile", scoped(billingH.SaveProfile)...)
	app.Get("/v1/tenants/:tenantId/invoices", scoped(billingH.ListInvoices)...)
	app.Post("/v1/tenants/:tenantId/invoices", scoped(billingH.IssueInvoice)...)
	app.Post("/v1/tenants/:tenantId/invoices/:invoiceId/checkout", scoped(billingH.Checkout)...)
	app.Get("/v1/tenants/:tenantId/invoices/:invoiceId/electronic", scoped(billingH.Electronic)...)
	app.Post("/v1/tenants/:tenantId/invoices/:invoiceId/electronic/retry", scoped(billingH.RetryElectronic)...)
	// Wompi calls this; authenticated by the event signature, then reconciled
	// against Wompi's API. Public URL: https://www.zuseguridad360.com/api/billing/webhooks/wompi
	app.Post("/v1/webhooks/wompi", billingH.WompiWebhook)

	go func() {
		if err := app.Listen(":" + cfg.Port); err != nil {
			log.Printf("Billing HTTP server stopped: %v", err)
		}
	}()
	log.Printf("ZU Billing on :%s — wompi=%s dian_env=%s issuer_configured=%v", cfg.Port, wompiClient.Environment(), cfg.DianEnv, dianCfg.Configured())

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	cancel()
	_ = app.ShutdownWithTimeout(5 * time.Second)
	_ = sqlDB.Close()
}

func declareExchange(url, name string) error {
	conn, err := amqp.Dial(url)
	if err != nil {
		return err
	}
	defer conn.Close()
	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	if err := ch.ExchangeDeclare(name, "topic", true, false, false, false, nil); err != nil {
		return errors.New("declare " + name + ": " + err.Error())
	}
	return nil
}
