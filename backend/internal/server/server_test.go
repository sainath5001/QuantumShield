package server_test

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/quantumshield/qguard/backend/internal/apierror"
	"github.com/quantumshield/qguard/backend/internal/blockchain"
	"github.com/quantumshield/qguard/backend/internal/config"
	"github.com/quantumshield/qguard/backend/internal/domain"
	"github.com/quantumshield/qguard/backend/internal/history"
	"github.com/quantumshield/qguard/backend/internal/server"
	"github.com/quantumshield/qguard/backend/internal/service"
	"github.com/quantumshield/qguard/backend/internal/version"
)

func testConfig() config.Config {
	return config.Config{
		AppEnv:                config.EnvTest,
		LogLevel:              "error",
		ServerHost:            "127.0.0.1",
		ServerPort:            8080,
		EthereumRPCURL:        "https://eth.example/rpc",
		ChainID:               1,
		EtherscanAPIKey:       "test-key",
		EtherscanBaseURL:      "https://api.etherscan.io/v2/api",
		EtherscanAPIVersion:   "v2",
		RequestTimeout:        2 * time.Second,
		ScanTimeout:           2 * time.Minute,
		MaxHistoryPageSize:    100,
		MaxHistoryPages:       10,
		MaxTransactions:       500,
		RateLimitRPS:          4,
		RateLimitBurst:        8,
		HTTPRateLimitRPS:      0, // disable inbound limiter in unit tests
		HTTPRateLimitBurst:    0,
		EtherscanMaxRetries:   1,
		EtherscanRetryBackoff: time.Millisecond,
		BlockHashCheckLimit:   0,
		CORSOrigins:           []string{"http://localhost:3000"},
	}
}

type fakeRPC struct {
	ok bool
}

func (f fakeRPC) Close() {}
func (f fakeRPC) VerifyConnection(ctx context.Context) error {
	if !f.ok {
		return blockchain.ErrChainMismatch{Configured: 1, Observed: 5}
	}
	return nil
}
func (f fakeRPC) ChainID(ctx context.Context) (*big.Int, error) { return big.NewInt(1), nil }
func (f fakeRPC) BlockNumber(ctx context.Context) (uint64, error) {
	return 123, nil
}
func (f fakeRPC) BalanceAt(ctx context.Context, addr common.Address) (*big.Int, error) {
	return big.NewInt(1), nil
}
func (f fakeRPC) CodeAt(ctx context.Context, addr common.Address) ([]byte, error) { return nil, nil }
func (f fakeRPC) BlockHashByNumber(ctx context.Context, number uint64) (common.Hash, error) {
	return common.Hash{}, nil
}
func (f fakeRPC) TransactionByHash(ctx context.Context, hash common.Hash) (*types.Transaction, bool, error) {
	return nil, false, nil
}
func (f fakeRPC) ClassifyAccount(ctx context.Context, addr common.Address) (domain.AccountSnapshot, error) {
	return domain.AccountSnapshot{
		Address:     addr,
		Kind:        domain.AccountLikelyEOA,
		BalanceWei:  big.NewInt(1),
		ChainID:     1,
		BlockNumber: 123,
		Limitations: domain.AccountClassificationLimitations,
	}, nil
}

type fakeHistory struct{}

func (fakeHistory) FetchNormalTransactions(ctx context.Context, address string) (domain.TransactionHistory, error) {
	return domain.TransactionHistory{
		Address:      common.HexToAddress(address).Hex(),
		Transactions: nil,
		Coverage: domain.HistoryCoverage{
			Status:             domain.CoverageCompleteWithinScope,
			Provider:           "etherscan",
			PaginationComplete: true,
		},
	}, nil
}

func TestHealthEndpoint(t *testing.T) {
	engine := server.NewEngine(server.Options{
		Config: testConfig(),
		Log:    slog.New(slog.NewTextHandler(io.Discard, nil)),
		RPC:    fakeRPC{ok: true},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var body map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["version"] != version.Version {
		t.Fatalf("version = %v", body["version"])
	}
}

func TestReadyEndpointOK(t *testing.T) {
	engine := server.NewEngine(server.Options{
		Config: testConfig(),
		Log:    slog.New(slog.NewTextHandler(io.Discard, nil)),
		RPC:    fakeRPC{ok: true},
	})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/ready", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	var body struct {
		Ready  bool `json:"ready"`
		Checks map[string]struct {
			Status string `json:"status"`
		} `json:"checks"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if !body.Ready || body.Checks["ethereum_rpc"].Status != "ok" {
		t.Fatalf("body = %+v", body)
	}
	if body.Checks["etherscan"].Status == "ok" || body.Checks["etherscan"].Status == "healthy" {
		t.Fatal("must not claim etherscan connectivity healthy from /ready")
	}
}

func TestReadyEndpointNotReadyWithoutRPC(t *testing.T) {
	engine := server.NewEngine(server.Options{
		Config: testConfig(),
		Log:    slog.New(slog.NewTextHandler(io.Discard, nil)),
	})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/ready", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestInfoEndpointCapabilities(t *testing.T) {
	engine := server.NewEngine(server.Options{
		Config: testConfig(),
		Log:    slog.New(slog.NewTextHandler(io.Discard, nil)),
		RPC:    fakeRPC{ok: true},
	})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/info", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	var body struct {
		Capabilities struct {
			HistoryRetrieval       bool `json:"history_retrieval"`
			AddressScanImplemented bool `json:"address_scan_implemented"`
			PublicKeyRecovery      bool `json:"public_key_recovery"`
			RiskAssessment         bool `json:"risk_assessment"`
			PrivateKeysAccepted    bool `json:"private_keys_accepted"`
		} `json:"capabilities"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if !body.Capabilities.HistoryRetrieval || !body.Capabilities.PublicKeyRecovery {
		t.Fatal("expected history_retrieval and public_key_recovery true")
	}
	if !body.Capabilities.AddressScanImplemented || !body.Capabilities.RiskAssessment {
		t.Fatal("expected address scan and risk assessment true")
	}
	if body.Capabilities.PrivateKeysAccepted {
		t.Fatal("private keys must never be accepted")
	}
}

func TestAccountAndHistoryEndpoints(t *testing.T) {
	acct := &service.AccountService{
		RPC:     fakeRPC{ok: true},
		History: &history.Fetcher{Source: fakeHistory{}, MaxHashChecks: 0},
	}
	engine := server.NewEngine(server.Options{
		Config:  testConfig(),
		Log:     slog.New(slog.NewTextHandler(io.Discard, nil)),
		RPC:     fakeRPC{ok: true},
		Account: acct,
	})

	addr := "0x2449ecef5012f0a0e153b278ef4fcc9625bc4c78"
	req := httptest.NewRequest(http.MethodGet, "/api/v1/accounts/"+addr, nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("account status = %d body=%s", rec.Code, rec.Body.String())
	}

	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/accounts/"+addr+"/transactions", nil)
	rec2 := httptest.NewRecorder()
	engine.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Fatalf("history status = %d body=%s", rec2.Code, rec2.Body.String())
	}
	var hist struct {
		Coverage struct {
			Status string `json:"status"`
		} `json:"coverage"`
	}
	_ = json.Unmarshal(rec2.Body.Bytes(), &hist)
	if hist.Coverage.Status != string(domain.CoverageCompleteWithinScope) {
		t.Fatalf("coverage = %s", hist.Coverage.Status)
	}
}

func TestInvalidAccountAddress(t *testing.T) {
	acct := &service.AccountService{RPC: fakeRPC{ok: true}}
	engine := server.NewEngine(server.Options{
		Config:  testConfig(),
		Log:     slog.New(slog.NewTextHandler(io.Discard, nil)),
		Account: acct,
	})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/accounts/0x1234", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
	var body apierror.Response
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body.Error.Code != apierror.CodeInvalidRequest {
		t.Fatalf("code = %s", body.Error.Code)
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
}
