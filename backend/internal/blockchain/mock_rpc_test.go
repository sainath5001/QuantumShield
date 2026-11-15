package blockchain

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/quantumshield/qguard/backend/internal/domain"
)

type mockRPC struct {
	chainID     int64
	blockNumber uint64
	balance     *big.Int
	code        []byte
	blockHashes map[uint64]common.Hash
	failChain   bool
}

func (m *mockRPC) Close() {}
func (m *mockRPC) VerifyConnection(ctx context.Context) error {
	if m.failChain {
		return ErrChainMismatch{Configured: MainnetChainID, Observed: 5}
	}
	if m.chainID != MainnetChainID {
		return ErrChainMismatch{Configured: MainnetChainID, Observed: m.chainID}
	}
	return nil
}
func (m *mockRPC) ChainID(ctx context.Context) (*big.Int, error) {
	return big.NewInt(m.chainID), nil
}
func (m *mockRPC) BlockNumber(ctx context.Context) (uint64, error) { return m.blockNumber, nil }
func (m *mockRPC) BalanceAt(ctx context.Context, addr common.Address) (*big.Int, error) {
	return m.balance, nil
}
func (m *mockRPC) CodeAt(ctx context.Context, addr common.Address) ([]byte, error) {
	return m.code, nil
}
func (m *mockRPC) BlockHashByNumber(ctx context.Context, number uint64) (common.Hash, error) {
	if h, ok := m.blockHashes[number]; ok {
		return h, nil
	}
	return common.Hash{}, context.Canceled
}
func (m *mockRPC) ClassifyAccount(ctx context.Context, addr common.Address) (domain.AccountSnapshot, error) {
	if err := m.VerifyConnection(ctx); err != nil {
		return domain.AccountSnapshot{}, err
	}
	snap := domain.AccountSnapshot{
		Address:     addr,
		BalanceWei:  m.balance,
		ChainID:     MainnetChainID,
		BlockNumber: m.blockNumber,
		Limitations: append([]string{}, domain.AccountClassificationLimitations...),
	}
	if len(m.code) == 0 {
		snap.Kind = domain.AccountLikelyEOA
	} else {
		snap.Kind = domain.AccountContract
		snap.HasCode = true
	}
	return snap, nil
}

func TestMockClassifyEOA(t *testing.T) {
	m := &mockRPC{chainID: 1, blockNumber: 100, balance: big.NewInt(42), code: nil}
	snap, err := m.ClassifyAccount(context.Background(), common.HexToAddress("0x2449ecef5012f0a0e153b278ef4fcc9625bc4c78"))
	if err != nil {
		t.Fatal(err)
	}
	if snap.Kind != domain.AccountLikelyEOA || snap.HasCode {
		t.Fatalf("snap = %+v", snap)
	}
	if snap.BalanceWei.String() != "42" {
		t.Fatalf("balance = %s", snap.BalanceWei.String())
	}
	if len(snap.Limitations) == 0 {
		t.Fatal("expected limitations")
	}
}

func TestMockClassifyContract(t *testing.T) {
	m := &mockRPC{chainID: 1, blockNumber: 100, balance: big.NewInt(0), code: []byte{0x60, 0x00}}
	snap, err := m.ClassifyAccount(context.Background(), common.HexToAddress("0xc5102fe9359fd9a28f877a67e36b0f050d81a3cc"))
	if err != nil {
		t.Fatal(err)
	}
	if snap.Kind != domain.AccountContract || !snap.HasCode {
		t.Fatalf("snap = %+v", snap)
	}
}

func TestMockChainMismatch(t *testing.T) {
	m := &mockRPC{failChain: true}
	err := m.VerifyConnection(context.Background())
	if err == nil {
		t.Fatal("expected mismatch")
	}
}

func TestParseAddressInvalid(t *testing.T) {
	_, err := ParseAddress("0x1234")
	if err == nil {
		t.Fatal("expected invalid")
	}
}

func TestDialRejectsNonMainnetConfig(t *testing.T) {
	_, err := Dial(context.Background(), "http://127.0.0.1:8545", 11155111)
	if err == nil {
		t.Fatal("expected rejection")
	}
}
