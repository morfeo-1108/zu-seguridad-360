package repository

import (
	"log"
	"os"

	"github.com/zu-seguridad/drone/internal/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// MustConnect abre la conexión a Postgres (DSN desde env) y migra los modelos
// del dominio drone. En dev sin BD disponible, degrada con log y devuelve nil;
// los servicios que reciben nil deben operar en modo degradado.
func MustConnect() *gorm.DB {
	dsn := os.Getenv("DRONE_DB_DSN")
	if dsn == "" {
		dsn = "host=postgres user=zu password=zu dbname=zu_drone port=5432 sslmode=disable"
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Printf("⚠️ drone: BD no disponible (%v) — modo degradado sin persistencia", err)
		return nil
	}
	if err := db.AutoMigrate(&model.Drone{}, &model.DroneInspectionReport{}); err != nil {
		log.Printf("⚠️ drone: migración falló: %v", err)
	}
	return db
}
