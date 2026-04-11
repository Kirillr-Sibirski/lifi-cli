package cli

import (
	"context"
	"os"
	"testing"

	"github.com/Kirillr-Sibirski/lifi-cli/internal/config"
)

func TestLiveSmokeReadPath(t *testing.T) {
	if os.Getenv("LIFI_SMOKE") != "1" {
		t.Skip("set LIFI_SMOKE=1 to run live smoke checks")
	}

	cfg, err := config.Load(config.GlobalOptions{})
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	rt := newRuntime(cfg)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	chains, err := rt.loadChains(ctx)
	if err != nil {
		t.Fatalf("load chains: %v", err)
	}
	if len(chains) == 0 {
		t.Fatalf("expected at least one chain")
	}

	vaults, err := loadAndFilterVaults(cfg, "base", "USDC", "", "apy", "desc", "", "", true)
	if err != nil {
		t.Fatalf("load vaults: %v", err)
	}
	if len(vaults) == 0 {
		t.Fatalf("expected at least one vault on base")
	}
}
