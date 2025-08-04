package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Profile represents a color profile with optional colors
type Profile struct {
	Tab        string `toml:"tab,omitempty"`
	Foreground string `toml:"fg,omitempty"`
	Background string `toml:"bg,omitempty"`
}

// Config represents the TOML configuration file structure
type Config struct {
	Profiles map[string]Profile `toml:"profiles"`
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
		return &Config{Profiles: make(map[string]Profile)}, nil
	}

	var config Config
	if _, err := toml.DecodeFile(configPath, &config); err != nil {
		return nil, fmt.Errorf("error parsing config file %s: %v", configPath, err)
	}

	// Initialize profiles map if nil
	if config.Profiles == nil {
		config.Profiles = make(map[string]Profile)
	}

	return &config, nil
}

// getProfile retrieves a specific profile by name
func getProfile(profileName string) (*Profile, error) {
	config, err := loadConfig()
	if err != nil {
		return nil, err
	}

	profile, exists := config.Profiles[profileName]
	if !exists {
		return nil, fmt.Errorf("profile %q not found", profileName)
	}

	return &profile, nil
}

// applyProfile applies a profile's colors using the existing runSetColor function
func applyProfile(profile *Profile) error {
	// Set tab color if specified
	if profile.Tab != "" {
		if err := runSetColor(TabColor, profile.Tab); err != nil {
			return fmt.Errorf("error setting tab color from profile: %v", err)
		}
	}

	// Set foreground color if specified
	if profile.Foreground != "" {
		if err := runSetColor(ForegroundColor, profile.Foreground); err != nil {
			return fmt.Errorf("error setting foreground color from profile: %v", err)
		}
	}

	// Set background color if specified
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
