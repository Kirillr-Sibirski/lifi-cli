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
