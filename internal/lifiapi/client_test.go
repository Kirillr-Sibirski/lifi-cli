package lifiapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGetQuoteAndStatus(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/quote":
			_, _ = w.Write([]byte(`{
				"tool":"lifi",
				"toolDetails":{"name":"LI.FI"},
				"action":{"fromAmount":"100","fromChainId":10,"toChainId":10,"fromToken":{"symbol":"USDC","decimals":6},"toToken":{"symbol":"aUSDC","decimals":6}},
				"estimate":{"approvalAddress":"0xabc","toAmount":"95","toAmountMin":"94"},
				"transactionRequest":{"to":"0xdef","value":"0","data":"0x","chainId":10,"gasPrice":"1","gasLimit":"21000"}
			}`))
		case "/status":
			_, _ = w.Write([]byte(`{"status":"DONE","bridge":"lifi"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := &Client{
		baseURL: server.URL,
		http:    &http.Client{Timeout: time.Second},
	}
	quote, err := client.GetQuote(context.Background(), QuoteRequest{
		FromChain:   10,
		ToChain:     10,
		FromToken:   "0xusdc",
		ToToken:     "0xvault",
		FromAddress: "0x1111111111111111111111111111111111111111",
		ToAddress:   "0x1111111111111111111111111111111111111111",
		FromAmount:  "100",
	})
	if err != nil {
		t.Fatalf("GetQuote returned error: %v", err)
	}
	if quote.Tool != "lifi" {
		t.Fatalf("expected tool lifi, got %q", quote.Tool)
	}

	status, err := client.GetStatus(context.Background(), StatusRequest{TxHash: "0xabc"})
	if err != nil {
		t.Fatalf("GetStatus returned error: %v", err)
	}
	if status["status"] != "DONE" {
		t.Fatalf("expected DONE status, got %v", status["status"])
	}
}
