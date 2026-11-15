package report

import (
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/quantumshield/qguard/backend/internal/domain"
	"github.com/quantumshield/qguard/backend/internal/version"
)

func TestBuildElevatedReport(t *testing.T) {
	addr := common.HexToAddress("0x2449ecef5012f0a0e153b278ef4fcc9625bc4c78")
	in := BuildInput{
		ScanID: "scan-1",
		Account: domain.AccountSnapshot{
			Address: addr, Kind: domain.AccountLikelyEOA, BalanceWei: big.NewInt(1), ChainID: 1,
		},
		History: domain.TransactionHistory{
			Coverage: domain.HistoryCoverage{Status: domain.CoverageCompleteWithinScope, PaginationComplete: true},
		},
		Exposure: domain.WalletExposureReport{
			Finding: domain.FindingPublicKeyExposed, ConfirmedCount: 1, DistinctPublicKeys: 1,
			Evidence: []domain.ExposureEvidence{{
				TransactionHash: common.HexToHash("0xabc"),
				RecoveryStatus:  domain.RecoveryConfirmed,
			}},
			Limitations: domain.ExposureAnalysisLimitations,
		},
		Risk: domain.RiskAssessment{
			Classification:        domain.RiskElevatedExposure,
			PublicKeyExposed:      true,
			VerifiedExposureCount: 1,
			DistinctPublicKeys:    1,
			Curve:                 "secp256k1",
			SignatureScheme:       "ecdsa",
			AccountKind:           domain.AccountLikelyEOA,
			HistoryCoverage:       domain.CoverageCompleteWithinScope,
			EvidenceQuality:       domain.EvidenceHigh,
			RecommendedNextSteps:  []string{"step"},
			Limitations:           []string{"lim"},
			Disclaimers:           domain.StandardDisclaimers,
			RiskModelVersion:      version.RiskModelVersion,
		},
		Now: time.Unix(1700000000, 0).UTC(),
	}
	rep := Build(in)
	if rep.RiskModelVersion != version.RiskModelVersion {
		t.Fatalf("version = %s", rep.RiskModelVersion)
	}
	if !strings.Contains(rep.ExecutiveSummary, "ELEVATED_EXPOSURE") {
		t.Fatalf("summary = %s", rep.ExecutiveSummary)
	}
	if !rep.ExposureFindings.Confirmed {
		t.Fatal("expected confirmed exposure findings")
	}
	if len(rep.Disclaimers) == 0 || len(rep.QuantumThreat.Bullets) == 0 {
		t.Fatal("expected disclaimers and quantum notes")
	}
	if rep.ScanID != "scan-1" || rep.Wallet.Address != addr {
		t.Fatalf("ids/address mismatch")
	}
}

func TestBuildContractReportAvoidsFakePQAdvice(t *testing.T) {
	in := BuildInput{
		ScanID: "scan-2",
		Account: domain.AccountSnapshot{
			Address: common.HexToAddress("0x1"), Kind: domain.AccountContract, HasCode: true, BalanceWei: big.NewInt(0), ChainID: 1,
		},
		History:  domain.TransactionHistory{Coverage: domain.HistoryCoverage{Status: domain.CoverageCompleteWithinScope}},
		Exposure: domain.WalletExposureReport{Finding: domain.FindingRequiresManualReview},
		Risk: domain.RiskAssessment{
			Classification:       domain.RiskRequiresReview,
			AccountKind:          domain.AccountContract,
			EvidenceQuality:      domain.EvidenceManual,
			RecommendedNextSteps: []string{"Review contract verification"},
			RiskModelVersion:     version.RiskModelVersion,
		},
		Now: time.Unix(1700000000, 0).UTC(),
	}
	rep := Build(in)
	joined := strings.Join(rep.Limitations, " ")
	if !strings.Contains(joined, "No unsupported post-quantum migration") {
		t.Fatalf("expected contract PQ limitation, got %v", rep.Limitations)
	}
}
