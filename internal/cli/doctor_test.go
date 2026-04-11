package cli

import (
	"testing"

	"github.com/Kirillr-Sibirski/lifi-cli/internal/config"
)

func TestWalletChecksMatchDerivedAddress(t *testing.T) {
	cfg := &config.Config{
		WalletPrivateKey: "4f3edf983ac636a65a842ce7c78d9aa706d3b113bce036f4ee5e04b6c74f8321",
		WalletAddress:    "0xe4c9F49C7681e81Bd05952150E4C42e9B2e02d5F",
		WalletKeyEnvName: "LIFI_WALLET_PRIVATE_KEY",
	}

	checks := walletChecks(cfg)
	if len(checks) < 4 {
		t.Fatalf("expected wallet checks, got %d", len(checks))
	}

	last := checks[len(checks)-1]
	if last.Name != "wallet address consistency" {
		t.Fatalf("unexpected last check: %#v", last)
	}
	if last.Status != "ok" {
		t.Fatalf("expected ok status, got %#v", last)
	}
}

func TestWalletChecksDetectMismatch(t *testing.T) {
	cfg := &config.Config{
		WalletPrivateKey: "4f3edf983ac636a65a842ce7c78d9aa706d3b113bce036f4ee5e04b6c74f8321",
		WalletAddress:    "0x1111111111111111111111111111111111111111",
		WalletKeyEnvName: "LIFI_WALLET_PRIVATE_KEY",
	}

	checks := walletChecks(cfg)
	last := checks[len(checks)-1]
	if last.Name != "wallet address consistency" {
		t.Fatalf("unexpected last check: %#v", last)
	}
	if last.Status != "fail" {
		t.Fatalf("expected fail status, got %#v", last)
	}
}
