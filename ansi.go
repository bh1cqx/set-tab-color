package main

import (
	"fmt"
	"strconv"
)

// hexToRGB converts a hex color string to RGB values
func hexToRGB(hex string) (r, g, b int, err error) {
	// Remove # prefix if present
	if hex[0] == '#' {
		hex = hex[1:]
	}

	// Parse hex values
	if len(hex) != 6 {
		return 0, 0, 0, fmt.Errorf("invalid hex color length")
	}

	rVal, err := strconv.ParseInt(hex[0:2], 16, 0)
	if err != nil {
		return 0, 0, 0, err
	}

	gVal, err := strconv.ParseInt(hex[2:4], 16, 0)
	if err != nil {
		return 0, 0, 0, err
	}

	bVal, err := strconv.ParseInt(hex[4:6], 16, 0)
	if err != nil {
		return 0, 0, 0, err
	}

	return int(rVal), int(gVal), int(bVal), nil
}

// colorText applies ANSI color formatting to text using hex color
func colorText(text, hexColor string) string {
	r, g, b, err := hexToRGB(hexColor)
	if err != nil {
		// If color conversion fails, return uncolored text
		return text
	}

	// Use 24-bit RGB color escape sequence: \033[38;2;r;g;bm for foreground
	// Reset sequence: \033[0m
	return fmt.Sprintf("\033[38;2;%d;%d;%dm%s\033[0m", r, g, b, text)
}
