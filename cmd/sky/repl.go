package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/mburakmmm/sky-lang/internal/interpreter"
	"github.com/mburakmmm/sky-lang/internal/lexer"
	"github.com/mburakmmm/sky-lang/internal/parser"
	"github.com/mburakmmm/sky-lang/internal/sema"
)

const replPrompt = "sky> "
const replContinue = "...  "

func replCommand(args []string) {
	fmt.Printf("SKY REPL v%s\n", version)
	fmt.Println("Type 'exit' or 'quit' to exit, 'help' for help")
	fmt.Println()

	interp := interpreter.New()
	scanner := bufio.NewScanner(os.Stdin)
	
	lineBuffer := ""
	inBlock := false
	indentLevel := 0

	for {
		// Prompt
		if inBlock {
			fmt.Print(replContinue)
		} else {
			fmt.Print(replPrompt)
		}

		if !scanner.Scan() {
			break
		}

		line := scanner.Text()

		// Handle special commands
		if !inBlock {
			switch strings.TrimSpace(line) {
			case "exit", "quit":
				fmt.Println("Goodbye!")
				return
			case "help":
				printReplHelp()
				continue
			case "":
				continue
			}
		}

		// Check for block start/continuation
		trimmed := strings.TrimSpace(line)
		if strings.HasSuffix(trimmed, ":") || 
		   strings.HasPrefix(trimmed, "if ") ||
		   strings.HasPrefix(trimmed, "while ") ||
		   strings.HasPrefix(trimmed, "for ") ||
		   strings.HasPrefix(trimmed, "function ") ||
		   strings.HasPrefix(trimmed, "class ") {
			inBlock = true
			indentLevel++
		}

		// Add line to buffer
		if lineBuffer != "" {
			lineBuffer += "\n"
		}
		lineBuffer += line

		// Check for block end
		if inBlock && (trimmed == "end" || trimmed == "") {
			if trimmed == "end" {
				indentLevel--
			}
			if indentLevel <= 0 {
				inBlock = false
			} else if trimmed == "" {
				continue
			}
		}

		// If not in block, evaluate
		if !inBlock {
			result, err := evalREPL(lineBuffer, interp)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			} else if result != "" {
				fmt.Println(result)
			}
			lineBuffer = ""
		}
	}
}

func evalREPL(input string, interp *interpreter.Interpreter) (string, error) {
	// Lexer
	l := lexer.New(input, "repl")
	
	// Parser
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		return "", fmt.Errorf("parse error: %s", strings.Join(p.Errors(), ", "))
	}

	// Semantic check
	checker := sema.NewChecker()
	errors := checker.Check(program)
	if len(errors) > 0 {
		errorMsgs := make([]string, len(errors))
		for i, err := range errors {
			errorMsgs[i] = err.Error()
		}
		return "", fmt.Errorf("semantic error: %s", strings.Join(errorMsgs, ", "))
	}

	// Evaluate
	err := interp.Eval(program)
	if err != nil {
		return "", err
	}

	return "", nil
}

func printReplHelp() {
	fmt.Println(`SKY REPL Commands:

  exit, quit    Exit the REPL
  help          Show this help
  
You can type any SKY code:

  let x = 10
  print(x)
  
  function greet(name)
    print("Hello, " + name)
  end
  
  greet("World")

Multi-line statements:
  
  if x > 5
    print("big")
  end

Note: REPL maintains state between statements.
`)
}

