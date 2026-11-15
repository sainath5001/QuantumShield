package risk

import (
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/quantumshield/qguard/backend/internal/domain"
	"github.com/quantumshield/qguard/backend/internal/version"
)

func baseAccount(kind domain.AccountKind) domain.AccountSnapshot {
	return domain.AccountSnapshot{
		Address:     common.HexToAddress("0x2449ecef5012f0a0e153b278ef4fcc9625bc4c78"),
		Kind:        kind,
		HasCode:     kind == domain.AccountContract,
		BalanceWei:  big.NewInt(0),
		ChainID:     1,
		BlockNumber: 1,
		Limitations: domain.AccountClassificationLimitations,
	}
}

func TestElevatedOnConfirmedExposure(t *testing.T) {
	in := Input{
		Account: baseAccount(domain.AccountLikelyEOA),
		History: domain.TransactionHistory{Coverage: domain.HistoryCoverage{Status: domain.CoveragePartial}},
		Exposure: domain.WalletExposureReport{
			Finding:               domain.FindingPublicKeyExposed,
			ConfirmedCount:        1,
			ConfirmedExposure:     true,
			AnalyzedTxCount:       1,
			SenderTxCount:         1,
			HistoryCoverageStatus: domain.CoveragePartial,
			Limitations:           domain.ExposureAnalysisLimitations,
		},
		Now: time.Unix(1700000000, 0).UTC(),
	}
	out := Assess(in)
	if out.Classification != domain.RiskElevatedExposure {
		t.Fatalf("got %s", out.Classification)
	}
	if out.RiskModelVersion != version.RiskModelVersion {
		t.Fatalf("version = %s", out.RiskModelVersion)
	}
	if out.EvidenceQuality != domain.EvidenceMedium {
		t.Fatalf("quality = %s (partial history should be MEDIUM)", out.EvidenceQuality)
	}
	if len(out.Disclaimers) == 0 {
		t.Fatal("expected disclaimers")
	}
}

func TestLowerExposureCompleteNoExposure(t *testing.T) {
	in := Input{
		Account: baseAccount(domain.AccountLikelyEOA),
		History: domain.TransactionHistory{Coverage: domain.HistoryCoverage{Status: domain.CoverageCompleteWithinScope, PaginationComplete: true}},
		Exposure: domain.WalletExposureReport{
			Finding:               domain.FindingNoExposureObserved,
			AnalyzedTxCount:       3,
			SenderTxCount:         3,
			HistoryCoverageStatus: domain.CoverageCompleteWithinScope,
			Limitations:           domain.ExposureAnalysisLimitations,
		},
		Now: time.Unix(1700000000, 0).UTC(),
	}
	out := Assess(in)
	if out.Classification != domain.RiskLowerExposure {
		t.Fatalf("got %s reason=%s", out.Classification, out.Reason)
	}
	if out.EvidenceQuality != domain.EvidenceHigh {
		t.Fatalf("quality = %s", out.EvidenceQuality)
	}
}

func TestLowerExposureEmptyHistory(t *testing.T) {
	in := Input{
		Account: baseAccount(domain.AccountLikelyEOA),
		History: domain.TransactionHistory{Coverage: domain.HistoryCoverage{Status: domain.CoverageCompleteWithinScope, PaginationComplete: true}},
		Exposure: domain.WalletExposureReport{
			Finding:               domain.FindingNoExposureObserved,
			SenderTxCount:         0,
			HistoryCoverageStatus: domain.CoverageCompleteWithinScope,
			Limitations:           domain.ExposureAnalysisLimitations,
		},
		Now: time.Unix(1700000000, 0).UTC(),
	}
	out := Assess(in)
	if out.Classification != domain.RiskLowerExposure {
		t.Fatalf("got %s", out.Classification)
	}
}

func TestUnknownPartialNoExposure(t *testing.T) {
	in := Input{
		Account: baseAccount(domain.AccountLikelyEOA),
		History: domain.TransactionHistory{Coverage: domain.HistoryCoverage{Status: domain.CoveragePartial}},
		Exposure: domain.WalletExposureReport{
			Finding:               domain.FindingInsufficientEvidence,
			SenderTxCount:         2,
			AnalyzedTxCount:       2,
			HistoryCoverageStatus: domain.CoveragePartial,
		},
		Now: time.Unix(1700000000, 0).UTC(),
	}
	out := Assess(in)
	if out.Classification != domain.RiskUnknown {
		t.Fatalf("got %s", out.Classification)
	}
}

func TestUnknownRateLimited(t *testing.T) {
	in := Input{
		Account:  baseAccount(domain.AccountLikelyEOA),
		History:  domain.TransactionHistory{Coverage: domain.HistoryCoverage{Status: domain.CoverageRateLimited}},
		Exposure: domain.WalletExposureReport{SenderTxCount: 1, AnalyzedTxCount: 1},
		Now:      time.Unix(1700000000, 0).UTC(),
	}
	out := Assess(in)
	if out.Classification != domain.RiskUnknown {
		t.Fatalf("got %s", out.Classification)
	}
}

func TestUnknownHighUnsupportedRatio(t *testing.T) {
	in := Input{
		Account: baseAccount(domain.AccountLikelyEOA),
		History: domain.TransactionHistory{Coverage: domain.HistoryCoverage{Status: domain.CoverageCompleteWithinScope}},
		Exposure: domain.WalletExposureReport{
			SenderTxCount:         10,
			AnalyzedTxCount:       2,
			UnsupportedCount:      8,
			ConfirmedCount:        0,
			HistoryCoverageStatus: domain.CoverageCompleteWithinScope,
		},
		Now: time.Unix(1700000000, 0).UTC(),
	}
	out := Assess(in)
	if out.Classification != domain.RiskUnknown {
		t.Fatalf("got %s reason=%s", out.Classification, out.Reason)
	}
}

func TestContractRequiresReview(t *testing.T) {
	in := Input{
		Account:  baseAccount(domain.AccountContract),
		History:  domain.TransactionHistory{Coverage: domain.HistoryCoverage{Status: domain.CoverageCompleteWithinScope}},
		Exposure: domain.WalletExposureReport{Finding: domain.FindingRequiresManualReview},
		Now:      time.Unix(1700000000, 0).UTC(),
	}
	out := Assess(in)
	if out.Classification != domain.RiskRequiresReview {
		t.Fatalf("got %s", out.Classification)
	}
	if out.EvidenceQuality != domain.EvidenceManual {
		t.Fatalf("quality = %s", out.EvidenceQuality)
	}
}

func TestConflictingSenderMismatchStillElevatedIfConfirmed(t *testing.T) {
	// Confirmed count takes precedence; mismatches alone don't elevate.
	in := Input{
		Account: baseAccount(domain.AccountLikelyEOA),
		History: domain.TransactionHistory{Coverage: domain.HistoryCoverage{Status: domain.CoverageCompleteWithinScope}},
		Exposure: domain.WalletExposureReport{
			ConfirmedCount:        1,
			SenderMismatchCount:   2,
			AnalyzedTxCount:       3,
			SenderTxCount:         3,
			HistoryCoverageStatus: domain.CoverageCompleteWithinScope,
		},
		Now: time.Unix(1700000000, 0).UTC(),
	}
	out := Assess(in)
	if out.Classification != domain.RiskElevatedExposure {
		t.Fatalf("got %s", out.Classification)
	}
}

func TestFailedHistoryUnknown(t *testing.T) {
	in := Input{
		Account:  baseAccount(domain.AccountLikelyEOA),
		History:  domain.TransactionHistory{Coverage: domain.HistoryCoverage{Status: domain.CoverageFailed}},
		Exposure: domain.WalletExposureReport{},
		Now:      time.Unix(1700000000, 0).UTC(),
	}
	out := Assess(in)
	if out.Classification != domain.RiskUnknown {
		t.Fatalf("got %s", out.Classification)
	}
}
