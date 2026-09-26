package main

import (
	"log/slog"
	"os"

	"github.com/sudarshanpokhrell/air/internal/config"
	"github.com/sudarshanpokhrell/air/internal/db"
	"github.com/sudarshanpokhrell/air/internal/storage"
	"github.com/sudarshanpokhrell/air/internal/store"
)

const version = "0.1.0"

type application struct {
	config  config.Config
	logger  *slog.Logger
	store   store.Store
	storage storage.Storage
}

func main() {
	cfg := config.MustLoad()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	database, err := db.Connect(cfg.DBUrl)
	if err != nil {
		logger.Error("failed to connect to database", "err", err)
		os.Exit(1)
	}
	defer database.Close()
	logger.Info("database connection pool established")

	app := &application{
		config: cfg,
		logger: logger,
		store:  store.NewStore(database),
		// storage: TODO storage.NewDisk(...) / storage.NewR2(...)
	}

	if err := app.serve(); err != nil {
		logger.Error("server error", "err", err)
		os.Exit(1)
	}
}
