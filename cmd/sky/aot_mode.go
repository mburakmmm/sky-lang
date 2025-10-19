//go:build llvm

package main

import (
	"fmt"
	"os/exec"

	"github.com/mburakmmm/sky-lang/internal/aot"
	"github.com/mburakmmm/sky-lang/internal/ast"
)

func compileAOT(program *ast.Program, outputPath string) error {
	compiler := aot.NewCompiler()
	compiler.SetOptimization(true)

	objFile := outputPath + ".o"
	if err := compiler.CompileToObjectFile(program, objFile); err != nil {
		return err
	}

	// Link object file to executable using clang
	// clang test_app.o.o -o test_app
	cmd := exec.Command("clang", objFile+".o", "-o", outputPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("linking failed: %v\nOutput: %s", err, string(output))
	}

	fmt.Printf("  → Generated: %s.o.bc (bitcode)\n", objFile)
	fmt.Printf("  → Generated: %s.o.s (assembly)\n", objFile)
	fmt.Printf("  → Generated: %s.o.o (object)\n", objFile)
	fmt.Printf("  → Linked: %s (executable)\n", outputPath)

	return nil
}
