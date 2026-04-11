package flow

import (
	"testing"

	"github.com/Kirillr-Sibirski/lifi-cli/internal/earn"
)

func TestPortfolioShowsDepositWithExactMatch(t *testing.T) {
	t.Parallel()

	vault := earn.Vault{
		Address: "0xVault",
		ChainID: 10,
		Protocol: earn.Protocol{
			Name: "morpho-v1",
		},
		UnderlyingTokens: []earn.UnderlyingToken{{Symbol: "USDC", Address: "0xUsdc"}},
	}
	positions := []map[string]any{
		{
			"protocolName":  "morpho-v1",
			"chainId":       10,
			"balanceNative": "1.0",
			"asset": map[string]any{
				"symbol":  "USDC",
				"address": "0xUsdc",
			},
		},
	}

	if !PortfolioShowsDeposit(positions, vault, 0, 1) {
		t.Fatalf("expected exact portfolio match to count as detected")
	}
}

func TestPortfolioShowsDepositUsesDeltaFallback(t *testing.T) {
	t.Parallel()

	vault := earn.Vault{
		ChainID:          10,
		UnderlyingTokens: []earn.UnderlyingToken{{Symbol: "USDC", Address: "0xUsdc"}},
	}
	positions := []map[string]any{
		{
			"chainId":       10,
			"balanceNative": "1.5",
			"asset": map[string]any{
				"symbol":  "USDC",
				"address": "0xUsdc",
			},
		},
	}

	if !PortfolioShowsDeposit(positions, vault, 1.0, 0.5) {
		t.Fatalf("expected delta fallback to detect deposit")
	}
}
