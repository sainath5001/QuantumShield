package domain

import (
	"time"

	"github.com/ethereum/go-ethereum/common"
)

// RecoveryStatus classifies a single-transaction recovery attempt.
type RecoveryStatus string

const (
	RecoveryConfirmed       RecoveryStatus = "confirmed"
	RecoveryFailed          RecoveryStatus = "recovery_failed"
	RecoverySenderMismatch  RecoveryStatus = "sender_mismatch"
	RecoveryUnsupportedType RecoveryStatus = "unsupported_type"
	RecoveryMissingSig      RecoveryStatus = "missing_signature"
	RecoveryInvalidSig      RecoveryStatus = "invalid_signature"
	RecoveryInconclusive    RecoveryStatus = "inconclusive"
	RecoveryFetchFailed     RecoveryStatus = "fetch_failed"
)

// SenderVerificationStatus reports whether the recovered key matches the expected sender.
type SenderVerificationStatus string

const (
	SenderVerified   SenderVerificationStatus = "verified"
	SenderMismatch   SenderVerificationStatus = "mismatch"
	SenderNotChecked SenderVerificationStatus = "not_checked"
)

// PubkeyEncoding identifies how a recovered public key is serialized in evidence.
type PubkeyEncoding string

const (
	PubkeyUncompressedSEC1 PubkeyEncoding = "sec1_uncompressed_0x04" // 65 bytes: 0x04 || X || Y
	PubkeyXY64             PubkeyEncoding = "xy_64_no_prefix"        // 64 bytes used for address hash
)

// ExposureEvidence is a structured record of a verified or attempted public-key recovery.
// Confirmed exposures require successful recovery AND sender verification.
type ExposureEvidence struct {
	WalletAddress         common.Address           `json:"wallet_address"`
	RecoveredPublicKeyHex string                   `json:"recovered_public_key_hex,omitempty"`
	PublicKeyFingerprint  string                   `json:"public_key_fingerprint,omitempty"` // keccak256(pubkey)[:8] hex
	PublicKeyEncoding     PubkeyEncoding           `json:"public_key_encoding,omitempty"`
	Curve                 string                   `json:"curve"`
	SignatureScheme       string                   `json:"signature_scheme"`
	TransactionHash       common.Hash              `json:"transaction_hash"`
	BlockNumber           *uint64                  `json:"block_number,omitempty"`
	TransactionType       *uint8                   `json:"transaction_type,omitempty"`
	SigningDigestHex      string                   `json:"signing_digest_hex,omitempty"`
	RecoveryStatus        RecoveryStatus           `json:"recovery_status"`
	SenderVerification    SenderVerificationStatus `json:"sender_verification_status"`
	ExpectedSender        common.Address           `json:"expected_sender"`
	DerivedSender         *common.Address          `json:"derived_sender,omitempty"`
	EvidenceSource        string                   `json:"evidence_source"`
	Detail                string                   `json:"detail,omitempty"`
	AnalyzedAt            time.Time                `json:"analyzed_at"`
}

// WalletExposureFinding summarizes exposure analysis without assigning a risk score.
type WalletExposureFinding string

const (
	FindingPublicKeyExposed     WalletExposureFinding = "public_key_exposed"
	FindingNoExposureObserved   WalletExposureFinding = "no_exposure_observed_in_scope"
	FindingInsufficientEvidence WalletExposureFinding = "insufficient_evidence"
	FindingRequiresManualReview WalletExposureFinding = "requires_manual_review"
)

// WalletExposureReport is the wallet-level analysis result (not a risk score).
type WalletExposureReport struct {
	Address               common.Address        `json:"address"`
	AccountKind           AccountKind           `json:"account_kind"`
	Finding               WalletExposureFinding `json:"finding"`
	ConfirmedExposure     bool                  `json:"confirmed_exposure"`
	DistinctPublicKeys    int                   `json:"distinct_public_keys"`
	ConfirmedCount        int                   `json:"confirmed_count"`
	FailedCount           int                   `json:"failed_count"`
	UnsupportedCount      int                   `json:"unsupported_count"`
	SenderMismatchCount   int                   `json:"sender_mismatch_count"`
	AnalyzedTxCount       int                   `json:"analyzed_tx_count"`
	SenderTxCount         int                   `json:"sender_tx_count"`
	HistoryCoverageStatus HistoryCoverageStatus `json:"history_coverage_status"`
	Evidence              []ExposureEvidence    `json:"evidence"`
	Limitations           []string              `json:"limitations"`
	AnalyzedAt            time.Time             `json:"analyzed_at"`
}

// ExposureAnalysisLimitations are always attached to wallet exposure reports.
var ExposureAnalysisLimitations = []string{
	"Confirmed exposure means a secp256k1 public key was recovered from a valid signature and verified against the sender address.",
	"No confirmed exposure within scope does not prove the public key was never exposed elsewhere (other chains, off-chain signatures, incomplete history).",
	"Public-key exposure is not present-day private-key compromise and is not evidence of a quantum attack.",
	"Signature recovery uses go-ethereum canonical signing digests; unsupported transaction types are counted explicitly.",
}
