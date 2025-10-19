package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mburakmmm/sky-lang/internal/formatter"
)

func fmtCommand(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Error: no input files specified")
		fmt.Fprintln(os.Stderr, "Usage: sky fmt <files...>")
		os.Exit(1)
	}

	check := false
	files := []string{}

	for _, arg := range args {
		if arg == "--check" || arg == "-c" {
			check = true
		} else {
			files = append(files, arg)
		}
	}

	if len(files) == 0 {
		fmt.Fprintln(os.Stderr, "Error: no input files specified")
		os.Exit(1)
	}

	hasChanges := false

	for _, filename := range files {
		// Read file
		content, err := os.ReadFile(filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", filename, err)
			os.Exit(1)
		}

		// Format
		formatted, err := formatter.FormatFile(filename, string(content))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error formatting %s: %v\n", filename, err)
			os.Exit(1)
		}

		if check {
			// Check mode: compare without writing
			if string(content) != formatted {
				fmt.Fprintf(os.Stderr, "%s: not formatted\n", filename)
				hasChanges = true
			}
		} else {
			// Write mode: save formatted output
			if string(content) != formatted {
				if err := os.WriteFile(filename, []byte(formatted), 0644); err != nil {
					fmt.Fprintf(os.Stderr, "Error writing %s: %v\n", filename, err)
					os.Exit(1)
				}
				fmt.Printf("Formatted: %s\n", filename)
			} else {
				fmt.Printf("Already formatted: %s\n", filename)
			}
		}
	}

	if check && hasChanges {
		os.Exit(1)
	}
}

func fmtAllCommand() {
	// Find all .sky files
	var skyFiles []string
	
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".sky" {
			skyFiles = append(skyFiles, path)
		}
		return nil
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding files: %v\n", err)
		os.Exit(1)
	}

	if len(skyFiles) == 0 {
		fmt.Println("No .sky files found")
		return
	}

	fmt.Printf("Formatting %d files...\n", len(skyFiles))
	fmtCommand(skyFiles)
}

