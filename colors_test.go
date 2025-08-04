package main

import (
	"encoding/json"
	"testing"
)

// TestNormalizeColor tests the color normalization function
func TestNormalizeColor(t *testing.T) {
	// Initialize cssColors for testing
	if err := json.Unmarshal(cssColorsJSON, &cssColors); err != nil {
		t.Fatalf("Failed to parse embedded CSS colors: %v", err)
	}

	tests := []struct {
		input    string
		expected string
	}{
		// Hex colors
		{"#ff0000", "ff0000"},
		{"ff0000", "ff0000"},
		{"#f80", "ff8800"},
		{"f80", "ff8800"},
		{"#FF0000", "ff0000"}, // uppercase
		{"FF0000", "ff0000"},  // uppercase without #

		// CSS color names (testing a few known ones)
		{"red", "ff0000"},
		{"blue", "0000ff"},
		{"green", "008000"},
		{"white", "ffffff"},
		{"black", "000000"},

		// Special case
		{"default", "default"},

		// Invalid colors
		{"invalid", ""},
		{"#gg0000", ""},
		{"#ff00", ""}, // wrong length
	}

	for _, test := range tests {
		result := normalizeColor(test.input)
		if result != test.expected {
			t.Errorf("normalizeColor(%q) = %q, expected %q", test.input, result, test.expected)
		}
	}
}

// TestExpandHex3 tests the 3-digit hex expansion
func TestExpandHex3(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"f80", "ff8800"},
		{"123", "112233"},
		{"abc", "aabbcc"},
		{"000", "000000"},
		{"fff", "ffffff"},
	}

	for _, test := range tests {
		result := expandHex3(test.input)
		if result != test.expected {
			t.Errorf("expandHex3(%q) = %q, expected %q", test.input, result, test.expected)
		}
	}
}

// TestIsHex tests the hex validation function
func TestIsHex(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"ff0000", true},
		{"123abc", true},
		{"000000", true},
		{"ffffff", true},
		{"gg0000", false},
		{"ff00zz", false},
		{"", true}, // empty string is valid (edge case)
	}

	for _, test := range tests {
		result := isHex(test.input)
		if result != test.expected {
			t.Errorf("isHex(%q) = %v, expected %v", test.input, result, test.expected)
		}
	}
}

// TestInitColors tests the color map initialization
func TestInitColors(t *testing.T) {
	// Reset cssColors to test initialization
	originalColors := cssColors
	cssColors = nil
	defer func() { cssColors = originalColors }()

	err := initColors()
	if err != nil {
		t.Fatalf("initColors() failed: %v", err)
	}

	if cssColors == nil {
		t.Error("cssColors should be initialized after initColors()")
	}

	// Test that some known colors exist
	knownColors := []string{"red", "blue", "green", "white", "black"}
	for _, color := range knownColors {
		if _, ok := cssColors[color]; !ok {
			t.Errorf("Expected color %q to be in cssColors map", color)
		}
	}
}
