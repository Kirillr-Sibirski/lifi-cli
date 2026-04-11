package config

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"gopkg.in/yaml.v3"
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

type fileConfig struct {
	Profile string `yaml:"profile"`
	API     struct {
		LifiAPIKey string `yaml:"lifi_api_key"`
	} `yaml:"api"`
	Defaults struct {
		FromChain   string `yaml:"from_chain"`
		SlippageBPS string `yaml:"slippage_bps"`
		Output      string `yaml:"output"`
	} `yaml:"defaults"`
	Wallet struct {
		Address       string `yaml:"address"`
		PrivateKeyEnv string `yaml:"private_key_env"`
	} `yaml:"wallet"`
	RPCs map[string]string `yaml:"rpcs"`
}

type Config struct {
	Global             GlobalOptions
	ResolvedConfigPath string
	DotEnvPath         string
	FileConfig         fileConfig
	APIKey             string
	WalletPrivateKey   string
	WalletAddress      string
	DefaultFromChain   string
	DefaultSlippageBPS string
	RPCs               map[string]string
}

func Load(global GlobalOptions) (*Config, error) {
	if global.Profile == "" {
		global.Profile = "default"
	}

	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("resolve working directory: %w", err)
	}

	dotEnvPath := filepath.Join(cwd, ".env")
	_ = loadDotEnv(dotEnvPath)

	resolvedPath, err := ResolveConfigPath(global.ConfigPath)
	if err != nil {
		return nil, err
	}

	fileCfg, err := loadFileConfig(resolvedPath)
	if err != nil {
		return nil, err
	}

	apiKey := firstNonEmpty(
		os.Getenv("LIFI_API_KEY"),
		fileCfg.API.LifiAPIKey,
	)

	privateKey := firstNonEmpty(
		os.Getenv("LIFI_WALLET_PRIVATE_KEY"),
		func() string {
			if fileCfg.Wallet.PrivateKeyEnv == "" {
				return ""
			}
			return os.Getenv(fileCfg.Wallet.PrivateKeyEnv)
		}(),
	)

	walletAddress := firstNonEmpty(
		os.Getenv("LIFI_WALLET_ADDRESS"),
		fileCfg.Wallet.Address,
	)
	if walletAddress == "" && privateKey != "" {
		derived, err := deriveWalletAddress(privateKey)
		if err == nil {
			walletAddress = derived
		}
	}

	cfg := &Config{
		Global:             global,
		ResolvedConfigPath: resolvedPath,
		DotEnvPath:         dotEnvPath,
		FileConfig:         fileCfg,
		APIKey:             apiKey,
		WalletPrivateKey:   privateKey,
		WalletAddress:      walletAddress,
		DefaultFromChain: firstNonEmpty(
			os.Getenv("LIFI_DEFAULT_FROM_CHAIN"),
			fileCfg.Defaults.FromChain,
		),
		DefaultSlippageBPS: firstNonEmpty(
			os.Getenv("LIFI_DEFAULT_SLIPPAGE_BPS"),
			fileCfg.Defaults.SlippageBPS,
		),
		RPCs: mergeRPCs(fileCfg.RPCs, loadRPCsFromEnv()),
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

func loadFileConfig(path string) (fileConfig, error) {
	var cfg fileConfig
	if path == "" {
		return cfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}
		return cfg, fmt.Errorf("read config file: %w", err)
	}

	if len(data) == 0 {
		return cfg, nil
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("parse config file: %w", err)
	}

	if cfg.RPCs == nil {
		cfg.RPCs = map[string]string{}
	}

	return cfg, nil
}

func loadDotEnv(path string) error {
	data, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	defer data.Close()

	scanner := bufio.NewScanner(data)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		line = strings.TrimPrefix(line, "export ")
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		value = strings.Trim(value, `"'`)
		if key == "" {
			continue
		}
		if _, exists := os.LookupEnv(key); exists {
			continue
		}
		if err := os.Setenv(key, value); err != nil {
			return err
		}
	}

	return scanner.Err()
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

		rpcs[normalizeKey(key)] = parts[1]
	}

	return rpcs
}

func mergeRPCs(fileRPCs, envRPCs map[string]string) map[string]string {
	merged := map[string]string{}

	for key, value := range fileRPCs {
		if value == "" {
			continue
		}
		merged[normalizeKey(key)] = value
	}

	for key, value := range envRPCs {
		if value == "" {
			continue
		}
		merged[normalizeKey(key)] = value
	}

	return merged
}

func normalizeKey(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, "-", "")
	value = strings.ReplaceAll(value, "_", "")
	value = strings.ReplaceAll(value, " ", "")
	return value
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func deriveWalletAddress(privateKeyHex string) (string, error) {
	privateKeyHex = strings.TrimPrefix(strings.TrimSpace(privateKeyHex), "0x")
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return "", err
	}

	address := crypto.PubkeyToAddress(privateKey.PublicKey)
	return common.HexToAddress(address.Hex()).Hex(), nil
}

func (c *Config) ConfigExists() bool {
	_, err := os.Stat(c.ResolvedConfigPath)
	return err == nil
}

func (c *Config) DotEnvExists() bool {
	if c.DotEnvPath == "" {
		return false
	}
	_, err := os.Stat(c.DotEnvPath)
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

func (c *Config) LookupRPC(key string) string {
	return c.RPCs[normalizeKey(key)]
}

func DefaultConfigTemplate() string {
	return strings.TrimSpace(`
profile: default
api:
  lifi_api_key: ""
defaults:
  from_chain: base
  slippage_bps: "50"
  output: table
wallet:
  address: "0x..."
  private_key_env: "LIFI_WALLET_PRIVATE_KEY"
rpcs:
  base: "https://mainnet.base.org"
  arbitrum: "https://arb1.arbitrum.io/rpc"
`) + "\n"
}
