package blockchain

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/quantumshield/qguard/backend/internal/domain"
)

// MainnetChainID is the only chain supported in the MVP.
const MainnetChainID int64 = 1

// RPCClient is the Ethereum JSON-RPC surface used by Q-Guard.
type RPCClient interface {
	Close()
	VerifyConnection(ctx context.Context) error
	ChainID(ctx context.Context) (*big.Int, error)
	BlockNumber(ctx context.Context) (uint64, error)
	BalanceAt(ctx context.Context, addr common.Address) (*big.Int, error)
	CodeAt(ctx context.Context, addr common.Address) ([]byte, error)
	BlockHashByNumber(ctx context.Context, number uint64) (common.Hash, error)
	ClassifyAccount(ctx context.Context, addr common.Address) (domain.AccountSnapshot, error)
}

// EthClient wraps go-ethereum ethclient with mainnet chain enforcement.
type EthClient struct {
	inner         *ethclient.Client
	expectedChain int64
	verifiedChain bool
	observedChain int64
}

// Dial connects to an Ethereum JSON-RPC endpoint and verifies the chain ID.
func Dial(ctx context.Context, rpcURL string, expectedChainID int64) (*EthClient, error) {
	if expectedChainID != MainnetChainID {
		return nil, fmt.Errorf("MVP supports Ethereum Mainnet only (chain id %d); got configured chain id %d", MainnetChainID, expectedChainID)
	}
	inner, err := ethclient.DialContext(ctx, rpcURL)
	if err != nil {
		return nil, fmt.Errorf("dial ethereum RPC: %w", err)
	}
	c := &EthClient{inner: inner, expectedChain: expectedChainID}
	if err := c.VerifyConnection(ctx); err != nil {
		inner.Close()
		return nil, err
	}
	return c, nil
}

// NewEthClient wraps an existing ethclient (primarily for tests).
// Call VerifyConnection before relying on chain-sensitive methods.
func NewEthClient(inner *ethclient.Client, expectedChainID int64) *EthClient {
	return &EthClient{inner: inner, expectedChain: expectedChainID}
}

// Close releases the underlying RPC connection.
func (c *EthClient) Close() {
	if c != nil && c.inner != nil {
		c.inner.Close()
	}
}

// VerifyConnection checks RPC reachability and enforces expected chain ID.
func (c *EthClient) VerifyConnection(ctx context.Context) error {
	id, err := c.inner.ChainID(ctx)
	if err != nil {
		return fmt.Errorf("ethereum RPC chain id: %w", err)
	}
	observed := id.Int64()
	if observed != c.expectedChain {
		return ErrChainMismatch{Configured: c.expectedChain, Observed: observed}
	}
	if _, err := c.inner.BlockNumber(ctx); err != nil {
		return fmt.Errorf("ethereum RPC block number: %w", err)
	}
	c.verifiedChain = true
	c.observedChain = observed
	return nil
}

func (c *EthClient) ensureVerified(ctx context.Context) error {
	if c.verifiedChain {
		return nil
	}
	return c.VerifyConnection(ctx)
}

// ChainID returns the chain ID from the node after verification.
func (c *EthClient) ChainID(ctx context.Context) (*big.Int, error) {
	if err := c.ensureVerified(ctx); err != nil {
		return nil, err
	}
	return big.NewInt(c.observedChain), nil
}

// BlockNumber returns the latest block number.
func (c *EthClient) BlockNumber(ctx context.Context) (uint64, error) {
	if err := c.ensureVerified(ctx); err != nil {
		return 0, err
	}
	return c.inner.BlockNumber(ctx)
}

// BalanceAt returns the account balance at the latest block.
func (c *EthClient) BalanceAt(ctx context.Context, addr common.Address) (*big.Int, error) {
	if err := c.ensureVerified(ctx); err != nil {
		return nil, err
	}
	return c.inner.BalanceAt(ctx, addr, nil)
}

// CodeAt returns account code at the latest block.
func (c *EthClient) CodeAt(ctx context.Context, addr common.Address) ([]byte, error) {
	if err := c.ensureVerified(ctx); err != nil {
		return nil, err
	}
	return c.inner.CodeAt(ctx, addr, nil)
}

// BlockHashByNumber returns the canonical block hash for a number.
func (c *EthClient) BlockHashByNumber(ctx context.Context, number uint64) (common.Hash, error) {
	if err := c.ensureVerified(ctx); err != nil {
		return common.Hash{}, err
	}
	header, err := c.inner.HeaderByNumber(ctx, new(big.Int).SetUint64(number))
	if err != nil {
		return common.Hash{}, err
	}
	return header.Hash(), nil
}

// ClassifyAccount loads balance/code and returns an initial account kind signal.
func (c *EthClient) ClassifyAccount(ctx context.Context, addr common.Address) (domain.AccountSnapshot, error) {
	if err := c.ensureVerified(ctx); err != nil {
		return domain.AccountSnapshot{}, err
	}
	block, err := c.inner.BlockNumber(ctx)
	if err != nil {
		return domain.AccountSnapshot{}, fmt.Errorf("block number: %w", err)
	}
	bal, err := c.inner.BalanceAt(ctx, addr, nil)
	if err != nil {
		return domain.AccountSnapshot{}, fmt.Errorf("balance: %w", err)
	}
	code, err := c.inner.CodeAt(ctx, addr, nil)
	if err != nil {
		return domain.AccountSnapshot{}, fmt.Errorf("code: %w", err)
	}

	snap := domain.AccountSnapshot{
		Address:     addr,
		BalanceWei:  bal,
		ChainID:     c.expectedChain,
		BlockNumber: block,
		Limitations: append([]string{}, domain.AccountClassificationLimitations...),
	}
	if len(code) == 0 {
		snap.Kind = domain.AccountLikelyEOA
		snap.HasCode = false
	} else {
		snap.Kind = domain.AccountContract
		snap.HasCode = true
		snap.CodeHash = crypto.Keccak256Hash(code)
	}
	return snap, nil
}

// ParseAddress validates and parses a hex address with 0x prefix.
func ParseAddress(raw string) (common.Address, error) {
	if !IsHexAddress(raw) {
		return common.Address{}, ErrInvalidAddress{Address: raw}
	}
	return common.HexToAddress(raw), nil
}
