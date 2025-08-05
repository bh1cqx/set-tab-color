package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// ColorTarget represents the type of color to set
type ColorTarget string

const (
	TabColor        ColorTarget = "tab"
	ForegroundColor ColorTarget = "fg"
	BackgroundColor ColorTarget = "bg"
)

// runSetColor executes it2setcolor with the given color and target
func runSetColor(target ColorTarget, color string) error {
	// Initialize CSS colors if not already done
	if err := initColors(); err != nil {
		return err
	}

	// Normalize user input
	normalizedColor := normalizeColor(color)
	if normalizedColor == "" {
		return fmt.Errorf("unknown color: %s", color)
	}

	// Locate and check existence of custom it2setcolor in ~/.iterm2/
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("could not get home dir: %v", err)
	}
	it2bin := filepath.Join(home, ".iterm2", "it2setcolor")

	if _, err := os.Stat(it2bin); os.IsNotExist(err) {
		return fmt.Errorf("it2setcolor not found at %s", it2bin)
	}

	// Execute it2setcolor with the normalized hex
	cmd := exec.Command(it2bin, string(target), normalizedColor)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// runSetPreset executes it2setcolor preset with the given preset name
func runSetPreset(presetName string) error {
	// Locate and check existence of custom it2setcolor in ~/.iterm2/
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("could not get home dir: %v", err)
	}
	it2bin := filepath.Join(home, ".iterm2", "it2setcolor")

	if _, err := os.Stat(it2bin); os.IsNotExist(err) {
		return fmt.Errorf("it2setcolor not found at %s", it2bin)
	}

	// Execute it2setcolor preset with the preset name
	cmd := exec.Command(it2bin, "preset", presetName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
