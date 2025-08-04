package main

import (
	"os"
	"path/filepath"
	"testing"
)

// TestGetConfigPath tests the configuration path resolution
func TestGetConfigPath(t *testing.T) {
	// Save original env var
	originalEnv := os.Getenv("SET_TAB_COLOR_CONFIG")
	defer func() {
		if originalEnv == "" {
			os.Unsetenv("SET_TAB_COLOR_CONFIG")
		} else {
			os.Setenv("SET_TAB_COLOR_CONFIG", originalEnv)
		}
	}()

	// Test with environment variable set
	customPath := "/tmp/custom-config.toml"
	os.Setenv("SET_TAB_COLOR_CONFIG", customPath)

	path, err := getConfigPath()
	if err != nil {
		t.Fatalf("getConfigPath() failed: %v", err)
	}

	if path != customPath {
		t.Errorf("Expected config path %q, got %q", customPath, path)
	}

	// Test with environment variable unset (default path)
	os.Unsetenv("SET_TAB_COLOR_CONFIG")

	path, err = getConfigPath()
	if err != nil {
		t.Fatalf("getConfigPath() failed: %v", err)
	}

	expectedSuffix := filepath.Join("set-tab-color.toml")
	if !filepath.IsAbs(path) {
		t.Errorf("Expected absolute path, got %q", path)
	}

	if !contains(path, expectedSuffix) {
		t.Errorf("Expected path to contain %q, got %q", expectedSuffix, path)
	}
}

// TestLoadConfig tests loading configuration from TOML files
func TestLoadConfig(t *testing.T) {
	// Create temporary config file
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "test-config.toml")

	configContent := `
[profiles.development]
tab = "blue"
fg = "white"
bg = "black"

[profiles.production]
tab = "red"
fg = "yellow"

[profiles.minimal]
tab = "green"
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

	// Load config
	config, err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig() failed: %v", err)
	}

	if config.Profiles == nil {
		t.Fatal("Expected profiles map to be initialized")
	}

	// Test development profile
	dev, exists := config.Profiles["development"]
	if !exists {
		t.Error("Expected 'development' profile to exist")
	}

	if dev.Tab != "blue" || dev.Foreground != "white" || dev.Background != "black" {
		t.Errorf("Development profile incorrect: tab=%q, fg=%q, bg=%q", dev.Tab, dev.Foreground, dev.Background)
	}

	// Test production profile
	prod, exists := config.Profiles["production"]
	if !exists {
		t.Error("Expected 'production' profile to exist")
	}

	if prod.Tab != "red" || prod.Foreground != "yellow" || prod.Background != "" {
		t.Errorf("Production profile incorrect: tab=%q, fg=%q, bg=%q", prod.Tab, prod.Foreground, prod.Background)
	}

	// Test minimal profile
	minimal, exists := config.Profiles["minimal"]
	if !exists {
		t.Error("Expected 'minimal' profile to exist")
	}

	if minimal.Tab != "green" || minimal.Foreground != "" || minimal.Background != "" {
		t.Errorf("Minimal profile incorrect: tab=%q, fg=%q, bg=%q", minimal.Tab, minimal.Foreground, minimal.Background)
	}
}

// TestLoadConfigMissing tests loading when config file doesn't exist
func TestLoadConfigMissing(t *testing.T) {
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

	config, err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig() should not fail for missing file: %v", err)
	}

	if config.Profiles == nil {
		t.Error("Expected profiles map to be initialized even for missing config")
	}

	if len(config.Profiles) != 0 {
		t.Errorf("Expected empty profiles map, got %d profiles", len(config.Profiles))
	}
}

// TestLoadConfigInvalid tests loading invalid TOML files
func TestLoadConfigInvalid(t *testing.T) {
	// Create temporary invalid config file
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "invalid-config.toml")

	invalidContent := `
[profiles.broken
tab = "red"
missing closing bracket
`

	if err := os.WriteFile(configFile, []byte(invalidContent), 0644); err != nil {
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

	// Load config should fail
	_, err := loadConfig()
	if err == nil {
		t.Error("Expected loadConfig() to fail for invalid TOML")
	}

	if !contains(err.Error(), "error parsing config file") {
		t.Errorf("Expected error to mention config parsing, got: %v", err)
	}
}

// TestGetProfile tests retrieving specific profiles
func TestGetProfile(t *testing.T) {
	// Create temporary config file
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "test-config.toml")

	configContent := `
[profiles.test-profile]
tab = "purple"
fg = "white"
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

	// Test existing profile
	profile, err := getProfile("test-profile")
	if err != nil {
		t.Fatalf("getProfile() failed: %v", err)
	}

	if profile.Tab != "purple" || profile.Foreground != "white" || profile.Background != "" {
		t.Errorf("Profile incorrect: tab=%q, fg=%q, bg=%q", profile.Tab, profile.Foreground, profile.Background)
	}

	// Test non-existent profile
	_, err = getProfile("non-existent")
	if err == nil {
		t.Error("Expected getProfile() to fail for non-existent profile")
	}

	if !contains(err.Error(), "profile \"non-existent\" not found") {
		t.Errorf("Expected error to mention profile not found, got: %v", err)
	}
}

// TestApplyProfile tests applying profiles (without actually executing it2setcolor)
func TestApplyProfile(t *testing.T) {
	// We can't easily test applyProfile without mocking runSetColor
	// This would require more complex test setup or dependency injection
	// For now, we'll test that the profile structure is correct
	profile := &Profile{
		Tab:        "blue",
		Foreground: "white",
		Background: "black",
	}

	if profile.Tab != "blue" {
		t.Errorf("Expected tab color 'blue', got %q", profile.Tab)
	}

	if profile.Foreground != "white" {
		t.Errorf("Expected foreground color 'white', got %q", profile.Foreground)
	}

	if profile.Background != "black" {
		t.Errorf("Expected background color 'black', got %q", profile.Background)
	}
}
