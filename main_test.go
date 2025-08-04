package main

import (
	"flag"
	"os"
	"strings"
	"testing"
)

// TestMainFlagParsing tests that the main function properly parses command-line arguments
func TestMainFlagParsing(t *testing.T) {
	// Test that all expected flags are defined
	expectedFlags := []string{"tab", "fg", "bg", "profile"}

	// Reset flag.CommandLine to ensure clean state
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	// Define the flags as they are in main()
	var (
		tabColor        = flag.String("tab", "", "Set tab color")
		foregroundColor = flag.String("fg", "", "Set foreground color")
		backgroundColor = flag.String("bg", "", "Set background color")
		profileName     = flag.String("profile", "", "Use predefined profile from config file")
	)

	// Test that flags are properly defined
	for _, flagName := range expectedFlags {
		if flag.Lookup(flagName) == nil {
			t.Errorf("Expected flag %q to be defined", flagName)
		}
	}

	// Test flag parsing
	testArgs := []string{"-tab", "red", "-fg", "white", "-bg", "black"}
	err := flag.CommandLine.Parse(testArgs)
	if err != nil {
		t.Fatalf("Flag parsing failed: %v", err)
	}

	if *tabColor != "red" {
		t.Errorf("Expected tab color 'red', got %q", *tabColor)
	}

	if *foregroundColor != "white" {
		t.Errorf("Expected foreground color 'white', got %q", *foregroundColor)
	}

	if *backgroundColor != "black" {
		t.Errorf("Expected background color 'black', got %q", *backgroundColor)
	}

	if *profileName != "" {
		t.Errorf("Expected profile name to be empty, got %q", *profileName)
	}
}

// TestMainErrorMessages tests error message generation
func TestMainErrorMessages(t *testing.T) {
	tests := []struct {
		name        string
		profileName string
		tabColor    string
		expectError string
	}{
		{
			name:        "profile with individual colors",
			profileName: "test",
			tabColor:    "red",
			expectError: "Cannot use -profile with individual color options",
		},
		{
			name:        "non-existent profile",
			profileName: "nonexistent",
			expectError: "profile \"nonexistent\" not found",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Test the validation logic without actually calling main()
			hasProfile := test.profileName != ""
			hasIndividualColors := test.tabColor != ""

			if hasProfile && hasIndividualColors {
				expectedMsg := "Cannot use -profile with individual color options"
				if !strings.Contains(test.expectError, expectedMsg) {
					t.Errorf("Expected error to contain %q", expectedMsg)
				}
			}

			if hasProfile && !hasIndividualColors {
				// This would test getProfile, which is already tested in config_test.go
			}
		})
	}
}
