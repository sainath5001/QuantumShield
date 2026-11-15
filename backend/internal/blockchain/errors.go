package blockchain

import "fmt"

// ErrChainMismatch is returned when the RPC node chain ID does not match config.
type ErrChainMismatch struct {
	Configured int64
	Observed   int64
}

func (e ErrChainMismatch) Error() string {
	return fmt.Sprintf("ethereum chain ID mismatch: configured %d, RPC reported %d; refusing to continue", e.Configured, e.Observed)
}

// ErrInvalidAddress is returned for malformed addresses.
type ErrInvalidAddress struct {
	Address string
}

func (e ErrInvalidAddress) Error() string {
	return fmt.Sprintf("invalid Ethereum address: %q", e.Address)
}
