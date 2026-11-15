package history

import (
	"context"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/quantumshield/qguard/backend/internal/domain"
)

type stubSource struct {
	hist domain.TransactionHistory
	err  error
}

func (s stubSource) FetchNormalTransactions(ctx context.Context, address string) (domain.TransactionHistory, error) {
	return s.hist, s.err
}

type stubVerifier struct {
	hashes map[uint64]common.Hash
}

func (s stubVerifier) BlockHashByNumber(ctx context.Context, number uint64) (common.Hash, error) {
	h, ok := s.hashes[number]
	if !ok {
		return common.Hash{}, context.Canceled
	}
	return h, nil
}

func TestFetchInvalidAddress(t *testing.T) {
	f := &Fetcher{Source: stubSource{}}
	_, err := f.Fetch(context.Background(), "bad")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestBlockHashMismatchMarksPartial(t *testing.T) {
	bn := uint64(10)
	want := common.HexToHash("0x1111111111111111111111111111111111111111111111111111111111111111")
	got := common.HexToHash("0x2222222222222222222222222222222222222222222222222222222222222222")
	hist := domain.TransactionHistory{
		Address: "0x2449EcEF5012f0A0e153B278Ef4Fcc9625BC4C78",
		Coverage: domain.HistoryCoverage{
			Status:             domain.CoverageCompleteWithinScope,
			PaginationComplete: true,
		},
		Transactions: []domain.NormalizedTransaction{{
			Hash:        common.HexToHash("0xf9db905d77704596d3600816bc70201586cfeec13bcf576320e2f38d6ca851a0"),
			BlockNumber: &bn,
			BlockHash:   &want,
			From:        common.HexToAddress("0x2449ecef5012f0a0e153b278ef4fcc9625bc4c78"),
		}},
	}
	f := &Fetcher{
		Source:        stubSource{hist: hist},
		Verifier:      stubVerifier{hashes: map[uint64]common.Hash{10: got}},
		MaxHashChecks: 5,
	}
	out, err := f.Fetch(context.Background(), "0x2449ecef5012f0a0e153b278ef4fcc9625bc4c78")
	if err != nil {
		t.Fatal(err)
	}
	if out.Coverage.BlockHashChecks.Mismatched != 1 {
		t.Fatalf("checks = %+v", out.Coverage.BlockHashChecks)
	}
	if out.Coverage.Status != domain.CoveragePartial {
		t.Fatalf("status = %s", out.Coverage.Status)
	}
}

func TestBlockHashMatchKeepsComplete(t *testing.T) {
	bn := uint64(10)
	h := common.HexToHash("0x1111111111111111111111111111111111111111111111111111111111111111")
	hist := domain.TransactionHistory{
		Coverage: domain.HistoryCoverage{Status: domain.CoverageCompleteWithinScope, PaginationComplete: true},
		Transactions: []domain.NormalizedTransaction{{
			Hash:        common.HexToHash("0xf9db905d77704596d3600816bc70201586cfeec13bcf576320e2f38d6ca851a0"),
			BlockNumber: &bn,
			BlockHash:   &h,
			From:        common.HexToAddress("0x2449ecef5012f0a0e153b278ef4fcc9625bc4c78"),
		}},
	}
	f := &Fetcher{
		Source:        stubSource{hist: hist},
		Verifier:      stubVerifier{hashes: map[uint64]common.Hash{10: h}},
		MaxHashChecks: 5,
	}
	out, err := f.Fetch(context.Background(), "0x2449ecef5012f0a0e153b278ef4fcc9625bc4c78")
	if err != nil {
		t.Fatal(err)
	}
	if out.Coverage.Status != domain.CoverageCompleteWithinScope {
		t.Fatalf("status = %s", out.Coverage.Status)
	}
	if out.Coverage.BlockHashChecks.Matched != 1 {
		t.Fatalf("checks = %+v", out.Coverage.BlockHashChecks)
	}
}
