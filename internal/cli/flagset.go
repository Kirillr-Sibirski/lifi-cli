package cli

import (
	"flag"
	"io"
	"strings"
)

type FlagSet struct {
	*flag.FlagSet
	bools   map[string]bool
	strings map[string]bool
	ints    map[string]bool
}

func NewFlagSet(name string) *FlagSet {
	return &FlagSet{
		FlagSet: flag.NewFlagSet(name, flag.ContinueOnError),
		bools:   map[string]bool{},
		strings: map[string]bool{},
		ints:    map[string]bool{},
	}
}

func (f *FlagSet) SetOutput(w io.Writer) {
	f.FlagSet.SetOutput(w)
}

func (f *FlagSet) StringVar(p *string, name, value, usage string) {
	f.FlagSet.StringVar(p, name, value, usage)
	f.strings["--"+name] = true
}

func (f *FlagSet) BoolVar(p *bool, name string, value bool, usage string) {
	f.FlagSet.BoolVar(p, name, value, usage)
	f.bools["--"+name] = true
}

func (f *FlagSet) IntVar(p *int, name string, value int, usage string) {
	f.FlagSet.IntVar(p, name, value, usage)
	f.ints["--"+name] = true
}

func (f *FlagSet) Parse(args []string) error {
	return f.FlagSet.Parse(reorderKnownFlags(args, f.knownFlags()))
}

func (f *FlagSet) knownFlags() map[string]bool {
	flags := map[string]bool{}
	for name := range f.bools {
		flags[name] = false
	}
	for name := range f.strings {
		flags[name] = true
	}
	for name := range f.ints {
		flags[name] = true
	}
	return flags
}

func reorderKnownFlags(args []string, known map[string]bool) []string {
	flags := make([]string, 0, len(args))
	positionals := make([]string, 0, len(args))

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if !strings.HasPrefix(arg, "--") {
			positionals = append(positionals, arg)
			continue
		}

		name := arg
		if idx := strings.Index(arg, "="); idx >= 0 {
			name = arg[:idx]
		}

		takesValue, ok := known[name]
		if !ok {
			positionals = append(positionals, arg)
			continue
		}

		flags = append(flags, arg)
		if takesValue && !strings.Contains(arg, "=") && i+1 < len(args) {
			flags = append(flags, args[i+1])
			i++
		}
	}

	return append(flags, positionals...)
}
