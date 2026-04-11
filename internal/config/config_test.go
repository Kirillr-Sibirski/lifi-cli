package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadUsesProfileAndEnvPrecedence(t *testing.T) {
	tmpDir := t.TempDir()
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(oldWD)
	})
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	configPath := filepath.Join(tmpDir, "config.yaml")
	configBody := `
profile: default
api:
  lifi_api_key: root-key
defaults:
  from_chain: base
profiles:
  prod:
    api:
      lifi_api_key: profile-key
    defaults:
      from_chain: optimism
    wallet:
      address: "0x123"
      private_key_env: "CUSTOM_WALLET_KEY"
    rpcs:
      optimism: "https://optimism.example"
`
	if err := os.WriteFile(configPath, []byte(configBody), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, ".env"), []byte("LIFI_API_KEY=dotenv-key\n"), 0o644); err != nil {
		t.Fatalf("write dotenv: %v", err)
	}

	t.Setenv("LIFI_DEFAULT_FROM_CHAIN", "arbitrum")
	t.Setenv("CUSTOM_WALLET_KEY", "abc123")

	cfg, err := Load(GlobalOptions{ConfigPath: configPath, Profile: "prod"})
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.ProfileName != "prod" {
		t.Fatalf("expected selected profile prod, got %q", cfg.ProfileName)
	}
	if cfg.APIKey != "dotenv-key" {
		t.Fatalf("expected dotenv api key precedence, got %q", cfg.APIKey)
	}
	if cfg.DefaultFromChain != "arbitrum" {
		t.Fatalf("expected env default chain precedence, got %q", cfg.DefaultFromChain)
	}
	if cfg.WalletKeyEnvName != "CUSTOM_WALLET_KEY" {
		t.Fatalf("expected custom wallet env name, got %q", cfg.WalletKeyEnvName)
	}
	if cfg.LookupRPC("optimism") != "https://optimism.example" {
		t.Fatalf("expected optimism rpc from profile, got %q", cfg.LookupRPC("optimism"))
	}
}
