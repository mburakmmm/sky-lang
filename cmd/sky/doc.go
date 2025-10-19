package main

import (
	"fmt"
	"os"

	"github.com/mburakmmm/sky-lang/internal/docgen"
)

func docCommand(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Error: no input files specified")
		fmt.Fprintln(os.Stderr, "Usage: sky doc <files...>")
		os.Exit(1)
	}

	output := ""

	for _, filename := range args {
		content, err := os.ReadFile(filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", filename, err)
			os.Exit(1)
		}

		doc, err := docgen.GenerateDoc(filename, string(content))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating docs for %s: %v\n", filename, err)
			os.Exit(1)
		}

		output += doc
	}

	fmt.Print(output)
}
