package cli

import (
	"fmt"
	"strings"
)

const lifiBanner = `
   ██████                     ██████
  ████████━━━━━━━━━━━━━━━━━━━████████
  ████████  c r o s s ─ c h a i n
  ████████━━━━━━━━━━━━━━━━━━━████████
   ██████                     ██████

 __    ___ _____ ___    ____ _     ___
|  |  |_ _|  ___|_ _|  / ___| |   |_ _|
|  |   | || |_   | |  | |   | |    | |
|  |___| ||  _|  | |  | |___| |___ | |
|______|___|_|   |___|  \____|_____|___|
`

func colorize(value, code string, noColor bool) string {
	if noColor || strings.TrimSpace(value) == "" {
		return value
	}
	return "\x1b[" + code + "m" + value + "\x1b[0m"
}

func brandBanner(noColor bool) string {
	if noColor {
		banner := strings.TrimLeft(lifiBanner, "\n")
		return banner + "li.fi cli | earn + composer for builders\n"
	}

	p1 := func(s string) string { return colorize(s, "38;5;213;1", false) }
	p2 := func(s string) string { return colorize(s, "38;5;205;1", false) }
	dim := func(s string) string { return colorize(s, "90", false) }

	bridge := strings.Join([]string{
		"   " + p1("██████") + "                     " + p1("██████"),
		"  " + p2("████████") + dim("━━━━━━━━━━━━━━━━━━━") + p2("████████"),
		"  " + p2("████████") + "  " + dim("c r o s s ─ c h a i n"),
		"  " + p2("████████") + dim("━━━━━━━━━━━━━━━━━━━") + p2("████████"),
		"   " + p1("██████") + "                     " + p1("██████"),
	}, "\n")

	logo := strings.Join([]string{
		p1(` __    ___ _____ ___    ____ _     ___`),
		p2(`|  |  |_ _|  ___|_ _|  / ___| |   |_ _|`),
		p1(`|  |   | || |_   | |  | |   | |    | |`),
		p2(`|  |___| ||  _|  | |  | |___| |___ | |`),
		p1(`|______|___|_|   |___|  \____|_____|___|`),
	}, "\n")

	tagline := colorize("li.fi cli", "97;1", false) + " " + colorize("| earn + composer for builders", "38;5;250", false)
	return bridge + "\n\n" + logo + "\n" + tagline + "\n"
}

func sectionTitle(title string, noColor bool) string {
	label := "/// " + strings.ToUpper(strings.TrimSpace(title))
	if noColor {
		return label
	}
	return colorize(label, "38;5;205;1", false)
}

func sectionRule(noColor bool) string {
	rule := strings.Repeat("—", 32)
	if noColor {
		rule = strings.Repeat("-", 32)
	}
	return colorize(rule, "90", noColor)
}

func printSectionHeader(title string, noColor bool) {
	fmt.Println(sectionTitle(title, noColor))
	fmt.Println(sectionRule(noColor))
}

func accentValue(value string, noColor bool) string {
	return colorize(value, "38;5;219;1", noColor)
}

func subtleValue(value string, noColor bool) string {
	return colorize(value, "90", noColor)
}

func statusLabel(status string, noColor bool) string {
	label := "[" + strings.ToLower(strings.TrimSpace(status)) + "]"
	if noColor {
		return label
	}
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "ok", "yes", "done", "completed":
		return colorize(label, "32;1", false)
	case "warn", "pending":
		return colorize(label, "33;1", false)
	case "fail", "error", "no":
		return colorize(label, "31;1", false)
	default:
		return colorize(label, "38;5;219;1", false)
	}
}
