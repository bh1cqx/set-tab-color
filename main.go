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
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nOptions:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nColor formats supported:\n")
		fmt.Fprintf(os.Stderr, "  - Hex colors: #f80, #ff8800\n")
		fmt.Fprintf(os.Stderr, "  - CSS color names: red, blue, lightblue, etc.\n")
		fmt.Fprintf(os.Stderr, "  - default: restore default color\n")
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s -tab red\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -fg white -bg black\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -tab #ff8800 -fg lightblue\n", os.Args[0])
	}

	flag.Parse()

	// Check if at least one color option was provided
	if *tabColor == "" && *foregroundColor == "" && *backgroundColor == "" {
		fmt.Fprintf(os.Stderr, "Error: At least one color option must be specified\n\n")
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
