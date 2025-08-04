package main

import (
	"os"
	"path/filepath"
	"testing"
)

// TestListProfileNames tests listing profile names
func TestListProfileNames(t *testing.T) {
	// Create temporary config file
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "test-config.toml")

	configContent := `
[profiles.development]
tab = "blue"
fg = "white"

[profiles.production]
tab = "red"

[profiles.test-profile]
bg = "black"
`

	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Set environment variable to use test config
	originalEnv := os.Getenv("SET_TAB_COLOR_CONFIG")
	os.Setenv("SET_TAB_COLOR_CONFIG", configFile)
	defer func() {
		if originalEnv == "" {
			os.Unsetenv("SET_TAB_COLOR_CONFIG")
		} else {
			os.Setenv("SET_TAB_COLOR_CONFIG", originalEnv)
		}
	}()

	profiles, err := listProfileNames()
	if err != nil {
		t.Fatalf("listProfileNames() failed: %v", err)
	}

	expectedProfiles := []string{"development", "production", "test-profile"}
	if len(profiles) != len(expectedProfiles) {
		t.Errorf("Expected %d profiles, got %d", len(expectedProfiles), len(profiles))
	}

	// Check that all expected profiles are present
	profileMap := make(map[string]bool)
	for _, profile := range profiles {
		profileMap[profile] = true
	}

	for _, expected := range expectedProfiles {
		if !profileMap[expected] {
			t.Errorf("Expected profile %q not found in results", expected)
		}
	}
}

// TestListProfileNamesEmpty tests listing when no profiles exist
func TestListProfileNamesEmpty(t *testing.T) {
	// Set environment variable to non-existent file
	nonExistentPath := "/tmp/non-existent-config.toml"
	originalEnv := os.Getenv("SET_TAB_COLOR_CONFIG")
	os.Setenv("SET_TAB_COLOR_CONFIG", nonExistentPath)
	defer func() {
		if originalEnv == "" {
			os.Unsetenv("SET_TAB_COLOR_CONFIG")
		} else {
			os.Setenv("SET_TAB_COLOR_CONFIG", originalEnv)
		}
	}()

	profiles, err := listProfileNames()
	if err != nil {
		t.Fatalf("listProfileNames() should not fail for missing config: %v", err)
	}

	if len(profiles) != 0 {
		t.Errorf("Expected 0 profiles for empty config, got %d", len(profiles))
	}
}

// TestListCSSColorNames tests listing CSS color names
func TestListCSSColorNames(t *testing.T) {
	colors, err := listCSSColorNames()
	if err != nil {
		t.Fatalf("listCSSColorNames() failed: %v", err)
	}

	if len(colors) == 0 {
		t.Error("Expected CSS colors to be loaded, got 0")
	}

	// Check that some known colors are present
	knownColors := []string{"red", "blue", "green", "white", "black"}
	colorMap := make(map[string]bool)
	for _, color := range colors {
		colorMap[color] = true
	}

	for _, expected := range knownColors {
		if !colorMap[expected] {
			t.Errorf("Expected CSS color %q not found in results", expected)
		}
	}

	// Verify colors are just names, not hex values
	for _, color := range colors {
		if contains(color, "#") {
			t.Errorf("Color name should not contain hex values, got: %s", color)
		}
	}
}
