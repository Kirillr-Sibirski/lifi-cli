package cli

import (
	"context"
	"testing"

	"github.com/Kirillr-Sibirski/lifi-cli/internal/config"
	"github.com/Kirillr-Sibirski/lifi-cli/internal/lifiapi"
)

func TestParseGlobalOptionsAnywhere(t *testing.T) {
	t.Parallel()

	global, remaining, err := parseGlobalOptions([]string{"deposit", "--json", "--profile", "prod", "--config", "/tmp/lifi.yaml"})
	if err != nil {
		t.Fatalf("parseGlobalOptions returned error: %v", err)
	}

	if !global.JSON {
		t.Fatalf("expected JSON flag to be true")
	}
	if global.Profile != "prod" {
		t.Fatalf("expected profile prod, got %q", global.Profile)
	}
	if global.ConfigPath != "/tmp/lifi.yaml" {
		t.Fatalf("expected config path to be captured, got %q", global.ConfigPath)
	}
	if len(remaining) != 1 || remaining[0] != "deposit" {
		t.Fatalf("unexpected remaining args: %#v", remaining)
	}
}

func TestResolveChainAndToken(t *testing.T) {
	t.Parallel()

	rt := &runtime{
		cfg: &config.Config{DefaultFromChain: "base"},
		chains: []lifiapi.Chain{
			{
				ID:   8453,
				Name: "Base",
				Key:  "base",
				Coin: "ETH",
				NativeToken: lifiapi.Token{
					ChainID:  8453,
					Address:  "0x0000000000000000000000000000000000000000",
					Symbol:   "ETH",
					Name:     "Ether",
					Decimals: 18,
					CoinKey:  "ETH",
				},
			},
		},
		tokensByChain: map[int][]lifiapi.Token{
			8453: {
				{
					ChainID:  8453,
					Address:  "0x833589fcd6edb6e08f4c7c32d4f71b54bda02913",
					Symbol:   "USDC",
					Name:     "USD Coin",
					Decimals: 6,
					CoinKey:  "USDC",
				},
			},
		},
	}

	chain, err := rt.resolveChain(context.Background(), "Base")
	if err != nil {
		t.Fatalf("resolveChain returned error: %v", err)
	}
	if chain.ID != 8453 {
		t.Fatalf("expected base chain id 8453, got %d", chain.ID)
	}

	token, err := rt.resolveToken(context.Background(), chain, "usd coin")
	if err != nil {
		t.Fatalf("resolveToken returned error: %v", err)
	}
	if token.Symbol != "USDC" {
		t.Fatalf("expected USDC, got %s", token.Symbol)
	}
}
