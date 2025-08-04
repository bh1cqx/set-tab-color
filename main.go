package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

//go:embed css-color-names/css-color-names.json
var cssColorsJSON []byte

var cssColors map[string]string

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

// runSetTabColor executes it2setcolor with the given color
func runSetTabColor(color string) error {
	// Parse embedded CSS color map
	if cssColors == nil {
		if err := json.Unmarshal(cssColorsJSON, &cssColors); err != nil {
			return fmt.Errorf("error parsing embedded css-colors.json: %v", err)
		}
	}

	// Normalize user input
	normalizedColor := normalizeColor(color)
	if normalizedColor == "" {
		return fmt.Errorf("unknown color: %s", color)
	}

	// Locate and check existence of custom it2setcolor in ~/.iterm2/
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("could not get home dir: %v", err)
	}
	it2bin := filepath.Join(home, ".iterm2", "it2setcolor")

	if _, err := os.Stat(it2bin); os.IsNotExist(err) {
		return fmt.Errorf("it2setcolor not found at %s", it2bin)
	}

	// Execute it2setcolor with the normalized hex
	cmd := exec.Command(it2bin, "tab", normalizedColor)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s COLOR\n", os.Args[0])
		os.Exit(1)
	}

	if err := runSetTabColor(os.Args[1]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
