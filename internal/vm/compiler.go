package vm

import (
	"fmt"

	"github.com/mburakmmm/sky-lang/internal/ast"
)

// Compiler compiles AST to bytecode
type Compiler struct {
	instructions []Instruction
	constants    []interface{}
	symbolTable  *SymbolTable
	scopeDepth   int
	functions    map[string]*CompiledFunction // Compiled functions
}

// SymbolTable tracks variables and their stack slots
type SymbolTable struct {
	outer   *SymbolTable
	symbols map[string]int // name -> stack slot
	numDefs int            // number of definitions in this scope
}

// NewSymbolTable creates a new symbol table
func NewSymbolTable(outer *SymbolTable) *SymbolTable {
	return &SymbolTable{
		outer:   outer,
		symbols: make(map[string]int),
		numDefs: 0,
	}
}

// Define adds a new symbol
func (st *SymbolTable) Define(name string) int {
	slot := st.numDefs
	st.symbols[name] = slot
	st.numDefs++
	return slot
}

// Resolve looks up a symbol
func (st *SymbolTable) Resolve(name string) (int, bool) {
	slot, ok := st.symbols[name]
	if ok {
		return slot, true
	}
	if st.outer != nil {
		return st.outer.Resolve(name)
	}
	return 0, false
}

// NewCompiler creates a new bytecode compiler
func NewCompiler() *Compiler {
	return &Compiler{
		instructions: make([]Instruction, 0),
		constants:    make([]interface{}, 0),
		symbolTable:  NewSymbolTable(nil),
		scopeDepth:   0,
		functions:    make(map[string]*CompiledFunction),
	}
}

// Compile compiles an AST program to bytecode
func (c *Compiler) Compile(program *ast.Program) (*Bytecode, error) {
	// Find and compile main function first
	var mainFunc *ast.FunctionStatement
	for _, stmt := range program.Statements {
		if fn, ok := stmt.(*ast.FunctionStatement); ok {
			if fn.Name.Value == "main" {
				mainFunc = fn
			} else {
				// Compile other functions as separate bytecode chunks
				if err := c.compileStatement(stmt); err != nil {
					return nil, err
				}
			}
		} else {
			if err := c.compileStatement(stmt); err != nil {
				return nil, err
			}
		}
	}

	// Compile main function body
	if mainFunc != nil {
		c.enterScope()
		for _, stmt := range mainFunc.Body.Statements {
			if err := c.compileStatement(stmt); err != nil {
				return nil, err
			}
		}
		c.leaveScope()
	}

	// Add halt at the end
	c.emit(Instruction{Op: OpHalt})

	return &Bytecode{
		Instructions: c.instructions,
		Constants:    c.constants,
		Functions:    c.functions,
	}, nil
}

func (c *Compiler) compileStatement(stmt ast.Statement) error {
	switch s := stmt.(type) {
	case *ast.LetStatement:
		return c.compileLetStatement(s)
	case *ast.ConstStatement:
		return c.compileConstStatement(s)
	case *ast.ReturnStatement:
		return c.compileReturnStatement(s)
	case *ast.ExpressionStatement:
		if err := c.compileExpression(s.Expression); err != nil {
			return err
		}
		c.emit(Instruction{Op: OpPop}) // Discard expression result
		return nil
	case *ast.IfStatement:
		return c.compileIfStatement(s)
	case *ast.WhileStatement:
		return c.compileWhileStatement(s)
	case *ast.ForStatement:
		return c.compileForStatement(s)
	case *ast.FunctionStatement:
		return c.compileFunctionStatement(s)
	default:
		return fmt.Errorf("unknown statement type: %T", stmt)
	}
}

func (c *Compiler) compileLetStatement(stmt *ast.LetStatement) error {
	// Compile the value expression
	if err := c.compileExpression(stmt.Value); err != nil {
		return err
	}

	// Define the variable
	slot := c.symbolTable.Define(stmt.Name.Value)
	c.emit(Instruction{
		Op:      OpSetLocal,
		Operand: slot,
		Name:    stmt.Name.Value,
	})

	return nil
}

func (c *Compiler) compileConstStatement(stmt *ast.ConstStatement) error {
	// Same as let for now (const checking done in sema phase)
	if err := c.compileExpression(stmt.Value); err != nil {
		return err
	}

	slot := c.symbolTable.Define(stmt.Name.Value)
	c.emit(Instruction{
		Op:      OpSetLocal,
		Operand: slot,
		Name:    stmt.Name.Value,
	})

	return nil
}

func (c *Compiler) compileReturnStatement(stmt *ast.ReturnStatement) error {
	if stmt.ReturnValue != nil {
		if err := c.compileExpression(stmt.ReturnValue); err != nil {
			return err
		}
	} else {
		c.emit(Instruction{Op: OpNil})
	}
	c.emit(Instruction{Op: OpReturn})
	return nil
}

func (c *Compiler) compileIfStatement(stmt *ast.IfStatement) error {
	// Compile condition
	if err := c.compileExpression(stmt.Condition); err != nil {
		return err
	}

	// Jump if false (we'll patch this later)
	jumpIfFalse := c.emitJump(OpJumpIfFalse)

	// Compile consequence
	for _, s := range stmt.Consequence.Statements {
		if err := c.compileStatement(s); err != nil {
			return err
		}
	}

	// Jump over alternative (if exists)
	jumpOver := c.emitJump(OpJump)

	// Patch jump-if-false to here
	c.patchJump(jumpIfFalse)

	// Compile elif/else
	if len(stmt.Elif) > 0 || stmt.Alternative != nil {
		// Handle elif chains
		for _, elif := range stmt.Elif {
			if err := c.compileExpression(elif.Condition); err != nil {
				return err
			}
			elifJumpIfFalse := c.emitJump(OpJumpIfFalse)
			for _, s := range elif.Consequence.Statements {
				if err := c.compileStatement(s); err != nil {
					return err
				}
			}
			elifJumpOver := c.emitJump(OpJump)
			c.patchJump(elifJumpIfFalse)
			defer c.patchJump(elifJumpOver)
		}

		// Compile else
		if stmt.Alternative != nil {
			for _, s := range stmt.Alternative.Statements {
				if err := c.compileStatement(s); err != nil {
					return err
				}
			}
		}
	}

	// Patch jump-over to here
	c.patchJump(jumpOver)

	return nil
}

func (c *Compiler) compileWhileStatement(stmt *ast.WhileStatement) error {
	loopStart := len(c.instructions)

	// Compile condition
	if err := c.compileExpression(stmt.Condition); err != nil {
		return err
	}

	// Jump if false (exit loop)
	exitJump := c.emitJump(OpJumpIfFalse)

	// Compile body
	for _, s := range stmt.Body.Statements {
		if err := c.compileStatement(s); err != nil {
			return err
		}
	}

	// Loop back to start
	c.emitLoop(loopStart)

	// Patch exit jump
	c.patchJump(exitJump)

	return nil
}

func (c *Compiler) compileForStatement(stmt *ast.ForStatement) error {
	// Compile iterable expression
	if err := c.compileExpression(stmt.Iterable); err != nil {
		return err
	}

	// For now, assume range() returns a list
	// TODO: proper iterator protocol

	loopStart := len(c.instructions)

	// Get next element (simplified)
	// TODO: proper iterator handling

	// Define loop variable
	slot := c.symbolTable.Define(stmt.Iterator.Value)
	c.emit(Instruction{
		Op:      OpSetLocal,
		Operand: slot,
		Name:    stmt.Iterator.Value,
	})

	// Compile body
	for _, s := range stmt.Body.Statements {
		if err := c.compileStatement(s); err != nil {
			return err
		}
	}

	// Loop back
	c.emitLoop(loopStart)

	return nil
}

func (c *Compiler) compileFunctionStatement(stmt *ast.FunctionStatement) error {
	funcName := stmt.Name.Value
	arity := len(stmt.Parameters)

	// Create a new compiler for this function
	funcCompiler := NewCompiler()
	funcCompiler.symbolTable = NewSymbolTable(nil)

	// Define parameters as locals
	for _, param := range stmt.Parameters {
		funcCompiler.symbolTable.Define(param.Name.Value)
	}

	// Compile function body
	for _, bodyStmt := range stmt.Body.Statements {
		if err := funcCompiler.compileStatement(bodyStmt); err != nil {
			return err
		}
	}

	// Add implicit return nil if no explicit return
	funcCompiler.emit(Instruction{Op: OpNil})
	funcCompiler.emit(Instruction{Op: OpReturn})

	// Create compiled function
	compiledFunc := &CompiledFunction{
		Name:         funcName,
		Arity:        arity,
		Async:        stmt.Async, // Store async flag
		Instructions: funcCompiler.instructions,
		Constants:    funcCompiler.constants,
		LocalCount:   funcCompiler.symbolTable.numDefs,
	}

	// Store in compiler's function table
	c.functions[funcName] = compiledFunc

	// Define function name in current scope
	c.symbolTable.Define(funcName)

	return nil
}

func (c *Compiler) compileExpression(expr ast.Expression) error {
	switch e := expr.(type) {
	case *ast.IntegerLiteral:
		idx := c.addConstant(e.Value)
		c.emit(Instruction{Op: OpConstant, Operand: idx})
		return nil

	case *ast.FloatLiteral:
		idx := c.addConstant(e.Value)
		c.emit(Instruction{Op: OpConstant, Operand: idx})
		return nil

	case *ast.StringLiteral:
		idx := c.addConstant(e.Value)
		c.emit(Instruction{Op: OpConstant, Operand: idx})
		return nil

	case *ast.BooleanLiteral:
		if e.Value {
			c.emit(Instruction{Op: OpTrue})
		} else {
			c.emit(Instruction{Op: OpFalse})
		}
		return nil

	case *ast.Identifier:
		// Variables and non-call identifiers
		if slot, ok := c.symbolTable.Resolve(e.Value); ok {
			c.emit(Instruction{
				Op:      OpGetLocal,
				Operand: slot,
				Name:    e.Value,
			})
		} else {
			c.emit(Instruction{
				Op:   OpGetGlobal,
				Name: e.Value,
			})
		}
		return nil

	case *ast.InfixExpression:
		return c.compileInfixExpression(e)

	case *ast.PrefixExpression:
		return c.compilePrefixExpression(e)

	case *ast.CallExpression:
		return c.compileCallExpression(e)

	case *ast.AwaitExpression:
		return c.compileAwaitExpression(e)

	case *ast.YieldExpression:
		return c.compileYieldExpression(e)

	default:
		return fmt.Errorf("unknown expression type: %T", expr)
	}
}

func (c *Compiler) compileInfixExpression(expr *ast.InfixExpression) error {
	// Handle assignment operators
	if expr.Operator == "=" || expr.Operator == "+=" || expr.Operator == "-=" ||
		expr.Operator == "*=" || expr.Operator == "/=" || expr.Operator == "%=" {
		// Compile right side
		if err := c.compileExpression(expr.Right); err != nil {
			return err
		}

		// Get left identifier
		ident, ok := expr.Left.(*ast.Identifier)
		if !ok {
			return fmt.Errorf("invalid assignment target")
		}

		// For compound assignments (+=, -=, etc.), need to load current value first
		if expr.Operator != "=" {
			// Load current value
			if slot, ok := c.symbolTable.Resolve(ident.Value); ok {
				c.emit(Instruction{Op: OpGetLocal, Operand: slot, Name: ident.Value})
			} else {
				c.emit(Instruction{Op: OpGetGlobal, Name: ident.Value})
			}

			// Swap stack so we have: [current, new] instead of [new, current]
			// Then emit operation
			switch expr.Operator {
			case "+=":
				c.emit(Instruction{Op: OpAdd})
			case "-=":
				c.emit(Instruction{Op: OpSub})
			case "*=":
				c.emit(Instruction{Op: OpMul})
			case "/=":
				c.emit(Instruction{Op: OpDiv})
			case "%=":
				c.emit(Instruction{Op: OpMod})
			}
		}

		// Store value
		if slot, ok := c.symbolTable.Resolve(ident.Value); ok {
			c.emit(Instruction{Op: OpSetLocal, Operand: slot, Name: ident.Value})
		} else {
			c.emit(Instruction{Op: OpSetGlobal, Name: ident.Value})
		}

		return nil
	}

	// Short-circuit for && and ||
	if expr.Operator == "&&" {
		if err := c.compileExpression(expr.Left); err != nil {
			return err
		}
		endJump := c.emitJump(OpJumpIfFalse)
		c.emit(Instruction{Op: OpPop}) // Pop left value
		if err := c.compileExpression(expr.Right); err != nil {
			return err
		}
		c.patchJump(endJump)
		return nil
	}

	if expr.Operator == "||" {
		if err := c.compileExpression(expr.Left); err != nil {
			return err
		}
		endJump := c.emitJump(OpJumpIfTrue)
		c.emit(Instruction{Op: OpPop}) // Pop left value
		if err := c.compileExpression(expr.Right); err != nil {
			return err
		}
		c.patchJump(endJump)
		return nil
	}

	// Compile left and right operands
	if err := c.compileExpression(expr.Left); err != nil {
		return err
	}
	if err := c.compileExpression(expr.Right); err != nil {
		return err
	}

	// Emit operator
	switch expr.Operator {
	case "+":
		c.emit(Instruction{Op: OpAdd})
	case "-":
		c.emit(Instruction{Op: OpSub})
	case "*":
		c.emit(Instruction{Op: OpMul})
	case "/":
		c.emit(Instruction{Op: OpDiv})
	case "%":
		c.emit(Instruction{Op: OpMod})
	case "==":
		c.emit(Instruction{Op: OpEqual})
	case "!=":
		c.emit(Instruction{Op: OpNotEqual})
	case ">":
		c.emit(Instruction{Op: OpGreaterThan})
	case ">=":
		c.emit(Instruction{Op: OpGreaterEqual})
	case "<":
		c.emit(Instruction{Op: OpLessThan})
	case "<=":
		c.emit(Instruction{Op: OpLessEqual})
	default:
		return fmt.Errorf("unknown operator: %s", expr.Operator)
	}

	return nil
}

func (c *Compiler) compilePrefixExpression(expr *ast.PrefixExpression) error {
	if err := c.compileExpression(expr.Right); err != nil {
		return err
	}

	switch expr.Operator {
	case "!":
		c.emit(Instruction{Op: OpNot})
	case "-":
		c.emit(Instruction{Op: OpNegate})
	default:
		return fmt.Errorf("unknown prefix operator: %s", expr.Operator)
	}

	return nil
}

func (c *Compiler) compileCallExpression(expr *ast.CallExpression) error {
	// Check for built-in functions
	if ident, ok := expr.Function.(*ast.Identifier); ok {
		switch ident.Value {
		case "print":
			// Compile arguments
			for _, arg := range expr.Arguments {
				if err := c.compileExpression(arg); err != nil {
					return err
				}
			}
			c.emit(Instruction{Op: OpPrint, Operand: len(expr.Arguments)})
			return nil

		case "len":
			if len(expr.Arguments) > 0 {
				if err := c.compileExpression(expr.Arguments[0]); err != nil {
					return err
				}
			}
			c.emit(Instruction{Op: OpLen})
			return nil

		case "range":
			if len(expr.Arguments) > 0 {
				if err := c.compileExpression(expr.Arguments[0]); err != nil {
					return err
				}
			}
			c.emit(Instruction{Op: OpRange})
			return nil
		}

		// User-defined function call
		// Push function name as constant (for OpCall to lookup)
		idx := c.addConstant(ident.Value)
		c.emit(Instruction{Op: OpConstant, Operand: idx})

		// Compile arguments
		for _, arg := range expr.Arguments {
			if err := c.compileExpression(arg); err != nil {
				return err
			}
		}

		c.emit(Instruction{
			Op:      OpCall,
			Operand: len(expr.Arguments),
		})

		return nil
	}

	// Non-identifier function (lambda, etc.)
	if err := c.compileExpression(expr.Function); err != nil {
		return err
	}

	// Compile arguments
	for _, arg := range expr.Arguments {
		if err := c.compileExpression(arg); err != nil {
			return err
		}
	}

	c.emit(Instruction{
		Op:      OpCall,
		Operand: len(expr.Arguments),
	})

	return nil
}

func (c *Compiler) compileAssignExpression(expr *ast.InfixExpression) error {
	// Compile right side
	if err := c.compileExpression(expr.Right); err != nil {
		return err
	}

	// Get left identifier
	ident, ok := expr.Left.(*ast.Identifier)
	if !ok {
		return fmt.Errorf("invalid assignment target")
	}

	// Emit set instruction
	if slot, ok := c.symbolTable.Resolve(ident.Value); ok {
		c.emit(Instruction{
			Op:      OpSetLocal,
			Operand: slot,
			Name:    ident.Value,
		})
	} else {
		c.emit(Instruction{
			Op:   OpSetGlobal,
			Name: ident.Value,
		})
	}

	return nil
}

// Helper methods
func (c *Compiler) emit(ins Instruction) int {
	pos := len(c.instructions)
	c.instructions = append(c.instructions, ins)
	return pos
}

func (c *Compiler) addConstant(value interface{}) int {
	c.constants = append(c.constants, value)
	return len(c.constants) - 1
}

func (c *Compiler) emitJump(op OpCode) int {
	return c.emit(Instruction{
		Op:      op,
		Operand: 9999, // Placeholder
	})
}

func (c *Compiler) patchJump(offset int) {
	jump := len(c.instructions) - offset
	c.instructions[offset].Operand = len(c.instructions)
	_ = jump // For backward compat
}

func (c *Compiler) emitLoop(loopStart int) {
	c.emit(Instruction{
		Op:      OpLoop,
		Operand: loopStart,
	})
}

func (c *Compiler) enterScope() {
	c.symbolTable = NewSymbolTable(c.symbolTable)
	c.scopeDepth++
}

func (c *Compiler) leaveScope() {
	c.symbolTable = c.symbolTable.outer
	c.scopeDepth--
}

func (c *Compiler) compileAwaitExpression(expr *ast.AwaitExpression) error {
	// Compile the expression being awaited
	if err := c.compileExpression(expr.Expression); err != nil {
		return err
	}

	// Emit OpAwait instruction
	c.emit(Instruction{Op: OpAwait})
	return nil
}

func (c *Compiler) compileYieldExpression(expr *ast.YieldExpression) error {
	// Compile the value to yield
	if expr.Value != nil {
		if err := c.compileExpression(expr.Value); err != nil {
			return err
		}
	} else {
		// Yield nil if no value
		c.emit(Instruction{Op: OpNil})
	}

	// Emit OpYield instruction
	c.emit(Instruction{Op: OpYield})
	return nil
}
