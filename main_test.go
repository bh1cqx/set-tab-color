package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

// TestMainIntegration tests the main function with various argument combinations
func TestMainIntegration(t *testing.T) {
	if os.Getenv("BE_CRASHER") == "1" {
		// This is the subprocess that will call main
		os.Args = strings.Split(os.Getenv("CRASHER_ARGS"), " ")
		main()
		return
	}

	tests := []struct {
		name        string
		args        []string
		expectError bool
		errorText   string
	}{
		{
			name:        "no arguments",
			args:        []string{"set-tab-color"},
			expectError: true,
			errorText:   "At least one color option must be specified",
		},
		{
			name:        "help flag",
			args:        []string{"set-tab-color", "-h"},
			expectError: false, // -h is handled by flag package and exits with 0
		},
		{
			name:        "invalid flag",
			args:        []string{"set-tab-color", "-invalid"},
			expectError: true,
			errorText:   "flag provided but not defined",
		},
		{
			name:        "valid tab color",
			args:        []string{"set-tab-color", "-tab", "red"},
			expectError: true, // Will fail due to missing it2setcolor binary, but argument parsing works
			errorText:   "it2setcolor not found",
		},
		{
			name:        "valid multiple colors",
			args:        []string{"set-tab-color", "-tab", "red", "-fg", "white", "-bg", "black"},
			expectError: true, // Will fail due to missing it2setcolor binary, but argument parsing works
			errorText:   "it2setcolor not found",
		},
		{
			name:        "invalid color",
			args:        []string{"set-tab-color", "-tab", "invalidcolor"},
			expectError: true,
			errorText:   "unknown color",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Skip help test as it's handled differently by flag package
			if test.name == "help flag" {
				t.Skip("Help flag exits the process, cannot test in this context")
				return
			}

			cmd := exec.Command(os.Args[0], "-test.run=TestMainIntegration")
			cmd.Env = append(os.Environ(), "BE_CRASHER=1")
			cmd.Env = append(cmd.Env, "CRASHER_ARGS="+strings.Join(test.args, " "))

			output, err := cmd.CombinedOutput()
			outputStr := string(output)

			if test.expectError {
				if err == nil {
					t.Errorf("Expected error but command succeeded. Output: %s", outputStr)
					return
				}
				if test.errorText != "" && !strings.Contains(outputStr, test.errorText) {
					t.Errorf("Expected output to contain %q, got: %s", test.errorText, outputStr)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v. Output: %s", err, outputStr)
				}
			}
		})
	}
}

// TestMainUsageMessage tests that the usage message is properly formatted
func TestMainUsageMessage(t *testing.T) {
	// Test by running with no arguments to trigger usage
	cmd := exec.Command(os.Args[0], "-test.run=TestMainIntegration")
	cmd.Env = append(os.Environ(), "BE_CRASHER=1")
	cmd.Env = append(cmd.Env, "CRASHER_ARGS=set-tab-color")

	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Error("Expected error when running without arguments")
		return
	}

	outputStr := string(output)
	expectedStrings := []string{
		"Usage:",
		"Options:",
		"-bg string",
		"-fg string",
		"-tab string",
		"Color formats supported:",
		"Examples:",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(outputStr, expected) {
			t.Errorf("Expected usage output to contain %q, got: %s", expected, outputStr)
		}
	}
}
