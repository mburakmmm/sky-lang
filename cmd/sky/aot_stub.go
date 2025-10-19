//go:build !llvm

package main

import (
	"fmt"

	"github.com/mburakmmm/sky-lang/internal/ast"
)

func compileAOT(program *ast.Program, outputPath string) error {
	return fmt.Errorf("AOT compilation not available: build with -tags llvm")
}

