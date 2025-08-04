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

// detectTerminalType detects the current terminal type by walking up the process ancestry
// and returning the first terminal type encountered
func detectTerminalType() TerminalType {
	// Get current process
	currentPid := int32(os.Getpid())
	proc, err := process.NewProcess(currentPid)
	if err != nil {
		return TerminalTypeUnknown
	}

	// Walk up the process tree looking for terminal types
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

		// Check for terminal types in priority order
		// Case-sensitive for all except iTerm

		// Check for SSH first (highest priority)
		if matchesTerminalName(name, "sshd", true) {
			return TerminalTypeSSH
		}

		// Check for tmux
		if matchesTerminalName(name, "tmux", true) {
			return TerminalTypeTmux
		}

		// Check for ETTerminal
		if matchesTerminalName(name, "etterminal", true) {
			return TerminalTypeETTerminal
		}

		// Check for iTerm2 (case-insensitive)
		if matchesTerminalName(name, "iterm2", false) {
			return TerminalTypeITerm2
		}

		// Check for VSCode (case-insensitive for "Code Helper")
		if matchesTerminalName(name, "Code Helper", false) {
			return TerminalTypeVSCode
		}
	}

	return TerminalTypeUnknown
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

	// Prefix match with space
	if strings.HasPrefix(name, terminal+" ") {
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
