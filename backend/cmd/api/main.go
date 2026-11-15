// Command api is the Q-Guard HTTP API server entrypoint.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/quantumshield/qguard/backend/internal/config"
	"github.com/quantumshield/qguard/backend/internal/server"
	"github.com/quantumshield/qguard/backend/internal/version"
)

func main() {
	// Load .env from cwd or repo root when present (local development). Missing file is fine.
	_ = godotenv.Load()
	_ = godotenv.Load("../.env")
	_ = godotenv.Load("../../.env")

	cfg, err := config.Load()
	if err != nil {
		slog.Error("configuration error", "err", err)
		os.Exit(1)
	}

	log := newLogger(cfg.LogLevel)
	log.Info("starting Q-Guard API",
		"app", version.App,
		"version", version.Version,
		"env", cfg.AppEnv,
		"addr", cfg.Address(),
		"chain_id", cfg.ChainID,
	)

	engine := server.NewEngine(server.Options{Config: cfg, Log: log})
	httpServer := server.NewHTTPServer(cfg, engine)

	go func() {
		log.Info("HTTP server listening", "addr", cfg.Address())
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("HTTP server failed", "err", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	log.Info("shutting down")
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Error("shutdown error", "err", err)
		os.Exit(1)
	}
}

func newLogger(level string) *slog.Logger {
	var lv slog.Level
	switch strings.ToLower(level) {
	case "debug":
		lv = slog.LevelDebug
	case "warn":
		lv = slog.LevelWarn
	case "error":
		lv = slog.LevelError
	default:
		lv = slog.LevelInfo
	}
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lv})
	return slog.New(handler)
}
