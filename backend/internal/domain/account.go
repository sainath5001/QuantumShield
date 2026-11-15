// Package domain holds shared domain types for Q-Guard analysis.
package domain

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// AccountKind is an initial classification signal, not absolute proof of wallet type.
type AccountKind string

const (
	AccountLikelyEOA AccountKind = "likely_eoa"
	AccountContract  AccountKind = "contract_account"
	AccountUnknown   AccountKind = "unknown"
)

// AccountSnapshot is RPC-derived account state at a point in time.
type AccountSnapshot struct {
	Address     common.Address `json:"address"`
	Kind        AccountKind    `json:"kind"`
	HasCode     bool           `json:"has_code"`
	CodeHash    common.Hash    `json:"code_hash,omitempty"`
	BalanceWei  *big.Int       `json:"balance_wei"`
	ChainID     int64          `json:"chain_id"`
	BlockNumber uint64         `json:"block_number"`
	Limitations []string       `json:"limitations"`
}

// AccountClassificationLimitations are always attached to classification results.
var AccountClassificationLimitations = []string{
	"Classification uses current eth_getCode as an initial signal, not absolute proof of wallet type.",
	"Account-abstraction smart accounts, proxies, and upgradeable contracts may look like contracts while signing authority resides elsewhere.",
	"Historical behavior (for example an EOA that later deployed code via CREATE) is not reconstructed in this milestone.",
	"Absence of code does not prove a single secp256k1 key controls the address forever.",
}
