// Package history orchestrates transaction-history retrieval and coverage reporting.
package history

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/quantumshield/qguard/backend/internal/blockchain"
	"github.com/quantumshield/qguard/backend/internal/domain"
)

// HistorySource retrieves normalized normal transactions for an address.
type HistorySource interface {
	FetchNormalTransactions(ctx context.Context, address string) (domain.TransactionHistory, error)
}

// BlockHashVerifier looks up canonical block hashes for consistency checks.
type BlockHashVerifier interface {
	BlockHashByNumber(ctx context.Context, number uint64) (common.Hash, error)
}

// Fetcher combines indexer history with optional RPC block-hash verification.
type Fetcher struct {
	Source        HistorySource
	Verifier      BlockHashVerifier
	MaxHashChecks int // max distinct blocks to verify; 0 disables
}

// Fetch returns transaction history and updates coverage with block-hash checks when configured.
func (f *Fetcher) Fetch(ctx context.Context, address string) (domain.TransactionHistory, error) {
	if _, err := blockchain.ParseAddress(address); err != nil {
		return domain.TransactionHistory{}, err
	}
	hist, err := f.Source.FetchNormalTransactions(ctx, address)
	if err != nil {
		return hist, err
	}
	if f.Verifier == nil || f.MaxHashChecks <= 0 {
		return hist, nil
	}
	hist.Coverage.BlockHashChecks = f.verifyBlockHashes(ctx, hist.Transactions)
	if hist.Coverage.BlockHashChecks.Mismatched > 0 {
		hist.Coverage.Notes = append(hist.Coverage.Notes,
			"One or more indexer block hashes did not match the RPC canonical hash; data may be inconsistent or reorganized.")
		if hist.Coverage.Status == domain.CoverageCompleteWithinScope {
			hist.Coverage.Status = domain.CoveragePartial
		}
	}
	return hist, nil
}

func (f *Fetcher) verifyBlockHashes(ctx context.Context, txs []domain.NormalizedTransaction) domain.BlockHashCheckSummary {
	summary := domain.BlockHashCheckSummary{}
	seenBlocks := make(map[uint64]common.Hash)
	for _, tx := range txs {
		if tx.BlockNumber == nil || tx.BlockHash == nil {
			continue
		}
		if _, exists := seenBlocks[*tx.BlockNumber]; exists {
			continue
		}
		seenBlocks[*tx.BlockNumber] = *tx.BlockHash
		if len(seenBlocks) >= f.MaxHashChecks {
			break
		}
	}
	for number, indexedHash := range seenBlocks {
		if err := ctx.Err(); err != nil {
			summary.Unavailable++
			continue
		}
		summary.Attempted++
		rpcHash, err := f.Verifier.BlockHashByNumber(ctx, number)
		if err != nil {
			summary.Unavailable++
			continue
		}
		if rpcHash == indexedHash {
			summary.Matched++
		} else {
			summary.Mismatched++
		}
	}
	return summary
}
