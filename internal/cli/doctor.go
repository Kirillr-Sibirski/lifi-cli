package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/Kirillr-Sibirski/lifi-cli/internal/config"
	"github.com/Kirillr-Sibirski/lifi-cli/internal/evm"
	"github.com/Kirillr-Sibirski/lifi-cli/internal/lifiapi"
)

type doctorCommand struct{}

type doctorCheck struct {
	Category string `json:"category,omitempty"`
	Name     string `json:"name"`
	Status   string `json:"status"`
	Detail   string `json:"detail,omitempty"`
}

type doctorResult struct {
	ConfigPath string        `json:"config_path"`
	Profile    string        `json:"profile"`
	Checks     []doctorCheck `json:"checks"`
}

func newDoctorCommand() Command {
	return doctorCommand{}
}

func (doctorCommand) Name() string {
	return "doctor"
}

func (doctorCommand) Summary() string {
	return "Check environment, API reachability, and wallet readiness"
}

func (doctorCommand) Usage() string {
	return "lifi doctor [--write-checks] [--chain <chain>] [--rpc-url <url>] [--json]"
}

func (doctorCommand) Run(cfg *config.Config, args []string) error {
	fs := newFlagSet("doctor")
	var writeChecks bool
	var chain string
	var rpcURL string

	fs.BoolVar(&writeChecks, "write-checks", false, "Validate wallet and RPC requirements for write commands")
	fs.StringVar(&chain, "chain", "", "Check a specific chain")
	fs.StringVar(&rpcURL, "rpc-url", "", "Override the RPC URL used for a chain check")

	fs.Usage = func() {
		fmt.Println("Usage:")
		fmt.Println("  " + doctorCommand{}.Usage())
	}

	if err := fs.Parse(args); err != nil {
		return err
	}

	checks := []doctorCheck{
		httpCheck("api", "earn.li.fi", "https://earn.li.fi/v1/earn/chains"),
		httpCheck("api", "li.quest", "https://li.quest/v1/chains"),
		envCheck("config", "LIFI_API_KEY", cfg.APIKey),
		dotEnvCheck(cfg),
		configPathCheck(cfg),
	}

	var expectedChain *lifiapi.Chain
	if strings.TrimSpace(chain) != "" {
		rt := newRuntime(cfg)
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		resolved, err := rt.resolveChain(ctx, chain)
		if err != nil {
			checks = append(checks, doctorCheck{
				Category: "rpc",
				Name:     "requested chain",
				Status:   "warn",
				Detail:   err.Error(),
			})
		} else {
			expectedChain = resolved
		}
	}

	if writeChecks {
		checks = append(checks, walletChecks(cfg)...)
	}

	if chain != "" || rpcURL != "" {
		targetChain := chain
		targetRPC := rpcURL

		if targetRPC == "" && expectedChain != nil {
			targetRPC = cfg.LookupRPC(expectedChain.Name)
			if targetRPC == "" {
				targetRPC = cfg.LookupRPC(expectedChain.Key)
			}
			if targetRPC == "" {
				targetRPC = cfg.LookupRPC(fmt.Sprintf("%d", expectedChain.ID))
			}
			if targetRPC == "" && len(expectedChain.Metamask.RPCURLs) > 0 {
				targetRPC = expectedChain.Metamask.RPCURLs[0]
			}
		}

		if targetRPC == "" && targetChain != "" {
			targetRPC = cfg.LookupRPC(targetChain)
		}

		checks = append(checks, rpcCheck(targetChain, targetRPC, expectedChain))
		if writeChecks && targetRPC != "" && cfg.WalletAddress != "" {
			checks = append(checks, nativeBalanceCheck(targetRPC, cfg.WalletAddress, expectedChain))
		}
	}

	result := doctorResult{
		ConfigPath: cfg.ResolvedConfigPath,
		Profile:    cfg.Global.Profile,
		Checks:     checks,
	}

	if cfg.Global.JSON {
		encoder := json.NewEncoder(osStdout{})
		encoder.SetIndent("", "  ")
		return encoder.Encode(result)
	}

	fmt.Printf("Config path: %s\n", cfg.ResolvedConfigPath)
	fmt.Printf("Profile:     %s\n\n", cfg.Global.Profile)
	categories := []string{"config", "api", "wallet", "rpc"}
	for _, category := range categories {
		filtered := make([]doctorCheck, 0)
		for _, check := range checks {
			if check.Category == category {
				filtered = append(filtered, check)
			}
		}
		if len(filtered) == 0 {
			continue
		}
		printSectionHeader(category, cfg.Global.NoColor)
		for _, check := range filtered {
			fmt.Printf("  [%s] %s", check.Status, check.Name)
			if check.Detail != "" {
				fmt.Printf(": %s", check.Detail)
			}
			fmt.Println()
		}
		fmt.Println()
	}

	return nil
}

func httpCheck(category, name, url string) doctorCheck {
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return doctorCheck{Category: category, Name: name, Status: "fail", Detail: err.Error()}
	}

	resp, err := client.Do(req)
	if err != nil {
		return doctorCheck{Category: category, Name: name, Status: "fail", Detail: err.Error()}
	}
	defer resp.Body.Close()

	status := "ok"
	if resp.StatusCode >= 400 {
		status = "warn"
	}

	return doctorCheck{
		Category: category,
		Name:     name,
		Status:   status,
		Detail:   fmt.Sprintf("HTTP %d", resp.StatusCode),
	}
}

func envCheck(category, name, value string) doctorCheck {
	if value == "" {
		return doctorCheck{Category: category, Name: name, Status: "warn", Detail: "not set"}
	}

	return doctorCheck{Category: category, Name: name, Status: "ok", Detail: "set"}
}

func configPathCheck(cfg *config.Config) doctorCheck {
	if cfg.ConfigExists() {
		return doctorCheck{Category: "config", Name: "config file", Status: "ok", Detail: cfg.ResolvedConfigPath}
	}

	return doctorCheck{Category: "config", Name: "config file", Status: "warn", Detail: "not found at " + cfg.ResolvedConfigPath}
}

func dotEnvCheck(cfg *config.Config) doctorCheck {
	if cfg.DotEnvExists() {
		return doctorCheck{Category: "config", Name: ".env file", Status: "ok", Detail: cfg.DotEnvPath}
	}
	return doctorCheck{Category: "config", Name: ".env file", Status: "warn", Detail: "not found at " + cfg.DotEnvPath}
}

func walletChecks(cfg *config.Config) []doctorCheck {
	checks := []doctorCheck{
		envCheck("wallet", cfg.WalletKeyEnvName, cfg.WalletPrivateKey),
		envCheck("wallet", "LIFI_WALLET_ADDRESS", cfg.WalletAddress),
	}

	if strings.TrimSpace(cfg.WalletPrivateKey) == "" {
		return checks
	}

	wallet, err := evm.WalletFromHex(cfg.WalletPrivateKey)
	if err != nil {
		checks = append(checks, doctorCheck{
			Category: "wallet",
			Name:     "private key",
			Status:   "fail",
			Detail:   err.Error(),
		})
		return checks
	}

	checks = append(checks, doctorCheck{
		Category: "wallet",
		Name:     "derived wallet address",
		Status:   "ok",
		Detail:   wallet.Address.Hex(),
	})

	if strings.TrimSpace(cfg.WalletAddress) == "" {
		return checks
	}

	status := "ok"
	detail := "matches private key"
	if !strings.EqualFold(cfg.WalletAddress, wallet.Address.Hex()) {
		status = "fail"
		detail = fmt.Sprintf("configured %s but derived %s", cfg.WalletAddress, wallet.Address.Hex())
	}

	checks = append(checks, doctorCheck{
		Category: "wallet",
		Name:     "wallet address consistency",
		Status:   status,
		Detail:   detail,
	})

	return checks
}

func rpcCheck(chain, rpcURL string, expectedChain *lifiapi.Chain) doctorCheck {
	name := "rpc"
	if chain != "" {
		name = "rpc:" + chain
	}

	if rpcURL == "" {
		return doctorCheck{Category: "rpc", Name: name, Status: "warn", Detail: "no RPC configured"}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, err := evm.DialRPC(ctx, rpcURL)
	if err != nil {
		return doctorCheck{Category: "rpc", Name: name, Status: "fail", Detail: err.Error()}
	}
	defer client.Close()

	chainID, err := client.ChainID(ctx)
	if err != nil {
		return doctorCheck{Category: "rpc", Name: name, Status: "fail", Detail: err.Error()}
	}

	detail := fmt.Sprintf("%s (chain id %s)", rpcURL, chainID.String())
	status := "ok"
	if expectedChain != nil && chainID.Int64() != int64(expectedChain.ID) {
		status = "fail"
		detail = fmt.Sprintf("%s (chain id %s, expected %d)", rpcURL, chainID.String(), expectedChain.ID)
	}

	return doctorCheck{Category: "rpc", Name: name, Status: status, Detail: detail}
}

func nativeBalanceCheck(rpcURL, walletAddress string, expectedChain *lifiapi.Chain) doctorCheck {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, err := evm.DialRPC(ctx, rpcURL)
	if err != nil {
		return doctorCheck{Category: "rpc", Name: "native gas balance", Status: "fail", Detail: err.Error()}
	}
	defer client.Close()

	balance, err := evm.Balance(ctx, client, "", common.HexToAddress(walletAddress))
	if err != nil {
		return doctorCheck{Category: "rpc", Name: "native gas balance", Status: "fail", Detail: err.Error()}
	}

	symbol := "native"
	decimals := 18
	if expectedChain != nil {
		symbol = expectedChain.NativeToken.Symbol
		decimals = expectedChain.NativeToken.Decimals
	}

	status := "ok"
	detail := fmt.Sprintf("%s %s", formatAmount(balance.String(), decimals, 6), symbol)
	if balance.Sign() == 0 {
		status = "warn"
		detail = "0 " + symbol
	}

	return doctorCheck{Category: "rpc", Name: "native gas balance", Status: status, Detail: detail}
}
