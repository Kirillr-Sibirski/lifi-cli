package cli

import (
	"fmt"
	"strings"
)

const lifiBanner = `
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
	banner := strings.TrimLeft(lifiBanner, "\n")
	if noColor {
		return banner + "li.fi cli | earn + composer for builders\n"
	}
	lines := strings.Split(strings.TrimRight(banner, "\n"), "\n")
	for i, line := range lines {
		switch {
		case i == 0 || i == len(lines)-1:
			lines[i] = colorize(line, "38;5;213;1", false)
		default:
			lines[i] = colorize(line, "38;5;205;1", false)
		}
	}
	tagline := colorize("li.fi cli", "97;1", false) + " " + colorize("| earn + composer for builders", "38;5;250", false)
	return strings.Join(lines, "\n") + "\n" + tagline + "\n"
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
