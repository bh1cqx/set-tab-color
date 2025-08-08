package main

import (
	"os"
	"path/filepath"
	"testing"
)

// getProfileWithTerminalOverride is a test helper function that can either auto-detect
// terminal info (when terminalOverride is empty) or use the specified terminal override
func getProfileWithTerminalOverride(profileName string, terminalOverride string) (*Profile, error) {
	// Detect terminal and shell info with optional terminal override
	terminalInfo := detectTerminalAndShell(terminalOverride)
	return getProfileWithTerminalInfo(profileName, &terminalInfo)
}

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

	// Test development profile exists
	_, exists := config.Profiles["development"]
	if !exists {
		t.Error("Expected 'development' profile to exist")
	}

	// Test production profile exists
	_, exists = config.Profiles["production"]
	if !exists {
		t.Error("Expected 'production' profile to exist")
	}

	// Test minimal profile exists
	_, exists = config.Profiles["minimal"]
	if !exists {
		t.Error("Expected 'minimal' profile to exist")
	}

	// Now test the actual profile values using getProfile
	devProfile, err := getProfileWithTerminalInfo("development", &TerminalShellInfo{
		Terminals: []TerminalType{},
		Shell:     ShellTypeUnknown,
		Valid:     false,
	})
	if err != nil {
		t.Fatalf("Failed to get development profile: %v", err)
	}
	if devProfile.Tab != "blue" || devProfile.Foreground != "white" || devProfile.Background != "black" {
		t.Errorf("Development profile incorrect: tab=%q, fg=%q, bg=%q", devProfile.Tab, devProfile.Foreground, devProfile.Background)
	}

	// Test production profile values
	prodProfile, err := getProfileWithTerminalInfo("production", &TerminalShellInfo{
		Terminals: []TerminalType{},
		Shell:     ShellTypeUnknown,
		Valid:     false,
	})
	if err != nil {
		t.Fatalf("Failed to get production profile: %v", err)
	}
	if prodProfile.Tab != "red" || prodProfile.Foreground != "yellow" || prodProfile.Background != "" {
		t.Errorf("Production profile incorrect: tab=%q, fg=%q, bg=%q", prodProfile.Tab, prodProfile.Foreground, prodProfile.Background)
	}

	// Test minimal profile values
	minimalProfile, err := getProfileWithTerminalInfo("minimal", &TerminalShellInfo{
		Terminals: []TerminalType{},
		Shell:     ShellTypeUnknown,
		Valid:     false,
	})
	if err != nil {
		t.Fatalf("Failed to get minimal profile: %v", err)
	}
	if minimalProfile.Tab != "green" || minimalProfile.Foreground != "" || minimalProfile.Background != "" {
		t.Errorf("Minimal profile incorrect: tab=%q, fg=%q, bg=%q", minimalProfile.Tab, minimalProfile.Foreground, minimalProfile.Background)
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

	// Test existing profile with no terminal/shell info to avoid sub-profile interference
	profile, err := getProfileWithTerminalInfo("test-profile", &TerminalShellInfo{
		Terminals: []TerminalType{},
		Shell:     ShellTypeUnknown,
		Valid:     false,
	})
	if err != nil {
		t.Fatalf("getProfileWithTerminalInfo() failed: %v", err)
	}

	if profile.Tab != "purple" || profile.Foreground != "white" || profile.Background != "" {
		t.Errorf("Profile incorrect: tab=%q, fg=%q, bg=%q", profile.Tab, profile.Foreground, profile.Background)
	}

	// Test non-existent profile
	_, err = getProfileWithTerminalInfo("non-existent", &TerminalShellInfo{
		Terminals: []TerminalType{},
		Shell:     ShellTypeUnknown,
		Valid:     false,
	})
	if err == nil {
		t.Error("Expected getProfileWithTerminalInfo() to fail for non-existent profile")
	}

	if !contains(err.Error(), "profile \"non-existent\" not found") {
		t.Errorf("Expected error to mention profile not found, got: %v", err)
	}

	// Test that the helper function works with auto-detection (backward compatibility)
	profile, err = getProfileWithTerminalOverride("test-profile", "")
	if err != nil {
		t.Fatalf("getProfileWithTerminalOverride() failed: %v", err)
	}

	// The result might vary based on current terminal, but it should at least contain the base values
	// We'll just verify that it returns a profile and doesn't fail
	if profile == nil {
		t.Error("Expected getProfileWithTerminalOverride() to return a profile")
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

// TestOverlayProfile tests the profile overlay functionality
func TestOverlayProfile(t *testing.T) {
	base := Profile{
		Tab:        "blue",
		Foreground: "white",
		Background: "black",
	}

	// Test overlay with all fields
	overlay1 := Profile{
		Tab:        "red",
		Foreground: "yellow",
		Background: "green",
	}

	result1 := overlayProfile(base, overlay1)
	if result1.Tab != "red" || result1.Foreground != "yellow" || result1.Background != "green" {
		t.Errorf("Full overlay failed: got tab=%q, fg=%q, bg=%q", result1.Tab, result1.Foreground, result1.Background)
	}

	// Test partial overlay (only some fields)
	overlay2 := Profile{
		Tab: "purple",
		// Foreground and Background are empty, should keep base values
	}

	result2 := overlayProfile(base, overlay2)
	if result2.Tab != "purple" || result2.Foreground != "white" || result2.Background != "black" {
		t.Errorf("Partial overlay failed: got tab=%q, fg=%q, bg=%q", result2.Tab, result2.Foreground, result2.Background)
	}

	// Test empty overlay (no changes)
	overlay3 := Profile{}

	result3 := overlayProfile(base, overlay3)
	if result3.Tab != "blue" || result3.Foreground != "white" || result3.Background != "black" {
		t.Errorf("Empty overlay failed: got tab=%q, fg=%q, bg=%q", result3.Tab, result3.Foreground, result3.Background)
	}
}

// TestGetProfileWithSubProfiles tests sub-profile functionality
func TestGetProfileWithSubProfiles(t *testing.T) {
	// Create temporary config file with sub-profiles
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "test-sub-profiles.toml")

	configContent := `
[profiles.dev]
tab = "blue"
fg = "white"
bg = "black"

[profiles.dev.zsh]
tab = "cyan"
fg = "yellow"

[profiles.dev.iterm2]
tab = "purple"
bg = "darkgray"

[profiles.prod]
tab = "red"
fg = "white"

[profiles.prod.ssh]
tab = "brightred"
bg = "black"

[profiles.prod.bash]
fg = "yellow"
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

	// Test base profile (no terminal/shell info)
	profile, err := getProfileWithTerminalInfo("dev", &TerminalShellInfo{
		Terminals: []TerminalType{},
		Shell:     ShellTypeUnknown,
		Valid:     false,
	})
	if err != nil {
		t.Fatalf("getProfileWithTerminalInfo() failed: %v", err)
	}
	if profile.Tab != "blue" || profile.Foreground != "white" || profile.Background != "black" {
		t.Errorf("Base dev profile incorrect: tab=%q, fg=%q, bg=%q", profile.Tab, profile.Foreground, profile.Background)
	}

	// Test shell-only overlay
	profile, err = getProfileWithTerminalInfo("dev", &TerminalShellInfo{
		Terminals: []TerminalType{},
		Shell:     ShellTypeZsh,
		Valid:     true,
	})
	if err != nil {
		t.Fatalf("getProfileWithTerminalInfo() with zsh failed: %v", err)
	}
	if profile.Tab != "cyan" || profile.Foreground != "yellow" || profile.Background != "black" {
		t.Errorf("dev.zsh overlay failed: tab=%q, fg=%q, bg=%q", profile.Tab, profile.Foreground, profile.Background)
	}

	// Test terminal-only overlay
	profile, err = getProfileWithTerminalInfo("dev", &TerminalShellInfo{
		Terminals: []TerminalType{TerminalTypeITerm2},
		Shell:     ShellTypeUnknown,
		Valid:     true,
	})
	if err != nil {
		t.Fatalf("getProfileWithTerminalInfo() with iterm2 failed: %v", err)
	}
	if profile.Tab != "purple" || profile.Foreground != "white" || profile.Background != "darkgray" {
		t.Errorf("dev.iterm2 overlay failed: tab=%q, fg=%q, bg=%q", profile.Tab, profile.Foreground, profile.Background)
	}

	// Test both shell and terminal overlays (terminal should take priority)
	profile, err = getProfileWithTerminalInfo("dev", &TerminalShellInfo{
		Terminals: []TerminalType{TerminalTypeITerm2},
		Shell:     ShellTypeZsh,
		Valid:     true,
	})
	if err != nil {
		t.Fatalf("getProfileWithTerminalInfo() with zsh+iterm2 failed: %v", err)
	}
	// Expected: tab=purple (terminal), fg=yellow (shell), bg=darkgray (terminal)
	if profile.Tab != "purple" || profile.Foreground != "yellow" || profile.Background != "darkgray" {
		t.Errorf("dev with zsh+iterm2 overlay failed: tab=%q, fg=%q, bg=%q", profile.Tab, profile.Foreground, profile.Background)
	}

	// Test SSH terminal override
	profile, err = getProfileWithTerminalInfo("prod", &TerminalShellInfo{
		Terminals: []TerminalType{TerminalTypeSSH},
		Shell:     ShellTypeUnknown,
		Valid:     true,
	})
	if err != nil {
		t.Fatalf("getProfileWithTerminalInfo() with SSH failed: %v", err)
	}
	if profile.Tab != "brightred" || profile.Foreground != "white" || profile.Background != "black" {
		t.Errorf("prod.ssh overlay failed: tab=%q, fg=%q, bg=%q", profile.Tab, profile.Foreground, profile.Background)
	}
}
