package main

import (
	"os"
	"strings"

	"github.com/shirou/gopsutil/v3/process"
)

// TerminalType represents different terminal types
type TerminalType string

const (
	TerminalTypeUnknown    TerminalType = "unknown"
	TerminalTypeITerm2     TerminalType = "iterm2"
	TerminalTypeETTerminal TerminalType = "etterminal"
	TerminalTypeSSH        TerminalType = "ssh"
	TerminalTypeTmux       TerminalType = "tmux"
	TerminalTypeVSCode     TerminalType = "vscode"
)

// ShellType represents different shell types
type ShellType string

const (
	ShellTypeUnknown ShellType = "unknown"
	ShellTypeBash    ShellType = "bash"
	ShellTypeZsh     ShellType = "zsh"
	ShellTypeFish    ShellType = "fish"
	ShellTypeTcsh    ShellType = "tcsh"
	ShellTypeCsh     ShellType = "csh"
	ShellTypeKsh     ShellType = "ksh"
	ShellTypeSh      ShellType = "sh"
)

// TerminalShellInfo contains both terminal and shell detection results
type TerminalShellInfo struct {
	Terminals []TerminalType // All terminals found in process chain, in order
	Shell     ShellType
	Valid     bool // true if shell comes before terminal in the process chain
}

// detectTerminalAndShell detects both terminal and shell types with validation
// that shell should come before terminal in the process ancestry
func detectTerminalAndShell() TerminalShellInfo {
	// Get current process
	currentPid := int32(os.Getpid())
	proc, err := process.NewProcess(currentPid)
	if err != nil {
		return TerminalShellInfo{
			Terminals: []TerminalType{},
			Shell:     ShellTypeUnknown,
			Valid:     false,
		}
	}

	var foundShell ShellType = ShellTypeUnknown
	var terminals []TerminalType
	var shellFoundFirst bool

	// Walk up the process tree looking for both shell and terminal types
	for {
		// Get parent process first (skip current process)
		parentPid, err := proc.Ppid()
		if err != nil || parentPid <= 1 {
			break
		}

		// Move to parent process
		proc, err = process.NewProcess(parentPid)
		if err != nil {
			break
		}

		// Get process name
		name, err := proc.Name()
		if err != nil {
			continue
		}

		// Check for shell types first (if we haven't found one yet)
		if foundShell == ShellTypeUnknown {
			if matchesTerminalName(name, "zsh", true) {
				foundShell = ShellTypeZsh
				shellFoundFirst = (len(terminals) == 0)
			} else if matchesTerminalName(name, "bash", true) {
				foundShell = ShellTypeBash
				shellFoundFirst = (len(terminals) == 0)
			} else if matchesTerminalName(name, "fish", true) {
				foundShell = ShellTypeFish
				shellFoundFirst = (len(terminals) == 0)
			} else if matchesTerminalName(name, "tcsh", true) {
				foundShell = ShellTypeTcsh
				shellFoundFirst = (len(terminals) == 0)
			} else if matchesTerminalName(name, "csh", true) {
				foundShell = ShellTypeCsh
				shellFoundFirst = (len(terminals) == 0)
			} else if matchesTerminalName(name, "ksh", true) {
				foundShell = ShellTypeKsh
				shellFoundFirst = (len(terminals) == 0)
			} else if matchesTerminalName(name, "sh", true) {
				foundShell = ShellTypeSh
				shellFoundFirst = (len(terminals) == 0)
			}
		}

		// Check for terminal types and collect all of them
		if matchesTerminalName(name, "sshd", true) {
			terminals = append(terminals, TerminalTypeSSH)
		} else if matchesTerminalName(name, "tmux", true) {
			terminals = append(terminals, TerminalTypeTmux)
		} else if matchesTerminalName(name, "etterminal", true) {
			terminals = append(terminals, TerminalTypeETTerminal)
		} else if matchesTerminalName(name, "iterm2", false) {
			terminals = append(terminals, TerminalTypeITerm2)
		} else if matchesTerminalName(name, "Code Helper", false) {
			terminals = append(terminals, TerminalTypeVSCode)
		}
	}

	return TerminalShellInfo{
		Terminals: terminals,
		Shell:     foundShell,
		Valid:     shellFoundFirst || (foundShell != ShellTypeUnknown && len(terminals) == 0),
	}
}

// terminalChainDetector is a function type that can be mocked in tests
var terminalChainDetector = detectAllTerminalsInChainImpl

// detectAllTerminalsInChain detects all terminal types in the process ancestry chain
func detectAllTerminalsInChain() []TerminalType {
	return terminalChainDetector()
}

// detectAllTerminalsInChainImpl is the actual implementation
func detectAllTerminalsInChainImpl() []TerminalType {
	// Get current process
	currentPid := int32(os.Getpid())
	proc, err := process.NewProcess(currentPid)
	if err != nil {
		return nil
	}

	var terminals []TerminalType

	// Walk up the process tree looking for all terminal types
	for {
		// Get parent process first (skip current process)
		parentPid, err := proc.Ppid()
		if err != nil || parentPid <= 1 {
			break
		}

		// Move to parent process
		proc, err = process.NewProcess(parentPid)
		if err != nil {
			break
		}

		// Get process name
		name, err := proc.Name()
		if err != nil {
			continue
		}

		// Check for terminal types
		if matchesTerminalName(name, "sshd", true) {
			terminals = append(terminals, TerminalTypeSSH)
		} else if matchesTerminalName(name, "tmux", true) {
			terminals = append(terminals, TerminalTypeTmux)
		} else if matchesTerminalName(name, "etterminal", true) {
			terminals = append(terminals, TerminalTypeETTerminal)
		} else if matchesTerminalName(name, "iterm2", false) {
			terminals = append(terminals, TerminalTypeITerm2)
		} else if matchesTerminalName(name, "Code Helper", false) {
			terminals = append(terminals, TerminalTypeVSCode)
		}
	}

	return terminals
}

// detectShellType detects shell type for backwards compatibility
func detectShellType() ShellType {
	info := detectTerminalAndShell()
	return info.Shell
}

// matchesTerminalName checks if a process name matches a terminal name
// either exactly or as a prefix followed by a space
func matchesTerminalName(processName, terminalName string, caseSensitive bool) bool {
	var name, terminal string

	if caseSensitive {
		name = processName
		terminal = terminalName
	} else {
		name = strings.ToLower(processName)
		terminal = strings.ToLower(terminalName)
	}

	// Exact match
	if name == terminal {
		return true
	}

	// Prefix match with space or colon (e.g., "tmux: server")
	if strings.HasPrefix(name, terminal+" ") || strings.HasPrefix(name, terminal+":") {
		return true
	}

	return false
}

// isTerminalInAncestorChain checks if a specific terminal name appears in the process ancestor chain
func isTerminalInAncestorChain(terminalName string) bool {
	// Get current process
	currentPid := int32(os.Getpid())
	proc, err := process.NewProcess(currentPid)
	if err != nil {
		return false
	}

	// Use case-insensitive matching for iterm, case-sensitive for others
	caseSensitive := strings.ToLower(terminalName) != "iterm"

	// Walk up the process tree looking for the terminal
	for {
		// Get process name
		name, err := proc.Name()
		if err != nil {
			break
		}

		// Check if the process name matches the terminal name
		if matchesTerminalName(name, terminalName, caseSensitive) {
			return true
		}

		// Get parent process
		parentPid, err := proc.Ppid()
		if err != nil || parentPid <= 1 {
			break
		}

		// Move to parent process
		proc, err = process.NewProcess(parentPid)
		if err != nil {
			break
		}
	}

	return false
}

// getProcessAncestorChain returns the full ancestor chain for debugging/logging purposes
func getProcessAncestorChain() ([]string, error) {
	var chain []string
	currentPid := int32(os.Getpid())
	proc, err := process.NewProcess(currentPid)
	if err != nil {
		return nil, err
	}

	for {
		// Get process name
		name, err := proc.Name()
		if err != nil {
			break
		}

		chain = append(chain, name)

		// Get parent process
		parentPid, err := proc.Ppid()
		if err != nil || parentPid <= 1 {
			break
		}

		// Move to parent process
		proc, err = process.NewProcess(parentPid)
		if err != nil {
			break
		}
	}

	return chain, nil
}

// ProcessInfo contains information about a process in the ancestor chain
type ProcessInfo struct {
	PID  int32
	Name string
}

// getProcessAncestorChainDetailed returns detailed information about the process ancestor chain
func getProcessAncestorChainDetailed() ([]ProcessInfo, error) {
	var chain []ProcessInfo
	currentPid := int32(os.Getpid())
	proc, err := process.NewProcess(currentPid)
	if err != nil {
		return nil, err
	}

	for {
		// Get process name
		name, err := proc.Name()
		if err != nil {
			break
		}

		chain = append(chain, ProcessInfo{
			PID:  proc.Pid,
			Name: name,
		})

		// Get parent process
		parentPid, err := proc.Ppid()
		if err != nil || parentPid <= 1 {
			break
		}

		// Move to parent process
		proc, err = process.NewProcess(parentPid)
		if err != nil {
			break
		}
	}

	return chain, nil
}
