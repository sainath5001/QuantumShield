// Package service contains application use-cases that do not belong in HTTP handlers.
package service

import (
	"github.com/quantumshield/qguard/backend/internal/config"
	"github.com/quantumshield/qguard/backend/internal/version"
)

// MetaService exposes health, readiness, and capability metadata.
type MetaService struct {
	cfg config.Config
}

// NewMetaService constructs a MetaService.
func NewMetaService(cfg config.Config) *MetaService {
	return &MetaService{cfg: cfg}
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

// ReadyResponse describes configuration readiness without false connectivity claims.
type ReadyResponse struct {
	Status  string                        `json:"status"`
	Ready   bool                          `json:"ready"`
	Checks  map[string]config.CheckResult `json:"checks"`
	Message string                        `json:"message"`
}

// Ready reports whether required configuration is present and validated.
// External RPC/indexer connectivity is explicitly marked as not probed.
func (s *MetaService) Ready() ReadyResponse {
	checks := s.cfg.ReadyChecks()
	return ReadyResponse{
		Status:  "ready",
		Ready:   true,
		Checks:  checks,
		Message: "Essential configuration is present. External dependency connectivity is not probed in this milestone.",
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
}

// Info returns static capability metadata for the foundation milestone.
func (s *MetaService) Info() InfoResponse {
	return InfoResponse{
		App:              version.App,
		Version:          version.Version,
		RiskModelVersion: version.RiskModelVersion,
		Chain:            chainInfoFor(s.cfg.ChainID),
		SupportedTxTypes: []TxTypeInfo{
			{Type: 0, Name: "legacy", Implemented: false, Notes: "Planned; signing-digest recovery not implemented in foundation milestone"},
			{Type: 1, Name: "eip2930_access_list", Implemented: false, Notes: "Planned; not implemented in foundation milestone"},
			{Type: 2, Name: "eip1559_dynamic_fee", Implemented: false, Notes: "Planned; not implemented in foundation milestone"},
			{Type: 3, Name: "eip4844_blob", Implemented: false, Notes: "Planned; not implemented in foundation milestone"},
		},
		Capabilities: Capabilities{
			AddressScanImplemented: false,
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
		},
		Limitations: []string{
			"Blockchain scanning is not implemented in this foundation milestone.",
			"Public-key recovery and risk classification are not yet available via the API.",
			"Readiness does not verify live Ethereum RPC or Etherscan connectivity.",
			"Absence of observed exposure will never be reported as quantum-safe.",
		},
	}
}

func chainInfoFor(chainID int64) ChainInfo {
	switch chainID {
	case 1:
		return ChainInfo{ChainID: 1, Name: "ethereum-mainnet", NativeAsset: "ETH"}
	case 11155111:
		return ChainInfo{ChainID: 11155111, Name: "ethereum-sepolia", NativeAsset: "ETH"}
	default:
		return ChainInfo{ChainID: chainID, Name: "unknown", NativeAsset: "ETH"}
	}
}
