package main

import (
	"fmt"
	"os"

	"github.com/mburakmmm/sky-lang/internal/linter"
)

func lintCommand(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Error: no input files specified")
		fmt.Fprintln(os.Stderr, "Usage: sky lint <files...>")
		os.Exit(1)
	}

	hasErrors := false

	for _, filename := range args {
		content, err := os.ReadFile(filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", filename, err)
			os.Exit(1)
		}

		issues, err := linter.LintFile(filename, string(content))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error linting %s: %v\n", filename, err)
			os.Exit(1)
		}

		if len(issues) > 0 {
			for _, issue := range issues {
				fmt.Printf("%s:%d:%d: %s [%s] %s\n",
					issue.File, issue.Line, issue.Column,
					issue.Severity, issue.Rule, issue.Message)

				if issue.Severity == "error" {
					hasErrors = true
				}
			}
		} else {
			fmt.Printf("%s: OK\n", filename)
		}
	}

	if hasErrors {
		os.Exit(1)
	}
}
