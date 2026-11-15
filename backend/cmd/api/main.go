// Command api is the Q-Guard HTTP API server entrypoint.
package main

import (
	"context"
	"errors"
	"log/slog"
	"math/big"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/quantumshield/qguard/backend/internal/blockchain"
	"github.com/quantumshield/qguard/backend/internal/config"
	"github.com/quantumshield/qguard/backend/internal/etherscan"
	"github.com/quantumshield/qguard/backend/internal/exposure"
	"github.com/quantumshield/qguard/backend/internal/history"
	"github.com/quantumshield/qguard/backend/internal/server"
	"github.com/quantumshield/qguard/backend/internal/service"
	"github.com/quantumshield/qguard/backend/internal/version"
)

func main() {
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

	ctx, cancelInit := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancelInit()

	var rpc blockchain.RPCClient
	eth, err := blockchain.Dial(ctx, cfg.EthereumRPCURL, cfg.ChainID)
	if err != nil {
		log.Warn("ethereum RPC not ready at startup; /ready will report not_ready until fixed", "err", err)
	} else {
		rpc = eth
		defer eth.Close()
		log.Info("ethereum RPC connected", "chain_id", cfg.ChainID)
	}

	esClient, err := etherscan.NewClient(etherscan.Config{
		BaseURL:     cfg.EtherscanBaseURL,
		APIKey:      cfg.EtherscanAPIKey,
		ChainID:     cfg.ChainID,
		PageSize:    cfg.MaxHistoryPageSize,
		MaxPages:    cfg.MaxHistoryPages,
		MaxTxs:      cfg.MaxTransactions,
		HTTPClient:  &http.Client{Timeout: cfg.RequestTimeout},
		RateLimit:   etherscan.NewTokenBucket(cfg.RateLimitRPS, cfg.RateLimitBurst),
		MaxRetries:  cfg.EtherscanMaxRetries,
		BaseBackoff: cfg.EtherscanRetryBackoff,
	})
	if err != nil {
		log.Error("etherscan client configuration error", "err", err)
		os.Exit(1)
	}

	fetcher := &history.Fetcher{
		Source:        esClient,
		Verifier:      rpc,
		MaxHashChecks: cfg.BlockHashCheckLimit,
	}
	var exp *exposure.Analyzer
	if rpc != nil {
		exp = &exposure.Analyzer{Txs: rpc, ChainID: big.NewInt(cfg.ChainID)}
	}
	accountSvc := &service.AccountService{
		RPC:      rpc,
		History:  fetcher,
		Exposure: exp,
		ChainID:  big.NewInt(cfg.ChainID),
		MaxTxs:   cfg.MaxTransactions,
	}

	engine := server.NewEngine(server.Options{
		Config:  cfg,
		Log:     log,
		RPC:     rpc,
		Account: accountSvc,
	})
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

	ctxShutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	log.Info("shutting down")
	if err := httpServer.Shutdown(ctxShutdown); err != nil {
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
