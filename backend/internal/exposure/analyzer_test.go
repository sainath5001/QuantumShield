package exposure

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/quantumshield/qguard/backend/internal/domain"
)

const testPrivHex = "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"

type mapTxProvider map[common.Hash]*types.Transaction

func (m mapTxProvider) TransactionByHash(ctx context.Context, hash common.Hash) (*types.Transaction, bool, error) {
	tx, ok := m[hash]
	return tx, ok, nil
}

func testKey(t *testing.T) *ecdsa.PrivateKey {
	t.Helper()
	key, err := crypto.HexToECDSA(testPrivHex)
	if err != nil {
		t.Fatal(err)
	}
	return key
}

func signedDynamic(t *testing.T, key *ecdsa.PrivateKey, nonce uint64) *types.Transaction {
	t.Helper()
	to := common.HexToAddress("0x0000000000000000000000000000000000000001")
	tx := types.NewTx(&types.DynamicFeeTx{
		ChainID:   big.NewInt(1),
		Nonce:     nonce,
		To:        &to,
		Value:     big.NewInt(1),
		Gas:       21000,
		GasFeeCap: big.NewInt(2),
		GasTipCap: big.NewInt(1),
	})
	signed, err := types.SignTx(tx, types.LatestSignerForChainID(big.NewInt(1)), key)
	if err != nil {
		t.Fatal(err)
	}
	return signed
}

func TestAnalyzeConfirmedExposure(t *testing.T) {
	key := testKey(t)
	addr := crypto.PubkeyToAddress(key.PublicKey)
	tx1 := signedDynamic(t, key, 1)
	tx2 := signedDynamic(t, key, 2)
	provider := mapTxProvider{tx1.Hash(): tx1, tx2.Hash(): tx2}

	bn := uint64(100)
	hist := domain.TransactionHistory{
		Coverage: domain.HistoryCoverage{Status: domain.CoverageCompleteWithinScope, PaginationComplete: true},
		Transactions: []domain.NormalizedTransaction{
			{Hash: tx1.Hash(), From: addr, BlockNumber: &bn, Source: domain.TxSourceEtherscan},
			{Hash: tx2.Hash(), From: addr, BlockNumber: &bn, Source: domain.TxSourceEtherscan},
		},
	}

	a := &Analyzer{Txs: provider, ChainID: big.NewInt(1), Now: func() time.Time { return time.Unix(1700000000, 0).UTC() }}
	report, err := a.Analyze(context.Background(), AnalyzeInput{
		Address:     addr,
		AccountKind: domain.AccountLikelyEOA,
		History:     hist,
	})
	if err != nil {
		t.Fatal(err)
	}
	if report.Finding != domain.FindingPublicKeyExposed || !report.ConfirmedExposure {
		t.Fatalf("report = %+v", report)
	}
	if report.ConfirmedCount != 2 || report.DistinctPublicKeys != 1 {
		t.Fatalf("confirmed=%d keys=%d", report.ConfirmedCount, report.DistinctPublicKeys)
	}
}

func TestAnalyzeContractRequiresReview(t *testing.T) {
	a := &Analyzer{Txs: mapTxProvider{}, ChainID: big.NewInt(1)}
	report, err := a.Analyze(context.Background(), AnalyzeInput{
		Address:     common.HexToAddress("0x0000000000000000000000000000000000000001"),
		AccountKind: domain.AccountContract,
		History:     domain.TransactionHistory{Coverage: domain.HistoryCoverage{Status: domain.CoverageCompleteWithinScope}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if report.Finding != domain.FindingRequiresManualReview {
		t.Fatalf("finding = %s", report.Finding)
	}
	if report.ConfirmedExposure {
		t.Fatal("contract must not get confirmed EOA exposure")
	}
	if report.AnalyzedTxCount != 0 {
		t.Fatal("contract path must not run EOA recovery")
	}
}

func TestAnalyzeNoSenderTxs(t *testing.T) {
	addr := crypto.PubkeyToAddress(testKey(t).PublicKey)
	a := &Analyzer{Txs: mapTxProvider{}, ChainID: big.NewInt(1)}
	report, err := a.Analyze(context.Background(), AnalyzeInput{
		Address:     addr,
		AccountKind: domain.AccountLikelyEOA,
		History: domain.TransactionHistory{
			Coverage:     domain.HistoryCoverage{Status: domain.CoverageCompleteWithinScope, PaginationComplete: true},
			Transactions: nil,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if report.Finding != domain.FindingNoExposureObserved {
		t.Fatalf("finding = %s", report.Finding)
	}
}

func TestAnalyzePartialHistoryInsufficient(t *testing.T) {
	key := testKey(t)
	addr := crypto.PubkeyToAddress(key.PublicKey)
	tx := signedDynamic(t, key, 1)
	// History lists a tx but provider cannot fetch it.
	a := &Analyzer{Txs: mapTxProvider{}, ChainID: big.NewInt(1)}
	report, err := a.Analyze(context.Background(), AnalyzeInput{
		Address:     addr,
		AccountKind: domain.AccountLikelyEOA,
		History: domain.TransactionHistory{
			Coverage: domain.HistoryCoverage{Status: domain.CoveragePartial},
			Transactions: []domain.NormalizedTransaction{
				{Hash: tx.Hash(), From: addr},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if report.Finding != domain.FindingInsufficientEvidence {
		t.Fatalf("finding = %s", report.Finding)
	}
}

func TestAnalyzeIgnoresIncomingOnly(t *testing.T) {
	key := testKey(t)
	addr := crypto.PubkeyToAddress(key.PublicKey)
	other := common.HexToAddress("0x00000000000000000000000000000000000000bb")
	tx := signedDynamic(t, key, 1)
	a := &Analyzer{Txs: mapTxProvider{tx.Hash(): tx}, ChainID: big.NewInt(1)}
	report, err := a.Analyze(context.Background(), AnalyzeInput{
		Address:     other, // not the signer
		AccountKind: domain.AccountLikelyEOA,
		History: domain.TransactionHistory{
			Coverage: domain.HistoryCoverage{Status: domain.CoverageCompleteWithinScope, PaginationComplete: true},
			Transactions: []domain.NormalizedTransaction{
				{Hash: tx.Hash(), From: addr}, // incoming relative to `other`
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if report.SenderTxCount != 0 || report.ConfirmedExposure {
		t.Fatalf("incoming txs must not count as sender exposure: %+v", report)
	}
}
