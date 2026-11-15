// Package config loads and validates environment-based application settings.
package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// Environment names.
const (
	EnvDevelopment = "development"
	EnvTest        = "test"
	EnvProduction  = "production"
)

// Config holds validated runtime configuration.
type Config struct {
	AppEnv   string
	LogLevel string

	ServerHost string
	ServerPort int

	EthereumRPCURL string
	ChainID        int64

	EtherscanAPIKey     string
	EtherscanBaseURL    string
	EtherscanAPIVersion string

	RequestTimeout time.Duration
	ScanTimeout    time.Duration

	MaxHistoryPageSize int
	MaxHistoryPages    int
	MaxTransactions    int

	RateLimitRPS   float64
	RateLimitBurst int

	// Inbound HTTP rate limit (per client IP). Separate from outbound Etherscan limits.
	HTTPRateLimitRPS   float64
	HTTPRateLimitBurst int

	EtherscanMaxRetries   int
	EtherscanRetryBackoff time.Duration
	BlockHashCheckLimit   int

	CORSOrigins []string
}

// Address returns host:port for the HTTP server.
func (c Config) Address() string {
	return fmt.Sprintf("%s:%d", c.ServerHost, c.ServerPort)
}

// Load reads configuration from the process environment.
// It does not load secrets from source files other than the environment
// (callers may use godotenv before Load for local development).
func Load() (Config, error) {
	cfg := Config{
		AppEnv:                getString("QGUARD_APP_ENV", EnvDevelopment),
		LogLevel:              getString("QGUARD_LOG_LEVEL", "info"),
		ServerHost:            getString("QGUARD_SERVER_HOST", "127.0.0.1"),
		ServerPort:            getInt("QGUARD_SERVER_PORT", 8080),
		EthereumRPCURL:        strings.TrimSpace(os.Getenv("QGUARD_ETH_RPC_URL")),
		ChainID:               int64(getInt("QGUARD_CHAIN_ID", 1)),
		EtherscanAPIKey:       strings.TrimSpace(os.Getenv("QGUARD_ETHERSCAN_API_KEY")),
		EtherscanBaseURL:      getString("QGUARD_ETHERSCAN_API_URL", "https://api.etherscan.io/v2/api"),
		EtherscanAPIVersion:   getString("QGUARD_ETHERSCAN_API_VERSION", "v2"),
		RequestTimeout:        getDuration("QGUARD_REQUEST_TIMEOUT", 15*time.Second),
		ScanTimeout:           getDuration("QGUARD_SCAN_TIMEOUT", 120*time.Second),
		MaxHistoryPageSize:    getInt("QGUARD_MAX_HISTORY_PAGE_SIZE", 100),
		MaxHistoryPages:       getInt("QGUARD_ETHERSCAN_MAX_PAGES", 10),
		MaxTransactions:       getInt("QGUARD_SCAN_MAX_TXS", 500),
		RateLimitRPS:          getFloat("QGUARD_RATE_LIMIT_RPS", 4),
		RateLimitBurst:        getInt("QGUARD_RATE_LIMIT_BURST", 8),
		HTTPRateLimitRPS:      getFloat("QGUARD_HTTP_RATE_LIMIT_RPS", 20),
		HTTPRateLimitBurst:    getInt("QGUARD_HTTP_RATE_LIMIT_BURST", 40),
		EtherscanMaxRetries:   getInt("QGUARD_ETHERSCAN_MAX_RETRIES", 3),
		EtherscanRetryBackoff: getDuration("QGUARD_ETHERSCAN_RETRY_BACKOFF", 200*time.Millisecond),
		BlockHashCheckLimit:   getInt("QGUARD_BLOCK_HASH_CHECK_LIMIT", 5),
		CORSOrigins:           splitCSV(getString("QGUARD_CORS_ORIGINS", "http://localhost:3000")),
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// Validate ensures required settings are present and coherent.
func (c Config) Validate() error {
	var errs []string

	switch c.AppEnv {
	case EnvDevelopment, EnvTest, EnvProduction:
	default:
		errs = append(errs, fmt.Sprintf("QGUARD_APP_ENV must be one of %s, %s, %s (got %q)", EnvDevelopment, EnvTest, EnvProduction, c.AppEnv))
	}

	switch strings.ToLower(c.LogLevel) {
	case "debug", "info", "warn", "error":
	default:
		errs = append(errs, fmt.Sprintf("QGUARD_LOG_LEVEL must be debug, info, warn, or error (got %q)", c.LogLevel))
	}

	if c.ServerHost == "" {
		errs = append(errs, "QGUARD_SERVER_HOST must not be empty")
	}
	if c.ServerPort < 1 || c.ServerPort > 65535 {
		errs = append(errs, fmt.Sprintf("QGUARD_SERVER_PORT must be between 1 and 65535 (got %d)", c.ServerPort))
	}

	if c.EthereumRPCURL == "" {
		errs = append(errs, "QGUARD_ETH_RPC_URL is required")
	} else if _, err := url.ParseRequestURI(c.EthereumRPCURL); err != nil {
		errs = append(errs, fmt.Sprintf("QGUARD_ETH_RPC_URL is not a valid URL: %v", err))
	}

	if c.ChainID != 1 {
		errs = append(errs, fmt.Sprintf("MVP supports Ethereum Mainnet only; QGUARD_CHAIN_ID must be 1 (got %d)", c.ChainID))
	}

	if c.EtherscanAPIKey == "" {
		errs = append(errs, "QGUARD_ETHERSCAN_API_KEY is required")
	}
	if c.EtherscanBaseURL == "" {
		errs = append(errs, "QGUARD_ETHERSCAN_API_URL must not be empty")
	} else if _, err := url.ParseRequestURI(c.EtherscanBaseURL); err != nil {
		errs = append(errs, fmt.Sprintf("QGUARD_ETHERSCAN_API_URL is not a valid URL: %v", err))
	}
	if c.EtherscanAPIVersion == "" {
		errs = append(errs, "QGUARD_ETHERSCAN_API_VERSION must not be empty")
	}

	if c.RequestTimeout <= 0 {
		errs = append(errs, "QGUARD_REQUEST_TIMEOUT must be greater than zero")
	}
	if c.ScanTimeout <= 0 {
		errs = append(errs, "QGUARD_SCAN_TIMEOUT must be greater than zero")
	}
	if c.MaxHistoryPageSize < 1 || c.MaxHistoryPageSize > 10000 {
		errs = append(errs, fmt.Sprintf("QGUARD_MAX_HISTORY_PAGE_SIZE must be between 1 and 10000 (got %d)", c.MaxHistoryPageSize))
	}
	if c.MaxHistoryPages < 1 {
		errs = append(errs, "QGUARD_ETHERSCAN_MAX_PAGES must be at least 1")
	}
	if c.MaxTransactions < 1 {
		errs = append(errs, "QGUARD_SCAN_MAX_TXS must be at least 1")
	}
	if c.RateLimitRPS <= 0 {
		errs = append(errs, "QGUARD_RATE_LIMIT_RPS must be greater than zero")
	}
	if c.RateLimitBurst < 1 {
		errs = append(errs, "QGUARD_RATE_LIMIT_BURST must be at least 1")
	}
	if c.HTTPRateLimitRPS < 0 {
		errs = append(errs, "QGUARD_HTTP_RATE_LIMIT_RPS must be >= 0 (0 disables inbound limiting)")
	}
	if c.HTTPRateLimitRPS > 0 && c.HTTPRateLimitBurst < 1 {
		errs = append(errs, "QGUARD_HTTP_RATE_LIMIT_BURST must be at least 1 when inbound limiting is enabled")
	}
	if c.EtherscanMaxRetries < 0 {
		errs = append(errs, "QGUARD_ETHERSCAN_MAX_RETRIES must be >= 0")
	}
	if c.EtherscanRetryBackoff < 0 {
		errs = append(errs, "QGUARD_ETHERSCAN_RETRY_BACKOFF must be >= 0")
	}
	if c.BlockHashCheckLimit < 0 {
		errs = append(errs, "QGUARD_BLOCK_HASH_CHECK_LIMIT must be >= 0")
	}
	if !strings.EqualFold(c.EtherscanAPIVersion, "v2") {
		errs = append(errs, fmt.Sprintf("QGUARD_ETHERSCAN_API_VERSION must be v2 for the current Etherscan API (got %q)", c.EtherscanAPIVersion))
	}
	if len(c.CORSOrigins) == 0 {
		errs = append(errs, "QGUARD_CORS_ORIGINS must include at least one origin")
	}

	if len(errs) > 0 {
		return fmt.Errorf("invalid configuration:\n  - %s", strings.Join(errs, "\n  - "))
	}
	return nil
}

// ReadyChecks returns configuration readiness without probing external networks.
// Presence of required settings means "configured"; connectivity is explicitly unchecked.
func (c Config) ReadyChecks() map[string]CheckResult {
	return map[string]CheckResult{
		"config": {
			Status:  "ok",
			Message: "Configuration loaded and validated",
		},
		"ethereum_rpc": {
			Status:  "configured",
			Message: "RPC URL present; connectivity is not probed in the foundation milestone",
		},
		"etherscan": {
			Status:  "configured",
			Message: "API key and base URL present; connectivity is not probed in the foundation milestone",
		},
	}
}

// CheckResult describes a readiness check outcome.
type CheckResult struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func getString(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		v = strings.TrimSpace(v)
		if v != "" {
			return v
		}
	}
	return fallback
}

func getInt(key string, fallback int) int {
	raw, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(raw) == "" {
		return fallback
	}
	n, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		// Leave invalid value for Validate to catch via sentinel or re-parse.
		// Store as 0 so Validate fails with a clear message when appropriate.
		return 0
	}
	return n
}

func getFloat(key string, fallback float64) float64 {
	raw, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(raw) == "" {
		return fallback
	}
	n, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
	if err != nil {
		return 0
	}
	return n
}

func getDuration(key string, fallback time.Duration) time.Duration {
	raw, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(raw) == "" {
		return fallback
	}
	d, err := time.ParseDuration(strings.TrimSpace(raw))
	if err != nil {
		return 0
	}
	return d
}

func splitCSV(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
