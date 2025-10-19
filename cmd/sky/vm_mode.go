package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/mburakmmm/sky-lang/internal/lexer"
	"github.com/mburakmmm/sky-lang/internal/parser"
	"github.com/mburakmmm/sky-lang/internal/sema"
	"github.com/mburakmmm/sky-lang/internal/vm"
)

// runWithVM runs SKY program using bytecode VM (for recursion support)
func runWithVM(filename string) error {
	// Read file
	content, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("error reading file: %v", err)
	}

	// Lex & Parse (use same API as main.go)
	l := lexer.New(string(content), filename)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		return fmt.Errorf("parse errors:\n%v", strings.Join(p.Errors(), "\n"))
	}

	// Semantic check
	checker := sema.NewChecker()
	errors := checker.Check(program)
	if len(errors) > 0 {
		errMsgs := make([]string, len(errors))
		for i, e := range errors {
			errMsgs[i] = e.Error()
		}
		return fmt.Errorf("semantic errors:\n%v", strings.Join(errMsgs, "\n"))
	}

	// Compile to bytecode
	compiler := vm.NewCompiler()
	bytecode, err := compiler.Compile(program)
	if err != nil {
		return fmt.Errorf("compile error: %v", err)
	}

	// Run on VM
	machine := vm.NewVM(bytecode)
	if err := machine.Run(); err != nil {
		return fmt.Errorf("runtime error: %v", err)
	}

	return nil
}

// dumpBytecode shows compiled bytecode
func dumpBytecode(filename string) error {
	// Read file
	content, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	// Lex & Parse (use same API as main.go)
	l := lexer.New(string(content), filename)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		return fmt.Errorf("parse errors")
	}

	// Compile
	compiler := vm.NewCompiler()
	bytecode, err := compiler.Compile(program)
	if err != nil {
		return err
	}

	// Disassemble
	bytecode.Disassemble(filename)

	return nil
}
