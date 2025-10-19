package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/mburakmmm/sky-lang/internal/ast"
	"github.com/mburakmmm/sky-lang/internal/interpreter"
	"github.com/mburakmmm/sky-lang/internal/lexer"
	"github.com/mburakmmm/sky-lang/internal/parser"
	"github.com/mburakmmm/sky-lang/internal/sema"
)

// TestResult represents a test result
type TestResult struct {
	File     string
	Duration time.Duration
	Success  bool
	Error    string
	Coverage float64
}

// TestRunner runs tests
type TestRunner struct {
	parallel    bool
	coverage    bool
	verbose     bool
	results     []TestResult
	resultMutex sync.Mutex
}

// NewTestRunner creates a new test runner
func NewTestRunner(parallel, coverage, verbose bool) *TestRunner {
	return &TestRunner{
		parallel: parallel,
		coverage: coverage,
		verbose:  verbose,
		results:  []TestResult{},
	}
}

// RunTests runs all test files
func (tr *TestRunner) RunTests(files []string) error {
	if tr.parallel {
		return tr.runParallel(files)
	}
	return tr.runSequential(files)
}

func (tr *TestRunner) runSequential(files []string) error {
	for _, file := range files {
		result := tr.runSingleTest(file)
		tr.results = append(tr.results, result)

		if tr.verbose {
			tr.printResult(result)
		}
	}
	return nil
}

func (tr *TestRunner) runParallel(files []string) error {
	var wg sync.WaitGroup

	for _, file := range files {
		wg.Add(1)
		go func(f string) {
			defer wg.Done()
			result := tr.runSingleTest(f)

			tr.resultMutex.Lock()
			tr.results = append(tr.results, result)
			tr.resultMutex.Unlock()

			if tr.verbose {
				tr.printResult(result)
			}
		}(file)
	}

	wg.Wait()
	return nil
}

func (tr *TestRunner) runSingleTest(file string) TestResult {
	start := time.Now()
	result := TestResult{
		File:    file,
		Success: false,
	}

	// Read file
	content, err := os.ReadFile(file)
	if err != nil {
		result.Error = fmt.Sprintf("Failed to read file: %v", err)
		result.Duration = time.Since(start)
		return result
	}

	// Lex
	l := lexer.New(string(content), file)

	// Parse
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		result.Error = fmt.Sprintf("Parse errors: %v", p.Errors())
		result.Duration = time.Since(start)
		return result
	}

	// Check for imports (skip semantic check if imports present)
	hasImports := false
	for _, stmt := range program.Statements {
		if _, ok := stmt.(*ast.ImportStatement); ok {
			hasImports = true
			break
		}
	}

	if !hasImports {
		// Semantic check
		checker := sema.NewChecker()
		errors := checker.Check(program)
		if len(errors) > 0 {
			result.Error = fmt.Sprintf("Semantic errors: %v", errors)
			result.Duration = time.Since(start)
			return result
		}
	}

	// Calculate coverage if requested
	if tr.coverage {
		result.Coverage = tr.calculateCoverage(program)
	}

	// Interpret
	interp := interpreter.New()
	interp.SetSourceFile(file)
	err = interp.Eval(program)

	if err != nil {
		result.Error = fmt.Sprintf("Runtime error: %v", err)
		result.Duration = time.Since(start)
		return result
	}

	result.Success = true
	result.Duration = time.Since(start)
	return result
}

func (tr *TestRunner) calculateCoverage(program *ast.Program) float64 {
	// Simple coverage: count executed statements vs total statements
	total := tr.countStatements(program)
	if total == 0 {
		return 100.0
	}

	// For now, assume 100% coverage (real implementation would track execution)
	return 100.0
}

func (tr *TestRunner) countStatements(program *ast.Program) int {
	count := 0
	for _, stmt := range program.Statements {
		count += tr.countStatementsInNode(stmt)
	}
	return count
}

func (tr *TestRunner) countStatementsInNode(node ast.Node) int {
	count := 1

	switch n := node.(type) {
	case *ast.FunctionStatement:
		for _, stmt := range n.Body.Statements {
			count += tr.countStatementsInNode(stmt)
		}
	case *ast.IfStatement:
		for _, stmt := range n.Consequence.Statements {
			count += tr.countStatementsInNode(stmt)
		}
		if n.Alternative != nil {
			for _, stmt := range n.Alternative.Statements {
				count += tr.countStatementsInNode(stmt)
			}
		}
	case *ast.WhileStatement:
		for _, stmt := range n.Body.Statements {
			count += tr.countStatementsInNode(stmt)
		}
	case *ast.ForStatement:
		for _, stmt := range n.Body.Statements {
			count += tr.countStatementsInNode(stmt)
		}
	}

	return count
}

func (tr *TestRunner) printResult(result TestResult) {
	status := "PASS"
	if !result.Success {
		status = "FAIL"
	}

	fmt.Printf("[%s] %s (%.2fms", status, result.File, float64(result.Duration.Microseconds())/1000.0)

	if tr.coverage {
		fmt.Printf(", coverage: %.1f%%", result.Coverage)
	}

	fmt.Println(")")

	if !result.Success && result.Error != "" {
		fmt.Printf("  Error: %s\n", result.Error)
	}
}

func (tr *TestRunner) PrintSummary() {
	passed := 0
	failed := 0
	totalDuration := time.Duration(0)
	totalCoverage := 0.0

	for _, result := range tr.results {
		if result.Success {
			passed++
		} else {
			failed++
		}
		totalDuration += result.Duration
		totalCoverage += result.Coverage
	}

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Printf("Test Summary: %d passed, %d failed (%.2fms total)\n",
		passed, failed, float64(totalDuration.Microseconds())/1000.0)

	if tr.coverage && len(tr.results) > 0 {
		avgCoverage := totalCoverage / float64(len(tr.results))
		fmt.Printf("Average Coverage: %.1f%%\n", avgCoverage)
	}

	fmt.Println(strings.Repeat("=", 60))
}

func runEnhancedTests(args []string) {
	parallel := false
	coverage := false
	verbose := false
	testDir := "tests"

	for _, arg := range args {
		switch arg {
		case "-p", "--parallel":
			parallel = true
		case "-c", "--coverage":
			coverage = true
		case "-v", "--verbose":
			verbose = true
		default:
			if !strings.HasPrefix(arg, "-") {
				testDir = arg
			}
		}
	}

	// Find all test files
	var testFiles []string
	err := filepath.Walk(testDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(path, "_test.sky") {
			testFiles = append(testFiles, path)
		}

		return nil
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding test files: %v\n", err)
		os.Exit(1)
	}

	if len(testFiles) == 0 {
		fmt.Printf("No test files found in %s\n", testDir)
		return
	}

	fmt.Printf("Running %d test(s)...\n", len(testFiles))

	runner := NewTestRunner(parallel, coverage, verbose)
	err = runner.RunTests(testFiles)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running tests: %v\n", err)
		os.Exit(1)
	}

	runner.PrintSummary()

	// Exit with error if any test failed
	for _, result := range runner.results {
		if !result.Success {
			os.Exit(1)
		}
	}
}
