package cli

import (
	"fmt"
	"strings"
)

const lifiBanner = `
 _      _  __ _
| |    (_)/ _(_)
| |     _| |_ _   li.fi earn + composer
| |    | |  _| |  route yield from the terminal
| |____| | | | |
|______|_|_| |_|
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
		return banner
	}
	lines := strings.Split(strings.TrimRight(banner, "\n"), "\n")
	for i, line := range lines {
		switch {
		case i == 0 || i == len(lines)-1:
			lines[i] = colorize(line, "36;1", false)
		case strings.Contains(line, "li.fi earn + composer"):
			lines[i] = colorize(line, "97;1", false)
		case strings.Contains(line, "route yield from the terminal"):
			lines[i] = colorize(line, "90", false)
		default:
			lines[i] = colorize(line, "34;1", false)
		}
	}
	return strings.Join(lines, "\n") + "\n"
}

func sectionTitle(title string, noColor bool) string {
	label := "/// " + strings.ToUpper(strings.TrimSpace(title))
	if noColor {
		return label
	}
	return colorize(label, "36;1", false)
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
	return colorize(value, "97;1", noColor)
}

func subtleValue(value string, noColor bool) string {
	return colorize(value, "90", noColor)
}
