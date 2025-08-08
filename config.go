package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Global verbose flag for debugging output
var verboseMode bool

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

	// Prefer ~/.config/set-tab-color.toml on all platforms (including macOS)
	homeDir, err := os.UserHomeDir()
	if err == nil {
		configPath := filepath.Join(homeDir, ".config", "set-tab-color.toml")
		// Check if ~/.config directory exists or the config file exists
		if _, err := os.Stat(filepath.Dir(configPath)); err == nil {
			return configPath, nil
		}
		if _, err := os.Stat(configPath); err == nil {
			return configPath, nil
		}
	}

	// Fall back to OS-specific config directory (~/Library/Application Support on macOS)
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

	if verboseMode {
		fmt.Fprintf(os.Stderr, "Using base profile: %q\n", profileName)
		fmt.Fprintf(os.Stderr, "  Base profile values: tab=%q, fg=%q, bg=%q, preset=%q\n",
			baseProfile.Tab, baseProfile.Foreground, baseProfile.Background, baseProfile.Preset)
	}

	// Start with base profile
	result := *baseProfile

	// Get the nested map for this profile to look for sub-profiles
	profileMap, ok := baseData.(map[string]interface{})
	if !ok {
		// No nested structure, just return base profile
		if verboseMode {
			fmt.Fprintf(os.Stderr, "No sub-profiles available for profile %q\n", profileName)
		}
		return &result, nil
	}

	// Use provided terminal info (caller must always provide it)
	terminalShellInfo := *terminalInfo
	if verboseMode {
		fmt.Fprintf(os.Stderr, "Terminal detection: %v\n", terminalShellInfo.Terminals)
		fmt.Fprintf(os.Stderr, "Shell detection: %s\n", terminalShellInfo.Shell)
		fmt.Fprintf(os.Stderr, "Detection valid: %v", terminalShellInfo.Valid)
		if !terminalShellInfo.Valid {
			fmt.Fprintf(os.Stderr, " (shell should come before terminal)")
		}
		fmt.Fprintf(os.Stderr, "\n")

		if chain, err := getProcessAncestorChain(); err == nil {
			fmt.Fprintf(os.Stderr, "Process ancestor chain:\n")
			for i, processName := range chain {
				fmt.Fprintf(os.Stderr, "  %d: %s\n", i, processName)
			}
		}
		fmt.Fprintf(os.Stderr, "\n")
	}

	// Apply shell-specific overlay first (if it exists)
	if terminalShellInfo.Shell != ShellTypeUnknown {
		shellKey := string(terminalShellInfo.Shell)
		if shellData, exists := profileMap[shellKey]; exists {
			if shellProfile, err := extractProfile(shellData); err == nil {
				if verboseMode {
					fmt.Fprintf(os.Stderr, "Applying shell-specific sub-profile: %s.%s\n", profileName, shellKey)
					fmt.Fprintf(os.Stderr, "  Shell sub-profile values: tab=%q, fg=%q, bg=%q, preset=%q\n",
						shellProfile.Tab, shellProfile.Foreground, shellProfile.Background, shellProfile.Preset)
				}
				result = overlayProfile(result, *shellProfile)
			}
		} else if verboseMode {
			fmt.Fprintf(os.Stderr, "No shell-specific sub-profile found for: %s.%s\n", profileName, shellKey)
		}
	}

	// Apply terminal-specific overlay last (takes priority)
	// Try terminals in order until we find one with a subprofile
	var appliedTerminalProfile bool
	if verboseMode {
		fmt.Fprintf(os.Stderr, "Checking terminals for sub-profiles: %v\n", terminalShellInfo.Terminals)
	}

	for _, terminal := range terminalShellInfo.Terminals {
		terminalKey := string(terminal)
		if terminalData, exists := profileMap[terminalKey]; exists {
			if terminalProfile, err := extractProfile(terminalData); err == nil {
				if verboseMode {
					fmt.Fprintf(os.Stderr, "Applying terminal-specific sub-profile: %s.%s\n", profileName, terminalKey)
					fmt.Fprintf(os.Stderr, "  Terminal sub-profile values: tab=%q, fg=%q, bg=%q, preset=%q\n",
						terminalProfile.Tab, terminalProfile.Foreground, terminalProfile.Background, terminalProfile.Preset)
				}
				result = overlayProfile(result, *terminalProfile)
				appliedTerminalProfile = true
				break // Use the first terminal that has a subprofile
			}
		} else if verboseMode {
			fmt.Fprintf(os.Stderr, "No terminal-specific sub-profile found for: %s.%s\n", profileName, terminalKey)
		}
	}

	if !appliedTerminalProfile && len(terminalShellInfo.Terminals) > 0 && verboseMode {
		fmt.Fprintf(os.Stderr, "No terminal sub-profiles found for any terminal in the process chain\n")
	}

	if verboseMode {
		fmt.Fprintf(os.Stderr, "Final profile values after overlays: tab=%q, fg=%q, bg=%q, preset=%q\n",
			result.Tab, result.Foreground, result.Background, result.Preset)
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
	if verboseMode {
		fmt.Fprintf(os.Stderr, "\nApplying profile settings:\n")
	}

	// Apply preset first if specified (so individual colors can override it)
	if profile.Preset != "" {
		if verboseMode {
			fmt.Fprintf(os.Stderr, "  Setting preset: %q\n", profile.Preset)
		}
		if err := runSetPreset(profile.Preset); err != nil {
			return fmt.Errorf("error setting preset from profile: %v", err)
		}
	}

	// Set tab color if specified (overrides preset)
	if profile.Tab != "" {
		if verboseMode {
			fmt.Fprintf(os.Stderr, "  Setting tab color: %q\n", profile.Tab)
		}
		if err := runSetColor(TabColor, profile.Tab); err != nil {
			return fmt.Errorf("error setting tab color from profile: %v", err)
		}
	}

	// Set foreground color if specified (overrides preset)
	if profile.Foreground != "" {
		if verboseMode {
			fmt.Fprintf(os.Stderr, "  Setting foreground color: %q\n", profile.Foreground)
		}
		if err := runSetColor(ForegroundColor, profile.Foreground); err != nil {
			return fmt.Errorf("error setting foreground color from profile: %v", err)
		}
	}

	// Set background color if specified (overrides preset)
	if profile.Background != "" {
		if verboseMode {
			fmt.Fprintf(os.Stderr, "  Setting background color: %q\n", profile.Background)
		}
		if err := runSetColor(BackgroundColor, profile.Background); err != nil {
			return fmt.Errorf("error setting background color from profile: %v", err)
		}
	}

	if verboseMode {
		fmt.Fprintf(os.Stderr, "Profile application complete.\n")
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
