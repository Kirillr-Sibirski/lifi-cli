package cli

import (
	"fmt"

	"github.com/Kirillr-Sibirski/defi-mullet/internal/config"
)

type completionCommand struct{}

func newCompletionCommand() Command {
	return completionCommand{}
}

func (completionCommand) Name() string {
	return "completion"
}

func (completionCommand) Summary() string {
	return "Show shell completion setup guidance"
}

func (completionCommand) Usage() string {
	return "lifi completion <bash|zsh|fish>"
}

func (completionCommand) Run(cfg *config.Config, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("completion requires a shell argument")
	}

	shell := args[0]
	switch shell {
	case "bash", "zsh", "fish":
	default:
		return fmt.Errorf("unsupported shell %q", shell)
	}

	fmt.Printf("Shell completion generation for %s is scaffolded and will be added next.\n", shell)
	return nil
}
