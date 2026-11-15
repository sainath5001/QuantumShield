package domain

import (
	"time"

	"github.com/ethereum/go-ethereum/common"
)

// SecurityReport is the complete API/frontend-facing scan report.
type SecurityReport struct {
	ID               string             `json:"id"`
	ScanID           string             `json:"scan_id"`
	GeneratedAt      time.Time          `json:"generated_at"`
	RiskModelVersion string             `json:"risk_model_version"`
	ExecutiveSummary string             `json:"executive_summary"`
	Wallet           ReportWalletInfo   `json:"wallet"`
	Cryptography     ReportCryptoInfo   `json:"cryptography"`
	ExposureFindings ReportExposureInfo `json:"exposure_findings"`
	Evidence         []ExposureEvidence `json:"evidence"`
	QuantumThreat    ReportQuantumInfo  `json:"quantum_threat"`
	Risk             RiskAssessment     `json:"risk"`
	Coverage         ReportCoverageInfo `json:"coverage"`
	RecommendedSteps []string           `json:"recommended_next_steps"`
	Limitations      []string           `json:"limitations"`
	Disclaimers      []string           `json:"disclaimers"`
}

// ReportWalletInfo summarizes the scanned account.
type ReportWalletInfo struct {
	Address     common.Address `json:"address"`
	AccountKind AccountKind    `json:"account_kind"`
	HasCode     bool           `json:"has_code"`
	BalanceWei  string         `json:"balance_wei,omitempty"`
	ChainID     int64          `json:"chain_id"`
	BlockNumber uint64         `json:"block_number,omitempty"`
}

// ReportCryptoInfo describes the observed cryptographic scheme.
type ReportCryptoInfo struct {
	Curve           string `json:"curve,omitempty"`
	SignatureScheme string `json:"signature_scheme,omitempty"`
	Notes           string `json:"notes"`
}

// ReportExposureInfo summarizes exposure status for the report.
type ReportExposureInfo struct {
	Confirmed             bool                  `json:"confirmed"`
	Finding               WalletExposureFinding `json:"finding"`
	VerifiedExposureCount int                   `json:"verified_exposure_count"`
	DistinctPublicKeys    int                   `json:"distinct_public_keys"`
}

// ReportQuantumInfo is educational threat context tied to the classification.
type ReportQuantumInfo struct {
	ShorRelevance   string   `json:"shor_relevance"`
	GroverRelevance string   `json:"grover_relevance"`
	PracticalNote   string   `json:"practical_note"`
	Bullets         []string `json:"bullets"`
}

// ReportCoverageInfo mirrors history/analysis coverage for the UI.
type ReportCoverageInfo struct {
	HistoryCoverage    HistoryCoverageStatus `json:"history_coverage"`
	EvidenceQuality    EvidenceQuality       `json:"evidence_quality"`
	AnalyzedTxCount    int                   `json:"analyzed_tx_count"`
	UnsupportedTxCount int                   `json:"unsupported_tx_count"`
	FailedTxCount      int                   `json:"failed_tx_count"`
	SenderTxCount      int                   `json:"sender_tx_count"`
	PaginationComplete bool                  `json:"pagination_complete,omitempty"`
}
