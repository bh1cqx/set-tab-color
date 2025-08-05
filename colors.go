package main

import (
	"strings"

	"github.com/bh1cqx/set-tab-color/generated"
)

var cssColors = generated.CSSColors

// initColors is no longer needed since cssColors is initialized directly
func initColors() error {
	return nil
}

// expandHex3 expands shorthand hex (#f80) → full hex (ff8800)
func expandHex3(s string) string {
	return strings.Repeat(string(s[0]), 2) +
		strings.Repeat(string(s[1]), 2) +
		strings.Repeat(string(s[2]), 2)
}

// isHex reports whether s consists only of lowercase hex digits
func isHex(s string) bool {
	for _, c := range s {
		if !strings.Contains("0123456789abcdef", string(c)) {
			return false
		}
	}
	return true
}

// normalizeColor handles #RGB, #RRGGBB, CSS names, and "default"
func normalizeColor(input string) string {
	clean := strings.ToLower(strings.TrimPrefix(input, "#"))
	if clean == "default" {
		return "default"
	}
	if len(clean) == 3 && isHex(clean) {
		return expandHex3(clean)
	}
	if len(clean) == 6 && isHex(clean) {
		return clean
	}
	if hex, ok := cssColors[clean]; ok {
		return strings.TrimPrefix(hex, "#")
	}
	return ""
}

// listCSSColorNames returns a list of all available CSS color names
func listCSSColorNames() ([]string, error) {
	// Initialize CSS colors if not already done
	if err := initColors(); err != nil {
		return nil, err
	}

	names := make([]string, 0, len(cssColors))
	for name := range cssColors {
		names = append(names, name)
	}

	return names, nil
}

// listCSSColorNamesFormatted returns a comma-separated string of all available CSS color names
// with each name colored according to its actual color value
func listCSSColorNamesFormatted() (string, error) {
	// Initialize CSS colors if not already done
	if err := initColors(); err != nil {
		return "", err
	}

	coloredNames := make([]string, 0, len(cssColors))
	for name, hexValue := range cssColors {
		// Remove the # prefix from hex value for our colorText function
		hex := strings.TrimPrefix(hexValue, "#")
		coloredName := colorText(name, hex)
		coloredNames = append(coloredNames, coloredName)
	}

	return strings.Join(coloredNames, ", "), nil
}
