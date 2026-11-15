package config

import (
	"os"
	"strings"
	"testing"
	"time"
)

func setEnv(t *testing.T, kv map[string]string) {
	t.Helper()
	for k, v := range kv {
		t.Setenv(k, v)
	}
}

func clearConfigEnv(t *testing.T) {
	t.Helper()
	keys := []string{
		"QGUARD_APP_ENV",
		"QGUARD_LOG_LEVEL",
		"QGUARD_SERVER_HOST",
		"QGUARD_SERVER_PORT",
		"QGUARD_ETH_RPC_URL",
		"QGUARD_CHAIN_ID",
		"QGUARD_ETHERSCAN_API_KEY",
		"QGUARD_ETHERSCAN_API_URL",
		"QGUARD_ETHERSCAN_API_VERSION",
		"QGUARD_REQUEST_TIMEOUT",
		"QGUARD_SCAN_TIMEOUT",
		"QGUARD_MAX_HISTORY_PAGE_SIZE",
		"QGUARD_ETHERSCAN_MAX_PAGES",
		"QGUARD_SCAN_MAX_TXS",
		"QGUARD_RATE_LIMIT_RPS",
		"QGUARD_RATE_LIMIT_BURST",
		"QGUARD_CORS_ORIGINS",
	}
	for _, k := range keys {
		_ = os.Unsetenv(k)
		t.Setenv(k, "")
		_ = os.Unsetenv(k)
	}
}

func validEnv(t *testing.T) {
	t.Helper()
	clearConfigEnv(t)
	setEnv(t, map[string]string{
		"QGUARD_ETH_RPC_URL":           "https://eth.example/rpc",
		"QGUARD_ETHERSCAN_API_KEY":     "test-key",
		"QGUARD_APP_ENV":               "test",
		"QGUARD_LOG_LEVEL":             "info",
		"QGUARD_SERVER_HOST":           "127.0.0.1",
		"QGUARD_SERVER_PORT":           "8080",
		"QGUARD_CHAIN_ID":              "1",
		"QGUARD_CORS_ORIGINS":          "http://localhost:3000",
		"QGUARD_REQUEST_TIMEOUT":       "15s",
		"QGUARD_SCAN_TIMEOUT":          "2m",
		"QGUARD_MAX_HISTORY_PAGE_SIZE": "100",
		"QGUARD_ETHERSCAN_MAX_PAGES":   "10",
		"QGUARD_SCAN_MAX_TXS":          "500",
		"QGUARD_RATE_LIMIT_RPS":        "4",
		"QGUARD_RATE_LIMIT_BURST":      "8",
	})
}

func TestLoadValidConfig(t *testing.T) {
	validEnv(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}
	if cfg.EthereumRPCURL != "https://eth.example/rpc" {
		t.Fatalf("EthereumRPCURL = %q", cfg.EthereumRPCURL)
	}
	if cfg.EtherscanAPIKey != "test-key" {
		t.Fatalf("EtherscanAPIKey = %q", cfg.EtherscanAPIKey)
	}
	if cfg.ServerPort != 8080 {
		t.Fatalf("ServerPort = %d", cfg.ServerPort)
	}
	if cfg.RequestTimeout != 15*time.Second {
		t.Fatalf("RequestTimeout = %s", cfg.RequestTimeout)
	}
	if cfg.Address() != "127.0.0.1:8080" {
		t.Fatalf("Address() = %q", cfg.Address())
	}
}

func TestLoadMissingRPCURL(t *testing.T) {
	validEnv(t)
	t.Setenv("QGUARD_ETH_RPC_URL", "")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for missing QGUARD_ETH_RPC_URL")
	}
	if !strings.Contains(err.Error(), "QGUARD_ETH_RPC_URL is required") {
		t.Fatalf("error = %v", err)
	}
}

func TestLoadMissingEtherscanKey(t *testing.T) {
	validEnv(t)
	t.Setenv("QGUARD_ETHERSCAN_API_KEY", "")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for missing QGUARD_ETHERSCAN_API_KEY")
	}
	if !strings.Contains(err.Error(), "QGUARD_ETHERSCAN_API_KEY is required") {
		t.Fatalf("error = %v", err)
	}
}

func TestLoadInvalidPort(t *testing.T) {
	validEnv(t)
	t.Setenv("QGUARD_SERVER_PORT", "70000")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for invalid port")
	}
	if !strings.Contains(err.Error(), "QGUARD_SERVER_PORT") {
		t.Fatalf("error = %v", err)
	}
}

func TestLoadInvalidURL(t *testing.T) {
	validEnv(t)
	t.Setenv("QGUARD_ETH_RPC_URL", "not-a-url")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for invalid RPC URL")
	}
	if !strings.Contains(err.Error(), "QGUARD_ETH_RPC_URL") {
		t.Fatalf("error = %v", err)
	}
}

func TestLoadInvalidAppEnv(t *testing.T) {
	validEnv(t)
	t.Setenv("QGUARD_APP_ENV", "staging")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for invalid app env")
	}
}

func TestLoadInvalidLogLevel(t *testing.T) {
	validEnv(t)
	t.Setenv("QGUARD_LOG_LEVEL", "verbose")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for invalid log level")
	}
}

func TestLoadInvalidDuration(t *testing.T) {
	validEnv(t)
	t.Setenv("QGUARD_REQUEST_TIMEOUT", "forever")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for invalid duration")
	}
}

func TestReadyChecksDoNotClaimConnectivity(t *testing.T) {
	validEnv(t)
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	checks := cfg.ReadyChecks()
	for _, key := range []string{"ethereum_rpc", "etherscan"} {
		c, ok := checks[key]
		if !ok {
			t.Fatalf("missing check %s", key)
		}
		if c.Status != "configured" {
			t.Fatalf("%s status = %q, want configured (must not claim healthy connectivity)", key, c.Status)
		}
		if !strings.Contains(strings.ToLower(c.Message), "not probed") {
			t.Fatalf("%s message should state connectivity not probed: %q", key, c.Message)
		}
	}
	if checks["config"].Status != "ok" {
		t.Fatalf("config status = %q", checks["config"].Status)
	}
}
