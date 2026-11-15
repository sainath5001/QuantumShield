package service

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/quantumshield/qguard/backend/internal/domain"
	"github.com/quantumshield/qguard/backend/internal/exposure"
	"github.com/quantumshield/qguard/backend/internal/history"
	"github.com/quantumshield/qguard/backend/internal/repository"
)

type stubRPC struct {
	kind domain.AccountKind
	code []byte
	txs  map[common.Hash]*types.Transaction
}

func (s stubRPC) Close()                                          {}
func (s stubRPC) VerifyConnection(ctx context.Context) error      { return nil }
func (s stubRPC) ChainID(ctx context.Context) (*big.Int, error)   { return big.NewInt(1), nil }
func (s stubRPC) BlockNumber(ctx context.Context) (uint64, error) { return 1, nil }
func (s stubRPC) BalanceAt(ctx context.Context, addr common.Address) (*big.Int, error) {
	return big.NewInt(0), nil
}
func (s stubRPC) CodeAt(ctx context.Context, addr common.Address) ([]byte, error) { return s.code, nil }
func (s stubRPC) BlockHashByNumber(ctx context.Context, number uint64) (common.Hash, error) {
	return common.Hash{}, nil
}
func (s stubRPC) TransactionByHash(ctx context.Context, hash common.Hash) (*types.Transaction, bool, error) {
	tx, ok := s.txs[hash]
	return tx, ok, nil
}
func (s stubRPC) ClassifyAccount(ctx context.Context, addr common.Address) (domain.AccountSnapshot, error) {
	return domain.AccountSnapshot{
		Address: addr, Kind: s.kind, HasCode: len(s.code) > 0,
		BalanceWei: big.NewInt(0), ChainID: 1, BlockNumber: 1,
		Limitations: domain.AccountClassificationLimitations,
	}, nil
}

type stubHistory struct {
	hist domain.TransactionHistory
}

func (s stubHistory) FetchNormalTransactions(ctx context.Context, address string) (domain.TransactionHistory, error) {
	return s.hist, nil
}

func TestScanLifecycleEOAElevated(t *testing.T) {
	key, err := crypto.HexToECDSA("ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80")
	if err != nil {
		t.Fatal(err)
	}
	addr := crypto.PubkeyToAddress(key.PublicKey)
	to := common.HexToAddress("0x0000000000000000000000000000000000000001")
	unsigned := types.NewTx(&types.DynamicFeeTx{
		ChainID: big.NewInt(1), Nonce: 1, To: &to, Value: big.NewInt(1),
		Gas: 21000, GasFeeCap: big.NewInt(2), GasTipCap: big.NewInt(1),
	})
	signed, err := types.SignTx(unsigned, types.LatestSignerForChainID(big.NewInt(1)), key)
	if err != nil {
		t.Fatal(err)
	}

	rpc := stubRPC{
		kind: domain.AccountLikelyEOA,
		txs:  map[common.Hash]*types.Transaction{signed.Hash(): signed},
	}
	histSrc := stubHistory{hist: domain.TransactionHistory{
		Address: addr.Hex(),
		Coverage: domain.HistoryCoverage{
			Status: domain.CoverageCompleteWithinScope, PaginationComplete: true, Provider: "etherscan", ChainID: 1,
		},
		Transactions: []domain.NormalizedTransaction{
			{Hash: signed.Hash(), From: addr, Source: domain.TxSourceEtherscan},
		},
	}}
	store := repository.NewMemory(repository.MemoryConfig{})
	svc := &ScanService{
		Store:    store,
		RPC:      rpc,
		History:  &history.Fetcher{Source: histSrc, MaxHashChecks: 0},
		Exposure: &exposure.Analyzer{Txs: rpc, ChainID: big.NewInt(1)},
		ChainID:  1,
		MaxTxs:   10,
		Timeout:  5 * time.Second,
	}

	job, err := svc.StartScan(context.Background(), addr.Hex())
	if err != nil {
		t.Fatal(err)
	}

	deadline := time.Now().Add(3 * time.Second)
	var final *domain.ScanJob
	for time.Now().Before(deadline) {
		final, err = svc.GetScan(context.Background(), job.ID)
		if err != nil {
			t.Fatal(err)
		}
		if final.Status == domain.ScanCompleted || final.Status == domain.ScanFailed {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if final == nil || final.Status != domain.ScanCompleted {
		t.Fatalf("final = %+v", final)
	}
	if final.Risk == nil || final.Risk.Classification != domain.RiskElevatedExposure {
		t.Fatalf("risk = %+v", final.Risk)
	}
	if final.ReportID == "" {
		t.Fatal("expected report id")
	}
	rep, err := svc.GetReport(context.Background(), final.ReportID)
	if err != nil || rep.Risk.Classification != domain.RiskElevatedExposure {
		t.Fatalf("report = %+v err=%v", rep, err)
	}
	risk, err := svc.LatestRisk(context.Background(), addr.Hex())
	if err != nil || risk.Classification != domain.RiskElevatedExposure {
		t.Fatalf("latest risk = %+v err=%v", risk, err)
	}
}

func TestScanLifecycleContractRequiresReview(t *testing.T) {
	addr := common.HexToAddress("0x00000000000000000000000000000000000000aa")
	rpc := stubRPC{kind: domain.AccountContract, code: []byte{0x60}, txs: map[common.Hash]*types.Transaction{}}
	histSrc := stubHistory{hist: domain.TransactionHistory{
		Coverage: domain.HistoryCoverage{Status: domain.CoverageCompleteWithinScope, PaginationComplete: true},
	}}
	store := repository.NewMemory(repository.MemoryConfig{})
	svc := &ScanService{
		Store: store, RPC: rpc,
		History:  &history.Fetcher{Source: histSrc, MaxHashChecks: 0},
		Exposure: &exposure.Analyzer{Txs: rpc, ChainID: big.NewInt(1)},
		ChainID:  1, Timeout: 5 * time.Second,
	}
	job, err := svc.StartScan(context.Background(), addr.Hex())
	if err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(3 * time.Second)
	var final *domain.ScanJob
	for time.Now().Before(deadline) {
		final, _ = svc.GetScan(context.Background(), job.ID)
		if final.Status == domain.ScanCompleted || final.Status == domain.ScanFailed {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if final.Status != domain.ScanCompleted || final.Risk.Classification != domain.RiskRequiresReview {
		t.Fatalf("final = %+v", final)
	}
}
