package main

import (
	"errors"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/sudarshanpokhrell/air/internal/config"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run main.go <up | down>")
	}

	cmd := os.Args[1]

	cfg := config.MustLoad()

	m, err := migrate.New("file://migrations", cfg.DBUrl)
	if err != nil {
		log.Fatalf("💥 Failed to initialize migration: %v", err)
	}

	defer m.Close()

	switch cmd {
	case "up":
		log.Println("🚀 Running database migration up...")

		if err := m.Up(); err != nil {
			if errors.Is(err, migrate.ErrNoChange) {
				log.Println("No new migrations to apply.")
				return
			}
			log.Fatalf(" Migration up failed: %v", err)
		}
		log.Println("✅ Migration up completed successfully!")

	case "down":
		log.Println("⚠️ Rolling back database migration down...")
		if err := m.Steps(-1); err != nil {
			if errors.Is(err, migrate.ErrNoChange) {
				log.Println("Nothing to rollback.")
				return
			}
			log.Fatalf("Migration down failed: %v", err)
		}
		log.Println("✅ Migration down completed successfully!")

	default:
		log.Fatalf("Unknown command: '%s'. Expected 'up' or 'down'.", cmd)
	}

}
