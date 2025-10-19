package jit

/*
#cgo LDFLAGS: -L/opt/homebrew/opt/llvm/lib -lLLVM
#cgo CFLAGS: -I/opt/homebrew/opt/llvm/include

#include <llvm-c/Core.h>
#include <llvm-c/ExecutionEngine.h>
#include <llvm-c/Target.h>
#include <llvm-c/Analysis.h>
#include <stdlib.h>
*/
import "C"
import (
	"fmt"
	"unsafe"
)

// Engine LLVM JIT execution engine
type Engine struct {
	engine  C.LLVMExecutionEngineRef
	module  C.LLVMModuleRef
	builder C.LLVMPassManagerRef
}

// NewEngine creates a new JIT engine
func NewEngine(module C.LLVMModuleRef) (*Engine, error) {
	// Initialize execution engine
	C.LLVMLinkInMCJIT()
	C.LLVMInitializeNativeTarget()
	C.LLVMInitializeNativeAsmPrinter()
	C.LLVMInitializeNativeAsmParser()

	var engine C.LLVMExecutionEngineRef
	var errMsg *C.char

	// Create execution engine
	if C.LLVMCreateExecutionEngineForModule(&engine, module, &errMsg) != 0 {
		defer C.LLVMDisposeMessage(errMsg)
		return nil, fmt.Errorf("failed to create execution engine: %s",
			C.GoString(errMsg))
	}

	// Create pass manager for optimizations
	pm := C.LLVMCreatePassManager()

	// Note: Transform passes require additional LLVM modules
	// For now, basic optimization only

	return &Engine{
		engine:  engine,
		module:  module,
		builder: pm,
	}, nil
}

// Optimize runs optimization passes on the module
func (e *Engine) Optimize() {
	C.LLVMRunPassManager(e.builder, e.module)
}

// RunFunction executes a function with given arguments
func (e *Engine) RunFunction(funcName string, args ...interface{}) (interface{}, error) {
	// Get function
	function := C.LLVMGetNamedFunction(e.module, C.CString(funcName))
	if function == (nil) {
		return nil, fmt.Errorf("function not found: %s", funcName)
	}

	// Convert arguments to LLVM generic values
	llvmArgs := make([]C.LLVMGenericValueRef, len(args))
	for i, arg := range args {
		llvmArgs[i] = e.convertToGenericValue(arg)
	}

	var argsPtr *C.LLVMGenericValueRef
	if len(llvmArgs) > 0 {
		argsPtr = &llvmArgs[0]
	}

	// Run function
	result := C.LLVMRunFunction(
		e.engine,
		function,
		C.uint(len(llvmArgs)),
		argsPtr,
	)

	// Convert result back
	return e.convertFromGenericValue(result, C.LLVMGetReturnType(C.LLVMGetElementType(C.LLVMTypeOf(function)))), nil
}

// RunFunctionAsMain executes a function as main (no args, int return)
func (e *Engine) RunFunctionAsMain(funcName string) (int, error) {
	function := C.LLVMGetNamedFunction(e.module, C.CString(funcName))
	if function == (nil) {
		return -1, fmt.Errorf("function not found: %s", funcName)
	}

	// Run as main function
	result := C.LLVMRunFunctionAsMain(
		e.engine,
		function,
		0,
		nil,
		nil,
	)

	return int(result), nil
}

// GetFunctionAddress gets the address of a compiled function
func (e *Engine) GetFunctionAddress(funcName string) (uint64, error) {
	function := C.LLVMGetNamedFunction(e.module, C.CString(funcName))
	if function == (nil) {
		return 0, fmt.Errorf("function not found: %s", funcName)
	}

	addr := C.LLVMGetFunctionAddress(e.engine, C.CString(funcName))
	return uint64(addr), nil
}

// AddGlobalMapping adds a global mapping for FFI
func (e *Engine) AddGlobalMapping(name string, addr unsafe.Pointer) {
	global := C.LLVMGetNamedGlobal(e.module, C.CString(name))
	if global != (nil) {
		C.LLVMAddGlobalMapping(e.engine, global, addr)
	}
}

// convertToGenericValue converts Go value to LLVM generic value
func (e *Engine) convertToGenericValue(val interface{}) C.LLVMGenericValueRef {
	switch v := val.(type) {
	case int:
		return C.LLVMCreateGenericValueOfInt(
			C.LLVMInt64Type(),
			C.ulonglong(v),
			0,
		)
	case int64:
		return C.LLVMCreateGenericValueOfInt(
			C.LLVMInt64Type(),
			C.ulonglong(v),
			0,
		)
	case float64:
		return C.LLVMCreateGenericValueOfFloat(
			C.LLVMDoubleType(),
			C.double(v),
		)
	case bool:
		var i C.ulonglong
		if v {
			i = 1
		}
		return C.LLVMCreateGenericValueOfInt(
			C.LLVMInt1Type(),
			i,
			0,
		)
	case string:
		cstr := C.CString(v)
		return C.LLVMCreateGenericValueOfPointer(unsafe.Pointer(cstr))
	default:
		return C.LLVMCreateGenericValueOfInt(C.LLVMInt64Type(), 0, 0)
	}
}

// convertFromGenericValue converts LLVM generic value to Go value
func (e *Engine) convertFromGenericValue(val C.LLVMGenericValueRef, typ C.LLVMTypeRef) interface{} {
	kind := C.LLVMGetTypeKind(typ)

	switch kind {
	case C.LLVMIntegerTypeKind:
		width := C.LLVMGetIntTypeWidth(typ)
		if width == 1 {
			// Bool
			return C.LLVMGenericValueToInt(val, 0) != 0
		}
		// Int
		return int64(C.LLVMGenericValueToInt(val, 0))

	case C.LLVMDoubleTypeKind:
		return float64(C.LLVMGenericValueToFloat(C.LLVMDoubleType(), val))

	case C.LLVMPointerTypeKind:
		ptr := C.LLVMGenericValueToPointer(val)
		if ptr != nil {
			return C.GoString((*C.char)(ptr))
		}
		return nil

	case C.LLVMVoidTypeKind:
		return nil

	default:
		return nil
	}
}

// FreeMachineCodeForFunction frees generated machine code
func (e *Engine) FreeMachineCodeForFunction(funcName string) {
	function := C.LLVMGetNamedFunction(e.module, C.CString(funcName))
	if function != (nil) {
		C.LLVMFreeMachineCodeForFunction(e.engine, function)
	}
}

// Dispose cleans up the execution engine
func (e *Engine) Dispose() {
	C.LLVMDisposePassManager(e.builder)
	C.LLVMDisposeExecutionEngine(e.engine)
}

// GetPointerToGlobal gets pointer to a global variable
func (e *Engine) GetPointerToGlobal(name string) unsafe.Pointer {
	global := C.LLVMGetNamedGlobal(e.module, C.CString(name))
	if global == (nil) {
		return nil
	}
	return C.LLVMGetPointerToGlobal(e.engine, global)
}

// RecompileAndRelinkFunction recompiles a function
func (e *Engine) RecompileAndRelinkFunction(funcName string) error {
	function := C.LLVMGetNamedFunction(e.module, C.CString(funcName))
	if function == (nil) {
		return fmt.Errorf("function not found: %s", funcName)
	}

	// Free old machine code
	C.LLVMFreeMachineCodeForFunction(e.engine, function)

	// Recompile (happens automatically on next call)
	return nil
}

// GetDataLayout gets the target data layout
func (e *Engine) GetDataLayout() string {
	layout := C.LLVMGetExecutionEngineTargetData(e.engine)
	layoutStr := C.LLVMCopyStringRepOfTargetData(layout)
	defer C.LLVMDisposeMessage(layoutStr)
	return C.GoString(layoutStr)
}

// GetTargetMachine gets the target machine
func (e *Engine) GetTargetMachine() C.LLVMTargetMachineRef {
	return C.LLVMGetExecutionEngineTargetMachine(e.engine)
}

// Statistics holds JIT compilation statistics
type Statistics struct {
	FunctionsCompiled int
	TotalCodeSize     uint64
	OptimizationLevel int
}

// GetStatistics returns JIT statistics
func (e *Engine) GetStatistics() Statistics {
	// TODO: Implement actual statistics gathering
	return Statistics{
		FunctionsCompiled: 0,
		TotalCodeSize:     0,
		OptimizationLevel: 2,
	}
}
