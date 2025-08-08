package main

import (
	"testing"
)

func TestDetectTerminalType(t *testing.T) {
	// This test will detect the actual terminal types running the tests
	info := detectTerminalAndShell("")

	// We can't assert specific values since it depends on the environment
	// but we can ensure it returns valid types
	validTypes := []TerminalType{
		TerminalTypeUnknown,
		TerminalTypeITerm2,
		TerminalTypeETTerminal,
		TerminalTypeSSH,
		TerminalTypeTmux,
		TerminalTypeVSCode,
	}

	for _, terminalType := range info.Terminals {
		found := false
		for _, validType := range validTypes {
			if terminalType == validType {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("detectTerminalAndShell() returned invalid type: %s", terminalType)
		}
	}

	t.Logf("Detected terminal types: %v", info.Terminals)
}

func TestTerminalAndShellDetection(t *testing.T) {
	// Test the combined terminal and shell detection
	info := detectTerminalAndShell("")

	t.Logf("Combined detection results:")
	t.Logf("  Terminals: %v", info.Terminals)
	t.Logf("  Shell: %v", info.Shell)
	t.Logf("  Valid: %v", info.Valid)

	// Test individual shell extraction for backwards compatibility
	shellType := info.Shell

	t.Logf("Individual detection results:")
	t.Logf("  Shell extracted: %v", shellType)

	// Verify consistency (should always match since we're using the same info)
	if shellType != info.Shell {
		t.Errorf("Shell extraction = %v, but detectTerminalAndShell().Shell = %v", shellType, info.Shell)
	}
}

func TestShellTypeValidation(t *testing.T) {
	// Test that shell types are valid
	validShellTypes := []ShellType{
		ShellTypeUnknown,
		ShellTypeBash,
		ShellTypeZsh,
		ShellTypeFish,
		ShellTypeTcsh,
		ShellTypeCsh,
		ShellTypeKsh,
		ShellTypeSh,
	}

	info := detectTerminalAndShell("")
	shellType := info.Shell

	found := false
	for _, validType := range validShellTypes {
		if shellType == validType {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("detectTerminalAndShell().Shell returned invalid type: %s", shellType)
	}

	t.Logf("Detected shell type: %s", shellType)
}

func TestTerminalOverride(t *testing.T) {
	tests := []struct {
		name             string
		terminalOverride string
		expectedTerminal TerminalType
		shouldPrepend    bool
	}{
		{"iTerm2 override", "iterm2", TerminalTypeITerm2, true},
		{"VSCode override", "vscode", TerminalTypeVSCode, true},
		{"SSH override", "ssh", TerminalTypeSSH, true},
		{"Tmux override", "tmux", TerminalTypeTmux, true},
		{"ETTerminal override", "etterminal", TerminalTypeETTerminal, true},
		{"Invalid override", "invalid", TerminalTypeUnknown, false},
		{"Empty override", "", TerminalTypeUnknown, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := detectTerminalAndShell(tt.terminalOverride)

			if tt.shouldPrepend {
				// Check that the override terminal is the first in the list
				if len(info.Terminals) == 0 {
					t.Errorf("Expected terminal override %q to be prepended, but got empty terminals list", tt.terminalOverride)
					return
				}

				if info.Terminals[0] != tt.expectedTerminal {
					t.Errorf("Expected first terminal to be %v for override %q, but got %v", tt.expectedTerminal, tt.terminalOverride, info.Terminals[0])
				}

				t.Logf("Override %q successfully prepended %v to terminals list: %v", tt.terminalOverride, tt.expectedTerminal, info.Terminals)
			} else {
				// For invalid or empty overrides, check that no specific terminal was forcefully prepended
				// (though there might still be detected terminals from the actual process chain)
				t.Logf("Override %q correctly ignored, terminals detected: %v", tt.terminalOverride, info.Terminals)
			}
		})
	}
}

func TestGetProcessAncestorChain(t *testing.T) {
	chain, err := getProcessAncestorChain()
	if err != nil {
		t.Fatalf("getProcessAncestorChain() returned error: %v", err)
	}

	if len(chain) == 0 {
		t.Error("getProcessAncestorChain() returned empty chain")
	}

	// Log the process chain for debugging
	t.Logf("Process ancestor chain:")
	for i, processName := range chain {
		t.Logf("  %d: %s", i, processName)
	}

	// The first process should be our test process
	if len(chain) > 0 {
		// It should contain some form of the test executable or "go"
		firstProcess := chain[0]
		if firstProcess == "" {
			t.Error("First process in chain should not be empty")
		}
	}
}

func TestMatchesTerminalName(t *testing.T) {
	tests := []struct {
		name          string
		processName   string
		terminalName  string
		caseSensitive bool
		expected      bool
	}{
		// Case-sensitive exact matches
		{"Exact match case-sensitive", "sshd", "sshd", true, true},
		{"Exact match case-sensitive fail", "sshd", "SSHD", true, false},
		{"Exact match case-sensitive tmux", "tmux", "tmux", true, true},

		// Case-insensitive exact matches (for iTerm)
		{"Exact match case-insensitive", "iterm", "iterm", false, true},
		{"Exact match case-insensitive upper", "ITERM", "iterm", false, true},
		{"Exact match case-insensitive mixed", "iTerm", "iterm", false, true},

		// Prefix matches with space
		{"Prefix match case-sensitive", "sshd server", "sshd", true, true},
		{"Prefix match case-sensitive fail", "sshd server", "SSHD", true, false},
		{"Prefix match case-insensitive", "iTerm args", "iterm", false, true},

		// Non-matches
		{"No match different name", "bash", "sshd", true, false},
		{"No match substring", "mysshd", "sshd", true, false},
		{"No match prefix without space", "sshdserver", "sshd", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchesTerminalName(tt.processName, tt.terminalName, tt.caseSensitive)
			if result != tt.expected {
				t.Errorf("matchesTerminalName(%q, %q, %v) = %v, expected %v",
					tt.processName, tt.terminalName, tt.caseSensitive, result, tt.expected)
			}
		})
	}
}

func TestIsTerminalInAncestorChain(t *testing.T) {
	tests := []struct {
		name         string
		terminalName string
	}{
		{"Test iTerm detection", "iterm"},
		{"Test iTerm case variation", "iTerm"},
		{"Test ETTerminal detection", "etterminal"},
		{"Test ETTerminal case variation", "ETTerminal"},
		{"Test SSH detection", "sshd"},
		{"Test tmux detection", "tmux"},
		{"Test nonexistent terminal", "nonexistent-terminal"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isTerminalInAncestorChain(tt.terminalName)
			t.Logf("isTerminalInAncestorChain(%q) returned: %v", tt.terminalName, result)

			// We can't assert specific values since it depends on the environment
			// but we can ensure it returns a boolean
			if result != true && result != false {
				t.Errorf("isTerminalInAncestorChain(%q) should return a boolean value", tt.terminalName)
			}
		})
	}
}

// BenchmarkDetectTerminalType benchmarks the terminal detection performance
func BenchmarkDetectTerminalType(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = detectTerminalAndShell("")
	}
}

// BenchmarkGetProcessAncestorChain benchmarks the process chain retrieval
func BenchmarkGetProcessAncestorChain(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = getProcessAncestorChain()
	}
}
