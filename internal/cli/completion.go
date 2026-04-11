package cli

import (
	"fmt"
	"sort"
	"strings"

	"github.com/Kirillr-Sibirski/lifi-cli/internal/config"
)

type completionCommand struct{}

func newCompletionCommand() Command {
	return completionCommand{}
}

func (completionCommand) Name() string {
	return "completion"
}

func (completionCommand) Summary() string {
	return "Generate shell completion scripts"
}

func (completionCommand) Usage() string {
	return "lifi completion <bash|zsh|fish>"
}

func (completionCommand) Run(cfg *config.Config, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("completion requires a shell argument")
	}

	switch args[0] {
	case "bash":
		fmt.Print(bashCompletionScript())
	case "zsh":
		fmt.Print(zshCompletionScript())
	case "fish":
		fmt.Print(fishCompletionScript())
	default:
		return fmt.Errorf("unsupported shell %q", args[0])
	}

	return nil
}

func completionSpecs() map[string][]string {
	return map[string][]string{
		"doctor": {
			"--json", "--write-checks", "--chain", "--rpc-url",
		},
		"chains": {
			"--search", "--evm-only", "--json",
		},
		"protocols": {
			"--search", "--supports", "--json",
		},
		"tokens": {
			"--chain", "--token", "--tags", "--json",
		},
		"vaults": {
			"--chain", "--asset", "--protocol", "--sort", "--order",
			"--min-tvl-usd", "--min-apy", "--transactional-only", "--limit", "--json",
		},
		"inspect": {
			"--json",
		},
		"recommend": {
			"--asset", "--from-chain", "--to-chain", "--strategy", "--min-tvl-usd", "--limit", "--json",
		},
		"quote": {
			"--vault", "--from-chain", "--to-chain", "--from-token", "--amount",
			"--amount-wei", "--from-address", "--to-address", "--slippage-bps",
			"--preset", "--allow-bridges", "--deny-bridges", "--allow-exchanges",
			"--deny-exchanges", "--json", "--raw", "--unsigned",
		},
		"allowance": {
			"--chain", "--token", "--owner", "--spender", "--amount", "--quote-file", "--json",
		},
		"approve": {
			"--chain", "--token", "--spender", "--amount", "--gas-policy", "--yes", "--json",
		},
		"deposit": {
			"--vault", "--from-chain", "--to-chain", "--from-token", "--amount",
			"--from-address", "--to-address", "--slippage-bps", "--approve",
			"--approval-amount", "--gas-policy", "--wait-timeout", "--portfolio-timeout",
			"--wait", "--verify-position", "--yes", "--dry-run", "--simulate",
			"--skip-simulate", "--json",
		},
		"portfolio": {
			"--chain", "--protocol", "--asset", "--json",
		},
		"status": {
			"--tx-hash", "--from-chain", "--to-chain", "--bridge", "--watch", "--interval", "--json",
		},
		"config": {
			"init", "show",
		},
		"completion": {
			"bash", "zsh", "fish",
		},
		"version": {},
	}
}

func completionCommands() []string {
	specs := completionSpecs()
	commands := make([]string, 0, len(specs))
	for name := range specs {
		commands = append(commands, name)
	}
	sort.Strings(commands)
	return commands
}

func bashCompletionScript() string {
	commandList := strings.Join(completionCommands(), " ")
	lines := []string{
		"_lifi_completions() {",
		"  local cur prev cmd",
		"  COMPREPLY=()",
		"  cur=\"${COMP_WORDS[COMP_CWORD]}\"",
		"  prev=\"${COMP_WORDS[COMP_CWORD-1]}\"",
		"  cmd=\"${COMP_WORDS[1]}\"",
		"",
		"  if [[ ${COMP_CWORD} -eq 1 ]]; then",
		fmt.Sprintf("    COMPREPLY=( $(compgen -W \"%s\" -- \"$cur\") )", commandList),
		"    return 0",
		"  fi",
		"",
		"  if [[ \"$cmd\" == \"config\" && ${COMP_CWORD} -eq 2 ]]; then",
		"    COMPREPLY=( $(compgen -W \"init show\" -- \"$cur\") )",
		"    return 0",
		"  fi",
		"",
		"  if [[ \"$cmd\" == \"completion\" && ${COMP_CWORD} -eq 2 ]]; then",
		"    COMPREPLY=( $(compgen -W \"bash zsh fish\" -- \"$cur\") )",
		"    return 0",
		"  fi",
		"",
		"  case \"$cmd\" in",
	}

	for _, command := range completionCommands() {
		spec := completionSpecs()[command]
		if len(spec) == 0 {
			continue
		}
		lines = append(lines,
			fmt.Sprintf("    %s)", command),
			fmt.Sprintf("      COMPREPLY=( $(compgen -W \"%s\" -- \"$cur\") )", strings.Join(spec, " ")),
			"      return 0",
			"      ;;",
		)
	}

	lines = append(lines,
		"  esac",
		"}",
		"complete -F _lifi_completions lifi",
	)

	return strings.Join(lines, "\n") + "\n"
}

func zshCompletionScript() string {
	return strings.TrimSpace(`
#compdef lifi

autoload -U +X bashcompinit && bashcompinit
`+"\n"+bashCompletionScript()) + "\n"
}

func fishCompletionScript() string {
	lines := []string{
		"complete -c lifi -f",
	}

	for _, command := range completionCommands() {
		lines = append(lines, fmt.Sprintf("complete -c lifi -n '__fish_use_subcommand' -a '%s'", command))
	}

	for command, values := range completionSpecs() {
		if len(values) == 0 {
			continue
		}
		if command == "config" || command == "completion" {
			for _, value := range values {
				lines = append(lines, fmt.Sprintf("complete -c lifi -n '__fish_seen_subcommand_from %s' -a '%s'", command, value))
			}
			continue
		}
		for _, value := range values {
			lines = append(lines, fmt.Sprintf("complete -c lifi -n '__fish_seen_subcommand_from %s' -l '%s'", command, strings.TrimPrefix(value, "--")))
		}
	}

	return strings.Join(lines, "\n") + "\n"
}
