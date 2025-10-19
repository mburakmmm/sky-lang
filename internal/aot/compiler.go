//go:build llvm

package aot

/*
#cgo LDFLAGS: -L/opt/homebrew/opt/llvm/lib -lLLVM
#cgo CFLAGS: -I/opt/homebrew/opt/llvm/include

#include <llvm-c/Core.h>
#include <llvm-c/Target.h>
#include <llvm-c/TargetMachine.h>
#include <llvm-c/Analysis.h>
#include <stdlib.h>
*/
import "C"
import (
	"fmt"
	"unsafe"

	"github.com/mburakmmm/sky-lang/internal/ast"
	"github.com/mburakmmm/sky-lang/internal/ir"
)

// Compiler AOT compiler for generating native binaries
type Compiler struct {
	targetTriple string
	optimize     bool
}

// NewCompiler creates a new AOT compiler
func NewCompiler() *Compiler {
	return &Compiler{
		targetTriple: C.GoString(C.LLVMGetDefaultTargetTriple()),
		optimize:     true,
	}
}

// CompileToObjectFile compiles a program to an object file (.o)
func (c *Compiler) CompileToObjectFile(program *ast.Program, outputPath string) error {
	// Build IR
	builder := ir.NewBuilder()
	module, err := builder.BuildModule(program)
	if err != nil {
		return fmt.Errorf("IR build error: %v", err)
	}

	// Initialize target
	C.LLVMInitializeAllTargetInfos()
	C.LLVMInitializeAllTargets()
	C.LLVMInitializeAllTargetMCs()
	C.LLVMInitializeAllAsmParsers()
	C.LLVMInitializeAllAsmPrinters()

	// Get target triple
	triple := C.CString(c.targetTriple)
	defer C.free(unsafe.Pointer(triple))

	// Get target
	var target C.LLVMTargetRef
	var errMsg *C.char
	if C.LLVMGetTargetFromTriple(triple, &target, &errMsg) != 0 {
		defer C.LLVMDisposeMessage(errMsg)
		return fmt.Errorf("failed to get target: %s", C.GoString(errMsg))
	}

	// Create target machine
	cpu := C.CString("generic")
	features := C.CString("")
	defer C.free(unsafe.Pointer(cpu))
	defer C.free(unsafe.Pointer(features))

	var optLevel C.LLVMCodeGenOptLevel
	if c.optimize {
		optLevel = C.LLVMCodeGenLevelDefault
	} else {
		optLevel = C.LLVMCodeGenLevelNone
	}

	targetMachine := C.LLVMCreateTargetMachine(
		target,
		triple,
		cpu,
		features,
		optLevel,
		C.LLVMRelocDefault,
		C.LLVMCodeModelDefault,
	)
	defer C.LLVMDisposeTargetMachine(targetMachine)

	// Emit object file
	outputFile := C.CString(outputPath)
	defer C.free(unsafe.Pointer(outputFile))

	if C.LLVMTargetMachineEmitToFile(
		targetMachine,
		module,
		outputFile,
		C.LLVMObjectFile,
		&errMsg,
	) != 0 {
		defer C.LLVMDisposeMessage(errMsg)
		return fmt.Errorf("failed to emit object file: %s", C.GoString(errMsg))
	}

	return nil
}

// SetOptimization enables or disables optimizations
func (c *Compiler) SetOptimization(enable bool) {
	c.optimize = enable
}

// SetTargetTriple sets the target triple for cross-compilation
func (c *Compiler) SetTargetTriple(triple string) {
	c.targetTriple = triple
}
