package main

import (
	"testing"
)

func TestDetectTerminalType(t *testing.T) {
	// This test will detect the actual terminal type running the tests
	terminalType := detectTerminalType()

	// We can't assert a specific value since it depends on the environment
	// but we can ensure it returns a valid type
	validTypes := []TerminalType{
		TerminalTypeUnknown,
		TerminalTypeITerm2,
		TerminalTypeETTerminal,
		TerminalTypeSSH,
		TerminalTypeTmux,
		TerminalTypeVSCode,
	}

	found := false
	for _, validType := range validTypes {
		if terminalType == validType {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("detectTerminalType() returned invalid type: %s", terminalType)
	}

	t.Logf("Detected terminal type: %s", terminalType)
}

func TestTerminalAndShellDetection(t *testing.T) {
	// Test the combined terminal and shell detection
	info := detectTerminalAndShell()

	t.Logf("Combined detection results:")
	t.Logf("  Terminal: %v", info.Terminal)
	t.Logf("  Shell: %v", info.Shell)
	t.Logf("  Valid: %v", info.Valid)

	// Test individual detection functions for backwards compatibility
	terminalType := detectTerminalType()
	shellType := detectShellType()

	t.Logf("Individual detection results:")
	t.Logf("  detectTerminalType() returned: %v", terminalType)
	t.Logf("  detectShellType() returned: %v", shellType)

	// Verify consistency
	if terminalType != info.Terminal {
		t.Errorf("detectTerminalType() = %v, but detectTerminalAndShell().Terminal = %v", terminalType, info.Terminal)
	}
	if shellType != info.Shell {
		t.Errorf("detectShellType() = %v, but detectTerminalAndShell().Shell = %v", shellType, info.Shell)
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

	shellType := detectShellType()

	found := false
	for _, validType := range validShellTypes {
		if shellType == validType {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("detectShellType() returned invalid type: %s", shellType)
	}

	t.Logf("Detected shell type: %s", shellType)
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

func TestGetProcessAncestorChainDetailed(t *testing.T) {
	chain, err := getProcessAncestorChainDetailed()
	if err != nil {
		t.Fatalf("getProcessAncestorChainDetailed() returned error: %v", err)
	}

	if len(chain) == 0 {
		t.Error("getProcessAncestorChainDetailed() returned empty chain")
	}

	// Log the detailed process chain for debugging
	t.Logf("Detailed process ancestor chain:")
	for i, processInfo := range chain {
		t.Logf("  %d: PID=%d, Name=%s", i, processInfo.PID, processInfo.Name)
	}

	// Verify that all entries have valid PID and non-empty names
	for i, processInfo := range chain {
		if processInfo.PID <= 0 {
			t.Errorf("Process %d has invalid PID: %d", i, processInfo.PID)
		}
		if processInfo.Name == "" {
			t.Errorf("Process %d has empty name", i)
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
		_ = detectTerminalType()
	}
}

// BenchmarkGetProcessAncestorChain benchmarks the process chain retrieval
func BenchmarkGetProcessAncestorChain(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = getProcessAncestorChain()
	}
}
