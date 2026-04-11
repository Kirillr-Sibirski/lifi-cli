package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"strings"

	"github.com/Kirillr-Sibirski/defi-mullet/internal/config"
)

type optionKind string

const (
	optionString optionKind = "string"
	optionBool   optionKind = "bool"
)

type optionSpec struct {
	name  string
	kind  optionKind
	usage string
}

type commandSpec struct {
	name        string
	summary     string
	usage       string
	description string
	argsHelp    string
	options     []optionSpec
}

type placeholderCommand struct {
	spec commandSpec
}

func newPlaceholderCommand(spec commandSpec) Command {
	return placeholderCommand{spec: spec}
}

func (p placeholderCommand) Name() string {
	return p.spec.name
}

func (p placeholderCommand) Summary() string {
	return p.spec.summary
}

func (p placeholderCommand) Usage() string {
	return p.spec.usage
}

func (p placeholderCommand) Run(cfg *config.Config, args []string) error {
	fs := newFlagSet(p.spec.name)
	stringValues := map[string]*string{}
	boolValues := map[string]*bool{}

	for _, option := range p.spec.options {
		switch option.kind {
		case optionString:
			stringValues[option.name] = fs.String(option.name, "", option.usage)
		case optionBool:
			boolValues[option.name] = fs.Bool(option.name, false, option.usage)
		}
	}

	fs.Usage = func() {
		fmt.Println("Usage:")
		fmt.Println("  " + p.spec.usage)
		if p.spec.description != "" {
			fmt.Println()
			fmt.Println(p.spec.description)
		}
	}

	if err := fs.Parse(args); err != nil {
		return err
	}

	if (p.spec.name == "inspect" || p.spec.name == "portfolio") && fs.NArg() == 0 {
		return fmt.Errorf("%s requires %s", p.spec.name, p.spec.argsHelp)
	}

	payload := map[string]any{
		"command":     p.spec.name,
		"summary":     p.spec.summary,
		"description": p.spec.description,
		"args":        fs.Args(),
		"string_flags": func() map[string]string {
			result := map[string]string{}
			for name, value := range stringValues {
				if *value != "" {
					result[name] = *value
				}
			}
			return result
		}(),
		"bool_flags": func() map[string]bool {
			result := map[string]bool{}
			for name, value := range boolValues {
				if *value {
					result[name] = *value
				}
			}
			return result
		}(),
		"implemented": false,
		"next_step":   "command scaffold is ready; LI.FI API integration comes next",
	}

	if cfg.Global.JSON {
		encoder := json.NewEncoder(osStdout{})
		encoder.SetIndent("", "  ")
		return encoder.Encode(payload)
	}

	fmt.Printf("%s\n\n", p.spec.summary)
	if p.spec.description != "" {
		fmt.Println(p.spec.description)
		fmt.Println()
	}
	fmt.Printf("Usage: %s\n", p.spec.usage)

	if len(fs.Args()) > 0 {
		fmt.Printf("Arguments: %s\n", strings.Join(fs.Args(), ", "))
	}

	if len(stringValues) > 0 || len(boolValues) > 0 {
		fmt.Println("Parsed flags:")

		fs.Visit(func(f *flag.Flag) {
			fmt.Printf("  --%s=%s\n", f.Name, f.Value.String())
		})
	}

	fmt.Println()
	fmt.Println("Command scaffold is ready; LI.FI API integration comes next.")
	return nil
}
