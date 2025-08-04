package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	// Define command-line flags
	var (
		tabColor        = flag.String("tab", "", "Set tab color")
		foregroundColor = flag.String("fg", "", "Set foreground color")
		backgroundColor = flag.String("bg", "", "Set background color")
		profileName     = flag.String("profile", "", "Use predefined profile from config file")
		listProfiles    = flag.Bool("list-profiles", false, "List all available profiles")
		listColors      = flag.Bool("list-colors", false, "List all available CSS color names")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nOptions:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nColor formats supported:\n")
		fmt.Fprintf(os.Stderr, "  - Hex colors: #f80, #ff8800\n")
		fmt.Fprintf(os.Stderr, "  - CSS color names: red, blue, lightblue, etc.\n")
		fmt.Fprintf(os.Stderr, "  - default: restore default color\n")
		fmt.Fprintf(os.Stderr, "\nConfiguration:\n")
		fmt.Fprintf(os.Stderr, "  Config file: ~/.config/set-tab-color.toml (or $SET_TAB_COLOR_CONFIG)\n")
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s -tab red\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -fg white -bg black\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -tab #ff8800 -fg lightblue\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -profile myprofile\n", os.Args[0])
	}

	flag.Parse()

	// Handle listing operations
	if *listProfiles {
		profiles, err := listProfileNames()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading profiles: %v\n", err)
			os.Exit(1)
		}

		if len(profiles) == 0 {
			fmt.Println("No profiles found.")
		} else {
			fmt.Println("Available profiles:")
			for _, name := range profiles {
				fmt.Printf("  %s\n", name)
			}
		}
		return
	}

	if *listColors {
		colors, err := listCSSColorNames()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading CSS colors: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Available CSS color names:")
		for _, name := range colors {
			fmt.Printf("  %s\n", name)
		}
		return
	}

	// Handle profile-based configuration
	if *profileName != "" {
		// Cannot mix profile with individual colors
		if *tabColor != "" || *foregroundColor != "" || *backgroundColor != "" {
			fmt.Fprintf(os.Stderr, "Error: Cannot use -profile with individual color options\n\n")
			flag.Usage()
			os.Exit(1)
		}

		profile, err := getProfile(*profileName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading profile: %v\n", err)
			os.Exit(1)
		}

		if err := applyProfile(profile); err != nil {
			fmt.Fprintf(os.Stderr, "Error applying profile: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Check if at least one color option was provided
	if *tabColor == "" && *foregroundColor == "" && *backgroundColor == "" {
		fmt.Fprintf(os.Stderr, "Error: At least one color option or profile must be specified\n\n")
		flag.Usage()
		os.Exit(1)
	}

	// Set colors based on provided arguments
	if *tabColor != "" {
		if err := runSetColor(TabColor, *tabColor); err != nil {
			fmt.Fprintf(os.Stderr, "Error setting tab color: %v\n", err)
			os.Exit(1)
		}
	}

	if *foregroundColor != "" {
		if err := runSetColor(ForegroundColor, *foregroundColor); err != nil {
			fmt.Fprintf(os.Stderr, "Error setting foreground color: %v\n", err)
			os.Exit(1)
		}
	}

	if *backgroundColor != "" {
		if err := runSetColor(BackgroundColor, *backgroundColor); err != nil {
			fmt.Fprintf(os.Stderr, "Error setting background color: %v\n", err)
			os.Exit(1)
		}
	}
}
