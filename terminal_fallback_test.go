package main

import (
	"os"
	"path/filepath"
	"testing"
)

// TestTerminalFallback tests the core fallback scenario:
// Process chain has tmux -> etterminal, but only etterminal has a subprofile
func TestTerminalFallback(t *testing.T) {
	// Create temporary config file with fallback scenario
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "fallback-config.toml")

	configContent := `
[profiles.work]
tab = "blue"
fg = "white"

# No tmux subprofile - should fallback to etterminal
[profiles.work.etterminal]
tab = "green"
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

	// Mock process hierarchy: detected terminal is tmux, but chain includes etterminal
	// Override detectAllTerminalsInChain for this test by providing manual terminal info
	terminalInfo := &TerminalShellInfo{
		Terminals: []TerminalType{TerminalTypeTmux, TerminalTypeETTerminal}, // Primary detected terminal (has no subprofile)
		Shell:     ShellTypeZsh,
		Valid:     true,
	}

	// Create a custom version that simulates the fallback scenario
	// We'll temporarily replace the terminalChainDetector function
	originalDetectFunc := terminalChainDetector
	terminalChainDetector = func() []TerminalType {
		return []TerminalType{TerminalTypeTmux, TerminalTypeETTerminal}
	}
	defer func() {
		terminalChainDetector = originalDetectFunc
	}()

	// Call the actual profile resolution logic
	profile, err := getProfileWithTerminalInfo("work", terminalInfo)
	if err != nil {
		t.Fatalf("getProfileWithTerminalInfo failed: %v", err)
	}

	// Verify fallback worked: should get etterminal subprofile values
	if profile.Tab != "green" {
		t.Errorf("Expected fallback to etterminal tab='green', got tab=%q", profile.Tab)
	}
	if profile.Foreground != "yellow" {
		t.Errorf("Expected fallback to etterminal fg='yellow', got fg=%q", profile.Foreground)
	}
}

// TestTerminalNoFallbackNeeded tests that when primary terminal has subprofile, no fallback occurs
func TestTerminalNoFallbackNeeded(t *testing.T) {
	// Create temporary config file where primary terminal has subprofile
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "no-fallback-config.toml")

	configContent := `
[profiles.dev]
tab = "blue"

[profiles.dev.tmux]
tab = "red"

[profiles.dev.etterminal]
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

	// Mock primary terminal as tmux (which has subprofile)
	terminalInfo := &TerminalShellInfo{
		Terminals: []TerminalType{TerminalTypeTmux},
		Shell:     ShellTypeZsh,
		Valid:     true,
	}

	// Call the actual profile resolution logic
	profile, err := getProfileWithTerminalInfo("dev", terminalInfo)
	if err != nil {
		t.Fatalf("getProfileWithTerminalInfo failed: %v", err)
	}

	// Should use tmux subprofile, not fallback to etterminal
	if profile.Tab != "red" {
		t.Errorf("Expected tmux subprofile tab='red', got tab=%q", profile.Tab)
	}
}

// TestTerminalFallbackOrder tests that fallback uses first available terminal in chain order
func TestTerminalFallbackOrder(t *testing.T) {
	// Create temporary config file with multiple fallback options
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "priority-config.toml")

	configContent := `
[profiles.test]
tab = "blue"

# First fallback option
[profiles.test.etterminal]
tab = "green"

# Second fallback option
[profiles.test.iterm2]
tab = "purple"
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

	// Mock primary terminal as tmux (no subprofile) with multiple fallback options
	terminalInfo := &TerminalShellInfo{
		Terminals: []TerminalType{TerminalTypeTmux, TerminalTypeETTerminal, TerminalTypeITerm2},
		Shell:     ShellTypeZsh,
		Valid:     true,
	}

	// Mock terminal chain: tmux, etterminal, iterm2
	originalDetectFunc := terminalChainDetector
	terminalChainDetector = func() []TerminalType {
		return []TerminalType{TerminalTypeTmux, TerminalTypeETTerminal, TerminalTypeITerm2}
	}
	defer func() {
		terminalChainDetector = originalDetectFunc
	}()

	// Call the actual profile resolution logic
	profile, err := getProfileWithTerminalInfo("test", terminalInfo)
	if err != nil {
		t.Fatalf("getProfileWithTerminalInfo failed: %v", err)
	}

	// Should use first available fallback (etterminal), not second (iterm2)
	if profile.Tab != "green" {
		t.Errorf("Expected first fallback etterminal tab='green', got tab=%q", profile.Tab)
	}
}
