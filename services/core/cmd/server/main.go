package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/zu-seguridad/core/config"
	"github.com/zu-seguridad/core/internal/handler"
	"github.com/zu-seguridad/core/internal/middleware"
	"github.com/zu-seguridad/core/internal/repository"
	"github.com/zu-seguridad/core/internal/service"
)

func main() {
	cfg := config.Load()
	if err := cfg.ValidateProduction(); err != nil {
		log.Fatalf("unsafe production configuration: %v", err)
	}
	if err := cfg.ValidateSyntheticPilot(); err != nil {
		log.Fatalf("unsafe synthetic pilot configuration: %v", err)
	}
	if err := cfg.ValidateGoogleSignup(); err != nil {
		log.Fatalf("unsafe Google signup configuration: %v", err)
	}

	// ── Infraestructura ──────────────────────────────────────
	db := repository.NewPostgresDB(cfg)
	if cfg.AutoMigrate {
		if err := repository.RunAllMigrations(db); err != nil {
			log.Fatalf("migration failed: %v", err)
		}
	} else {
		log.Println("AUTO_MIGRATE=false; startup assumes the database schema was provisioned externally")
	}
	redis := repository.NewRedisClient(cfg)
	amqp := repository.NewRabbitMQ(cfg)
	enforcer, err := middleware.NewCasbinEnforcer(db)
	if err != nil {
		log.Fatalf("authorization initialization failed: %v", err)
	}

	// ── Repositories ────────────────────────────────────────
	companyRepo := repository.NewCompanyRepository(db, cfg.SyntheticPilotEnabled)
	employeeRepo := repository.NewEmployeeRepository(db)
	userRepo := repository.NewUserRepository(db, cfg.SyntheticPilotEnabled)
	incidentRepo := repository.NewIncidentRepository(db)

	// ── Services ────────────────────────────────────────────
	authSvc, err := service.NewAuthService(userRepo, redis, cfg)
	if err != nil {
		log.Fatalf("authentication initialization failed: %v", err)
	}
	companySvc := service.NewCompanyService(companyRepo, amqp, cfg.SyntheticPilotEnabled)
	employeeSvc := service.NewEmployeeService(employeeRepo, amqp, cfg.SyntheticPilotEnabled)
	incidentSvc := service.NewIncidentService(incidentRepo, amqp, cfg.SyntheticPilotEnabled)
	trainingSvc := service.NewTrainingService(db, amqp, cfg.SyntheticPilotEnabled)
	riskSvc := service.NewRiskService(db)
	mgmtSvc := service.NewManagementSystemService(db)
	watchSvc := service.NewRegulatoryWatchService(db)

	// ── Handlers ────────────────────────────────────────────
	healthH := handler.NewHealthHandler(db, redis)
	authH := handler.NewAuthHandler(authSvc)
	companyH := handler.NewCompanyHandler(companySvc)
	employeeH := handler.NewEmployeeHandler(employeeSvc)
	incidentH := handler.NewIncidentHandler(incidentSvc)
	trainingH := handler.NewTrainingHandler(trainingSvc)
	mgmtH := handler.NewManagementHandler(mgmtSvc, watchSvc)
	metaH := handler.NewSystemMetaHandler()
	_ = riskSvc // handler en siguiente iteración

	// ── Fiber App ───────────────────────────────────────────
	app := fiber.New(fiber.Config{
		AppName:      "ZU Core v6.0",
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		BodyLimit:    50 * 1024 * 1024, // 50MB para uploads
		ErrorHandler: handler.GlobalErrorHandler,
	})

	app.Use(recover.New())
	app.Use(requestid.New())
	app.Use(compress.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.AllowedOrigins,
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Authorization,X-Request-ID",
		AllowCredentials: true,
	}))
	app.Use(limiter.New(limiter.Config{Max: 120, Expiration: 1 * time.Minute}))
	app.Use(middleware.OtelTracing())
	app.Use(middleware.Logger())

	// ── Routes ──────────────────────────────────────────────
	app.Get("/v1/health", healthH.Check)

	// Auth
	auth := app.Group("/v1/auth")
	auth.Post("/login", authH.Login)
	auth.Post("/refresh", authH.RefreshToken)
	auth.Get("/google/config", authH.GoogleConfig)
	auth.Post("/google/challenge", middleware.RequireAllowedOrigin(cfg.AllowedOrigins), limiter.New(limiter.Config{Max: 10, Expiration: time.Minute}), middleware.RequestBodyLimit(16*1024), authH.GoogleChallenge)
	auth.Post("/google/start", middleware.RequireAllowedOrigin(cfg.AllowedOrigins), limiter.New(limiter.Config{Max: 8, Expiration: time.Minute}), middleware.RequestBodyLimit(16*1024), authH.GoogleStart)
	auth.Post("/google/complete", middleware.RequireAllowedOrigin(cfg.AllowedOrigins), limiter.New(limiter.Config{Max: 5, Expiration: time.Minute}), middleware.RequestBodyLimit(16*1024), authH.GoogleComplete)
	auth.Post("/logout", middleware.JWTProtected(cfg), authH.Logout)
	auth.Get("/me", middleware.JWTProtected(cfg), authH.Me)
	auth.Post("/switch-company", middleware.JWTProtected(cfg), authH.SwitchCompany)

	// Protección JWT para todas las rutas siguientes
	jwt := middleware.JWTProtected(cfg)

	// Public certificate verification is not registered until its response is
	// minimized and independently reviewed for personal-data exposure.

	// Companies
	co := app.Group("/v1/companies", jwt, middleware.Authorize(enforcer, "companies"))
	co.Get("/", companyH.List)
	co.Post("/", companyH.Create)
	coByID := app.Group("/v1/companies/:id", jwt, middleware.CompanyScope("id"), middleware.CompanyMembershipScope("id", userRepo), middleware.Authorize(enforcer, "companies"))
	coByID.Get("/", companyH.GetByID)
	coByID.Put("/", companyH.Update)
	coByID.Delete("/", companyH.Delete)
	coByID.Get("/score", companyH.GetRiskScore)
	coByID.Post("/ocr-rut", middleware.SensitiveDataProcessingDisabled(), companyH.ProcessRUT)

	// Employees
	emp := app.Group("/v1/companies/:companyId/employees", jwt, middleware.CompanyScope("companyId"), middleware.CompanyMembershipScope("companyId", userRepo), middleware.Authorize(enforcer, "employees"), middleware.SensitiveDataProcessingGate(cfg.SyntheticPilotEnabled))
	emp.Get("/", employeeH.List)
	emp.Post("/", employeeH.Create)
	emp.Post("/import", middleware.SensitiveDataProcessingDisabled(), employeeH.BulkImport)
	emp.Get("/:id", employeeH.GetByID)
	emp.Put("/:id", employeeH.Update)
	emp.Delete("/:id", employeeH.Delete)
	emp.Get("/:id/profile", employeeH.GetOccupationalProfile)

	// Medical exam endpoints are intentionally not registered until consent checks,
	// company membership, medical-data minimization, and retention are reviewed.

	// Incidents
	inc := app.Group("/v1/companies/:companyId/incidents", jwt, middleware.CompanyScope("companyId"), middleware.CompanyMembershipScope("companyId", userRepo), middleware.Authorize(enforcer, "incidents"), middleware.SensitiveDataProcessingGate(cfg.SyntheticPilotEnabled))
	inc.Get("/", incidentH.List)
	inc.Post("/", incidentH.Create)
	inc.Get("/:id", incidentH.GetByID)
	inc.Put("/:id/investigate", incidentH.Investigate)
	inc.Put("/:id/close", incidentH.Close)
	inc.Post("/:id/furat", incidentH.GenerateFURAT)
	app.Get("/v1/incidents/open", jwt, middleware.CompanyScope("companyId"), middleware.CompanyMembershipScope("companyId", userRepo), middleware.Authorize(enforcer, "incidents"), middleware.SensitiveDataProcessingGate(cfg.SyntheticPilotEnabled), incidentH.GetOpenByTenant)

	// Trainings
	tr := app.Group("/v1/companies/:companyId/trainings", jwt, middleware.CompanyScope("companyId"), middleware.CompanyMembershipScope("companyId", userRepo), middleware.Authorize(enforcer, "trainings"), middleware.SensitiveDataProcessingGate(cfg.SyntheticPilotEnabled))
	tr.Get("/", trainingH.List)
	tr.Post("/", trainingH.Create)
	tr.Post("/:id/complete", trainingH.Complete)
	app.Get("/v1/trainings/expiring-certs", jwt, middleware.CompanyScope("companyId"), middleware.CompanyMembershipScope("companyId", userRepo), middleware.Authorize(enforcer, "trainings"), middleware.SensitiveDataProcessingGate(cfg.SyntheticPilotEnabled), trainingH.GetExpiring)

	// Zone checklists and recognition routes are intentionally not registered:
	// their legal references require review against current Colombian rules.

	// Evidence routes remain unregistered until every object lookup is scoped
	// through a verified company membership (the current evidence model is tenant-only).

	// ── Sistema de Gestión PHVA + cumplimiento ISO/Res.0312 ──
	// Company management routes remain unregistered pending an object-level
	// ownership test for every management-system ID.
	// ── Vigilancia normativa (humano en el lazo) ──
	regulatory := app.Group("/v1/regulatory", jwt)
	regulatory.Get("/pending", middleware.Authorize(enforcer, "regulatory_read"), mgmtH.PendingChanges)
	regulatory.Post("/changes/:id/review", middleware.Authorize(enforcer, "regulatory_review"), mgmtH.ReviewChange)

	// ── Metadata autenticada y de sólo lectura ──
	system := app.Group("/v1/system", jwt, middleware.Authorize(enforcer, "system"))
	system.Get("/roles", metaH.Roles)
	system.Get("/tenancy", metaH.Tenancy)
	system.Get("/capabilities", metaH.Capabilities)
	// Evidence and agent-event routes remain closed until object-level company
	// scope and data minimization are proven by integration tests.

	// ── Start ────────────────────────────────────────────────
	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("🚀 ZU Core v6.0 running on %s", addr)
	log.Printf("   Env: %s | DB: %s | Redis: %s", cfg.Env, cfg.DBHost, cfg.RedisAddr)

	go func() {
		if err := app.Listen(addr); err != nil {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = app.ShutdownWithContext(ctx)
	amqp.Close()
}
