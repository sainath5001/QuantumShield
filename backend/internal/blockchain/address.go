package blockchain

import (
	"strings"

	"github.com/ethereum/go-ethereum/common"
)

// IsHexAddress reports whether s is a 20-byte hex address with a 0x prefix.
// Q-Guard API inputs require the 0x prefix even though go-ethereum also accepts bare hex.
// This helper does not perform RPC lookups or scanning.
func IsHexAddress(s string) bool {
	if !strings.HasPrefix(s, "0x") && !strings.HasPrefix(s, "0X") {
		return false
	}
	return common.IsHexAddress(s)
}
