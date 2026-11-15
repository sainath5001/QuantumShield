package server_test

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/quantumshield/qguard/backend/internal/apierror"
	"github.com/quantumshield/qguard/backend/internal/config"
	"github.com/quantumshield/qguard/backend/internal/server"
	"github.com/quantumshield/qguard/backend/internal/version"
)

func testConfig() config.Config {
	return config.Config{
		AppEnv:              config.EnvTest,
		LogLevel:            "error",
		ServerHost:          "127.0.0.1",
		ServerPort:          8080,
		EthereumRPCURL:      "https://eth.example/rpc",
		ChainID:             1,
		EtherscanAPIKey:     "test-key",
		EtherscanBaseURL:    "https://api.etherscan.io/api",
		EtherscanAPIVersion: "v1",
		RequestTimeout:      2 * time.Second,
		ScanTimeout:         2 * time.Minute,
		MaxHistoryPageSize:  100,
		MaxHistoryPages:     10,
		MaxTransactions:     500,
		RateLimitRPS:        4,
		RateLimitBurst:      8,
		CORSOrigins:         []string{"http://localhost:3000"},
	}
}

func TestHealthEndpoint(t *testing.T) {
	engine := server.NewEngine(server.Options{
		Config: testConfig(),
		Log:    slog.New(slog.NewTextHandler(io.Discard, nil)),
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if rec.Header().Get("X-Request-ID") == "" {
		t.Fatal("expected X-Request-ID header")
	}

	var body map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["status"] != "ok" {
		t.Fatalf("status field = %v", body["status"])
	}
	if body["version"] != version.Version {
		t.Fatalf("version = %v", body["version"])
	}
	if body["risk_model_version"] != version.RiskModelVersion {
		t.Fatalf("risk_model_version = %v", body["risk_model_version"])
	}
}

func TestReadyEndpoint(t *testing.T) {
	engine := server.NewEngine(server.Options{
		Config: testConfig(),
		Log:    slog.New(slog.NewTextHandler(io.Discard, nil)),
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ready", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}

	var body struct {
		Status string `json:"status"`
		Ready  bool   `json:"ready"`
		Checks map[string]struct {
			Status  string `json:"status"`
			Message string `json:"message"`
		} `json:"checks"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if !body.Ready || body.Status != "ready" {
		t.Fatalf("ready response = %+v", body)
	}
	if body.Checks["ethereum_rpc"].Status != "configured" {
		t.Fatalf("ethereum_rpc check = %+v", body.Checks["ethereum_rpc"])
	}
	if body.Checks["etherscan"].Status == "healthy" || body.Checks["etherscan"].Status == "ok" {
		t.Fatal("ready must not claim etherscan connectivity is healthy/ok")
	}
}

func TestInfoEndpoint(t *testing.T) {
	engine := server.NewEngine(server.Options{
		Config: testConfig(),
		Log:    slog.New(slog.NewTextHandler(io.Discard, nil)),
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/info", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}

	var body struct {
		Capabilities struct {
			AddressScanImplemented bool `json:"address_scan_implemented"`
			PrivateKeysAccepted    bool `json:"private_keys_accepted"`
		} `json:"capabilities"`
		SupportedTxTypes []struct {
			Implemented bool `json:"implemented"`
		} `json:"supported_transaction_types"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Capabilities.AddressScanImplemented {
		t.Fatal("foundation must not claim address scan is implemented")
	}
	if body.Capabilities.PrivateKeysAccepted {
		t.Fatal("private keys must never be accepted")
	}
	for _, tx := range body.SupportedTxTypes {
		if tx.Implemented {
			t.Fatal("foundation must not claim tx type analysis is implemented")
		}
	}
}

func TestNotFoundJSONError(t *testing.T) {
	engine := server.NewEngine(server.Options{
		Config: testConfig(),
		Log:    slog.New(slog.NewTextHandler(io.Discard, nil)),
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/does-not-exist", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d", rec.Code)
	}

	var body apierror.Response
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Error.Code != apierror.CodeNotFound {
		t.Fatalf("code = %s", body.Error.Code)
	}
}

func TestCORSAllowList(t *testing.T) {
	engine := server.NewEngine(server.Options{
		Config: testConfig(),
		Log:    slog.New(slog.NewTextHandler(io.Discard, nil)),
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:3000" {
		t.Fatalf("ACA-Origin = %q", got)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	req2.Header.Set("Origin", "http://evil.example")
	rec2 := httptest.NewRecorder()
	engine.ServeHTTP(rec2, req2)
	if got := rec2.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("unexpected ACA-Origin for disallowed origin: %q", got)
	}
}

func TestRequestIDPropagation(t *testing.T) {
	engine := server.NewEngine(server.Options{
		Config: testConfig(),
		Log:    slog.New(slog.NewTextHandler(io.Discard, nil)),
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	req.Header.Set("X-Request-ID", "client-req-1")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if got := rec.Header().Get("X-Request-ID"); got != "client-req-1" {
		t.Fatalf("X-Request-ID = %q", got)
	}
}
