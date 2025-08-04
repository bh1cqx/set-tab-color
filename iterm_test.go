package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestRunSetColor tests the iTerm2 integration with mocked binary
func TestRunSetColor(t *testing.T) {
	// Initialize cssColors for testing
	if err := json.Unmarshal(cssColorsJSON, &cssColors); err != nil {
		t.Fatalf("Failed to parse embedded CSS colors: %v", err)
	}

	tests := []struct {
		name          string
		target        ColorTarget
		input         string
		expectedArgs  []string
		shouldError   bool
		errorContains string
	}{
		{
			name:         "tab hex color",
			target:       TabColor,
			input:        "#ff0000",
			expectedArgs: []string{"tab", "ff0000"},
			shouldError:  false,
		},
		{
			name:         "foreground short hex",
			target:       ForegroundColor,
			input:        "#f80",
			expectedArgs: []string{"fg", "ff8800"},
			shouldError:  false,
		},
		{
			name:         "background css color name",
			target:       BackgroundColor,
			input:        "red",
			expectedArgs: []string{"bg", "ff0000"},
			shouldError:  false,
		},
		{
			name:         "tab default color",
			target:       TabColor,
			input:        "default",
			expectedArgs: []string{"tab", "default"},
			shouldError:  false,
		},
		{
			name:          "invalid color",
			target:        TabColor,
			input:         "invalidcolor",
			expectedArgs:  nil,
			shouldError:   true,
			errorContains: "unknown color",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Create temp directory for mock setup
			tempDir := t.TempDir()

			// Mock the home directory
			originalHome := os.Getenv("HOME")
			os.Setenv("HOME", tempDir)
			defer os.Setenv("HOME", originalHome)

			// Create .iterm2 directory and mock binary
			iterm2Dir := filepath.Join(tempDir, ".iterm2")
			if err := os.MkdirAll(iterm2Dir, 0755); err != nil {
				t.Fatalf("Failed to create .iterm2 directory: %v", err)
			}

			mockBinary := filepath.Join(iterm2Dir, "it2setcolor")
			// Create a simple script that just exits successfully
			if err := os.WriteFile(mockBinary, []byte("#!/bin/sh\nexit 0\n"), 0755); err != nil {
				t.Fatalf("Failed to create mock binary: %v", err)
			}

			// Test the function - this will execute the command but with our mock binary
			err := runSetColor(test.target, test.input)

			if test.shouldError {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if test.errorContains != "" && !contains(err.Error(), test.errorContains) {
					t.Errorf("Expected error to contain %q, got %q", test.errorContains, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// Since we can't easily mock exec.Command directly in Go without changing the main code,
			// we verify the logic by testing the color normalization separately
			// The integration test above ensures the full flow works with our mock binary
			normalizedColor := normalizeColor(test.input)
			if normalizedColor != test.expectedArgs[1] {
				t.Errorf("Expected normalized color %q, got %q", test.expectedArgs[1], normalizedColor)
			}
		})
	}
}

// TestRunSetColorMissingBinary tests behavior when it2setcolor is missing
func TestRunSetColorMissingBinary(t *testing.T) {
	// Initialize cssColors for testing
	if err := json.Unmarshal(cssColorsJSON, &cssColors); err != nil {
		t.Fatalf("Failed to parse embedded CSS colors: %v", err)
	}

	// Create temp directory without the binary
	tempDir := t.TempDir()

	// Mock the home directory to a location without it2setcolor
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Test with valid color but missing binary
	err := runSetColor(TabColor, "red")
	if err == nil {
		t.Errorf("Expected error for missing binary, got none")
		return
	}

	expectedError := "it2setcolor not found"
	if !contains(err.Error(), expectedError) {
		t.Errorf("Expected error to contain %q, got %q", expectedError, err.Error())
	}
}

// TestColorTarget tests the ColorTarget enum values
func TestColorTarget(t *testing.T) {
	tests := []struct {
		target   ColorTarget
		expected string
	}{
		{TabColor, "tab"},
		{ForegroundColor, "fg"},
		{BackgroundColor, "bg"},
	}

	for _, test := range tests {
		if string(test.target) != test.expected {
			t.Errorf("ColorTarget %v = %q, expected %q", test.target, string(test.target), test.expected)
		}
	}
}
