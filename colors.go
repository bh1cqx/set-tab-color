package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
)

//go:embed css-color-names/css-color-names.json
var cssColorsJSON []byte

var cssColors map[string]string

// initColors initializes the CSS color map from embedded JSON
func initColors() error {
	if cssColors == nil {
		if err := json.Unmarshal(cssColorsJSON, &cssColors); err != nil {
			return fmt.Errorf("error parsing embedded css-colors.json: %v", err)
		}
	}
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
