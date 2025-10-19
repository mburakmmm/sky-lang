//go:build llvm

package main

import (
	"fmt"
	"os"

	"github.com/mburakmmm/sky-lang/internal/ir"
	"github.com/mburakmmm/sky-lang/internal/jit"
	"github.com/mburakmmm/sky-lang/internal/lexer"
	"github.com/mburakmmm/sky-lang/internal/parser"
	"github.com/mburakmmm/sky-lang/internal/sema"
)

func runWithJIT(filename string) error {
	// Read file
	content, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("error reading file: %v", err)
	}

	// Lex
	l := lexer.New(string(content), filename)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		return fmt.Errorf("parse errors: %v", p.Errors())
	}

	// Semantic check
	checker := sema.NewChecker()
	errors := checker.Check(program)
	if len(errors) > 0 {
		return fmt.Errorf("semantic errors: %v", errors)
	}

	// Build IR
	builder := ir.NewBuilder()
	module, err := builder.BuildModule(program)
	if err != nil {
		return fmt.Errorf("IR build error: %v", err)
	}

	// Create JIT engine
	engine, err := jit.NewEngine(module)
	if err != nil {
		return fmt.Errorf("JIT engine error: %v", err)
	}

	// Run main function
	if err := engine.RunMain(); err != nil {
		return fmt.Errorf("execution error: %v", err)
	}

	return nil
}
