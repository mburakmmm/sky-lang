//go:build llvm

package aot

/*
#cgo LDFLAGS: -L/opt/homebrew/opt/llvm/lib -lLLVM
#cgo CFLAGS: -I/opt/homebrew/opt/llvm/include

#include <llvm-c/Core.h>
#include <llvm-c/Target.h>
#include <llvm-c/TargetMachine.h>
#include <llvm-c/Analysis.h>
#include <llvm-c/BitWriter.h>
#include <stdlib.h>
*/
import "C"
import (
	"fmt"
	"unsafe"

	"github.com/mburakmmm/sky-lang/internal/ast"
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
	// Build IR directly in AOT package to avoid type issues
	context := C.LLVMContextCreate()
	defer C.LLVMContextDispose(context)

	moduleName := C.CString(outputPath)
	defer C.free(unsafe.Pointer(moduleName))
	module := C.LLVMModuleCreateWithNameInContext(moduleName, context)
	defer C.LLVMDisposeModule(module)

	builder := C.LLVMCreateBuilderInContext(context)
	defer C.LLVMDisposeBuilder(builder)

	// For now, create a minimal main function that returns 0
	// Full IR generation would use internal/ir package
	intType := C.LLVMInt32TypeInContext(context)
	mainType := C.LLVMFunctionType(intType, nil, 0, 0)
	mainFunc := C.LLVMAddFunction(module, C.CString("main"), mainType)

	entry := C.LLVMAppendBasicBlockInContext(context, mainFunc, C.CString("entry"))
	C.LLVMPositionBuilderAtEnd(builder, entry)
	C.LLVMBuildRet(builder, C.LLVMConstInt(intType, 0, 0))

	// Initialize all LLVM targets
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

	// Step 1: Write bitcode for debugging
	bitcodeFile := outputPath + ".bc"
	bcPath := C.CString(bitcodeFile)
	defer C.free(unsafe.Pointer(bcPath))

	if C.LLVMWriteBitcodeToFile(module, bcPath) != 0 {
		return fmt.Errorf("failed to write bitcode file")
	}

	// Step 2: Emit assembly for debugging
	asmFile := outputPath + ".s"
	asmPath := C.CString(asmFile)
	defer C.free(unsafe.Pointer(asmPath))

	// Try to emit assembly file (native code generation)
	var asmErr *C.char
	result := C.LLVMTargetMachineEmitToFile(
		targetMachine,
		module,
		asmPath,
		C.LLVMAssemblyFile,
		&asmErr,
	)

	if result != 0 {
		// Assembly emission failed, but continue
		if asmErr != nil {
			C.LLVMDisposeMessage(asmErr)
		}
	}

	// Step 3: Emit object file (the main goal)
	objFile := outputPath + ".o"
	objPath := C.CString(objFile)
	defer C.free(unsafe.Pointer(objPath))

	var objErr *C.char
	if C.LLVMTargetMachineEmitToFile(
		targetMachine,
		module,
		objPath,
		C.LLVMObjectFile,
		&objErr,
	) != 0 {
		defer C.LLVMDisposeMessage(objErr)
		return fmt.Errorf("failed to emit object file: %s", C.GoString(objErr))
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
