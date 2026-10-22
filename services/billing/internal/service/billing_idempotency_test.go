//go:build integration
// +build integration

package service

// Test de integración de la idempotencia de facturación (hallazgo de auditoría #2).
//
// NOTA: requiere una BD (SQLite en memoria o Postgres). Se ejecuta en CI donde
// las dependencias están disponibles. No corre en sandboxes sin acceso a módulos Go.
//
// Para correrlo en CI:
//   go test -tags=integration ./internal/service/ -run TestBillingIdempotency
//

import (
	"context"
	"testing"

	"github.com/zu-seguridad/billing/internal/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("no se pudo abrir BD de prueba: %v", err)
	}
	if err := db.AutoMigrate(&model.IdempotencyKey{}, &model.OutboxEvent{}); err != nil {
		t.Fatalf("migración falló: %v", err)
	}
	return db
}

func TestBillingIdempotency_PreventsDoubleCharge(t *testing.T) {
	db := setupTestDB(t)
	svc := NewBillingService(db)
	ctx := context.Background()
	key := "monthly-2026-06"

	// Primera ejecución: no procesada aún
	if svc.WasProcessed(ctx, key) {
		t.Error("clave nueva no debería estar procesada")
	}
	// Marcar como procesada
	if err := svc.MarkProcessed(ctx, key); err != nil {
		t.Fatalf("MarkProcessed falló: %v", err)
	}
	// Segunda ejecución (reintento de Wompi): YA procesada → no recobra
	if !svc.WasProcessed(ctx, key) {
		t.Error("tras MarkProcessed, la clave debe estar marcada (evita doble cobro)")
	}
}

func TestBillingIdempotency_DistinctKeysIndependent(t *testing.T) {
	db := setupTestDB(t)
	svc := NewBillingService(db)
	ctx := context.Background()

	svc.MarkProcessed(ctx, "monthly-2026-06")
	// Un mes distinto NO debe estar marcado
	if svc.WasProcessed(ctx, "monthly-2026-07") {
		t.Error("claves de meses distintos deben ser independientes")
	}
}
