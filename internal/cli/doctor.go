package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Kirillr-Sibirski/defi-mullet/internal/config"
)

type doctorCommand struct{}

type doctorCheck struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Detail string `json:"detail,omitempty"`
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
		httpCheck("earn.li.fi", "https://earn.li.fi"),
		httpCheck("li.quest", "https://li.quest/v1/info"),
		envCheck("LIFI_API_KEY", cfg.APIKey),
		configPathCheck(cfg),
	}

	if writeChecks {
		checks = append(checks, envCheck("LIFI_WALLET_PRIVATE_KEY", cfg.WalletPrivateKey))
		checks = append(checks, envCheck("LIFI_WALLET_ADDRESS", cfg.WalletAddress))
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
	for _, check := range checks {
		fmt.Printf("[%s] %s", check.Status, check.Name)
		if check.Detail != "" {
			fmt.Printf(": %s", check.Detail)
		}
		fmt.Println()
	}

	return nil
}

func httpCheck(name, url string) doctorCheck {
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return doctorCheck{Name: name, Status: "fail", Detail: err.Error()}
	}

	resp, err := client.Do(req)
	if err != nil {
		return doctorCheck{Name: name, Status: "fail", Detail: err.Error()}
	}
	defer resp.Body.Close()

	status := "ok"
	if resp.StatusCode >= 400 {
		status = "warn"
	}

	return doctorCheck{
		Name:   name,
		Status: status,
		Detail: fmt.Sprintf("HTTP %d", resp.StatusCode),
	}
}

func envCheck(name, value string) doctorCheck {
	if value == "" {
		return doctorCheck{Name: name, Status: "warn", Detail: "not set"}
	}

	return doctorCheck{Name: name, Status: "ok", Detail: "set"}
}

func configPathCheck(cfg *config.Config) doctorCheck {
	if cfg.ConfigExists() {
		return doctorCheck{Name: "config file", Status: "ok", Detail: cfg.ResolvedConfigPath}
	}

	return doctorCheck{Name: "config file", Status: "warn", Detail: "not found at " + cfg.ResolvedConfigPath}
}

func rpcCheck(chain, rpcURL string) doctorCheck {
	name := "rpc"
	if chain != "" {
		name = "rpc:" + chain
	}

	if rpcURL == "" {
		return doctorCheck{Name: name, Status: "warn", Detail: "no RPC configured"}
	}

	return doctorCheck{Name: name, Status: "ok", Detail: rpcURL}
}
