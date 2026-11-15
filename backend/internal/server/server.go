// Package server wires HTTP routing and middleware for the Q-Guard API.
package server

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/quantumshield/qguard/backend/internal/blockchain"
	"github.com/quantumshield/qguard/backend/internal/config"
	"github.com/quantumshield/qguard/backend/internal/handler"
	"github.com/quantumshield/qguard/backend/internal/middleware"
	"github.com/quantumshield/qguard/backend/internal/service"
)

// Options controls HTTP server construction.
type Options struct {
	Config  config.Config
	Log     *slog.Logger
	RPC     blockchain.RPCClient
	Account *service.AccountService
}

// NewEngine builds the Gin engine with middleware and routes.
func NewEngine(opts Options) *gin.Engine {
	if opts.Log == nil {
		opts.Log = slog.Default()
	}

	switch opts.Config.AppEnv {
	case config.EnvProduction:
		gin.SetMode(gin.ReleaseMode)
	case config.EnvTest:
		gin.SetMode(gin.TestMode)
	default:
		gin.SetMode(gin.DebugMode)
	}

	r := gin.New()
	r.Use(middleware.RequestID())
	r.Use(middleware.Recovery(opts.Log))
	r.Use(middleware.Logger(opts.Log))
	r.Use(middleware.CORS(opts.Config.CORSOrigins))
	r.Use(middleware.Timeout(opts.Config.RequestTimeout))

	var rpcProbe service.RPCProbe
	if opts.RPC != nil {
		rpcProbe = opts.RPC
	}
	meta := service.NewMetaService(opts.Config, rpcProbe)
	api := handler.NewAPI(meta, opts.Account)

	v1 := r.Group("/api/v1")
	{
		v1.GET("/health", api.Health)
		v1.GET("/ready", api.Ready)
		v1.GET("/info", api.Info)
		v1.GET("/accounts/:address", api.GetAccount)
		v1.GET("/accounts/:address/transactions", api.GetAccountTransactions)
		v1.GET("/accounts/:address/exposure", api.GetAccountExposure)
	}

	r.NoRoute(handler.NotFound)
	r.NoMethod(handler.MethodNotAllowed)

	return r
}

// NewHTTPServer creates an http.Server with sensible timeouts.
func NewHTTPServer(cfg config.Config, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              cfg.Address(),
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       cfg.RequestTimeout + 5*time.Second,
		WriteTimeout:      cfg.RequestTimeout + 5*time.Second,
		IdleTimeout:       60 * time.Second,
	}
}
