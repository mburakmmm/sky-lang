//go:build llvm

package main

import (
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

	// Link object file to executable
	// TODO: Call system linker (clang, gcc, or ld)
	// For now, just return the object file

	return nil
}
