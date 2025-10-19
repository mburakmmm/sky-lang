package main

import (
	"fmt"
	"log"
	"os"

	"github.com/mburakmmm/sky-lang/internal/lsp"
)

const version = "0.1.0"

func main() {
	if len(os.Args) > 1 {
		arg := os.Args[1]
		switch arg {
		case "version", "--version", "-v":
			fmt.Printf("skyls version %s\n", version)
			return
		case "help", "--help", "-h":
			printHelp()
			return
		}
	}

	// Setup logger (stderr için - stdout LSP için kullanılıyor)
	logger := log.New(os.Stderr, "[skyls] ", log.LstdFlags)
	logger.Println("SKY Language Server v" + version)
	logger.Println("Starting LSP server on stdio...")

	// LSP server oluştur ve başlat
	server := lsp.NewServer(os.Stdin, os.Stdout, logger)

	if err := server.Start(); err != nil {
		logger.Fatalf("Server error: %v", err)
	}
}

func printHelp() {
	fmt.Print(`skyls - SKY Language Server

The Language Server Protocol (LSP) implementation for SKY.

USAGE:
  skyls [options]

OPTIONS:
  --version, -v       Show version information
  --help, -h          Show this help message

The language server communicates via stdio and should be
configured in your editor's LSP client settings.

FEATURES (when implemented):
  - Syntax highlighting
  - Auto-completion
  - Go to definition
  - Find references
  - Hover information
  - Diagnostic errors/warnings
  - Code formatting
  - Rename symbols

For editor setup instructions, visit:
https://github.com/mburakmmm/sky-lang/docs/lsp
`)
}
