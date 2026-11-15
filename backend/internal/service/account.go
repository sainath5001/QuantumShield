package service

import (
	"context"
	"encoding/json"
	"math/big"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/quantumshield/qguard/backend/internal/blockchain"
	"github.com/quantumshield/qguard/backend/internal/domain"
	"github.com/quantumshield/qguard/backend/internal/exposure"
	"github.com/quantumshield/qguard/backend/internal/history"
)

// AccountService exposes account classification, history retrieval, and exposure analysis.
type AccountService struct {
	RPC      blockchain.RPCClient
	History  *history.Fetcher
	Exposure *exposure.Analyzer
	ChainID  *big.Int
	MaxTxs   int
}

// AccountView is a JSON-friendly account snapshot.
type AccountView struct {
	Address     string   `json:"address"`
	Kind        string   `json:"kind"`
	HasCode     bool     `json:"has_code"`
	CodeHash    string   `json:"code_hash,omitempty"`
	BalanceWei  string   `json:"balance_wei"`
	ChainID     int64    `json:"chain_id"`
	BlockNumber uint64   `json:"block_number"`
	Limitations []string `json:"limitations"`
}

// GetAccount classifies an address using RPC.
func (s *AccountService) GetAccount(ctx context.Context, rawAddress string) (AccountView, error) {
	addr, err := blockchain.ParseAddress(rawAddress)
	if err != nil {
		return AccountView{}, err
	}
	if s.RPC == nil {
		return AccountView{}, errRPCUnavailable
	}
	snap, err := s.RPC.ClassifyAccount(ctx, addr)
	if err != nil {
		return AccountView{}, err
	}
	return accountViewFromSnapshot(snap), nil
}

// GetHistory retrieves normal transaction history with explicit coverage.
func (s *AccountService) GetHistory(ctx context.Context, rawAddress string) (domain.TransactionHistory, error) {
	if _, err := blockchain.ParseAddress(rawAddress); err != nil {
		return domain.TransactionHistory{}, err
	}
	if s.History == nil {
		return domain.TransactionHistory{}, errHistoryUnavailable
	}
	return s.History.Fetch(ctx, rawAddress)
}

// GetExposure runs public-key recovery analysis for an address.
// It does not assign a risk score.
func (s *AccountService) GetExposure(ctx context.Context, rawAddress string) (domain.WalletExposureReport, error) {
	addr, err := blockchain.ParseAddress(rawAddress)
	if err != nil {
		return domain.WalletExposureReport{}, err
	}
	if s.RPC == nil {
		return domain.WalletExposureReport{}, errRPCUnavailable
	}
	if s.Exposure == nil {
		return domain.WalletExposureReport{}, errExposureUnavailable
	}
	if s.History == nil {
		return domain.WalletExposureReport{}, errHistoryUnavailable
	}

	snap, err := s.RPC.ClassifyAccount(ctx, addr)
	if err != nil {
		return domain.WalletExposureReport{}, err
	}

	hist := domain.TransactionHistory{
		Coverage: domain.HistoryCoverage{Status: domain.CoverageUnknown},
	}
	if snap.Kind != domain.AccountContract && snap.Kind != domain.AccountUnknown {
		hist, err = s.History.Fetch(ctx, rawAddress)
		if err != nil {
			return domain.WalletExposureReport{}, err
		}
	}

	return s.Exposure.Analyze(ctx, exposure.AnalyzeInput{
		Address:         addr,
		AccountKind:     snap.Kind,
		History:         hist,
		MaxTransactions: s.MaxTxs,
	})
}

func accountViewFromSnapshot(snap domain.AccountSnapshot) AccountView {
	bal := "0"
	if snap.BalanceWei != nil {
		bal = snap.BalanceWei.String()
	}
	view := AccountView{
		Address:     snap.Address.Hex(),
		Kind:        string(snap.Kind),
		HasCode:     snap.HasCode,
		BalanceWei:  bal,
		ChainID:     snap.ChainID,
		BlockNumber: snap.BlockNumber,
		Limitations: snap.Limitations,
	}
	if snap.HasCode {
		view.CodeHash = snap.CodeHash.Hex()
	}
	return view
}

// MarshalTxHistoryJSON marshals history using decimal strings for big integers.
func MarshalTxHistoryJSON(h domain.TransactionHistory) ([]byte, error) {
	type txDTO struct {
		Hash             string  `json:"hash"`
		BlockNumber      *uint64 `json:"block_number,omitempty"`
		BlockHash        string  `json:"block_hash,omitempty"`
		TransactionIndex *uint64 `json:"transaction_index,omitempty"`
		From             string  `json:"from"`
		To               string  `json:"to,omitempty"`
		Nonce            uint64  `json:"nonce"`
		ValueWei         string  `json:"value_wei"`
		Input            string  `json:"input"`
		Type             *uint8  `json:"type,omitempty"`
		ChainID          string  `json:"chain_id,omitempty"`
		Gas              uint64  `json:"gas"`
		GasPriceWei      string  `json:"gas_price_wei,omitempty"`
		GasUsed          *uint64 `json:"gas_used,omitempty"`
		StatusOK         *bool   `json:"status_ok,omitempty"`
		Timestamp        string  `json:"timestamp,omitempty"`
		V                string  `json:"v,omitempty"`
		R                string  `json:"r,omitempty"`
		S                string  `json:"s,omitempty"`
		Source           string  `json:"source"`
	}
	type dto struct {
		Address      string                 `json:"address"`
		Transactions []txDTO                `json:"transactions"`
		Coverage     domain.HistoryCoverage `json:"coverage"`
	}
	out := dto{Address: h.Address, Coverage: h.Coverage, Transactions: make([]txDTO, 0, len(h.Transactions))}
	for _, t := range h.Transactions {
		item := txDTO{
			Hash:             t.Hash.Hex(),
			BlockNumber:      t.BlockNumber,
			TransactionIndex: t.TxIndex,
			From:             t.From.Hex(),
			Nonce:            t.Nonce,
			ValueWei:         bigString(t.ValueWei),
			Input:            hexutil.Encode(t.Input),
			Type:             t.Type,
			Gas:              t.Gas,
			GasUsed:          t.GasUsed,
			StatusOK:         t.StatusOK,
			Source:           string(t.Source),
		}
		if t.BlockHash != nil {
			item.BlockHash = t.BlockHash.Hex()
		}
		if t.To != nil {
			item.To = t.To.Hex()
		}
		if t.ChainID != nil {
			item.ChainID = t.ChainID.String()
		}
		if t.GasPriceWei != nil {
			item.GasPriceWei = t.GasPriceWei.String()
		}
		if t.Timestamp != nil {
			item.Timestamp = t.Timestamp.UTC().Format("2006-01-02T15:04:05Z")
		}
		if t.V != nil {
			item.V = t.V.String()
		}
		if t.R != nil {
			item.R = t.R.String()
		}
		if t.S != nil {
			item.S = t.S.String()
		}
		out.Transactions = append(out.Transactions, item)
	}
	return json.Marshal(out)
}

func bigString(v *big.Int) string {
	if v == nil {
		return "0"
	}
	return v.String()
}

type staticError string

func (e staticError) Error() string { return string(e) }

const (
	errRPCUnavailable      staticError = "ethereum RPC client is not configured"
	errHistoryUnavailable  staticError = "history client is not configured"
	errExposureUnavailable staticError = "exposure analyzer is not configured"
)
