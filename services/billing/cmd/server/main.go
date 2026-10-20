package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/zu-seguridad/billing/config"
	"github.com/zu-seguridad/billing/internal/handler"
	"github.com/zu-seguridad/billing/internal/middleware"
	"github.com/zu-seguridad/billing/internal/model"
	"github.com/zu-seguridad/billing/internal/service"
	"github.com/zu-seguridad/billing/internal/wompi"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg := config.Load()
	if err := cfg.LoadVaultSecrets(); err != nil {
		log.Fatalf("vault secret loading failed: %v", err)
	}
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("Billing database connection failed: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Billing database handle failed: %v", err)
	}
	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("Billing database ping failed: %v", err)
	}
	if cfg.AutoMigrate {
		if err := db.AutoMigrate(&model.Subscription{}, &model.Invoice{}, &model.IdempotencyKey{}, &model.OutboxEvent{}); err != nil {
			log.Fatalf("Billing migration failed: %v", err)
		}
	} else {
		log.Println("AUTO_MIGRATE=false; startup assumes subscriptions/invoices schema was provisioned externally")
	}

	// Payment collection (Wompi), cron invoicing, and the outbox relay remain
	// disabled pending a sandbox payment flow with dedicated tests. The
	// company-quota check below is read-only (it creates/mutates nothing) and
	// is required by Core's company creation flow per the per-professional
	// plan-quota model (basico=10, estandar=20, premium=50 companies), so it
	// is enabled with real JWT + tenant-scope enforcement.
	app := fiber.New(fiber.Config{AppName: "ZU Billing"})
	app.Use(recover.New())
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "service": "zu-billing", "mode": "quota-check-only"})
	})

	wompiClient := wompi.NewClient(cfg.WompiURL, cfg.WompiKey)
	billingSvc := service.NewBillingService(db, wompiClient)
	billingH := handler.NewBillingHandler(billingSvc, cfg.WompiEventsSecret)
	jwt := middleware.JWTProtected(cfg)
	app.Get("/v1/tenants/:tenantId/can-add-company", jwt, middleware.TenantScope("tenantId"), billingH.CanAddCompany)
	// No JWT: Wompi itself calls this, authenticated only by its own event
	// signature (verified inside the handler).
	app.Post("/v1/webhooks/wompi", billingH.WompiWebhook)

	go func() {
		if err := app.Listen(":" + cfg.Port); err != nil {
			log.Printf("Billing HTTP server stopped: %v", err)
		}
	}()
	log.Printf("ZU Billing health-only mode on :%s; financial endpoints are disabled", cfg.Port)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	_ = app.Shutdown()
	_ = sqlDB.Close()
}
