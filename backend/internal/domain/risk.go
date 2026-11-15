package domain

import "time"

// RiskClassification is the qguard-risk-v1 label set.
type RiskClassification string

const (
	RiskLowerExposure    RiskClassification = "LOWER_EXPOSURE"
	RiskElevatedExposure RiskClassification = "ELEVATED_EXPOSURE"
	RiskUnknown          RiskClassification = "UNKNOWN"
	RiskRequiresReview   RiskClassification = "REQUIRES_REVIEW"
)

// EvidenceQuality describes completeness/quality of observed evidence — not attack probability.
type EvidenceQuality string

const (
	EvidenceHigh   EvidenceQuality = "HIGH"
	EvidenceMedium EvidenceQuality = "MEDIUM"
	EvidenceLow    EvidenceQuality = "LOW"
	EvidenceManual EvidenceQuality = "MANUAL"
)

// RiskAssessment is the deterministic output of the risk engine.
type RiskAssessment struct {
	Classification        RiskClassification    `json:"classification"`
	Reason                string                `json:"reason"`
	EvidenceQuality       EvidenceQuality       `json:"evidence_quality"`
	PublicKeyExposed      bool                  `json:"public_key_exposed"`
	SignatureScheme       string                `json:"signature_scheme,omitempty"`
	Curve                 string                `json:"curve,omitempty"`
	AccountKind           AccountKind           `json:"account_kind"`
	HistoryCoverage       HistoryCoverageStatus `json:"history_coverage"`
	AnalyzedTxCount       int                   `json:"analyzed_tx_count"`
	UnsupportedTxCount    int                   `json:"unsupported_tx_count"`
	FailedTxCount         int                   `json:"failed_tx_count"`
	VerifiedExposureCount int                   `json:"verified_exposure_count"`
	SenderMismatchCount   int                   `json:"sender_mismatch_count"`
	SenderTxCount         int                   `json:"sender_tx_count"`
	DistinctPublicKeys    int                   `json:"distinct_public_keys"`
	Limitations           []string              `json:"limitations"`
	Disclaimers           []string              `json:"disclaimers"`
	QuantumThreatNotes    []string              `json:"quantum_threat_notes"`
	RecommendedNextSteps  []string              `json:"recommended_next_steps"`
	RiskModelVersion      string                `json:"risk_model_version"`
	AssessedAt            time.Time             `json:"assessed_at"`
}

// StandardDisclaimers must appear on every assessment/report.
var StandardDisclaimers = []string{
	"Q-Guard never requests private keys, seed phrases, or wallet passwords.",
	"Public-key exposure does not mean the private key is known today or that the wallet has been compromised classically.",
	"No observed exposure within scope does not mean the wallet is quantum-safe.",
	"Analysis is limited to the configured chain, history provider responses, and supported transaction types; other chains and off-chain signatures are not covered.",
	"Q-Guard does not claim that a quantum computer has hacked any Ethereum wallet.",
	"Q-Guard does not estimate when a cryptographically relevant quantum computer will exist.",
}
