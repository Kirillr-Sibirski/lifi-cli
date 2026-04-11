package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Kirillr-Sibirski/defi-mullet/internal/config"
)

type configCommand struct{}

func newConfigCommand() Command {
	return configCommand{}
}

func (configCommand) Name() string {
	return "config"
}

func (configCommand) Summary() string {
	return "Initialize or inspect lifi config"
}

func (configCommand) Usage() string {
	return "lifi config <init|show> [flags]"
}

func (configCommand) Run(cfg *config.Config, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("config requires a subcommand: init or show")
	}

	switch args[0] {
	case "init":
		return runConfigInit(cfg, args[1:])
	case "show":
		return runConfigShow(cfg)
	case "--help", "-h", "help":
		fmt.Println("Usage:")
		fmt.Println("  lifi config init [--force]")
		fmt.Println("  lifi config show")
		return nil
	default:
		return fmt.Errorf("unknown config subcommand %q", args[0])
	}
}

func runConfigInit(cfg *config.Config, args []string) error {
	fs := newFlagSet("config init")
	var force bool
	fs.BoolVar(&force, "force", false, "Overwrite an existing config file")

	fs.Usage = func() {
		fmt.Println("Usage:")
		fmt.Println("  lifi config init [--force]")
	}

	if err := fs.Parse(args); err != nil {
		return err
	}

	if cfg.ConfigExists() && !force {
		return fmt.Errorf("config file already exists at %s (use --force to overwrite)", cfg.ResolvedConfigPath)
	}

	if err := os.MkdirAll(filepath.Dir(cfg.ResolvedConfigPath), 0o755); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	if err := os.WriteFile(cfg.ResolvedConfigPath, []byte(config.DefaultConfigTemplate()), 0o644); err != nil {
		return fmt.Errorf("write config file: %w", err)
	}

	if cfg.Global.JSON {
		return writeJSON(map[string]string{
			"status":      "ok",
			"config_path": cfg.ResolvedConfigPath,
		})
	}

	fmt.Printf("Wrote config file to %s\n", cfg.ResolvedConfigPath)
	return nil
}

func runConfigShow(cfg *config.Config) error {
	payload := map[string]any{
		"config_path":         cfg.ResolvedConfigPath,
		"config_exists":       cfg.ConfigExists(),
		"dotenv_path":         cfg.DotEnvPath,
		"dotenv_exists":       cfg.DotEnvExists(),
		"profile":             cfg.Global.Profile,
		"default_from_chain":  cfg.DefaultFromChain,
		"default_slippagebps": cfg.DefaultSlippageBPS,
		"wallet_address_set":  cfg.WalletAddress != "",
		"wallet_key_set":      cfg.WalletPrivateKey != "",
		"api_key_set":         cfg.APIKey != "",
		"rpc_keys":            cfg.RPCKeys(),
	}

	if cfg.Global.JSON {
		return writeJSON(payload)
	}

	fmt.Printf("Config path:          %s\n", cfg.ResolvedConfigPath)
	fmt.Printf("Config exists:        %t\n", cfg.ConfigExists())
	fmt.Printf(".env path:            %s\n", cfg.DotEnvPath)
	fmt.Printf(".env exists:          %t\n", cfg.DotEnvExists())
	fmt.Printf("Profile:              %s\n", cfg.Global.Profile)
	fmt.Printf("Default from chain:   %s\n", emptyFallback(cfg.DefaultFromChain))
	fmt.Printf("Default slippage bps: %s\n", emptyFallback(cfg.DefaultSlippageBPS))
	fmt.Printf("API key set:          %t\n", cfg.APIKey != "")
	fmt.Printf("Wallet address set:   %t\n", cfg.WalletAddress != "")
	fmt.Printf("Wallet key set:       %t\n", cfg.WalletPrivateKey != "")
	fmt.Printf("RPC keys:             %s\n", formatList(cfg.RPCKeys()))
	return nil
}

func writeJSON(v any) error {
	encoder := json.NewEncoder(osStdout{})
	encoder.SetIndent("", "  ")
	return encoder.Encode(v)
}

func emptyFallback(value string) string {
	if value == "" {
		return "(unset)"
	}

	return value
}
