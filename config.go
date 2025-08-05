package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Profile represents a color profile with optional colors and preset
type Profile struct {
	Tab        string `toml:"tab,omitempty"`
	Foreground string `toml:"fg,omitempty"`
	Background string `toml:"bg,omitempty"`
	Preset     string `toml:"preset,omitempty"`
}

// Config represents the TOML configuration file structure with nested profiles
type Config struct {
	Profiles map[string]interface{} `toml:"profiles"`
}

// getConfigPath returns the configuration file path, checking env var first
func getConfigPath() (string, error) {
	// Check environment variable first
	if configPath := os.Getenv("SET_TAB_COLOR_CONFIG"); configPath != "" {
		return configPath, nil
	}

	// Default to ~/.config/set-tab-color.toml
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("could not get config directory: %v", err)
	}

	return filepath.Join(configDir, "set-tab-color.toml"), nil
}

// loadConfig loads the TOML configuration file
func loadConfig() (*Config, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	// If config file doesn't exist, return empty config
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &Config{Profiles: make(map[string]interface{})}, nil
	}

	// Load config maintaining nested structure
	var config Config
	if _, err := toml.DecodeFile(configPath, &config); err != nil {
		return nil, fmt.Errorf("error parsing config file %s: %v", configPath, err)
	}

	// Initialize profiles map if nil
	if config.Profiles == nil {
		config.Profiles = make(map[string]interface{})
	}

	return &config, nil
}

// extractProfile dynamically extracts a profile from a nested map structure
func extractProfile(data interface{}) (*Profile, error) {
	m, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("expected map[string]interface{}, got %T", data)
	}

	// Check if this is a profile (has tab, fg, or bg keys)
	if !isProfileMap(m) {
		return nil, fmt.Errorf("not a profile map")
	}

	profile := &Profile{}

	if tab, ok := m["tab"]; ok {
		if tabStr, ok := tab.(string); ok {
			profile.Tab = tabStr
		}
	}

	if fg, ok := m["fg"]; ok {
		if fgStr, ok := fg.(string); ok {
			profile.Foreground = fgStr
		}
	}

	if bg, ok := m["bg"]; ok {
		if bgStr, ok := bg.(string); ok {
			profile.Background = bgStr
		}
	}

	if preset, ok := m["preset"]; ok {
		if presetStr, ok := preset.(string); ok {
			profile.Preset = presetStr
		}
	}

	return profile, nil
}

// isProfileMap checks if a map contains profile-like keys
func isProfileMap(m map[string]interface{}) bool {
	for key := range m {
		if key == "tab" || key == "fg" || key == "bg" || key == "preset" {
			return true
		}
	}
	return false
}

// getProfile retrieves a specific profile by name with sub-profile overlays
func getProfile(profileName string) (*Profile, error) {
	return getProfileWithTerminalInfo(profileName, nil)
}

// getProfileWithTerminalInfo retrieves a profile with optional terminal info override (for testing)
func getProfileWithTerminalInfo(profileName string, terminalInfo *TerminalShellInfo) (*Profile, error) {
	config, err := loadConfig()
	if err != nil {
		return nil, err
	}

	// Find base profile in nested structure
	baseData, exists := config.Profiles[profileName]
	if !exists {
		return nil, fmt.Errorf("profile %q not found", profileName)
	}

	// Extract base profile
	baseProfile, err := extractProfile(baseData)
	if err != nil {
		// Not a valid profile at top level, check if it's a nested structure
		return nil, fmt.Errorf("profile %q is not a valid profile", profileName)
	}

	// Start with base profile
	result := *baseProfile

	// Use provided terminal info or detect it
	var terminalShellInfo TerminalShellInfo
	if terminalInfo != nil {
		terminalShellInfo = *terminalInfo
	} else {
		terminalShellInfo = detectTerminalAndShell()
	}

	// Get the nested map for this profile to look for sub-profiles
	profileMap, ok := baseData.(map[string]interface{})
	if !ok {
		// No nested structure, just return base profile
		return &result, nil
	}

	// Apply shell-specific overlay first (if it exists)
	if terminalShellInfo.Shell != ShellTypeUnknown {
		shellKey := string(terminalShellInfo.Shell)
		if shellData, exists := profileMap[shellKey]; exists {
			if shellProfile, err := extractProfile(shellData); err == nil {
				result = overlayProfile(result, *shellProfile)
			}
		}
	}

	// Apply terminal-specific overlay last (takes priority)
	if terminalShellInfo.Terminal != TerminalTypeUnknown {
		terminalKey := string(terminalShellInfo.Terminal)
		if terminalData, exists := profileMap[terminalKey]; exists {
			if terminalProfile, err := extractProfile(terminalData); err == nil {
				result = overlayProfile(result, *terminalProfile)
			}
		}
	}

	return &result, nil
}

// overlayProfile applies overlay settings on top of base profile
func overlayProfile(base Profile, overlay Profile) Profile {
	result := base

	// Overlay non-empty values from overlay profile
	if overlay.Tab != "" {
		result.Tab = overlay.Tab
	}
	if overlay.Foreground != "" {
		result.Foreground = overlay.Foreground
	}
	if overlay.Background != "" {
		result.Background = overlay.Background
	}
	if overlay.Preset != "" {
		result.Preset = overlay.Preset
	}

	return result
}

// applyProfile applies a profile's colors using the existing runSetColor function
func applyProfile(profile *Profile) error {
	// Apply preset first if specified (so individual colors can override it)
	if profile.Preset != "" {
		if err := runSetPreset(profile.Preset); err != nil {
			return fmt.Errorf("error setting preset from profile: %v", err)
		}
	}

	// Set tab color if specified (overrides preset)
	if profile.Tab != "" {
		if err := runSetColor(TabColor, profile.Tab); err != nil {
			return fmt.Errorf("error setting tab color from profile: %v", err)
		}
	}

	// Set foreground color if specified (overrides preset)
	if profile.Foreground != "" {
		if err := runSetColor(ForegroundColor, profile.Foreground); err != nil {
			return fmt.Errorf("error setting foreground color from profile: %v", err)
		}
	}

	// Set background color if specified (overrides preset)
	if profile.Background != "" {
		if err := runSetColor(BackgroundColor, profile.Background); err != nil {
			return fmt.Errorf("error setting background color from profile: %v", err)
		}
	}

	return nil
}

// listProfileNames returns a list of all available profile names
func listProfileNames() ([]string, error) {
	config, err := loadConfig()
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(config.Profiles))
	for name := range config.Profiles {
		names = append(names, name)
	}

	return names, nil
}
