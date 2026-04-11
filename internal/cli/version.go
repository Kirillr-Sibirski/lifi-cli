package cli

import (
	"fmt"

	"github.com/Kirillr-Sibirski/defi-mullet/internal/config"
)

type versionCommand struct{}

func newVersionCommand() Command {
	return versionCommand{}
}

func (versionCommand) Name() string {
	return "version"
}

func (versionCommand) Summary() string {
	return "Print the lifi CLI version"
}

func (versionCommand) Usage() string {
	return "lifi version"
}

func (versionCommand) Run(cfg *config.Config, args []string) error {
	if cfg.Global.JSON {
		return writeJSON(map[string]string{
			"name":    "lifi",
			"version": version,
		})
	}

	fmt.Printf("lifi %s\n", version)
	return nil
}

type osStdout struct{}

func (osStdout) Write(p []byte) (int, error) {
	return fmt.Print(string(p))
}
