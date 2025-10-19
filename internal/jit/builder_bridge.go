//go:build llvm

package jit

/*
#cgo LDFLAGS: -L/opt/homebrew/opt/llvm/lib -lLLVM
#cgo CFLAGS: -I/opt/homebrew/opt/llvm/include

#include <llvm-c/Core.h>
#include <llvm-c/ExecutionEngine.h>
#include <llvm-c/Target.h>
#include <llvm-c/Analysis.h>
#include <stdio.h>
#include <stdlib.h>

// Export Go printf wrapper for LLVM
int sky_printf(const char* fmt, const char* str) {
    return printf(fmt, str);
}
*/
import "C"
import (
	"fmt"

	"github.com/mburakmmm/sky-lang/internal/ast"
	"github.com/mburakmmm/sky-lang/internal/sema"
)

// BuildAndRun builds IR from AST and executes with JIT
func BuildAndRun(program *ast.Program, moduleName string) error {
	// Initialize
	C.LLVMInitializeNativeTarget()
	C.LLVMInitializeNativeAsmPrinter()
	C.LLVMLinkInMCJIT()

	context := C.LLVMContextCreate()
	defer C.LLVMContextDispose(context)

	modName := C.CString(moduleName)
	module := C.LLVMModuleCreateWithNameInContext(modName, context)

	builder := C.LLVMCreateBuilderInContext(context)
	defer C.LLVMDisposeBuilder(builder)

	// Build IR from program
	irBuilder := &IRBuilder{
		context: context,
		module:  module,
		builder: builder,
		values:  make(map[string]C.LLVMValueRef),
		types:   make(map[string]C.LLVMTypeRef),
	}

	if err := irBuilder.buildProgram(program); err != nil {
		return fmt.Errorf("IR generation error: %v", err)
	}

	// Verify module
	var verifyErr *C.char
	if C.LLVMVerifyModule(module, C.LLVMReturnStatusAction, &verifyErr) != 0 {
		defer C.LLVMDisposeMessage(verifyErr)
		return fmt.Errorf("module verification failed: %s", C.GoString(verifyErr))
	}

	// Create execution engine (use Interpreter for external symbol resolution)
	var engine C.LLVMExecutionEngineRef
	var errMsg *C.char

	if C.LLVMCreateInterpreterForModule(&engine, module, &errMsg) != 0 {
		defer C.LLVMDisposeMessage(errMsg)
		return fmt.Errorf("failed to create execution engine: %s", C.GoString(errMsg))
	}
	defer C.LLVMDisposeExecutionEngine(engine)

	// Map sky_printf to C function
	skyPrintfFunc := C.LLVMGetNamedFunction(module, C.CString("sky_printf"))
	if skyPrintfFunc != nil {
		C.LLVMAddGlobalMapping(engine, skyPrintfFunc, C.sky_printf)
	}

	// Run main function
	mainFunc := C.LLVMGetNamedFunction(module, C.CString("main"))
	if mainFunc == nil {
		return fmt.Errorf("main function not found")
	}

	exitCode := C.LLVMRunFunctionAsMain(engine, mainFunc, 0, nil, nil)

	fmt.Printf("JIT execution completed (return value: %d)\n", exitCode)

	return nil
}

// IRBuilder builds LLVM IR from SKY AST
type IRBuilder struct {
	context     C.LLVMContextRef
	module      C.LLVMModuleRef
	builder     C.LLVMBuilderRef
	values      map[string]C.LLVMValueRef
	types       map[string]C.LLVMTypeRef
	currentFunc C.LLVMValueRef
}

func (b *IRBuilder) buildProgram(program *ast.Program) error {
	// Declare printf
	b.declarePrintf()

	// Generate all statements
	for _, stmt := range program.Statements {
		if err := b.buildStatement(stmt); err != nil {
			return err
		}
	}

	return nil
}

func (b *IRBuilder) declarePrintf() {
	i8PtrType := C.LLVMPointerType(C.LLVMInt8TypeInContext(b.context), 0)
	intType := C.LLVMInt32TypeInContext(b.context)

	// sky_printf(const char* fmt, const char* str)
	paramTypes := []C.LLVMTypeRef{i8PtrType, i8PtrType}
	printfType := C.LLVMFunctionType(intType, &paramTypes[0], 2, 0) // Not variadic
	printf := C.LLVMAddFunction(b.module, C.CString("sky_printf"), printfType)
	b.values["printf"] = printf
	b.types["printf"] = printfType
}

func (b *IRBuilder) buildStatement(stmt ast.Statement) error {
	switch s := stmt.(type) {
	case *ast.FunctionStatement:
		return b.buildFunction(s)
	default:
		// Skip unsupported for MVP
		return nil
	}
}

func (b *IRBuilder) buildFunction(stmt *ast.FunctionStatement) error {
	// Return type
	returnType := C.LLVMVoidTypeInContext(b.context)
	if stmt.ReturnType != nil {
		returnType = b.getType(sema.ResolveType(stmt.ReturnType))
	}

	// Create function
	funcType := C.LLVMFunctionType(returnType, nil, 0, 0)
	function := C.LLVMAddFunction(b.module, C.CString(stmt.Name.Value), funcType)
	b.values[stmt.Name.Value] = function
	b.currentFunc = function

	// Create entry block
	entry := C.LLVMAppendBasicBlockInContext(b.context, function, C.CString("entry"))
	C.LLVMPositionBuilderAtEnd(b.builder, entry)

	// Build body
	for _, bodyStmt := range stmt.Body.Statements {
		if err := b.buildBodyStatement(bodyStmt); err != nil {
			return err
		}
	}

	// Add return if missing
	if C.LLVMGetBasicBlockTerminator(C.LLVMGetInsertBlock(b.builder)) == nil {
		if returnType == C.LLVMVoidTypeInContext(b.context) {
			C.LLVMBuildRetVoid(b.builder)
		} else {
			C.LLVMBuildRet(b.builder, C.LLVMConstInt(returnType, 0, 0))
		}
	}

	return nil
}

func (b *IRBuilder) buildBodyStatement(stmt ast.Statement) error {
	switch s := stmt.(type) {
	case *ast.ExpressionStatement:
		_, err := b.buildExpression(s.Expression)
		return err
	case *ast.ReturnStatement:
		if s.ReturnValue != nil {
			val, err := b.buildExpression(s.ReturnValue)
			if err != nil {
				return err
			}
			C.LLVMBuildRet(b.builder, val)
		} else {
			C.LLVMBuildRetVoid(b.builder)
		}
		return nil
	default:
		return nil
	}
}

func (b *IRBuilder) buildExpression(expr ast.Expression) (C.LLVMValueRef, error) {
	switch e := expr.(type) {
	case *ast.Identifier:
		if val, ok := b.values[e.Value]; ok {
			return val, nil
		}
		return nil, fmt.Errorf("undefined variable: %s", e.Value)
	case *ast.CallExpression:
		return b.buildCall(e)
	case *ast.IntegerLiteral:
		return C.LLVMConstInt(C.LLVMInt64TypeInContext(b.context), C.ulonglong(e.Value), 0), nil
	case *ast.StringLiteral:
		str := C.CString(e.Value)
		return C.LLVMBuildGlobalStringPtr(b.builder, str, C.CString("str")), nil
	case *ast.InfixExpression:
		left, err := b.buildExpression(e.Left)
		if err != nil {
			return nil, err
		}
		right, err := b.buildExpression(e.Right)
		if err != nil {
			return nil, err
		}
		switch e.Operator {
		case "+":
			return C.LLVMBuildAdd(b.builder, left, right, C.CString("add")), nil
		case "-":
			return C.LLVMBuildSub(b.builder, left, right, C.CString("sub")), nil
		case "*":
			return C.LLVMBuildMul(b.builder, left, right, C.CString("mul")), nil
		case "/":
			return C.LLVMBuildSDiv(b.builder, left, right, C.CString("div")), nil
		default:
			return nil, fmt.Errorf("unsupported operator: %s", e.Operator)
		}
	default:
		return nil, fmt.Errorf("unsupported expression: %T", expr)
	}
}

func (b *IRBuilder) buildCall(expr *ast.CallExpression) (C.LLVMValueRef, error) {
	if ident, ok := expr.Function.(*ast.Identifier); ok {
		if ident.Value == "print" {
			// print() builtin
			if len(expr.Arguments) > 0 {
				arg, err := b.buildExpression(expr.Arguments[0])
				if err != nil {
					return nil, err
				}

				// Build format string "%s\n" for string arguments
				// For integers, use "%lld\n"
				formatStr := C.CString("%s\n")
				format := C.LLVMBuildGlobalStringPtr(b.builder, formatStr, C.CString("fmt"))

				// Call printf(format, arg)
				printf := b.values["printf"]
				printfType := b.types["printf"]
				args := []C.LLVMValueRef{format, arg}
				var zero C.LLVMValueRef
				C.LLVMBuildCall2(b.builder, printfType, printf, &args[0], 2, C.CString(""))
				return zero, nil
			}
		}
	}

	var zero C.LLVMValueRef
	return zero, fmt.Errorf("unsupported call")
}

func (b *IRBuilder) getType(t sema.Type) C.LLVMTypeRef {
	switch t {
	case sema.IntType:
		return C.LLVMInt64TypeInContext(b.context)
	case sema.FloatType:
		return C.LLVMDoubleTypeInContext(b.context)
	case sema.BoolType:
		return C.LLVMInt1TypeInContext(b.context)
	case sema.StringType:
		return C.LLVMPointerType(C.LLVMInt8TypeInContext(b.context), 0)
	default:
		return C.LLVMVoidTypeInContext(b.context)
	}
}
