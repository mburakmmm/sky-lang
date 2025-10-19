package ir

/*
#cgo LDFLAGS: -L/opt/homebrew/opt/llvm/lib -lLLVM
#cgo CFLAGS: -I/opt/homebrew/opt/llvm/include

#include <llvm-c/Core.h>
#include <llvm-c/ExecutionEngine.h>
#include <llvm-c/Target.h>
#include <llvm-c/Analysis.h>
#include <llvm-c/BitWriter.h>
#include <stdlib.h>
*/
import "C"
import (
	"fmt"

	"github.com/mburakmmm/sky-lang/internal/ast"
	"github.com/mburakmmm/sky-lang/internal/sema"
)

// Builder LLVM IR builder
type Builder struct {
	context C.LLVMContextRef
	module  C.LLVMModuleRef
	builder C.LLVMBuilderRef

	// Symbol table for LLVM values
	values map[string]C.LLVMValueRef

	// Current function being built
	currentFunc C.LLVMValueRef

	// Type cache
	intType    C.LLVMTypeRef
	floatType  C.LLVMTypeRef
	boolType   C.LLVMTypeRef
	stringType C.LLVMTypeRef
	voidType   C.LLVMTypeRef
}

// NewBuilder creates a new LLVM IR builder
func NewBuilder(moduleName string) *Builder {
	C.LLVMInitializeNativeTarget()
	C.LLVMInitializeNativeAsmPrinter()
	C.LLVMLinkInMCJIT()

	context := C.LLVMContextCreate()
	module := C.LLVMModuleCreateWithNameInContext(
		C.CString(moduleName),
		context,
	)
	builder := C.LLVMCreateBuilderInContext(context)

	b := &Builder{
		context: context,
		module:  module,
		builder: builder,
		values:  make(map[string]C.LLVMValueRef),
	}

	// Initialize types
	b.intType = C.LLVMInt64TypeInContext(context)
	b.floatType = C.LLVMDoubleTypeInContext(context)
	b.boolType = C.LLVMInt1TypeInContext(context)
	b.voidType = C.LLVMVoidTypeInContext(context)

	// String type as i8*
	b.stringType = C.LLVMPointerType(
		C.LLVMInt8TypeInContext(context),
		0,
	)

	return b
}

// GenerateIR generates LLVM IR from AST
func (b *Builder) GenerateIR(program *ast.Program) error {
	// Declare built-in functions
	b.declarePrintf()

	// Generate IR for all statements
	for _, stmt := range program.Statements {
		if err := b.generateStatement(stmt); err != nil {
			return err
		}
	}

	return nil
}

// declarePrintf declares the printf function
func (b *Builder) declarePrintf() {
	// int printf(const char* format, ...)
	printfType := C.LLVMFunctionType(
		C.LLVMInt32TypeInContext(b.context),
		&b.stringType,
		1,
		1, // variadic
	)

	printf := C.LLVMAddFunction(
		b.module,
		C.CString("printf"),
		printfType,
	)

	b.values["printf"] = printf
}

// generateStatement generates IR for a statement
func (b *Builder) generateStatement(stmt ast.Statement) error {
	switch s := stmt.(type) {
	case *ast.FunctionStatement:
		return b.generateFunction(s)
	case *ast.LetStatement:
		return b.generateLetStatement(s)
	case *ast.ReturnStatement:
		return b.generateReturnStatement(s)
	case *ast.ExpressionStatement:
		_, err := b.generateExpression(s.Expression)
		return err
	case *ast.IfStatement:
		return b.generateIfStatement(s)
	case *ast.WhileStatement:
		return b.generateWhileStatement(s)
	case *ast.ForStatement:
		return b.generateForStatement(s)
	default:
		return fmt.Errorf("unsupported statement type: %T", stmt)
	}
}

// generateFunction generates IR for a function
func (b *Builder) generateFunction(stmt *ast.FunctionStatement) error {
	// Build parameter types
	paramCount := len(stmt.Parameters)
	var paramTypes *C.LLVMTypeRef
	if paramCount > 0 {
		paramTypesSlice := make([]C.LLVMTypeRef, paramCount)
		for i := range stmt.Parameters {
			paramTypesSlice[i] = b.intType // Default to int
		}
		paramTypes = &paramTypesSlice[0]
	}

	// Return type
	returnType := b.voidType
	if stmt.ReturnType != nil {
		returnType = b.getLLVMType(sema.ResolveType(stmt.ReturnType))
	}

	// Create function type
	funcType := C.LLVMFunctionType(
		returnType,
		paramTypes,
		C.uint(paramCount),
		0,
	)

	// Create function
	function := C.LLVMAddFunction(
		b.module,
		C.CString(stmt.Name.Value),
		funcType,
	)

	b.values[stmt.Name.Value] = function
	b.currentFunc = function

	// Create entry block
	entry := C.LLVMAppendBasicBlockInContext(
		b.context,
		function,
		C.CString("entry"),
	)
	C.LLVMPositionBuilderAtEnd(b.builder, entry)

	// Generate function body
	if stmt.Body != nil {
		for _, bodyStmt := range stmt.Body.Statements {
			if err := b.generateStatement(bodyStmt); err != nil {
				return err
			}
		}
	}

	// Add return void if no explicit return
	if returnType == b.voidType {
		C.LLVMBuildRetVoid(b.builder)
	}

	// Verify function
	var errMsg *C.char
	if C.LLVMVerifyFunction(function, C.LLVMPrintMessageAction) != 0 {
		return fmt.Errorf("function verification failed: %s",
			C.GoString(errMsg))
	}

	return nil
}

// generateLetStatement generates IR for let statement
func (b *Builder) generateLetStatement(stmt *ast.LetStatement) error {
	value, err := b.generateExpression(stmt.Value)
	if err != nil {
		return err
	}

	// Allocate stack space for variable
	alloca := C.LLVMBuildAlloca(
		b.builder,
		C.LLVMTypeOf(value),
		C.CString(stmt.Name.Value),
	)

	// Store value
	C.LLVMBuildStore(b.builder, value, alloca)

	// Remember the alloca
	b.values[stmt.Name.Value] = alloca

	return nil
}

// generateReturnStatement generates IR for return statement
func (b *Builder) generateReturnStatement(stmt *ast.ReturnStatement) error {
	if stmt.ReturnValue != nil {
		value, err := b.generateExpression(stmt.ReturnValue)
		if err != nil {
			return err
		}
		C.LLVMBuildRet(b.builder, value)
	} else {
		C.LLVMBuildRetVoid(b.builder)
	}
	return nil
}

// generateExpression generates IR for an expression
func (b *Builder) generateExpression(expr ast.Expression) (C.LLVMValueRef, error) {
	switch e := expr.(type) {
	case *ast.IntegerLiteral:
		return C.LLVMConstInt(
			b.intType,
			C.ulonglong(e.Value),
			0,
		), nil

	case *ast.FloatLiteral:
		return C.LLVMConstReal(
			b.floatType,
			C.double(e.Value),
		), nil

	case *ast.BooleanLiteral:
		var val C.ulonglong
		if e.Value {
			val = 1
		}
		return C.LLVMConstInt(b.boolType, val, 0), nil

	case *ast.StringLiteral:
		return b.createGlobalString(e.Value), nil

	case *ast.Identifier:
		alloca, ok := b.values[e.Value]
		if !ok {
			var zero C.LLVMValueRef
			return zero, fmt.Errorf("undefined variable: %s", e.Value)
		}
		return C.LLVMBuildLoad2(
			b.builder,
			b.intType, // TODO: get actual type
			alloca,
			C.CString(""),
		), nil

	case *ast.InfixExpression:
		return b.generateInfixExpression(e)

	case *ast.CallExpression:
		return b.generateCallExpression(e)

	default:
		var zero C.LLVMValueRef
		return zero, fmt.Errorf("unsupported expression type: %T", expr)
	}
}

// generateInfixExpression generates IR for infix expression
func (b *Builder) generateInfixExpression(expr *ast.InfixExpression) (C.LLVMValueRef, error) {
	left, err := b.generateExpression(expr.Left)
	if err != nil {
		var zero C.LLVMValueRef
		return zero, err
	}

	right, err := b.generateExpression(expr.Right)
	if err != nil {
		var zero C.LLVMValueRef
		return zero, err
	}

	switch expr.Operator {
	case "+":
		return C.LLVMBuildAdd(b.builder, left, right, C.CString("addtmp")), nil
	case "-":
		return C.LLVMBuildSub(b.builder, left, right, C.CString("subtmp")), nil
	case "*":
		return C.LLVMBuildMul(b.builder, left, right, C.CString("multmp")), nil
	case "/":
		return C.LLVMBuildSDiv(b.builder, left, right, C.CString("divtmp")), nil
	case "%":
		return C.LLVMBuildSRem(b.builder, left, right, C.CString("modtmp")), nil
	case "==":
		return C.LLVMBuildICmp(b.builder, C.LLVMIntEQ, left, right, C.CString("eqtmp")), nil
	case "!=":
		return C.LLVMBuildICmp(b.builder, C.LLVMIntNE, left, right, C.CString("netmp")), nil
	case "<":
		return C.LLVMBuildICmp(b.builder, C.LLVMIntSLT, left, right, C.CString("lttmp")), nil
	case "<=":
		return C.LLVMBuildICmp(b.builder, C.LLVMIntSLE, left, right, C.CString("letmp")), nil
	case ">":
		return C.LLVMBuildICmp(b.builder, C.LLVMIntSGT, left, right, C.CString("gttmp")), nil
	case ">=":
		return C.LLVMBuildICmp(b.builder, C.LLVMIntSGE, left, right, C.CString("getmp")), nil
	default:
		var zero C.LLVMValueRef
		return zero, fmt.Errorf("unsupported operator: %s", expr.Operator)
	}
}

// generateCallExpression generates IR for function call
func (b *Builder) generateCallExpression(expr *ast.CallExpression) (C.LLVMValueRef, error) {
	// Get function
	funcIdent, ok := expr.Function.(*ast.Identifier)
	if !ok {
		var zero C.LLVMValueRef
		return zero, fmt.Errorf("only direct function calls supported")
	}

	function, ok := b.values[funcIdent.Value]
	if !ok {
		var zero C.LLVMValueRef
		return zero, fmt.Errorf("undefined function: %s", funcIdent.Value)
	}

	// Special handling for print
	if funcIdent.Value == "print" {
		return b.generatePrintCall(expr)
	}

	// Generate arguments
	args := make([]C.LLVMValueRef, len(expr.Arguments))
	for i, arg := range expr.Arguments {
		val, err := b.generateExpression(arg)
		if err != nil {
			var zero C.LLVMValueRef
			return zero, err
		}
		args[i] = val
	}

	var argsPtr *C.LLVMValueRef
	if len(args) > 0 {
		argsPtr = &args[0]
	}

	return C.LLVMBuildCall2(
		b.builder,
		C.LLVMGetElementType(C.LLVMTypeOf(function)),
		function,
		argsPtr,
		C.uint(len(args)),
		C.CString("calltmp"),
	), nil
}

// generatePrintCall generates IR for print() call using printf
func (b *Builder) generatePrintCall(expr *ast.CallExpression) (C.LLVMValueRef, error) {
	printf := b.values["printf"]

	for _, arg := range expr.Arguments {
		val, err := b.generateExpression(arg)
		if err != nil {
			var zero C.LLVMValueRef
			return zero, err
		}

		// Determine format string based on type
		var format string
		valType := C.LLVMTypeOf(val)

		if C.LLVMGetTypeKind(valType) == C.LLVMIntegerTypeKind {
			if C.LLVMGetIntTypeWidth(valType) == 1 {
				format = "%d" // bool as int
			} else {
				format = "%lld"
			}
		} else if C.LLVMGetTypeKind(valType) == C.LLVMDoubleTypeKind {
			format = "%f"
		} else if C.LLVMGetTypeKind(valType) == C.LLVMPointerTypeKind {
			format = "%s"
		} else {
			format = "%p"
		}

		formatStr := b.createGlobalString(format)

		args := []C.LLVMValueRef{formatStr, val}
		C.LLVMBuildCall2(
			b.builder,
			C.LLVMGetElementType(C.LLVMTypeOf(printf)),
			printf,
			&args[0],
			2,
			C.CString(""),
		)
	}

	// Print newline
	newline := b.createGlobalString("\n")
	args := []C.LLVMValueRef{newline}
	return C.LLVMBuildCall2(
		b.builder,
		C.LLVMGetElementType(C.LLVMTypeOf(printf)),
		printf,
		&args[0],
		1,
		C.CString(""),
	), nil
}

// generateIfStatement generates IR for if statement
func (b *Builder) generateIfStatement(stmt *ast.IfStatement) error {
	// Evaluate condition
	cond, err := b.generateExpression(stmt.Condition)
	if err != nil {
		return err
	}

	// Create blocks
	thenBlock := C.LLVMAppendBasicBlockInContext(
		b.context,
		b.currentFunc,
		C.CString("then"),
	)
	elseBlock := C.LLVMAppendBasicBlockInContext(
		b.context,
		b.currentFunc,
		C.CString("else"),
	)
	mergeBlock := C.LLVMAppendBasicBlockInContext(
		b.context,
		b.currentFunc,
		C.CString("ifcont"),
	)

	// Branch
	C.LLVMBuildCondBr(b.builder, cond, thenBlock, elseBlock)

	// Then block
	C.LLVMPositionBuilderAtEnd(b.builder, thenBlock)
	for _, s := range stmt.Consequence.Statements {
		if err := b.generateStatement(s); err != nil {
			return err
		}
	}
	C.LLVMBuildBr(b.builder, mergeBlock)

	// Else block
	C.LLVMPositionBuilderAtEnd(b.builder, elseBlock)
	if stmt.Alternative != nil {
		for _, s := range stmt.Alternative.Statements {
			if err := b.generateStatement(s); err != nil {
				return err
			}
		}
	}
	C.LLVMBuildBr(b.builder, mergeBlock)

	// Continue in merge block
	C.LLVMPositionBuilderAtEnd(b.builder, mergeBlock)

	return nil
}

// generateWhileStatement generates IR for while loop
func (b *Builder) generateWhileStatement(stmt *ast.WhileStatement) error {
	condBlock := C.LLVMAppendBasicBlockInContext(
		b.context,
		b.currentFunc,
		C.CString("loopcond"),
	)
	loopBlock := C.LLVMAppendBasicBlockInContext(
		b.context,
		b.currentFunc,
		C.CString("loop"),
	)
	afterBlock := C.LLVMAppendBasicBlockInContext(
		b.context,
		b.currentFunc,
		C.CString("afterloop"),
	)

	// Branch to condition
	C.LLVMBuildBr(b.builder, condBlock)

	// Condition block
	C.LLVMPositionBuilderAtEnd(b.builder, condBlock)
	cond, err := b.generateExpression(stmt.Condition)
	if err != nil {
		return err
	}
	C.LLVMBuildCondBr(b.builder, cond, loopBlock, afterBlock)

	// Loop body
	C.LLVMPositionBuilderAtEnd(b.builder, loopBlock)
	for _, s := range stmt.Body.Statements {
		if err := b.generateStatement(s); err != nil {
			return err
		}
	}
	C.LLVMBuildBr(b.builder, condBlock)

	// After loop
	C.LLVMPositionBuilderAtEnd(b.builder, afterBlock)

	return nil
}

// generateForStatement generates IR for for loop
func (b *Builder) generateForStatement(stmt *ast.ForStatement) error {
	// TODO: Implement proper iterator handling
	return fmt.Errorf("for loops not yet implemented in LLVM backend")
}

// createGlobalString creates a global string constant
func (b *Builder) createGlobalString(str string) C.LLVMValueRef {
	return C.LLVMBuildGlobalStringPtr(
		b.builder,
		C.CString(str),
		C.CString("str"),
	)
}

// getLLVMType converts sema.Type to LLVM type
func (b *Builder) getLLVMType(t sema.Type) C.LLVMTypeRef {
	switch t {
	case sema.IntType:
		return b.intType
	case sema.FloatType:
		return b.floatType
	case sema.BoolType:
		return b.boolType
	case sema.StringType:
		return b.stringType
	case sema.VoidType:
		return b.voidType
	default:
		return b.intType // Default
	}
}

// DumpIR dumps LLVM IR to string
func (b *Builder) DumpIR() string {
	cstr := C.LLVMPrintModuleToString(b.module)
	defer C.LLVMDisposeMessage(cstr)
	return C.GoString(cstr)
}

// Dispose cleans up LLVM resources
func (b *Builder) Dispose() {
	C.LLVMDisposeBuilder(b.builder)
	C.LLVMDisposeModule(b.module)
	C.LLVMContextDispose(b.context)
}

// GetModule returns the LLVM module
func (b *Builder) GetModule() C.LLVMModuleRef {
	return b.module
}

// WriteBitcodeToFile writes LLVM bitcode to file
func (b *Builder) WriteBitcodeToFile(filename string) error {
	if C.LLVMWriteBitcodeToFile(b.module, C.CString(filename)) != 0 {
		return fmt.Errorf("failed to write bitcode to file: %s", filename)
	}
	return nil
}
