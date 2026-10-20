package repository

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/zu-seguridad/core/config"
	"github.com/zu-seguridad/core/internal/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewPostgresDB(cfg *config.Config) *gorm.DB {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=America/Bogota",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort, cfg.DBSSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		TranslateError: true,
		Logger: logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: true,
			ParameterizedQueries:      true,
			Colorful:                  false,
		}),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}

	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Println("✅ PostgreSQL connected")
	return db
}

func RunMigrations(db *gorm.DB) error {
	return RunAllMigrations(db)
}

func RunAllMigrations(db *gorm.DB) error {
	for _, statement := range []string{
		`CREATE EXTENSION IF NOT EXISTS "pgcrypto"`,
		`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`,
	} {
		if err := db.Exec(statement).Error; err != nil {
			return fmt.Errorf("database extension setup failed: %w", err)
		}
	}
	if err := db.AutoMigrate(
		&model.User{},
		&model.Company{},
		&model.CompanyMembership{},
		&model.Employee{},
		&model.Incident{},
		&model.Training{},
		&model.Certificate{},
		&model.Certification{},
		&model.Document{},
		&model.MedicalExam{},
		&model.RiskEntry{},
		&model.EPPItem{},
		&model.EPPDelivery{},
		&model.ConsentRecord{},
		&model.Project{},
		&model.Zone{},
		&model.ChecklistItem{},
		&model.EvidenceRecord{},
		&model.AgentEvent{},
		// Dominio PHVA + cumplimiento normativo
		&model.ManagementSystem{},
		&model.ComplianceStandard{},
		&model.SSTObjective{},
		&model.CorrectiveAction{},
		// Vigilancia normativa
		&model.RegulatorySource{},
		&model.RegulatoryChange{},
	); err != nil {
		return err
	}
	return db.Exec(`INSERT INTO company_memberships
		(tenant_id, user_id, company_id, role, is_active, is_synthetic, created_at, updated_at)
		SELECT u.tenant_id, u.id, u.company_id, 'manager', u.is_active, u.is_synthetic, NOW(), NOW()
		FROM users u
		JOIN companies c ON c.id = u.company_id AND c.tenant_id = u.tenant_id
		WHERE u.company_id IS NOT NULL AND u.deleted_at IS NULL AND c.deleted_at IS NULL
		ON CONFLICT (tenant_id, user_id, company_id) DO NOTHING`).Error
}
