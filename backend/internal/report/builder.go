// Package report assembles SecurityReport DTOs from scan evidence and risk assessments.
package report

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/quantumshield/qguard/backend/internal/domain"
	"github.com/quantumshield/qguard/backend/internal/quantum"
	"github.com/quantumshield/qguard/backend/internal/version"
)

// BuildInput is the evidence required to assemble a complete security report.
type BuildInput struct {
	ScanID   string
	Account  domain.AccountSnapshot
	History  domain.TransactionHistory
	Exposure domain.WalletExposureReport
	Risk     domain.RiskAssessment
	Now      time.Time
}

// Build creates a SecurityReport. It does not invent chain data.
func Build(in BuildInput) domain.SecurityReport {
	now := in.Now
	if now.IsZero() {
		now = time.Now().UTC()
	}
	shor, grover, practical, bullets := quantum.ReportExplanation(in.Risk.PublicKeyExposed)

	bal := "0"
	if in.Account.BalanceWei != nil {
		bal = in.Account.BalanceWei.String()
	}

	paginationComplete := in.History.Coverage.PaginationComplete
	rep := domain.SecurityReport{
		ID:               uuid.NewString(),
		ScanID:           in.ScanID,
		GeneratedAt:      now,
		RiskModelVersion: version.RiskModelVersion,
		ExecutiveSummary: executiveSummary(in.Risk),
		Wallet: domain.ReportWalletInfo{
			Address:     in.Account.Address,
			AccountKind: in.Account.Kind,
			HasCode:     in.Account.HasCode,
			BalanceWei:  bal,
			ChainID:     in.Account.ChainID,
			BlockNumber: in.Account.BlockNumber,
		},
		Cryptography: domain.ReportCryptoInfo{
			Curve:           in.Risk.Curve,
			SignatureScheme: in.Risk.SignatureScheme,
			Notes:           cryptoNotes(in.Risk),
		},
		ExposureFindings: domain.ReportExposureInfo{
			Confirmed:             in.Risk.PublicKeyExposed,
			Finding:               in.Exposure.Finding,
			VerifiedExposureCount: in.Risk.VerifiedExposureCount,
			DistinctPublicKeys:    in.Risk.DistinctPublicKeys,
		},
		Evidence: append([]domain.ExposureEvidence{}, in.Exposure.Evidence...),
		QuantumThreat: domain.ReportQuantumInfo{
			ShorRelevance:   shor,
			GroverRelevance: grover,
			PracticalNote:   practical,
			Bullets:         bullets,
		},
		Risk: in.Risk,
		Coverage: domain.ReportCoverageInfo{
			HistoryCoverage:    in.Risk.HistoryCoverage,
			EvidenceQuality:    in.Risk.EvidenceQuality,
			AnalyzedTxCount:    in.Risk.AnalyzedTxCount,
			UnsupportedTxCount: in.Risk.UnsupportedTxCount,
			FailedTxCount:      in.Risk.FailedTxCount,
			SenderTxCount:      in.Risk.SenderTxCount,
			PaginationComplete: paginationComplete,
		},
		RecommendedSteps: append([]string{}, in.Risk.RecommendedNextSteps...),
		Limitations:      unique(append(append([]string{}, in.Risk.Limitations...), in.Exposure.Limitations...)),
		Disclaimers:      append([]string{}, domain.StandardDisclaimers...),
	}

	// Contract wallets: do not invent unsupported PQ migration advice.
	if in.Account.Kind == domain.AccountContract {
		rep.RecommendedSteps = filterContractSafeSteps(rep.RecommendedSteps)
		rep.Limitations = append(rep.Limitations,
			"No unsupported post-quantum migration recommendation is made for this contract wallet; review its authentication logic first.",
		)
	}
	return rep
}

func executiveSummary(r domain.RiskAssessment) string {
	switch r.Classification {
	case domain.RiskElevatedExposure:
		return fmt.Sprintf("Classification %s: a verified secp256k1 public key was recovered for this EOA within the examined scope. This elevates long-term harvest-now risk under a hypothetical cryptographically relevant quantum computer — not evidence of a present-day break.", r.Classification)
	case domain.RiskLowerExposure:
		return fmt.Sprintf("Classification %s: no verified public-key exposure was observed within the examined scope. This is not a quantum-safety certificate.", r.Classification)
	case domain.RiskRequiresReview:
		return fmt.Sprintf("Classification %s: the address appears to be a contract/smart account. Manual review of its signature verification logic is required before EOA-style quantum exposure conclusions.", r.Classification)
	default:
		return fmt.Sprintf("Classification %s: evidence quality or history coverage is insufficient for a confident lower/elevated exposure conclusion.", r.Classification)
	}
}

func cryptoNotes(r domain.RiskAssessment) string {
	if r.AccountKind == domain.AccountContract {
		return "Contract account: authentication scheme is defined by contract code; EOA ECDSA assumptions do not automatically apply."
	}
	if r.Curve == "" {
		return "Cryptographic scheme not established for this assessment."
	}
	return "Ethereum EOA transactions use ECDSA over secp256k1. Public keys are recovered from signatures using go-ethereum canonical signing digests."
}

func filterContractSafeSteps(steps []string) []string {
	out := make([]string, 0, len(steps))
	for _, s := range steps {
		out = append(out, s)
	}
	return out
}

func unique(in []string) []string {
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
