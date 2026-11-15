// Package etherscan implements the Etherscan API v2 account history client.
package etherscan

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/quantumshield/qguard/backend/internal/domain"
)

const (
	providerName = "etherscan"
	defaultSort  = "asc"
)

// HTTPDoer abstracts http.Client for tests.
type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Config configures the Etherscan v2 client.
type Config struct {
	BaseURL     string
	APIKey      string
	ChainID     int64
	PageSize    int
	MaxPages    int
	MaxTxs      int
	HTTPClient  HTTPDoer
	RateLimit   RateLimiter
	MaxRetries  int
	BaseBackoff time.Duration
	Clock       func() time.Time
}

// Client talks to Etherscan API v2 (`/v2/api` with chainid).
type Client struct {
	cfg Config
}

// NewClient constructs an Etherscan client.
func NewClient(cfg Config) (*Client, error) {
	if cfg.BaseURL == "" {
		return nil, fmt.Errorf("etherscan base URL is required")
	}
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("etherscan API key is required")
	}
	if cfg.ChainID != 1 {
		return nil, fmt.Errorf("MVP supports Ethereum Mainnet only (chain id 1); got %d", cfg.ChainID)
	}
	if cfg.PageSize < 1 {
		cfg.PageSize = 100
	}
	if cfg.MaxPages < 1 {
		cfg.MaxPages = 10
	}
	if cfg.MaxTxs < 1 {
		cfg.MaxTxs = 500
	}
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = &http.Client{Timeout: 15 * time.Second}
	}
	if cfg.RateLimit == nil {
		cfg.RateLimit = NewTokenBucket(4, 8)
	}
	if cfg.MaxRetries < 0 {
		cfg.MaxRetries = 0
	}
	if cfg.BaseBackoff <= 0 {
		cfg.BaseBackoff = 200 * time.Millisecond
	}
	if cfg.Clock == nil {
		cfg.Clock = time.Now
	}
	return &Client{cfg: cfg}, nil
}

type apiResponse struct {
	Status  string          `json:"status"`
	Message string          `json:"message"`
	Result  json.RawMessage `json:"result"`
}

type rawNormalTx struct {
	BlockNumber      string `json:"blockNumber"`
	BlockHash        string `json:"blockHash"`
	TimeStamp        string `json:"timeStamp"`
	Hash             string `json:"hash"`
	Nonce            string `json:"nonce"`
	TransactionIndex string `json:"transactionIndex"`
	From             string `json:"from"`
	To               string `json:"to"`
	Value            string `json:"value"`
	Gas              string `json:"gas"`
	GasPrice         string `json:"gasPrice"`
	Input            string `json:"input"`
	GasUsed          string `json:"gasUsed"`
	IsError          string `json:"isError"`
	TxReceiptStatus  string `json:"txreceipt_status"`
	Type             string `json:"type"`
	ChainID          string `json:"chainId"`
	V                string `json:"v"`
	R                string `json:"r"`
	S                string `json:"s"`
}

// FetchNormalTransactions paginates account txlist and returns normalized history + coverage.
// Token transfers are not fetched: pubkey exposure analysis needs sender signatures on normal txs.
// Internal transactions are not fetched for classification (eth_getCode is used instead).
func (c *Client) FetchNormalTransactions(ctx context.Context, address string) (domain.TransactionHistory, error) {
	if !common.IsHexAddress(address) || !(strings.HasPrefix(address, "0x") || strings.HasPrefix(address, "0X")) {
		return domain.TransactionHistory{}, newAPIError(CodeInvalidAddr, "address must be 0x-prefixed 20-byte hex", false)
	}

	coverage := domain.HistoryCoverage{
		Status:    domain.CoverageUnknown,
		Provider:  providerName,
		ChainID:   c.cfg.ChainID,
		PageSize:  c.cfg.PageSize,
		FetchedAt: c.cfg.Clock().UTC(),
		Notes: []string{
			"History is bounded by Etherscan indexed results, page/tx caps, and API plan limits — not a proof of complete chain history.",
			"Etherscan v2 txlist may omit transaction type and signature (v/r/s) fields; those are left empty until RPC enrichment in a later milestone.",
			"Token-transfer and internal-transaction endpoints are not queried in this milestone.",
		},
	}

	seen := make(map[common.Hash]struct{})
	out := make([]domain.NormalizedTransaction, 0, c.cfg.PageSize)
	var earliest, latest *uint64

	for page := 1; page <= c.cfg.MaxPages; page++ {
		if err := ctx.Err(); err != nil {
			coverage.Status = domain.CoverageFailed
			coverage.APIErrors = append(coverage.APIErrors, domain.HistoryAPIError{
				Page: page, Code: string(CodeCanceled), Message: err.Error(),
			})
			return domain.TransactionHistory{Address: address, Transactions: out, Coverage: finalizeCoverage(coverage, out, earliest, latest)}, err
		}

		if len(out) >= c.cfg.MaxTxs {
			coverage.TruncatedByTxCap = true
			coverage.PaginationComplete = false
			coverage.Status = domain.CoveragePartial
			coverage.Notes = append(coverage.Notes, "Stopped because QGUARD_SCAN_MAX_TXS was reached.")
			break
		}

		coverage.PagesRequested++
		rawPage, err := c.fetchTxListPage(ctx, address, page)
		if err != nil {
			ae, ok := err.(*APIError)
			if !ok {
				ae = newAPIError(CodeNetwork, err.Error(), true)
			}
			coverage.APIErrors = append(coverage.APIErrors, domain.HistoryAPIError{
				Page: page, Code: string(ae.Code), Message: ae.Message,
			})
			switch {
			case ae.Code == CodeRateLimited:
				coverage.Status = domain.CoverageRateLimited
				coverage.PaginationComplete = false
			case len(out) == 0:
				coverage.Status = domain.CoverageFailed
				coverage.PaginationComplete = false
			default:
				coverage.Status = domain.CoveragePartial
				coverage.PaginationComplete = false
			}
			return domain.TransactionHistory{Address: address, Transactions: out, Coverage: finalizeCoverage(coverage, out, earliest, latest)}, nil
		}

		if rawPage == nil {
			// Empty result set.
			coverage.PaginationComplete = true
			if len(out) == 0 {
				coverage.Status = domain.CoverageCompleteWithinScope
				coverage.Notes = append(coverage.Notes, "Provider returned no normal transactions for this address within the queried scope.")
			} else {
				coverage.Status = domain.CoverageCompleteWithinScope
			}
			break
		}

		if len(rawPage) == 0 {
			coverage.PaginationComplete = true
			coverage.Status = domain.CoverageCompleteWithinScope
			break
		}

		for _, raw := range rawPage {
			tx, err := normalizeRawTx(raw, c.cfg.ChainID)
			if err != nil {
				coverage.UnsupportedNotes = append(coverage.UnsupportedNotes, err.Error())
				continue
			}
			if _, exists := seen[tx.Hash]; exists {
				coverage.DuplicatesDropped++
				continue
			}
			seen[tx.Hash] = struct{}{}
			out = append(out, tx)
			if tx.BlockNumber != nil {
				bn := *tx.BlockNumber
				if earliest == nil || bn < *earliest {
					v := bn
					earliest = &v
				}
				if latest == nil || bn > *latest {
					v := bn
					latest = &v
				}
			}
			if len(out) >= c.cfg.MaxTxs {
				coverage.TruncatedByTxCap = true
				break
			}
		}

		if len(rawPage) < c.cfg.PageSize {
			coverage.PaginationComplete = true
			if coverage.TruncatedByTxCap {
				coverage.Status = domain.CoveragePartial
			} else {
				coverage.Status = domain.CoverageCompleteWithinScope
			}
			break
		}

		if page == c.cfg.MaxPages {
			coverage.TruncatedByPageCap = true
			coverage.PaginationComplete = false
			coverage.Status = domain.CoveragePartial
			coverage.Notes = append(coverage.Notes, "Stopped because QGUARD_ETHERSCAN_MAX_PAGES was reached; more pages may exist.")
		}
	}

	if coverage.Status == domain.CoverageUnknown {
		if coverage.TruncatedByPageCap || coverage.TruncatedByTxCap {
			coverage.Status = domain.CoveragePartial
		} else if coverage.PaginationComplete {
			coverage.Status = domain.CoverageCompleteWithinScope
		}
	}

	return domain.TransactionHistory{
		Address:      common.HexToAddress(address).Hex(),
		Transactions: out,
		Coverage:     finalizeCoverage(coverage, out, earliest, latest),
	}, nil
}

func finalizeCoverage(c domain.HistoryCoverage, txs []domain.NormalizedTransaction, earliest, latest *uint64) domain.HistoryCoverage {
	c.TransactionsRetrieved = len(txs) + c.DuplicatesDropped
	c.UniqueTransactions = len(txs)
	c.EarliestBlock = earliest
	c.LatestBlock = latest
	if c.Status == domain.CoverageUnknown {
		c.Status = domain.CoverageUnknown
		c.Notes = append(c.Notes, "Coverage status could not be determined confidently.")
	}
	return c
}

func (c *Client) fetchTxListPage(ctx context.Context, address string, page int) ([]rawNormalTx, error) {
	var lastErr error
	for attempt := 0; attempt <= c.cfg.MaxRetries; attempt++ {
		if err := ctx.Err(); err != nil {
			return nil, newAPIError(CodeCanceled, err.Error(), false)
		}
		if err := c.cfg.RateLimit.Wait(ctx); err != nil {
			return nil, newAPIError(CodeCanceled, err.Error(), false)
		}

		rows, empty, err := c.doTxList(ctx, address, page)
		if err == nil {
			if empty {
				return nil, nil
			}
			return rows, nil
		}
		lastErr = err
		ae, ok := err.(*APIError)
		if !ok || !ae.Retryable || attempt == c.cfg.MaxRetries {
			return nil, err
		}
		backoff := c.cfg.BaseBackoff * time.Duration(1<<attempt)
		timer := time.NewTimer(backoff)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, newAPIError(CodeCanceled, ctx.Err().Error(), false)
		case <-timer.C:
		}
	}
	return nil, lastErr
}

func (c *Client) doTxList(ctx context.Context, address string, page int) ([]rawNormalTx, bool, error) {
	u, err := url.Parse(c.cfg.BaseURL)
	if err != nil {
		return nil, false, newAPIError(CodeAPI, "invalid base URL", false)
	}
	q := u.Query()
	q.Set("chainid", strconv.FormatInt(c.cfg.ChainID, 10))
	q.Set("module", "account")
	q.Set("action", "txlist")
	q.Set("address", address)
	q.Set("startblock", "0")
	q.Set("endblock", "99999999")
	q.Set("page", strconv.Itoa(page))
	q.Set("offset", strconv.Itoa(c.cfg.PageSize))
	q.Set("sort", defaultSort)
	q.Set("apikey", c.cfg.APIKey)
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, false, newAPIError(CodeNetwork, err.Error(), true)
	}

	resp, err := c.cfg.HTTPClient.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return nil, false, newAPIError(CodeCanceled, ctx.Err().Error(), false)
		}
		return nil, false, newAPIError(CodeNetwork, err.Error(), true)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, false, newAPIError(CodeNetwork, err.Error(), true)
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, false, newAPIError(CodeRateLimited, "HTTP 429 too many requests", true)
	}
	if resp.StatusCode >= 500 {
		return nil, false, newAPIError(CodeHTTP, fmt.Sprintf("HTTP %d", resp.StatusCode), true)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, false, newAPIError(CodeHTTP, fmt.Sprintf("HTTP %d", resp.StatusCode), false)
	}

	var parsed apiResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, false, newAPIError(CodeUnsupported, "response JSON could not be parsed", false)
	}

	msg := strings.ToLower(parsed.Message)
	resultStr := strings.TrimSpace(string(parsed.Result))

	// Empty history: status 0 + "No transactions found"
	if parsed.Status == "0" && (strings.Contains(msg, "no transaction") || strings.Contains(strings.ToLower(resultStr), "no transaction")) {
		return nil, true, nil
	}

	if parsed.Status != "1" {
		return nil, false, interpretAPIFailure(parsed.Message, resultStr)
	}

	var rows []rawNormalTx
	if err := json.Unmarshal(parsed.Result, &rows); err != nil {
		// Sometimes result is a string even on odd success paths.
		return nil, false, newAPIError(CodeUnsupported, "txlist result was not a transaction array", false)
	}
	return rows, false, nil
}

func interpretAPIFailure(message, result string) error {
	combined := strings.ToLower(message + " " + result)
	switch {
	case strings.Contains(combined, "rate limit"):
		return newAPIError(CodeRateLimited, firstNonEmpty(result, message), true)
	case strings.Contains(combined, "invalid api key"):
		return newAPIError(CodeInvalidAPIKey, firstNonEmpty(result, message), false)
	case strings.Contains(combined, "invalid address"):
		return newAPIError(CodeInvalidAddr, firstNonEmpty(result, message), false)
	case strings.Contains(combined, "timeout"):
		return newAPIError(CodeTimeout, firstNonEmpty(result, message), true)
	case strings.Contains(combined, "unsupported chainid"), strings.Contains(combined, "missing or unsupported chain"):
		return newAPIError(CodeUnsupported, firstNonEmpty(result, message), false)
	default:
		return newAPIError(CodeAPI, firstNonEmpty(result, message, "unknown etherscan error"), false)
	}
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		v = strings.TrimSpace(v)
		if v != "" && v != `""` && v != "null" {
			// result may be JSON-quoted
			if strings.HasPrefix(v, `"`) && strings.HasSuffix(v, `"`) {
				if u, err := strconv.Unquote(v); err == nil {
					return u
				}
			}
			return v
		}
	}
	return "unknown error"
}

func normalizeRawTx(raw rawNormalTx, configuredChain int64) (domain.NormalizedTransaction, error) {
	if raw.Hash == "" || !common.IsHexAddress(raw.From) {
		return domain.NormalizedTransaction{}, fmt.Errorf("unsupported tx record: missing hash or from")
	}
	hash := common.HexToHash(raw.Hash)
	from := common.HexToAddress(raw.From)

	tx := domain.NormalizedTransaction{
		Hash:     hash,
		From:     from,
		Nonce:    mustUint64(raw.Nonce),
		ValueWei: mustBigInt(raw.Value),
		Input:    common.FromHex(raw.Input),
		Gas:      mustUint64(raw.Gas),
		Source:   domain.TxSourceEtherscan,
	}
	if raw.To != "" && common.IsHexAddress(raw.To) {
		to := common.HexToAddress(raw.To)
		tx.To = &to
	}
	if bn, ok := parseUint64(raw.BlockNumber); ok {
		tx.BlockNumber = &bn
	}
	if raw.BlockHash != "" {
		bh := common.HexToHash(raw.BlockHash)
		tx.BlockHash = &bh
	}
	if idx, ok := parseUint64(raw.TransactionIndex); ok {
		tx.TxIndex = &idx
	}
	if gp := mustBigInt(raw.GasPrice); gp != nil {
		tx.GasPriceWei = gp
	}
	if gu, ok := parseUint64(raw.GasUsed); ok {
		tx.GasUsed = &gu
	}
	if ts, ok := parseUint64(raw.TimeStamp); ok {
		t := time.Unix(int64(ts), 0).UTC()
		tx.Timestamp = &t
	}
	if raw.IsError == "0" || raw.TxReceiptStatus == "1" {
		ok := true
		tx.StatusOK = &ok
	} else if raw.IsError == "1" || raw.TxReceiptStatus == "0" {
		ok := false
		tx.StatusOK = &ok
	}
	if raw.Type != "" {
		if v, err := strconv.ParseUint(strings.TrimPrefix(raw.Type, "0x"), 0, 8); err == nil {
			t := uint8(v)
			tx.Type = &t
		}
	}
	if raw.ChainID != "" {
		if v, ok := new(big.Int).SetString(raw.ChainID, 0); ok {
			tx.ChainID = v
		}
	} else {
		tx.ChainID = big.NewInt(configuredChain)
	}
	tx.V = parseBigOptional(raw.V)
	tx.R = parseBigOptional(raw.R)
	tx.S = parseBigOptional(raw.S)
	return tx, nil
}

func parseUint64(s string) (uint64, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, false
	}
	v, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, false
	}
	return v, true
}

func mustUint64(s string) uint64 {
	v, _ := parseUint64(s)
	return v
}

func mustBigInt(s string) *big.Int {
	s = strings.TrimSpace(s)
	if s == "" {
		return big.NewInt(0)
	}
	v, ok := new(big.Int).SetString(s, 10)
	if !ok {
		v, ok = new(big.Int).SetString(s, 0)
		if !ok {
			return big.NewInt(0)
		}
	}
	return v
}

func parseBigOptional(s string) *big.Int {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	v, ok := new(big.Int).SetString(s, 0)
	if !ok {
		return nil
	}
	return v
}
