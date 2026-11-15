// Package service contains application use-cases that do not belong in HTTP handlers.
package service

import (
	"context"
	"time"

	"github.com/quantumshield/qguard/backend/internal/config"
	"github.com/quantumshield/qguard/backend/internal/version"
)

// RPCProbe verifies Ethereum RPC readiness without implying indexer health.
type RPCProbe interface {
	VerifyConnection(ctx context.Context) error
	BlockNumber(ctx context.Context) (uint64, error)
}

// MetaService exposes health, readiness, and capability metadata.
type MetaService struct {
	cfg      config.Config
	rpcProbe RPCProbe
}

// NewMetaService constructs a MetaService.
func NewMetaService(cfg config.Config, rpcProbe RPCProbe) *MetaService {
	return &MetaService{cfg: cfg, rpcProbe: rpcProbe}
}

// HealthResponse is returned by the liveness endpoint.
type HealthResponse struct {
	Status           string `json:"status"`
	App              string `json:"app"`
	Version          string `json:"version"`
	RiskModelVersion string `json:"risk_model_version"`
	ChainID          int64  `json:"chain_id"`
	Environment      string `json:"environment"`
}

// Health returns process liveness information. It does not probe external systems.
func (s *MetaService) Health() HealthResponse {
	return HealthResponse{
		Status:           "ok",
		App:              version.App,
		Version:          version.Version,
		RiskModelVersion: version.RiskModelVersion,
		ChainID:          s.cfg.ChainID,
		Environment:      s.cfg.AppEnv,
	}
}

// ReadyResponse describes configuration and optional live dependency checks.
type ReadyResponse struct {
	Status  string                        `json:"status"`
	Ready   bool                          `json:"ready"`
	Checks  map[string]config.CheckResult `json:"checks"`
	Message string                        `json:"message"`
}

// Ready reports readiness. RPC may be probed when a client is configured.
// Etherscan connectivity is never claimed healthy from configuration alone.
func (s *MetaService) Ready(ctx context.Context) ReadyResponse {
	checks := map[string]config.CheckResult{
		"config": {
			Status:  "ok",
			Message: "Configuration loaded and validated",
		},
		"etherscan": {
			Status:  "configured",
			Message: "API key and base URL present; live Etherscan connectivity is not probed by /ready",
		},
	}

	ready := true
	if s.rpcProbe == nil {
		checks["ethereum_rpc"] = config.CheckResult{
			Status:  "configured",
			Message: "RPC client not attached; connectivity not verified",
		}
		ready = false
	} else {
		probeCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		if err := s.rpcProbe.VerifyConnection(probeCtx); err != nil {
			checks["ethereum_rpc"] = config.CheckResult{
				Status:  "error",
				Message: err.Error(),
			}
			ready = false
		} else if bn, err := s.rpcProbe.BlockNumber(probeCtx); err != nil {
			checks["ethereum_rpc"] = config.CheckResult{
				Status:  "error",
				Message: err.Error(),
			}
			ready = false
		} else {
			checks["ethereum_rpc"] = config.CheckResult{
				Status:  "ok",
				Message: "RPC reachable; chain ID matches configured mainnet; latest block observed",
			}
			_ = bn
		}
	}

	status := "ready"
	msg := "Essential dependencies passed readiness checks."
	if !ready {
		status = "not_ready"
		msg = "One or more essential dependency checks failed or were not verified."
	}
	return ReadyResponse{
		Status:  status,
		Ready:   ready,
		Checks:  checks,
		Message: msg,
	}
}

// InfoResponse describes scanner capabilities for clients.
type InfoResponse struct {
	App              string              `json:"app"`
	Version          string              `json:"version"`
	RiskModelVersion string              `json:"risk_model_version"`
	Chain            ChainInfo           `json:"chain"`
	SupportedTxTypes []TxTypeInfo        `json:"supported_transaction_types"`
	Capabilities     Capabilities        `json:"capabilities"`
	HistoryProvider  HistoryProviderInfo `json:"history_provider"`
	Limitations      []string            `json:"limitations"`
}

// ChainInfo describes the configured chain.
type ChainInfo struct {
	ChainID     int64  `json:"chain_id"`
	Name        string `json:"name"`
	NativeAsset string `json:"native_asset"`
}

// TxTypeInfo describes a transaction envelope the scanner intends to support.
type TxTypeInfo struct {
	Type        uint8  `json:"type"`
	Name        string `json:"name"`
	Implemented bool   `json:"implemented"`
	Notes       string `json:"notes"`
}

// Capabilities describes what the API currently offers.
type Capabilities struct {
	AddressScanImplemented bool `json:"address_scan_implemented"`
	AccountLookup          bool `json:"account_lookup"`
	HistoryRetrieval       bool `json:"history_retrieval"`
	PublicKeyRecovery      bool `json:"public_key_recovery"`
	RiskAssessment         bool `json:"risk_assessment"`
	WalletConnectRequired  bool `json:"wallet_connect_required"`
	PrivateKeysAccepted    bool `json:"private_keys_accepted"`
}

// HistoryProviderInfo describes the configured history source (without secrets).
type HistoryProviderInfo struct {
	Name       string `json:"name"`
	BaseURL    string `json:"base_url"`
	APIVersion string `json:"api_version"`
	Configured bool   `json:"configured"`
	ChainID    int64  `json:"chain_id"`
}

// Info returns capability metadata for the current milestone.
func (s *MetaService) Info() InfoResponse {
	return InfoResponse{
		App:              version.App,
		Version:          version.Version,
		RiskModelVersion: version.RiskModelVersion,
		Chain:            ChainInfo{ChainID: 1, Name: "ethereum-mainnet", NativeAsset: "ETH"},
		SupportedTxTypes: []TxTypeInfo{
			{Type: 0, Name: "legacy", Implemented: false, Notes: "History retrieval available; signature recovery not implemented yet"},
			{Type: 1, Name: "eip2930_access_list", Implemented: false, Notes: "History retrieval available; signature recovery not implemented yet"},
			{Type: 2, Name: "eip1559_dynamic_fee", Implemented: false, Notes: "History retrieval available; signature recovery not implemented yet"},
			{Type: 3, Name: "eip4844_blob", Implemented: false, Notes: "History retrieval available; signature recovery not implemented yet"},
		},
		Capabilities: Capabilities{
			AddressScanImplemented: false,
			AccountLookup:          true,
			HistoryRetrieval:       true,
			PublicKeyRecovery:      false,
			RiskAssessment:         false,
			WalletConnectRequired:  false,
			PrivateKeysAccepted:    false,
		},
		HistoryProvider: HistoryProviderInfo{
			Name:       "etherscan",
			BaseURL:    s.cfg.EtherscanBaseURL,
			APIVersion: s.cfg.EtherscanAPIVersion,
			Configured: s.cfg.EtherscanAPIKey != "",
			ChainID:    s.cfg.ChainID,
		},
		Limitations: []string{
			"MVP supports Ethereum Mainnet (chain id 1) only; mismatched RPC chain IDs are rejected.",
			"Etherscan history is indexed and capped; complete_within_scope means complete within provider pages/caps, not absolute chain completeness.",
			"Public-key recovery and risk classification are not yet available.",
			"Token-transfer and internal-transaction history are not fetched in this milestone.",
			"Absence of observed exposure will never be reported as quantum-safe.",
		},
	}
}
