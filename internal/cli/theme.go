package cli

import (
	"fmt"
	"strings"
)

// lifiBannerLines is the "li.fi cli" logo in DOS Rebel block-font style.
var lifiBannerLines = []string{
	`█████        ███  ███████████  ███       █████████  █████       █████`,
	`░░███        ░░░  ░░███░░░░░░█ ░░░       ███░░░░░███░░███       ░░███ `,
	` ░███        ████  ░███   █ ░  ████     ███     ░░░  ░███        ░███ `,
	` ░███       ░░███  ░███████   ░░███    ░███          ░███        ░███ `,
	` ░███        ░███  ░███░░░█    ░███    ░███          ░███        ░███ `,
	` ░███      █ ░███  ░███  ░     ░███    ░░███     ███ ░███      █ ░███ `,
	` ███████████ █████ █████       █████    ░░█████████  ███████████ █████`,
	`░░░░░░░░░░░ ░░░░░ ░░░░░       ░░░░░      ░░░░░░░░░  ░░░░░░░░░░░ ░░░░░`,
}

func colorize(value, code string, noColor bool) string {
	if noColor || strings.TrimSpace(value) == "" {
		return value
	}
	return "\x1b[" + code + "m" + value + "\x1b[0m"
}

func brandBanner(noColor bool) string {
	tagline := "li.fi cli | earn + composer for builders"
	if noColor {
		return strings.Join(lifiBannerLines, "\n") + "\n" + tagline + "\n"
	}

	p1 := func(s string) string { return colorize(s, "38;5;213;1", false) }
	p2 := func(s string) string { return colorize(s, "38;5;205;1", false) }

	colored := make([]string, len(lifiBannerLines))
	for i, line := range lifiBannerLines {
		if i%2 == 0 {
			colored[i] = p1(line)
		} else {
			colored[i] = p2(line)
		}
	}

	styledTagline := colorize("li.fi cli", "97;1", false) + " " + colorize("| earn + composer for builders", "38;5;250", false)
	return strings.Join(colored, "\n") + "\n" + styledTagline + "\n"
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
