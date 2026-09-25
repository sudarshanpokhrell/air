package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sudarshanpokhrell/air/internal/config"
	"github.com/sudarshanpokhrell/air/internal/db"
)

func main() {
	cfg := config.MustLoad()

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	dbConn, err := db.Connect(cfg.DBUrl)
	if err != nil {
		log.Fatalf("main.go db:connect: %v", err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		slog.Info("server listening", "addr", srv.Addr)

		if err := srv.ListenAndServe(); err != nil &&
			err != http.ErrServerClosed {
			slog.Error("server failed", "err", err)
			stop()
		}
	}()

	<-ctx.Done()

	slog.Info("shutting down server")

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("server shutdown failed", "err", err)
	}

	if err := dbConn.Close(); err != nil {
		slog.Error("database close failed", "err", err)
	}

	slog.Info("server stopped")
}
