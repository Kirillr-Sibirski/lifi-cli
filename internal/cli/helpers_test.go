package cli

import "testing"

func TestTruncateAddrPreservesFullAddress(t *testing.T) {
	t.Parallel()

	addr := "0x1234567890abcdef1234567890abcdef12345678"
	if got := truncateAddr(addr); got != addr {
		t.Fatalf("expected full address, got %q", got)
	}
}

func TestPortfolioSummaryRowIncludesAssetAddress(t *testing.T) {
	t.Parallel()

	row := portfolioSummaryRow(1, map[string]any{
		"chainId":       8453,
		"protocolName":  "morpho-v1",
		"balanceNative": "1.0",
		"valueUsd":      "123.45",
		"asset": map[string]any{
			"symbol":  "USDC",
			"address": "0x1234567890abcdef1234567890abcdef12345678",
		},
	}, map[string]string{"8453": "Base"})

	if got := row[4]; got != "0x1234567890abcdef1234567890abcdef12345678" {
		t.Fatalf("expected address column to contain full asset address, got %q", got)
	}
}
