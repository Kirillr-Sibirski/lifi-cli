package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	defaultConfigDirName  = "lifi"
	defaultConfigFileName = "config.yaml"
)

type GlobalOptions struct {
	ConfigPath string
	Profile    string
	JSON       bool
	Verbose    bool
	Quiet      bool
	NoColor    bool
}

type Config struct {
	Global             GlobalOptions
	ResolvedConfigPath string
	APIKey             string
	WalletPrivateKey   string
	WalletAddress      string
	DefaultFromChain   string
	DefaultSlippageBPS string
	RPCs               map[string]string
}

func Load(global GlobalOptions) (*Config, error) {
	resolvedPath, err := ResolveConfigPath(global.ConfigPath)
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		Global:             global,
		ResolvedConfigPath: resolvedPath,
		APIKey:             os.Getenv("LIFI_API_KEY"),
		WalletPrivateKey:   os.Getenv("LIFI_WALLET_PRIVATE_KEY"),
		WalletAddress:      os.Getenv("LIFI_WALLET_ADDRESS"),
		DefaultFromChain:   os.Getenv("LIFI_DEFAULT_FROM_CHAIN"),
		DefaultSlippageBPS: os.Getenv("LIFI_DEFAULT_SLIPPAGE_BPS"),
		RPCs:               loadRPCsFromEnv(),
	}

	if cfg.Global.Profile == "" {
		cfg.Global.Profile = "default"
	}

	return cfg, nil
}

func ResolveConfigPath(explicit string) (string, error) {
	if explicit != "" {
		return explicit, nil
	}

	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve config dir: %w", err)
	}

	return filepath.Join(configDir, defaultConfigDirName, defaultConfigFileName), nil
}

func loadRPCsFromEnv() map[string]string {
	rpcs := map[string]string{}

	for _, env := range os.Environ() {
		if !strings.HasPrefix(env, "LIFI_RPC_") {
			continue
		}

		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 || parts[1] == "" {
			continue
		}

		key := strings.TrimPrefix(parts[0], "LIFI_RPC_")
		if key == "" {
			continue
		}

		rpcs[strings.ToLower(key)] = parts[1]
	}

	return rpcs
}

func (c *Config) ConfigExists() bool {
	_, err := os.Stat(c.ResolvedConfigPath)
	return err == nil
}

func (c *Config) RPCKeys() []string {
	keys := make([]string, 0, len(c.RPCs))
	for key := range c.RPCs {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func DefaultConfigTemplate() string {
	return strings.TrimSpace(`
profile: default
api:
  lifi_api_key: ""
defaults:
  from_chain: base
  slippage_bps: 50
  output: table
wallet:
  address: "0x..."
  private_key_env: "LIFI_WALLET_PRIVATE_KEY"
rpcs:
  base: "https://..."
  arbitrum: "https://..."
`) + "\n"
}
