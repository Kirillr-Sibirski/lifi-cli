package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Kirillr-Sibirski/lifi-cli/internal/config"
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

	if writeChecks {
		checks = append(checks, envCheck("wallet", cfg.WalletKeyEnvName, cfg.WalletPrivateKey))
		checks = append(checks, envCheck("wallet", "LIFI_WALLET_ADDRESS", cfg.WalletAddress))
	}

	if chain != "" || rpcURL != "" {
		targetChain := chain
		targetRPC := rpcURL

		if targetRPC == "" && targetChain != "" {
			targetRPC = cfg.RPCs[targetChain]
		}

		checks = append(checks, rpcCheck(targetChain, targetRPC))
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

func rpcCheck(chain, rpcURL string) doctorCheck {
	name := "rpc"
	if chain != "" {
		name = "rpc:" + chain
	}

	if rpcURL == "" {
		return doctorCheck{Category: "rpc", Name: name, Status: "warn", Detail: "no RPC configured"}
	}

	return doctorCheck{Category: "rpc", Name: name, Status: "ok", Detail: rpcURL}
}
