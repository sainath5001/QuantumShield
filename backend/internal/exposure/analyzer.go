// Package exposure aggregates verified public-key exposure evidence from transaction history.
package exposure

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/quantumshield/qguard/backend/internal/cryptography"
	"github.com/quantumshield/qguard/backend/internal/domain"
)

// TransactionProvider loads a full signed transaction by hash (typically Ethereum RPC).
type TransactionProvider interface {
	TransactionByHash(ctx context.Context, hash common.Hash) (*types.Transaction, bool, error)
}

// Analyzer performs wallet-level public-key exposure analysis.
type Analyzer struct {
	Txs     TransactionProvider
	ChainID *big.Int
	Now     func() time.Time
}

// AnalyzeInput is the wallet analysis request.
type AnalyzeInput struct {
	Address         common.Address
	AccountKind     domain.AccountKind
	History         domain.TransactionHistory
	MaxTransactions int // 0 = all sender txs in history
}

// Analyze examines sender transactions for recoverable secp256k1 public keys.
// Contract accounts are not analyzed with EOA recovery rules.
func (a *Analyzer) Analyze(ctx context.Context, in AnalyzeInput) (domain.WalletExposureReport, error) {
	now := time.Now().UTC()
	if a.Now != nil {
		now = a.Now().UTC()
	}
	if a.ChainID == nil || a.ChainID.Sign() <= 0 {
		return domain.WalletExposureReport{}, fmt.Errorf("chain ID is required")
	}
	if a.Txs == nil {
		return domain.WalletExposureReport{}, fmt.Errorf("transaction provider is required")
	}

	report := domain.WalletExposureReport{
		Address:               in.Address,
		AccountKind:           in.AccountKind,
		HistoryCoverageStatus: in.History.Coverage.Status,
		Evidence:              []domain.ExposureEvidence{},
		Limitations:           append([]string{}, domain.ExposureAnalysisLimitations...),
		AnalyzedAt:            now,
	}

	switch in.AccountKind {
	case domain.AccountContract:
		report.Finding = domain.FindingRequiresManualReview
		report.ConfirmedExposure = false
		report.Limitations = append(report.Limitations,
			"Address currently has contract code; EOA secp256k1 recovery rules are not applied as if this were an ordinary EOA.",
			"Owner keys and signature schemes may be defined by custom contract logic (multisig, account abstraction, EIP-1271, etc.).",
			"Do not infer quantum resistance or vulnerability without examining the contract's verification mechanism.",
		)
		return report, nil
	case domain.AccountUnknown:
		report.Finding = domain.FindingInsufficientEvidence
		report.Limitations = append(report.Limitations, "Account kind is unknown; exposure analysis was not performed.")
		return report, nil
	}

	// Likely EOA path.
	senderTxs := make([]domain.NormalizedTransaction, 0)
	for _, tx := range in.History.Transactions {
		if tx.From == in.Address {
			senderTxs = append(senderTxs, tx)
		}
	}
	report.SenderTxCount = len(senderTxs)

	limit := in.MaxTransactions
	if limit <= 0 || limit > len(senderTxs) {
		limit = len(senderTxs)
	}

	fingerprints := map[string]struct{}{}
	for i := 0; i < limit; i++ {
		if err := ctx.Err(); err != nil {
			return report, err
		}
		meta := senderTxs[i]
		report.AnalyzedTxCount++

		full, ok, err := a.Txs.TransactionByHash(ctx, meta.Hash)
		if err != nil {
			ev := domain.ExposureEvidence{
				WalletAddress:      in.Address,
				ExpectedSender:     in.Address,
				TransactionHash:    meta.Hash,
				BlockNumber:        meta.BlockNumber,
				TransactionType:    meta.Type,
				Curve:              "secp256k1",
				SignatureScheme:    "ecdsa",
				RecoveryStatus:     domain.RecoveryFetchFailed,
				SenderVerification: domain.SenderNotChecked,
				EvidenceSource:     "ethereum_rpc_transaction",
				Detail:             fmt.Sprintf("failed to fetch signed transaction: %v", err),
				AnalyzedAt:         now,
			}
			report.Evidence = append(report.Evidence, ev)
			report.FailedCount++
			continue
		}
		if !ok || full == nil {
			ev := domain.ExposureEvidence{
				WalletAddress:      in.Address,
				ExpectedSender:     in.Address,
				TransactionHash:    meta.Hash,
				BlockNumber:        meta.BlockNumber,
				Curve:              "secp256k1",
				SignatureScheme:    "ecdsa",
				RecoveryStatus:     domain.RecoveryFetchFailed,
				SenderVerification: domain.SenderNotChecked,
				EvidenceSource:     "ethereum_rpc_transaction",
				Detail:             "transaction not found on RPC",
				AnalyzedAt:         now,
			}
			report.Evidence = append(report.Evidence, ev)
			report.FailedCount++
			continue
		}

		ev := cryptography.RecoverPublicKey(full, cryptography.RecoverOptions{
			ChainID:        a.ChainID,
			ExpectedSender: in.Address,
			BlockNumber:    meta.BlockNumber,
			Now:            func() time.Time { return now },
		})
		report.Evidence = append(report.Evidence, ev)

		switch ev.RecoveryStatus {
		case domain.RecoveryConfirmed:
			report.ConfirmedCount++
			if ev.PublicKeyFingerprint != "" {
				fingerprints[ev.PublicKeyFingerprint] = struct{}{}
			}
		case domain.RecoveryUnsupportedType:
			report.UnsupportedCount++
		case domain.RecoverySenderMismatch:
			report.SenderMismatchCount++
		default:
			report.FailedCount++
		}
	}

	report.DistinctPublicKeys = len(fingerprints)
	report.ConfirmedExposure = report.ConfirmedCount > 0

	switch {
	case report.ConfirmedCount > 0:
		report.Finding = domain.FindingPublicKeyExposed
		if report.DistinctPublicKeys > 1 {
			report.Limitations = append(report.Limitations,
				"Multiple distinct recovered public keys verified to the same address — unexpected for a single EOA; treat as requiring investigation.")
		}
	case in.History.Coverage.Status == domain.CoverageCompleteWithinScope && report.SenderTxCount == 0:
		report.Finding = domain.FindingNoExposureObserved
		report.Limitations = append(report.Limitations,
			"No sender transactions were present in the examined history scope.")
	case in.History.Coverage.Status == domain.CoverageCompleteWithinScope &&
		report.AnalyzedTxCount > 0 &&
		report.ConfirmedCount == 0 &&
		report.UnsupportedCount == 0 &&
		report.FailedCount == 0 &&
		report.SenderMismatchCount == 0:
		report.Finding = domain.FindingNoExposureObserved
	case in.History.Coverage.Status == domain.CoveragePartial ||
		in.History.Coverage.Status == domain.CoverageRateLimited ||
		in.History.Coverage.Status == domain.CoverageFailed ||
		in.History.Coverage.Status == domain.CoverageUnknown ||
		report.FailedCount > 0 ||
		report.UnsupportedCount > 0:
		report.Finding = domain.FindingInsufficientEvidence
	default:
		report.Finding = domain.FindingNoExposureObserved
	}

	if report.Finding == domain.FindingNoExposureObserved {
		report.Limitations = append(report.Limitations,
			"No confirmed public-key exposure was observed in scope; this is not a quantum-safety certification.")
	}

	return report, nil
}
