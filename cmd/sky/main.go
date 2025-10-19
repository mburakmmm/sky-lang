package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/mburakmmm/sky-lang/internal/ast"
	"github.com/mburakmmm/sky-lang/internal/interpreter"
	"github.com/mburakmmm/sky-lang/internal/lexer"
	"github.com/mburakmmm/sky-lang/internal/parser"
	"github.com/mburakmmm/sky-lang/internal/sema"
)

const version = "0.1.0"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "run":
		runCommand(os.Args[2:])
	case "build":
		buildCommand(os.Args[2:])
	case "test":
		testCommand(os.Args[2:])
	case "repl":
		replCommand(os.Args[2:])
	case "dump":
		dumpCommand(os.Args[2:])
	case "check":
		checkCommand(os.Args[2:])
	case "version", "--version", "-v":
		fmt.Printf("SKY version %s\n", version)
	case "help", "--help", "-h":
		printHelp()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `Usage: sky <command> [options]

Commands:
  run <file>        Run a SKY program (JIT)
  build <file>      Build a SKY program (AOT)
  test              Run tests
  repl              Start interactive REPL
  dump <options>    Dump lexer/parser output
  check <file>      Check semantics without running
  version           Show version
  help              Show this help

Use "sky <command> --help" for more information about a command.
`)
}

func printHelp() {
	fmt.Print(`SKY Programming Language

A modern, safe programming language with JIT compilation,
optional static typing, and powerful concurrency features.

USAGE:
  sky <command> [options]

COMMANDS:
  run <file>              Run a SKY program using JIT compilation
  build <file>            Compile to native binary (AOT)
  test [path]             Run test files
  repl                    Start interactive REPL
  dump --tokens <file>    Show lexer tokens
  dump --ast <file>       Show AST structure
  check <file>            Type check without execution
  version                 Show version information
  help                    Show this help message

EXAMPLES:
  sky run hello.sky                 # Run a program
  sky dump --tokens hello.sky       # Show tokens
  sky dump --ast hello.sky          # Show AST
  sky check myprogram.sky           # Type check
  sky build -o myapp main.sky       # Build binary

For more information, visit: https://github.com/mburakmmm/sky-lang
`)
}

func runCommand(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Error: no input file specified")
		fmt.Fprintln(os.Stderr, "Usage: sky run [--vm|--jit] <file>")
		os.Exit(1)
	}

	// Check for mode flags
	useVMMode := false
	useJITMode := false
	filename := args[0]

	if args[0] == "--vm" {
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Error: no input file specified")
			fmt.Fprintln(os.Stderr, "Usage: sky run --vm <file>")
			os.Exit(1)
		}
		useVMMode = true
		filename = args[1]
	} else if args[0] == "--jit" {
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Error: no input file specified")
			fmt.Fprintln(os.Stderr, "Usage: sky run --jit <file>")
			os.Exit(1)
		}
		useJITMode = true
		filename = args[1]
	}

	// Use JIT mode if requested (LLVM backend)
	if useJITMode {
		if err := runWithJIT(filename); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		return
	}

	// Use VM mode if requested (better recursion support)
	if useVMMode {
		if err := runWithVM(filename); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		return
	}

	// Regular interpreter mode
	filename = args[0]
	content, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	// Lexer & Parser
	l := lexer.New(string(content), filename)
	p := parser.New(l)
	program := p.ParseProgram()

	// Parser hataları
	if len(p.Errors()) > 0 {
		fmt.Fprintln(os.Stderr, "Parse errors:")
		for _, err := range p.Errors() {
			fmt.Fprintf(os.Stderr, "  - %s\n", err)
		}
		os.Exit(1)
	}

	// Semantic checker (skip if imports present, as they're resolved at runtime)
	hasImports := false
	for _, stmt := range program.Statements {
		if _, ok := stmt.(*ast.ImportStatement); ok {
			hasImports = true
			break
		}
	}

	if !hasImports {
		checker := sema.NewChecker()
		errors := checker.Check(program)

		if len(errors) > 0 {
			fmt.Fprintln(os.Stderr, "Semantic errors:")
			for _, err := range errors {
				fmt.Fprintf(os.Stderr, "  - %s\n", err)
			}
			os.Exit(1)
		}
	}

	// Interpreter
	interp := interpreter.New()
	interp.SetSourceFile(filename)
	err = interp.Eval(program)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Runtime error: %v\n", err)
		os.Exit(1)
	}
}

func buildCommand(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Error: no input file specified")
		fmt.Fprintln(os.Stderr, "Usage: sky build [-o output] <file>")
		os.Exit(1)
	}

	// Parse flags
	outputFile := "a.out"
	filename := args[0]

	for i := 0; i < len(args); i++ {
		if args[i] == "-o" && i+1 < len(args) {
			outputFile = args[i+1]
			i++
		} else if args[i][0] != '-' {
			filename = args[i]
		}
	}

	fmt.Printf("Building %s...\n", filename)

	// Read and parse file
	content, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	l := lexer.New(string(content), filename)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		fmt.Fprintln(os.Stderr, "Parse errors:")
		for _, err := range p.Errors() {
			fmt.Fprintf(os.Stderr, "  - %s\n", err)
		}
		os.Exit(1)
	}

	// Semantic check
	hasImports := false
	for _, stmt := range program.Statements {
		if _, ok := stmt.(*ast.ImportStatement); ok {
			hasImports = true
			break
		}
	}

	if !hasImports {
		checker := sema.NewChecker()
		errors := checker.Check(program)

		if len(errors) > 0 {
			fmt.Fprintln(os.Stderr, "Semantic errors:")
			for _, err := range errors {
				fmt.Fprintf(os.Stderr, "  - %s\n", err)
			}
			os.Exit(1)
		}
	}

	// AOT compile
	if err := compileAOT(program, outputFile); err != nil {
		fmt.Fprintf(os.Stderr, "Build error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully built: %s\n", outputFile)
}

func testCommand(args []string) {
	testDir := "tests"
	if len(args) > 0 {
		testDir = args[0]
	}

	fmt.Printf("Running tests in %s/\n\n", testDir)

	// Find all .sky test files
	testFiles, err := filepath.Glob(filepath.Join(testDir, "*.sky"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding test files: %v\n", err)
		os.Exit(1)
	}

	if len(testFiles) == 0 {
		fmt.Printf("No test files found in %s/\n", testDir)
		return
	}

	passed := 0
	failed := 0

	for _, testFile := range testFiles {
		// Expected output file
		expectedFile := strings.TrimSuffix(testFile, ".sky") + ".expected"

		fmt.Printf("Testing %s... ", filepath.Base(testFile))

		// Run test
		result, err := runTest(testFile)
		if err != nil {
			fmt.Printf("❌ FAIL: %v\n", err)
			failed++
			continue
		}

		// Check expected output if exists
		if _, err := os.Stat(expectedFile); err == nil {
			expected, _ := os.ReadFile(expectedFile)
			if string(expected) != result {
				fmt.Printf("❌ FAIL: output mismatch\n")
				fmt.Printf("  Expected: %s\n", string(expected))
				fmt.Printf("  Got: %s\n", result)
				failed++
				continue
			}
		}

		fmt.Println("✅ PASS")
		passed++
	}

	fmt.Printf("\n%d passed, %d failed\n", passed, failed)

	if failed > 0 {
		os.Exit(1)
	}
}

func runTest(filename string) (string, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run
	l := lexer.New(string(content), filename)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		w.Close()
		os.Stdout = oldStdout
		return "", fmt.Errorf("parse error: %v", p.Errors())
	}

	checker := sema.NewChecker()
	errors := checker.Check(program)
	if len(errors) > 0 {
		w.Close()
		os.Stdout = oldStdout
		return "", fmt.Errorf("semantic error: %v", errors)
	}

	interp := interpreter.New()
	err = interp.Eval(program)

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read captured output
	var buf strings.Builder
	io.Copy(&buf, r)

	return buf.String(), err
}

// replCommand defined in repl.go

func checkCommand(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Error: no input file specified")
		fmt.Fprintln(os.Stderr, "Usage: sky check <file>")
		os.Exit(1)
	}

	filename := args[0]
	content, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Checking %s...\n\n", filename)

	// Lexer & Parser
	l := lexer.New(string(content), filename)
	p := parser.New(l)
	program := p.ParseProgram()

	// Parser hataları
	if len(p.Errors()) > 0 {
		fmt.Println("Parse errors:")
		for _, err := range p.Errors() {
			fmt.Printf("  - %s\n", err)
		}
		os.Exit(1)
	}

	// Semantic checker
	checker := sema.NewChecker()
	errors := checker.Check(program)

	if len(errors) > 0 {
		fmt.Println("Semantic errors:")
		for _, err := range errors {
			fmt.Printf("  - %s\n", err)
		}
		fmt.Printf("\n❌ Found %d error(s)\n", len(errors))
		os.Exit(1)
	}

	fmt.Println("✅ No errors found")
}
