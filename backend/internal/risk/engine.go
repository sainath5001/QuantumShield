// Package risk implements the deterministic qguard-risk-v1 classification engine.
// It is independent of HTTP handlers and performs no I/O.
package risk

import (
	"time"

	"github.com/quantumshield/qguard/backend/internal/domain"
	"github.com/quantumshield/qguard/backend/internal/quantum"
	"github.com/quantumshield/qguard/backend/internal/version"
)

const unsupportedRatioThreshold = 0.50

// Input is the pure evidence bundle consumed by Assess.
type Input struct {
	Account  domain.AccountSnapshot
	History  domain.TransactionHistory
	Exposure domain.WalletExposureReport
	Now      time.Time
}

// Assess applies qguard-risk-v1 rules and returns a RiskAssessment.
func Assess(in Input) domain.RiskAssessment {
	now := in.Now
	if now.IsZero() {
		now = time.Now().UTC()
	}

	base := domain.RiskAssessment{
		AccountKind:           in.Account.Kind,
		HistoryCoverage:       in.History.Coverage.Status,
		AnalyzedTxCount:       in.Exposure.AnalyzedTxCount,
		UnsupportedTxCount:    in.Exposure.UnsupportedCount,
		FailedTxCount:         in.Exposure.FailedCount,
		VerifiedExposureCount: in.Exposure.ConfirmedCount,
		SenderMismatchCount:   in.Exposure.SenderMismatchCount,
		SenderTxCount:         in.Exposure.SenderTxCount,
		DistinctPublicKeys:    in.Exposure.DistinctPublicKeys,
		PublicKeyExposed:      in.Exposure.ConfirmedCount > 0,
		Disclaimers:           append([]string{}, domain.StandardDisclaimers...),
		Limitations:           uniqueStrings(append(append([]string{}, in.Exposure.Limitations...), in.Account.Limitations...)),
		RiskModelVersion:      version.RiskModelVersion,
		AssessedAt:            now.UTC(),
		QuantumThreatNotes:    quantum.AssessmentNotes(),
	}

	// Contract wallets: always REQUIRES_REVIEW in v1.
	if in.Account.Kind == domain.AccountContract {
		base.Classification = domain.RiskRequiresReview
		base.EvidenceQuality = domain.EvidenceManual
		base.Reason = "Address has deployed code; authentication may use custom contract logic. EOA secp256k1 exposure rules are not applied as a final conclusion."
		base.RecommendedNextSteps = []string{
			"Review the contract's signature verification mechanism (owners, multisig threshold, EIP-1271, ERC-4337 validateUserOp).",
			"Identify which EOAs, if any, authorize the account and scan those addresses separately.",
			"Do not treat this contract address as quantum-safe or quantum-broken without that review.",
		}
		return base
	}

	if in.Account.Kind == domain.AccountUnknown {
		base.Classification = domain.RiskUnknown
		base.EvidenceQuality = domain.EvidenceLow
		base.Reason = "Account kind could not be determined; insufficient evidence for classification."
		base.RecommendedNextSteps = []string{"Retry the scan when RPC connectivity is available."}
		return base
	}

	// EOA path — scheme known when we analyze ECDSA txs.
	base.Curve = "secp256k1"
	base.SignatureScheme = "ecdsa"

	cov := in.History.Coverage.Status
	verified := in.Exposure.ConfirmedCount
	unsupported := in.Exposure.UnsupportedCount
	analyzed := in.Exposure.AnalyzedTxCount
	failed := in.Exposure.FailedCount
	senderTxs := in.Exposure.SenderTxCount

	// Rule: confirmed exposure → ELEVATED (even with partial history).
	if verified >= 1 {
		base.Classification = domain.RiskElevatedExposure
		base.PublicKeyExposed = true
		base.Reason = "At least one secp256k1 public key was recovered from a valid transaction signature and verified against the sender address."
		base.EvidenceQuality = domain.EvidenceHigh
		if unsupported > 0 || failed > 0 || cov == domain.CoveragePartial || cov == domain.CoverageRateLimited {
			base.EvidenceQuality = domain.EvidenceMedium
		}
		base.RecommendedNextSteps = elevatedSteps()
		base.Limitations = append(base.Limitations,
			"Elevated exposure refers to long-term risk under a hypothetical cryptographically relevant quantum computer (Shor/ECC-DL), not a present-day classical break.",
		)
		return base
	}

	// No verified exposure below this point.

	// Coverage none/failed/unknown without exposure → UNKNOWN
	switch cov {
	case domain.CoverageFailed, domain.CoverageUnknown, "":
		base.Classification = domain.RiskUnknown
		base.EvidenceQuality = domain.EvidenceLow
		base.Reason = "History coverage is failed or unknown and no verified public-key exposure was observed."
		base.RecommendedNextSteps = unknownSteps()
		return base
	case domain.CoveragePartial, domain.CoverageRateLimited:
		base.Classification = domain.RiskUnknown
		base.EvidenceQuality = domain.EvidenceLow
		base.Reason = "History is incomplete (partial or rate-limited) and no verified public-key exposure was observed; absence of evidence is not treated as lower exposure."
		base.RecommendedNextSteps = unknownSteps()
		return base
	case domain.CoverageUnsupported:
		base.Classification = domain.RiskUnknown
		base.EvidenceQuality = domain.EvidenceLow
		base.Reason = "History provider returned unsupported responses; exposure status cannot be established."
		base.RecommendedNextSteps = unknownSteps()
		return base
	}

	// complete_within_scope path
	if cov == domain.CoverageCompleteWithinScope {
		if unsupported > 0 && analyzed == 0 && senderTxs > 0 {
			base.Classification = domain.RiskUnknown
			base.EvidenceQuality = domain.EvidenceLow
			base.Reason = "Sender transactions exist but none could be analyzed as supported types; no verified exposure."
			base.RecommendedNextSteps = unknownSteps()
			return base
		}
		denom := senderTxs
		if denom == 0 {
			denom = analyzed + unsupported + failed
		}
		if denom > 0 && float64(unsupported)/float64(denom) >= unsupportedRatioThreshold && verified == 0 {
			base.Classification = domain.RiskUnknown
			base.EvidenceQuality = domain.EvidenceLow
			base.Reason = "Unsupported transactions are a high fraction of examined sender history; evidence is insufficient for lower-exposure classification."
			base.RecommendedNextSteps = unknownSteps()
			return base
		}
		if failed > 0 && analyzed == 0 && unsupported == 0 && senderTxs > 0 {
			base.Classification = domain.RiskUnknown
			base.EvidenceQuality = domain.EvidenceLow
			base.Reason = "All examined sender transactions failed recovery/fetch; evidence is insufficient."
			base.RecommendedNextSteps = unknownSteps()
			return base
		}

		// Lower exposure within scope.
		base.Classification = domain.RiskLowerExposure
		base.EvidenceQuality = domain.EvidenceHigh
		if unsupported > 0 || failed > 0 || in.Exposure.SenderMismatchCount > 0 {
			base.EvidenceQuality = domain.EvidenceMedium
		}
		if senderTxs == 0 {
			base.Reason = "No sender transactions were present in the examined history scope; no public-key exposure was observed. This does not prove the key was never exposed."
		} else {
			base.Reason = "No verified public-key exposure was observed across the supported, complete-within-scope examined history. This does not prove the key was never exposed."
		}
		base.RecommendedNextSteps = lowerSteps()
		return base
	}

	base.Classification = domain.RiskUnknown
	base.EvidenceQuality = domain.EvidenceLow
	base.Reason = "Evidence could not be mapped to a confident classification under qguard-risk-v1."
	base.RecommendedNextSteps = unknownSteps()
	return base
}

func elevatedSteps() []string {
	return []string{
		"Treat the recovered secp256k1 public key as harvestable for a future cryptographically relevant quantum adversary.",
		"Prefer moving funds to an address whose public key has not been observed in on-chain signatures, understanding migration and operational risks.",
		"Monitor Ethereum ecosystem post-quantum / account-abstraction migration guidance; do not invent unsupported PQ schemes for this wallet.",
		"Remember: exposure is not evidence of present-day classical theft.",
	}
}

func lowerSteps() []string {
	return []string{
		"Continue avoiding unnecessary signatures that would expose the public key.",
		"Re-scan periodically; new outbound transactions can change exposure status.",
		"Do not interpret this result as quantum-safe certification.",
	}
}

func unknownSteps() []string {
	return []string{
		"Retry the scan when RPC/indexer connectivity and rate limits allow fuller history.",
		"Increase page/tx caps only if provider limits and local resources allow.",
		"Treat the result as inconclusive rather than safe.",
	}
}

func uniqueStrings(in []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(in))
	for _, s := range in {
		if s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}
