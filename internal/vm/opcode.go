package vm

import "fmt"

// OpCode represents a bytecode instruction
type OpCode byte

const (
	// Stack operations
	OpConstant OpCode = iota // Push constant to stack
	OpPop                    // Pop from stack
	OpDup                    // Duplicate top of stack

	// Variables
	OpGetLocal  // Get local variable
	OpSetLocal  // Set local variable
	OpGetGlobal // Get global variable
	OpSetGlobal // Set global variable

	// Arithmetic
	OpAdd    // a + b
	OpSub    // a - b
	OpMul    // a * b
	OpDiv    // a / b
	OpMod    // a % b
	OpNegate // -a

	// Comparison
	OpEqual        // a == b
	OpNotEqual     // a != b
	OpGreaterThan  // a > b
	OpGreaterEqual // a >= b
	OpLessThan     // a < b
	OpLessEqual    // a <= b

	// Logic
	OpAnd // a && b
	OpOr  // a || b
	OpNot // !a

	// Control flow
	OpJump        // Unconditional jump
	OpJumpIfFalse // Jump if top of stack is false
	OpJumpIfTrue  // Jump if top of stack is true
	OpLoop        // Jump backward (for loops)
	OpBreak       // Break from loop
	OpContinue    // Continue to next iteration

	// Functions
	OpCall      // Call function
	OpCallAsync // Call async function (returns Promise)
	OpReturn    // Return from function

	// Async
	OpAwait // Await a promise
	OpYield // Yield a value (for coroutines)

	// Built-ins
	OpPrint // print() built-in
	OpLen   // len() built-in
	OpRange // range() built-in

	// Special
	OpTrue  // Push true
	OpFalse // Push false
	OpNil   // Push nil
	OpHalt  // Stop execution
)

// String returns the string representation of an opcode
func (op OpCode) String() string {
	switch op {
	case OpConstant:
		return "CONSTANT"
	case OpPop:
		return "POP"
	case OpDup:
		return "DUP"
	case OpGetLocal:
		return "GET_LOCAL"
	case OpSetLocal:
		return "SET_LOCAL"
	case OpGetGlobal:
		return "GET_GLOBAL"
	case OpSetGlobal:
		return "SET_GLOBAL"
	case OpAdd:
		return "ADD"
	case OpSub:
		return "SUB"
	case OpMul:
		return "MUL"
	case OpDiv:
		return "DIV"
	case OpMod:
		return "MOD"
	case OpNegate:
		return "NEGATE"
	case OpEqual:
		return "EQUAL"
	case OpNotEqual:
		return "NOT_EQUAL"
	case OpGreaterThan:
		return "GREATER"
	case OpGreaterEqual:
		return "GREATER_EQ"
	case OpLessThan:
		return "LESS"
	case OpLessEqual:
		return "LESS_EQ"
	case OpAnd:
		return "AND"
	case OpOr:
		return "OR"
	case OpNot:
		return "NOT"
	case OpJump:
		return "JUMP"
	case OpJumpIfFalse:
		return "JUMP_IF_FALSE"
	case OpJumpIfTrue:
		return "JUMP_IF_TRUE"
	case OpLoop:
		return "LOOP"
	case OpBreak:
		return "BREAK"
	case OpContinue:
		return "CONTINUE"
	case OpCall:
		return "CALL"
	case OpCallAsync:
		return "CALL_ASYNC"
	case OpReturn:
		return "RETURN"
	case OpAwait:
		return "AWAIT"
	case OpYield:
		return "YIELD"
	case OpPrint:
		return "PRINT"
	case OpLen:
		return "LEN"
	case OpRange:
		return "RANGE"
	case OpTrue:
		return "TRUE"
	case OpFalse:
		return "FALSE"
	case OpNil:
		return "NIL"
	case OpHalt:
		return "HALT"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", op)
	}
}

// Instruction represents a single bytecode instruction with optional operands
type Instruction struct {
	Op       OpCode
	Operand  int    // For jumps, local indices, constant indices
	Operand2 int    // For Call (arg count), etc.
	Name     string // For variable names
}

// String returns the string representation of an instruction
func (ins Instruction) String() string {
	switch ins.Op {
	case OpConstant:
		return fmt.Sprintf("%-16s %d", ins.Op, ins.Operand)
	case OpGetLocal, OpSetLocal:
		return fmt.Sprintf("%-16s %d (%s)", ins.Op, ins.Operand, ins.Name)
	case OpGetGlobal, OpSetGlobal:
		return fmt.Sprintf("%-16s %s", ins.Op, ins.Name)
	case OpJump, OpJumpIfFalse, OpJumpIfTrue, OpLoop:
		return fmt.Sprintf("%-16s -> %d", ins.Op, ins.Operand)
	case OpCall:
		return fmt.Sprintf("%-16s %d args", ins.Op, ins.Operand)
	default:
		return ins.Op.String()
	}
}

// Bytecode represents compiled bytecode
type Bytecode struct {
	Instructions []Instruction
	Constants    []interface{}                // int64, float64, string, bool
	Functions    map[string]*CompiledFunction // Compiled functions
}

// Disassemble prints bytecode in human-readable format
func (bc *Bytecode) Disassemble(name string) {
	fmt.Printf("== %s ==\n", name)

	// Print compiled functions
	if len(bc.Functions) > 0 {
		fmt.Println("\n=== Compiled Functions ===")
		for fname, fn := range bc.Functions {
			fmt.Printf("\nFunction: %s (arity: %d, locals: %d)\n", fname, fn.Arity, fn.LocalCount)
			for i, ins := range fn.Instructions {
				fmt.Printf("  %04d  %s\n", i, ins)
			}
		}
		fmt.Println("\n=== Main Code ===")
	}

	for i, ins := range bc.Instructions {
		fmt.Printf("%04d  %s\n", i, ins)
	}
	fmt.Println()
}
