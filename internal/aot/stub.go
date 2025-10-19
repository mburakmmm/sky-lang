//go:build !llvm

package aot

import (
	"fmt"

	"github.com/mburakmmm/sky-lang/internal/ast"
)

// Compiler stub for non-LLVM builds
type Compiler struct{}

func NewCompiler() *Compiler {
	return &Compiler{}
}

func (c *Compiler) CompileToObjectFile(program *ast.Program, outputPath string) error {
	return fmt.Errorf("AOT compilation not available: build with -tags llvm")
}

func (c *Compiler) SetOptimization(enable bool) {}

func (c *Compiler) SetTargetTriple(triple string) {}
