package cli

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/Kirillr-Sibirski/defi-mullet/internal/config"
)

const version = "0.1.0-dev"

type Command interface {
	Name() string
	Summary() string
	Usage() string
	Run(cfg *config.Config, args []string) error
}

func Run(args []string) int {
	commands := builtInCommands()

	global, remaining, err := parseGlobalOptions(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}

	cfg, err := config.Load(global)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	if len(remaining) == 0 {
		printRootUsage(commands)
		return 0
	}

	name := remaining[0]
	if name == "help" {
		printRootUsage(commands)
		return 0
	}

	cmd, ok := commands[name]
	if !ok {
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n", name)
		printRootUsage(commands)
		return 2
	}

	if err := cmd.Run(cfg, remaining[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	return 0
}

func parseGlobalOptions(args []string) (config.GlobalOptions, []string, error) {
	fs := flag.NewFlagSet("lifi", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	var global config.GlobalOptions
	fs.StringVar(&global.ConfigPath, "config", "", "Path to the config file")
	fs.StringVar(&global.Profile, "profile", "default", "Config profile name")
	fs.BoolVar(&global.JSON, "json", false, "Print machine-readable JSON output")
	fs.BoolVar(&global.Verbose, "verbose", false, "Enable verbose output")
	fs.BoolVar(&global.Quiet, "quiet", false, "Reduce non-essential output")
	fs.BoolVar(&global.NoColor, "no-color", false, "Disable ANSI color output")

	if err := fs.Parse(args); err != nil {
		return config.GlobalOptions{}, nil, err
	}

	return global, fs.Args(), nil
}

func builtInCommands() map[string]Command {
	commands := []Command{
		newDoctorCommand(),
		newPlaceholderCommand(commandSpec{
			name:    "chains",
			summary: "List LI.FI chains relevant to Earn and Composer execution",
			usage:   "lifi chains [--search <query>] [--evm-only] [--json]",
			description: "Lists chains relevant to LI.FI Earn and Composer execution. " +
				"The scaffold is in place and will be wired to LI.FI chain metadata next.",
			options: []optionSpec{
				{name: "search", kind: optionString, usage: "Filter chains by name or identifier"},
				{name: "evm-only", kind: optionBool, usage: "Only show EVM chains"},
			},
		}),
		newPlaceholderCommand(commandSpec{
			name:    "protocols",
			summary: "List supported Earn and Composer protocols",
			usage:   "lifi protocols [--search <query>] [--supports deposit|withdraw] [--json]",
			description: "Lists supported Earn and Composer protocols. " +
				"The scaffold is in place and will be wired to protocol metadata next.",
			options: []optionSpec{
				{name: "search", kind: optionString, usage: "Filter protocols by name"},
				{name: "supports", kind: optionString, usage: "Filter by supported action"},
			},
		}),
		newPlaceholderCommand(commandSpec{
			name:        "tokens",
			summary:     "Resolve tokens by symbol or address",
			usage:       "lifi tokens [--chain <chain>] [--token <symbol-or-address>] [--tags <tag[,tag]>] [--json]",
			description: "Resolves tokens through LI.FI metadata endpoints for quote preparation and validation.",
			options: []optionSpec{
				{name: "chain", kind: optionString, usage: "Filter tokens by chain"},
				{name: "token", kind: optionString, usage: "Resolve a token by symbol or address"},
				{name: "tags", kind: optionString, usage: "Filter by LI.FI token tags"},
			},
		}),
		newPlaceholderCommand(commandSpec{
			name:        "vaults",
			summary:     "List depositable vaults",
			usage:       "lifi vaults [--chain <chain>] [--asset <symbol-or-address>] [--protocol <name>] [--sort apy|apy30d|tvl|name] [--order asc|desc] [--min-tvl-usd <amount>] [--min-apy <percent>] [--transactional-only] [--limit <n>] [--json]",
			description: "Lists depositable vaults and supports the filtering model documented in the README.",
			options: []optionSpec{
				{name: "chain", kind: optionString, usage: "Filter vaults by chain"},
				{name: "asset", kind: optionString, usage: "Filter vaults by deposit asset"},
				{name: "protocol", kind: optionString, usage: "Filter vaults by protocol"},
				{name: "sort", kind: optionString, usage: "Sort by apy, apy30d, tvl, or name"},
				{name: "order", kind: optionString, usage: "Sort order: asc or desc"},
				{name: "min-tvl-usd", kind: optionString, usage: "Minimum TVL in USD"},
				{name: "min-apy", kind: optionString, usage: "Minimum APY percentage"},
				{name: "transactional-only", kind: optionBool, usage: "Only include transactional vaults"},
				{name: "limit", kind: optionString, usage: "Maximum number of results"},
			},
		}),
		newPlaceholderCommand(commandSpec{
			name:        "inspect",
			summary:     "Show full details for a vault",
			usage:       "lifi inspect <vault> [--json]",
			description: "Shows full vault details, including protocol, APY, TVL, and transactional metadata.",
			argsHelp:    "<vault>",
		}),
		newPlaceholderCommand(commandSpec{
			name:        "recommend",
			summary:     "Rank vaults for a target asset",
			usage:       "lifi recommend [--asset <symbol-or-address>] [--from-chain <chain>] [--to-chain <chain>] [--strategy highest-apy|safest|balanced] [--min-tvl-usd <amount>] [--limit <n>] [--json]",
			description: "Ranks vaults for a target asset using the selected strategy.",
			options: []optionSpec{
				{name: "asset", kind: optionString, usage: "Target asset"},
				{name: "from-chain", kind: optionString, usage: "Source chain"},
				{name: "to-chain", kind: optionString, usage: "Destination chain"},
				{name: "strategy", kind: optionString, usage: "Strategy: highest-apy, safest, or balanced"},
				{name: "min-tvl-usd", kind: optionString, usage: "Minimum TVL in USD"},
				{name: "limit", kind: optionString, usage: "Maximum number of results"},
			},
		}),
		newPlaceholderCommand(commandSpec{
			name:        "quote",
			summary:     "Generate a Composer quote for a vault deposit",
			usage:       "lifi quote --vault <address> --from-chain <chain> --from-token <symbol-or-address> --amount <human> --from-address <address> [options]",
			description: "Generates a Composer quote for depositing into a vault.",
			options: []optionSpec{
				{name: "vault", kind: optionString, usage: "Target vault address"},
				{name: "from-chain", kind: optionString, usage: "Source chain"},
				{name: "to-chain", kind: optionString, usage: "Destination chain; defaults to the vault chain"},
				{name: "from-token", kind: optionString, usage: "Source token symbol or address"},
				{name: "amount", kind: optionString, usage: "Human-readable amount"},
				{name: "amount-wei", kind: optionString, usage: "Raw amount in token base units"},
				{name: "from-address", kind: optionString, usage: "Source wallet address"},
				{name: "to-address", kind: optionString, usage: "Destination wallet address"},
				{name: "slippage-bps", kind: optionString, usage: "Allowed slippage in basis points"},
				{name: "preset", kind: optionString, usage: "Quote preset"},
				{name: "allow-bridges", kind: optionString, usage: "Allowlist bridges"},
				{name: "deny-bridges", kind: optionString, usage: "Denylist bridges"},
				{name: "allow-exchanges", kind: optionString, usage: "Allowlist exchanges"},
				{name: "deny-exchanges", kind: optionString, usage: "Denylist exchanges"},
				{name: "raw", kind: optionBool, usage: "Print raw transaction payload details"},
			},
		}),
		newPlaceholderCommand(commandSpec{
			name:        "allowance",
			summary:     "Check token allowance for a wallet and spender",
			usage:       "lifi allowance [--chain <chain>] [--token <symbol-or-address>] [--owner <address>] [--spender <address>] [--amount <human>] [--quote-file <path>] [--json]",
			description: "Checks whether a wallet has enough allowance for a token and spender.",
			options: []optionSpec{
				{name: "chain", kind: optionString, usage: "Chain name or ID"},
				{name: "token", kind: optionString, usage: "Token symbol or address"},
				{name: "owner", kind: optionString, usage: "Owner address"},
				{name: "spender", kind: optionString, usage: "Spender address"},
				{name: "amount", kind: optionString, usage: "Amount to validate"},
				{name: "quote-file", kind: optionString, usage: "Quote payload file"},
			},
		}),
		newPlaceholderCommand(commandSpec{
			name:        "approve",
			summary:     "Send an ERC-20 approval transaction",
			usage:       "lifi approve --chain <chain> --token <symbol-or-address> --spender <address> --amount <human|max> [--yes] [--json]",
			description: "Sends an ERC-20 approval transaction for a token and spender.",
			options: []optionSpec{
				{name: "chain", kind: optionString, usage: "Chain name or ID"},
				{name: "token", kind: optionString, usage: "Token symbol or address"},
				{name: "spender", kind: optionString, usage: "Spender address"},
				{name: "amount", kind: optionString, usage: "Approval amount or max"},
				{name: "yes", kind: optionBool, usage: "Skip confirmation prompts"},
			},
		}),
		newPlaceholderCommand(commandSpec{
			name:        "deposit",
			summary:     "Execute a full Earn deposit flow",
			usage:       "lifi deposit --vault <address> --from-chain <chain> --from-token <symbol-or-address> --amount <human> [options]",
			description: "Executes the full Earn deposit flow from vault resolution to optional verification.",
			options: []optionSpec{
				{name: "vault", kind: optionString, usage: "Target vault address"},
				{name: "from-chain", kind: optionString, usage: "Source chain"},
				{name: "to-chain", kind: optionString, usage: "Destination chain"},
				{name: "from-token", kind: optionString, usage: "Source token symbol or address"},
				{name: "amount", kind: optionString, usage: "Human-readable amount"},
				{name: "from-address", kind: optionString, usage: "Source wallet address"},
				{name: "to-address", kind: optionString, usage: "Destination wallet address"},
				{name: "slippage-bps", kind: optionString, usage: "Allowed slippage in basis points"},
				{name: "approve", kind: optionString, usage: "Approval mode: auto, always, or never"},
				{name: "wait", kind: optionBool, usage: "Wait for the source transaction receipt"},
				{name: "verify-position", kind: optionBool, usage: "Verify the resulting Earn position"},
				{name: "yes", kind: optionBool, usage: "Skip confirmation prompts"},
				{name: "dry-run", kind: optionBool, usage: "Stop after quote and readiness checks"},
			},
		}),
		newPlaceholderCommand(commandSpec{
			name:        "portfolio",
			summary:     "Show Earn positions for an address",
			usage:       "lifi portfolio <address> [--chain <chain>] [--protocol <name>] [--asset <symbol-or-address>] [--json]",
			description: "Shows the current Earn positions for an address.",
			argsHelp:    "<address>",
			options: []optionSpec{
				{name: "chain", kind: optionString, usage: "Filter by chain"},
				{name: "protocol", kind: optionString, usage: "Filter by protocol"},
				{name: "asset", kind: optionString, usage: "Filter by asset"},
			},
		}),
		newPlaceholderCommand(commandSpec{
			name:        "status",
			summary:     "Track LI.FI execution state for a transaction hash",
			usage:       "lifi status --tx-hash <hash> [--from-chain <chain>] [--to-chain <chain>] [--bridge <name>] [--watch] [--interval <duration>] [--json]",
			description: "Tracks LI.FI execution status for a transaction hash.",
			options: []optionSpec{
				{name: "tx-hash", kind: optionString, usage: "Transaction hash"},
				{name: "from-chain", kind: optionString, usage: "Source chain"},
				{name: "to-chain", kind: optionString, usage: "Destination chain"},
				{name: "bridge", kind: optionString, usage: "Bridge name"},
				{name: "watch", kind: optionBool, usage: "Poll for updates continuously"},
				{name: "interval", kind: optionString, usage: "Polling interval"},
			},
		}),
		newConfigCommand(),
		newCompletionCommand(),
		newVersionCommand(),
	}

	result := make(map[string]Command, len(commands))
	for _, cmd := range commands {
		result[cmd.Name()] = cmd
	}

	return result
}

func printRootUsage(commands map[string]Command) {
	names := make([]string, 0, len(commands))
	for name := range commands {
		names = append(names, name)
	}
	sort.Strings(names)

	fmt.Println("lifi")
	fmt.Println()
	fmt.Println("CLI for LI.FI Earn and Composer.")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  lifi [global flags] <command> [command flags] [arguments]")
	fmt.Println()
	fmt.Println("Global flags:")
	fmt.Println("  --config <path>   Path to the config file")
	fmt.Println("  --profile <name>  Config profile name")
	fmt.Println("  --json            Print machine-readable JSON output")
	fmt.Println("  --verbose         Enable verbose output")
	fmt.Println("  --quiet           Reduce non-essential output")
	fmt.Println("  --no-color        Disable ANSI color output")
	fmt.Println()
	fmt.Println("Commands:")

	longest := 0
	for _, name := range names {
		if len(name) > longest {
			longest = len(name)
		}
	}

	for _, name := range names {
		cmd := commands[name]
		fmt.Printf("  %-*s  %s\n", longest, name, cmd.Summary())
	}

	fmt.Println()
	fmt.Println("Run `lifi <command> --help` for command details.")
}

func newFlagSet(name string) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	return fs
}

func formatList(items []string) string {
	if len(items) == 0 {
		return "(none)"
	}

	return strings.Join(items, ", ")
}
