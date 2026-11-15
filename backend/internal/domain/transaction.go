package domain

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

// TxSource identifies where normalized transaction fields originated.
type TxSource string

const (
	TxSourceEtherscan TxSource = "etherscan"
	TxSourceRPC       TxSource = "rpc"
)

// NormalizedTransaction is a chain-normalized view of a transaction for analysis.
// Large numeric fields use big.Int / uint64 — never float64.
type NormalizedTransaction struct {
	Hash        common.Hash     `json:"hash"`
	BlockNumber *uint64         `json:"block_number,omitempty"`
	BlockHash   *common.Hash    `json:"block_hash,omitempty"`
	TxIndex     *uint64         `json:"transaction_index,omitempty"`
	From        common.Address  `json:"from"`
	To          *common.Address `json:"to,omitempty"` // nil for contract creation
	Nonce       uint64          `json:"nonce"`
	ValueWei    *big.Int        `json:"value_wei"`
	Input       []byte          `json:"input"`
	Type        *uint8          `json:"type,omitempty"` // nil if indexer did not provide type
	ChainID     *big.Int        `json:"chain_id,omitempty"`
	Gas         uint64          `json:"gas"`
	GasPriceWei *big.Int        `json:"gas_price_wei,omitempty"`
	GasUsed     *uint64         `json:"gas_used,omitempty"`
	StatusOK    *bool           `json:"status_ok,omitempty"`
	Timestamp   *time.Time      `json:"timestamp,omitempty"`

	// Signature fields when available from the source. Etherscan v2 txlist often omits these.
	V *big.Int `json:"v,omitempty"`
	R *big.Int `json:"r,omitempty"`
	S *big.Int `json:"s,omitempty"`

	Source TxSource `json:"source"`
}

// IsSender reports whether addr is the transaction sender.
func (t NormalizedTransaction) IsSender(addr common.Address) bool {
	return t.From == addr
}
