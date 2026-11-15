package etherscan

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/quantumshield/qguard/backend/internal/domain"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) Do(req *http.Request) (*http.Response, error) { return f(req) }

func jsonResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	}
}

func testClient(t *testing.T, doer HTTPDoer, pageSize, maxPages, maxTxs int) *Client {
	t.Helper()
	c, err := NewClient(Config{
		BaseURL:     "https://api.etherscan.io/v2/api",
		APIKey:      "test-key",
		ChainID:     1,
		PageSize:    pageSize,
		MaxPages:    maxPages,
		MaxTxs:      maxTxs,
		HTTPClient:  doer,
		RateLimit:   NopLimiter{},
		MaxRetries:  1,
		BaseBackoff: time.Millisecond,
		Clock:       func() time.Time { return time.Unix(1700000000, 0).UTC() },
	})
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestFetchNormalTransactionsSuccess(t *testing.T) {
	body := `{
	  "status":"1","message":"OK","result":[{
	    "blockNumber":"23467053",
	    "blockHash":"0xf5646226f819fbdd6b2f1cb99de6e5d17da3d876ec166e69f4e9736c8ebf7ab4",
	    "timeStamp":"1759129619",
	    "hash":"0xf9db905d77704596d3600816bc70201586cfeec13bcf576320e2f38d6ca851a0",
	    "nonce":"8",
	    "transactionIndex":"184",
	    "from":"0x2449ecef5012f0a0e153b278ef4fcc9625bc4c78",
	    "to":"0xc5102fe9359fd9a28f877a67e36b0f050d81a3cc",
	    "value":"0",
	    "gas":"73271",
	    "gasPrice":"238744402",
	    "input":"0x5c19a95c",
	    "gasUsed":"48847",
	    "txreceipt_status":"1",
	    "isError":"0"
	  }]
	}`
	client := testClient(t, roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Query().Get("chainid") != "1" {
			t.Fatalf("chainid = %s", r.URL.Query().Get("chainid"))
		}
		if r.URL.Query().Get("action") != "txlist" {
			t.Fatalf("action = %s", r.URL.Query().Get("action"))
		}
		return jsonResponse(200, body), nil
	}), 100, 10, 500)

	hist, err := client.FetchNormalTransactions(context.Background(), "0x2449ecef5012f0a0e153b278ef4fcc9625bc4c78")
	if err != nil {
		t.Fatal(err)
	}
	if hist.Coverage.Status != domain.CoverageCompleteWithinScope {
		t.Fatalf("coverage = %s", hist.Coverage.Status)
	}
	if !hist.Coverage.PaginationComplete {
		t.Fatal("expected pagination complete")
	}
	if len(hist.Transactions) != 1 {
		t.Fatalf("txs = %d", len(hist.Transactions))
	}
	if hist.Transactions[0].ValueWei.String() != "0" {
		t.Fatalf("value = %s", hist.Transactions[0].ValueWei.String())
	}
	if hist.Transactions[0].Type != nil {
		t.Fatal("fixture omits type; expected nil type (no invented fields)")
	}
}

func TestFetchEmptyHistory(t *testing.T) {
	body := `{"status":"0","message":"No transactions found","result":"No transactions found"}`
	client := testClient(t, roundTripFunc(func(*http.Request) (*http.Response, error) {
		return jsonResponse(200, body), nil
	}), 100, 10, 500)

	hist, err := client.FetchNormalTransactions(context.Background(), "0x0000000000000000000000000000000000000001")
	if err != nil {
		t.Fatal(err)
	}
	if hist.Coverage.Status != domain.CoverageCompleteWithinScope {
		t.Fatalf("status = %s", hist.Coverage.Status)
	}
	if len(hist.Transactions) != 0 {
		t.Fatalf("expected empty history")
	}
}

func TestFetchPaginationAndPartialCap(t *testing.T) {
	calls := 0
	mk := func(hash string) string {
		return `{"blockNumber":"1","blockHash":"0xabc","timeStamp":"1","hash":"` + hash + `","nonce":"0","transactionIndex":"0","from":"0x2449ecef5012f0a0e153b278ef4fcc9625bc4c78","to":"0xc5102fe9359fd9a28f877a67e36b0f050d81a3cc","value":"1","gas":"21000","gasPrice":"1","input":"0x","gasUsed":"21000","isError":"0","txreceipt_status":"1"}`
	}
	client := testClient(t, roundTripFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		page := r.URL.Query().Get("page")
		var result string
		switch page {
		case "1":
			result = "[" + mk("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa") + "," + mk("0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb") + "]"
		default:
			result = "[" + mk("0xcccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc") + "]"
		}
		return jsonResponse(200, `{"status":"1","message":"OK","result":`+result+`}`), nil
	}), 2, 1, 500)

	hist, err := client.FetchNormalTransactions(context.Background(), "0x2449ecef5012f0a0e153b278ef4fcc9625bc4c78")
	if err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatalf("calls = %d", calls)
	}
	if hist.Coverage.Status != domain.CoveragePartial {
		t.Fatalf("status = %s", hist.Coverage.Status)
	}
	if !hist.Coverage.TruncatedByPageCap {
		t.Fatal("expected truncated by page cap")
	}
	if hist.Coverage.PaginationComplete {
		t.Fatal("pagination should not be complete")
	}
}

func TestFetchRateLimit(t *testing.T) {
	client := testClient(t, roundTripFunc(func(*http.Request) (*http.Response, error) {
		return jsonResponse(200, `{"status":"0","message":"NOTOK","result":"Max rate limit reached, please use API Key for higher rate limit"}`), nil
	}), 100, 5, 500)
	// MaxRetries=1 means 2 attempts; both fail identically.
	hist, err := client.FetchNormalTransactions(context.Background(), "0x2449ecef5012f0a0e153b278ef4fcc9625bc4c78")
	if err != nil {
		t.Fatalf("rate limit should return coverage, not hard error: %v", err)
	}
	if hist.Coverage.Status != domain.CoverageRateLimited {
		t.Fatalf("status = %s", hist.Coverage.Status)
	}
	if len(hist.Coverage.APIErrors) == 0 {
		t.Fatal("expected api errors recorded")
	}
}

func TestFetchAPIFailure(t *testing.T) {
	client := testClient(t, roundTripFunc(func(*http.Request) (*http.Response, error) {
		return jsonResponse(200, `{"status":"0","message":"NOTOK","result":"Invalid API Key"}`), nil
	}), 100, 5, 500)
	hist, err := client.FetchNormalTransactions(context.Background(), "0x2449ecef5012f0a0e153b278ef4fcc9625bc4c78")
	if err != nil {
		t.Fatalf("expected soft failure with coverage, got %v", err)
	}
	if hist.Coverage.Status != domain.CoverageFailed {
		t.Fatalf("status = %s", hist.Coverage.Status)
	}
}

func TestFetchInvalidAddress(t *testing.T) {
	client := testClient(t, roundTripFunc(func(*http.Request) (*http.Response, error) {
		t.Fatal("should not call network")
		return nil, nil
	}), 100, 5, 500)
	_, err := client.FetchNormalTransactions(context.Background(), "not-an-address")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestFetchDuplicateResults(t *testing.T) {
	dup := `{"blockNumber":"1","blockHash":"0xabc","timeStamp":"1","hash":"0xf9db905d77704596d3600816bc70201586cfeec13bcf576320e2f38d6ca851a0","nonce":"8","transactionIndex":"1","from":"0x2449ecef5012f0a0e153b278ef4fcc9625bc4c78","to":"0xc5102fe9359fd9a28f877a67e36b0f050d81a3cc","value":"0","gas":"1","gasPrice":"1","input":"0x","gasUsed":"1","isError":"0","txreceipt_status":"1"}`
	body := `{"status":"1","message":"OK","result":[` + dup + `,` + dup + `]}`
	client := testClient(t, roundTripFunc(func(*http.Request) (*http.Response, error) {
		return jsonResponse(200, body), nil
	}), 100, 5, 500)
	hist, err := client.FetchNormalTransactions(context.Background(), "0x2449ecef5012f0a0e153b278ef4fcc9625bc4c78")
	if err != nil {
		t.Fatal(err)
	}
	if len(hist.Transactions) != 1 {
		t.Fatalf("unique txs = %d", len(hist.Transactions))
	}
	if hist.Coverage.DuplicatesDropped != 1 {
		t.Fatalf("duplicates = %d", hist.Coverage.DuplicatesDropped)
	}
}

func TestFetchUnsupportedResponse(t *testing.T) {
	client := testClient(t, roundTripFunc(func(*http.Request) (*http.Response, error) {
		return jsonResponse(200, `{"status":"1","message":"OK","result":"surprise-string"}`), nil
	}), 100, 5, 500)
	hist, err := client.FetchNormalTransactions(context.Background(), "0x2449ecef5012f0a0e153b278ef4fcc9625bc4c78")
	if err != nil {
		t.Fatalf("expected coverage result, got err %v", err)
	}
	if hist.Coverage.Status != domain.CoverageFailed {
		t.Fatalf("status = %s", hist.Coverage.Status)
	}
}

func TestContextCancel(t *testing.T) {
	client := testClient(t, roundTripFunc(func(*http.Request) (*http.Response, error) {
		time.Sleep(50 * time.Millisecond)
		return jsonResponse(200, `{"status":"1","message":"OK","result":[]}`), nil
	}), 100, 5, 500)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := client.FetchNormalTransactions(ctx, "0x2449ecef5012f0a0e153b278ef4fcc9625bc4c78")
	if err == nil {
		t.Fatal("expected cancel error")
	}
}

func TestRejectNonMainnetConfig(t *testing.T) {
	_, err := NewClient(Config{BaseURL: "https://api.etherscan.io/v2/api", APIKey: "k", ChainID: 11155111})
	if err == nil {
		t.Fatal("expected mainnet-only error")
	}
}
