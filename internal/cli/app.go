package cli

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/Kirillr-Sibirski/lifi-cli/internal/apperror"
	"github.com/Kirillr-Sibirski/lifi-cli/internal/config"
)

var version = "0.1.4"

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
	tableNoColor = global.NoColor
	if err != nil {
		if global.JSON {
			_ = writeJSON(apperror.JSONPayload(apperror.Wrap("config", apperror.ExitConfig, err)))
		} else {
			fmt.Fprintln(os.Stderr, err)
		}
		return apperror.ExitConfig
	}

	if len(remaining) == 0 {
		printRootUsage(commands, cfg.Global.NoColor)
		return 0
	}

	name := remaining[0]
	if name == "help" {
		printRootUsage(commands, cfg.Global.NoColor)
		return 0
	}

	cmd, ok := commands[name]
	if !ok {
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n", name)
		printRootUsage(commands, cfg.Global.NoColor)
		return 2
	}

	if err := cmd.Run(cfg, remaining[1:]); err != nil {
		appErr := apperror.Classify(err)
		if cfg.Global.JSON {
			_ = writeJSON(apperror.JSONPayload(appErr))
		} else {
			fmt.Fprintln(os.Stderr, appErr.Message)
		}
		return appErr.ExitCode()
	}

	return 0
}

func parseGlobalOptions(args []string) (config.GlobalOptions, []string, error) {
	global := config.GlobalOptions{
		Profile: "default",
	}
	remaining := make([]string, 0, len(args))

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--config":
			if i+1 >= len(args) {
				return config.GlobalOptions{}, nil, fmt.Errorf("missing value for --config")
			}
			i++
			global.ConfigPath = args[i]
		case strings.HasPrefix(arg, "--config="):
			global.ConfigPath = strings.TrimPrefix(arg, "--config=")
		case arg == "--profile":
			if i+1 >= len(args) {
				return config.GlobalOptions{}, nil, fmt.Errorf("missing value for --profile")
			}
			i++
			global.Profile = args[i]
		case strings.HasPrefix(arg, "--profile="):
			global.Profile = strings.TrimPrefix(arg, "--profile=")
		case arg == "--json":
			global.JSON = true
		case arg == "--verbose":
			global.Verbose = true
		case arg == "--quiet":
			global.Quiet = true
		case arg == "--no-color":
			global.NoColor = true
		default:
			remaining = append(remaining, arg)
		}
	}

	return global, remaining, nil
}

func builtInCommands() map[string]Command {
	commands := []Command{
		newDoctorCommand(),
		newChainsCommand(),
		newProtocolsCommand(),
		newTokensCommand(),
		newVaultsCommand(),
		newInspectCommand(),
		newRecommendCommand(),
		newQuoteCommand(),
		newAllowanceCommand(),
		newApproveCommand(),
		newDepositCommand(),
		newPortfolioCommand(),
		newStatusCommand(),
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

func printRootUsage(commands map[string]Command, noColor bool) {
	names := make([]string, 0, len(commands))
	for name := range commands {
		names = append(names, name)
	}
	sort.Strings(names)

	fmt.Print(brandBanner(noColor))
	fmt.Println()
	printSectionHeader("Usage", noColor)
	fmt.Println("  lifi [global flags] <command> [command flags] [arguments]")
	fmt.Println()
	printSectionHeader("Global Flags", noColor)
	fmt.Println("  --config <path>   Path to the config file")
	fmt.Println("  --profile <name>  Config profile name")
	fmt.Println("  --json            Print machine-readable JSON output")
	fmt.Println("  --verbose         Enable verbose output")
	fmt.Println("  --quiet           Reduce non-essential output")
	fmt.Println("  --no-color        Disable ANSI color output")
	fmt.Println()
	printSectionHeader("Commands", noColor)

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
	fmt.Println(subtleValue("Run `lifi <command> --help` for command details.", noColor))
}

func newFlagSet(name string) *FlagSet {
	fs := NewFlagSet(name)
	fs.SetOutput(os.Stderr)
	return fs
}

func formatList(items []string) string {
	if len(items) == 0 {
		return "(none)"
	}

	return strings.Join(items, ", ")
}
