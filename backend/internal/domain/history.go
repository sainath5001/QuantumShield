package domain

import "time"

// HistoryCoverageStatus describes how complete retrieved history is.
// HTTP 200 alone never implies complete coverage.
type HistoryCoverageStatus string

const (
	CoverageCompleteWithinScope HistoryCoverageStatus = "complete_within_scope"
	CoveragePartial             HistoryCoverageStatus = "partial"
	CoverageRateLimited         HistoryCoverageStatus = "rate_limited"
	CoverageFailed              HistoryCoverageStatus = "failed"
	CoverageUnsupported         HistoryCoverageStatus = "unsupported"
	CoverageUnknown             HistoryCoverageStatus = "unknown"
)

// HistoryCoverage records pagination and provider outcomes for a history fetch.
type HistoryCoverage struct {
	Status                HistoryCoverageStatus `json:"status"`
	Provider              string                `json:"provider"`
	ChainID               int64                 `json:"chain_id"`
	PagesRequested        int                   `json:"pages_requested"`
	PageSize              int                   `json:"page_size"`
	TransactionsRetrieved int                   `json:"transactions_retrieved"`
	UniqueTransactions    int                   `json:"unique_transactions"`
	DuplicatesDropped     int                   `json:"duplicates_dropped"`
	PaginationComplete    bool                  `json:"pagination_complete"`
	TruncatedByPageCap    bool                  `json:"truncated_by_page_cap"`
	TruncatedByTxCap      bool                  `json:"truncated_by_tx_cap"`
	StartBlock            *uint64               `json:"start_block,omitempty"`
	EndBlock              *uint64               `json:"end_block,omitempty"`
	EarliestBlock         *uint64               `json:"earliest_block,omitempty"`
	LatestBlock           *uint64               `json:"latest_block,omitempty"`
	APIErrors             []HistoryAPIError     `json:"api_errors,omitempty"`
	UnsupportedNotes      []string              `json:"unsupported_notes,omitempty"`
	BlockHashChecks       BlockHashCheckSummary `json:"block_hash_checks"`
	FetchedAt             time.Time             `json:"fetched_at"`
	Notes                 []string              `json:"notes,omitempty"`
}

// HistoryAPIError captures a provider error without leaking secrets.
type HistoryAPIError struct {
	Page    int    `json:"page,omitempty"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// BlockHashCheckSummary records optional RPC cross-checks of indexer block hashes.
type BlockHashCheckSummary struct {
	Attempted   int `json:"attempted"`
	Matched     int `json:"matched"`
	Mismatched  int `json:"mismatched"`
	Unavailable int `json:"unavailable"`
}

// TransactionHistory is the result of a history retrieval operation.
type TransactionHistory struct {
	Address      string                  `json:"address"`
	Transactions []NormalizedTransaction `json:"transactions"`
	Coverage     HistoryCoverage         `json:"coverage"`
}
