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

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s COLOR\n", os.Args[0])
		os.Exit(1)
	}

	// Parse embedded CSS color map
	if err := json.Unmarshal(cssColorsJSON, &cssColors); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing embedded css-colors.json: %v\n", err)
		os.Exit(1)
	}

	// Normalize user input
	color := normalizeColor(os.Args[1])
	if color == "" {
		fmt.Fprintf(os.Stderr, "Unknown color: %s\n", os.Args[1])
		os.Exit(1)
	}

	// Locate custom it2setcolor in ~/.iterm2/
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not get home dir: %v\n", err)
		os.Exit(1)
	}
	it2bin := filepath.Join(home, ".iterm2", "it2setcolor")

	// Execute it2setcolor with the normalized hex
	cmd := exec.Command(it2bin, "tab", color)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to run %s: %v\n", it2bin, err)
		os.Exit(1)
	}
}
