package vm

import (
	"fmt"

	"github.com/mburakmmm/sky-lang/internal/interpreter"
)

// VM is a stack-based virtual machine
type VM struct {
	bytecode *Bytecode
	stack    []interface{}
	sp       int // stack pointer
	globals  map[string]interface{}
	ip       int // instruction pointer
	frames   []*CallFrame
	fp       int // frame pointer
}

// CallFrame represents a function call frame
type CallFrame struct {
	function    string
	returnAddr  int
	basePointer int
	localCount  int
}

// NewVM creates a new virtual machine
func NewVM(bytecode *Bytecode) *VM {
	vm := &VM{
		bytecode: bytecode,
		stack:    make([]interface{}, 0, 2048),
		sp:       0,
		globals:  make(map[string]interface{}),
		ip:       0,
		frames:   make([]*CallFrame, 0, 1024),
		fp:       0,
	}

	// Load compiled functions into global namespace
	for name, fn := range bytecode.Functions {
		vm.globals[name] = fn
	}

	return vm
}

// Run executes the bytecode
func (vm *VM) Run() error {
	for vm.ip < len(vm.bytecode.Instructions) {
		ins := vm.bytecode.Instructions[vm.ip]
		vm.ip++

		switch ins.Op {
		case OpConstant:
			constant := vm.bytecode.Constants[ins.Operand]
			vm.push(constant)

		case OpPop:
			vm.pop()

		case OpDup:
			val := vm.peek(0)
			vm.push(val)

		case OpTrue:
			vm.push(true)

		case OpFalse:
			vm.push(false)

		case OpNil:
			vm.push(nil)

		case OpAdd:
			if err := vm.binaryOp("+"); err != nil {
				return err
			}

		case OpSub:
			if err := vm.binaryOp("-"); err != nil {
				return err
			}

		case OpMul:
			if err := vm.binaryOp("*"); err != nil {
				return err
			}

		case OpDiv:
			if err := vm.binaryOp("/"); err != nil {
				return err
			}

		case OpMod:
			if err := vm.binaryOp("%"); err != nil {
				return err
			}

		case OpNegate:
			val := vm.pop()
			switch v := val.(type) {
			case int64:
				vm.push(-v)
			case float64:
				vm.push(-v)
			default:
				return fmt.Errorf("cannot negate %T", val)
			}

		case OpNot:
			val := vm.pop()
			vm.push(!vm.isTruthy(val))

		case OpEqual:
			b := vm.pop()
			a := vm.pop()
			vm.push(vm.isEqual(a, b))

		case OpNotEqual:
			b := vm.pop()
			a := vm.pop()
			vm.push(!vm.isEqual(a, b))

		case OpGreaterThan:
			if err := vm.comparison(">"); err != nil {
				return err
			}

		case OpGreaterEqual:
			if err := vm.comparison(">="); err != nil {
				return err
			}

		case OpLessThan:
			if err := vm.comparison("<"); err != nil {
				return err
			}

		case OpLessEqual:
			if err := vm.comparison("<="); err != nil {
				return err
			}

		case OpGetLocal:
			// Get local variable from current frame
			basePtr := vm.frameBase()
			slot := basePtr + ins.Operand

			if slot >= len(vm.stack) {
				return fmt.Errorf("stack underflow: trying to access slot %d (sp=%d)", slot, vm.sp)
			}

			val := vm.stack[slot]
			vm.push(val)

		case OpSetLocal:
			val := vm.peek(0)
			basePtr := vm.frameBase()
			slot := basePtr + ins.Operand

			// Extend stack if needed
			for len(vm.stack) <= slot {
				vm.stack = append(vm.stack, nil)
			}
			vm.stack[slot] = val

		case OpGetGlobal:
			val, ok := vm.globals[ins.Name]
			if !ok {
				return fmt.Errorf("undefined variable: %s", ins.Name)
			}
			vm.push(val)

		case OpSetGlobal:
			val := vm.peek(0)
			vm.globals[ins.Name] = val

		case OpJump:
			vm.ip = ins.Operand

		case OpJumpIfFalse:
			condition := vm.peek(0)
			if !vm.isTruthy(condition) {
				vm.ip = ins.Operand
			}
			vm.pop()

		case OpJumpIfTrue:
			condition := vm.peek(0)
			if vm.isTruthy(condition) {
				vm.ip = ins.Operand
			}
			vm.pop()

		case OpLoop:
			vm.ip = ins.Operand

		case OpCall:
			argCount := ins.Operand

			// Arguments are on top of stack, function name is below them
			// Stack layout: [funcName, arg1, arg2, ...]

			// Pop arguments
			args := make([]interface{}, argCount)
			for i := argCount - 1; i >= 0; i-- {
				args[i] = vm.pop()
			}

			// Pop function name/reference
			funcVal := vm.pop()

			// Handle different function value types
			var compiledFunc *CompiledFunction
			var funcName string

			switch f := funcVal.(type) {
			case string:
				// Function name as string
				funcName = f
				var ok bool
				compiledFunc, ok = vm.bytecode.Functions[funcName]
				if !ok {
					return fmt.Errorf("undefined function: %s", funcName)
				}
			case *CompiledFunction:
				// Direct function reference
				compiledFunc = f
				funcName = f.Name
			default:
				return fmt.Errorf("expected function, got %T", funcVal)
			}

			// Check arity
			if argCount != compiledFunc.Arity {
				return fmt.Errorf("function %s expects %d arguments, got %d", funcName, compiledFunc.Arity, argCount)
			}

			// Push arguments back onto stack (they'll be locals in the function)
			for _, arg := range args {
				vm.push(arg)
			}

			// Create new call frame
			frame := &CallFrame{
				function:    funcName,
				returnAddr:  vm.ip,
				basePointer: vm.sp - argCount,
				localCount:  compiledFunc.LocalCount,
			}

			// Push frame
			vm.frames = append(vm.frames, frame)
			vm.fp = len(vm.frames)

			// Save current state and switch to function
			savedIP := vm.ip
			savedBytecode := vm.bytecode

			// Execute function bytecode
			funcVM := &VM{
				bytecode: &Bytecode{
					Instructions: compiledFunc.Instructions,
					Constants:    compiledFunc.Constants,
					Functions:    vm.bytecode.Functions,
				},
				stack:   vm.stack,
				sp:      vm.sp,
				globals: vm.globals,
				ip:      0,
				frames:  vm.frames,
				fp:      vm.fp,
			}

			if err := funcVM.Run(); err != nil {
				return err
			}

			// Restore state
			vm.stack = funcVM.stack
			vm.sp = funcVM.sp
			vm.ip = savedIP
			vm.bytecode = savedBytecode

			// Pop frame
			if len(vm.frames) > 0 {
				vm.frames = vm.frames[:len(vm.frames)-1]
				vm.fp = len(vm.frames)
			}

		case OpReturn:
			returnValue := vm.pop()

			// If we're in a function, return from it
			if len(vm.frames) > 0 {
				// Pop frame but leave return value
				vm.frames = vm.frames[:len(vm.frames)-1]
				vm.fp = len(vm.frames)

				// Push return value
				vm.push(returnValue)

				// Exit this VM run (we're done with this function)
				return nil
			}

			// Top-level return
			vm.push(returnValue)

		case OpPrint:
			argCount := ins.Operand
			args := make([]interface{}, argCount)
			for i := argCount - 1; i >= 0; i-- {
				args[i] = vm.pop()
			}
			for _, arg := range args {
				fmt.Println(vm.valueToString(arg))
			}
			vm.push(nil)

		case OpLen:
			val := vm.pop()
			switch v := val.(type) {
			case string:
				vm.push(int64(len(v)))
			case []interface{}:
				vm.push(int64(len(v)))
			default:
				return fmt.Errorf("len() not supported for %T", val)
			}

		case OpRange:
			val := vm.pop()
			n, ok := val.(int64)
			if !ok {
				return fmt.Errorf("range() expects integer, got %T", val)
			}
			list := make([]interface{}, n)
			for i := int64(0); i < n; i++ {
				list[i] = i
			}
			vm.push(list)

		case OpAwait:
			// For now, await just passes through the value
			// In a full implementation, this would suspend execution
			// and wait for promise resolution
			value := vm.pop()
			// If it's a promise-like structure, resolve it
			// For simplicity, we just push the value back
			vm.push(value)

		case OpYield:
			// For now, yield just returns the value
			// In a full implementation, this would suspend the coroutine
			// and yield control back to the caller
			value := vm.pop()
			vm.push(value)

		case OpBreak:
			// Signal break to loop (will be caught by loop compilation)
			return fmt.Errorf("__break__")

		case OpContinue:
			// Signal continue to loop (will be caught by loop compilation)
			return fmt.Errorf("__continue__")

		case OpHalt:
			return nil

		default:
			return fmt.Errorf("unknown opcode: %s", ins.Op)
		}
	}

	return nil
}

// Stack operations
func (vm *VM) push(val interface{}) {
	if vm.sp >= len(vm.stack) {
		vm.stack = append(vm.stack, val)
	} else {
		vm.stack[vm.sp] = val
	}
	vm.sp++
}

func (vm *VM) pop() interface{} {
	if vm.sp == 0 {
		return nil
	}
	vm.sp--
	return vm.stack[vm.sp]
}

func (vm *VM) peek(distance int) interface{} {
	return vm.stack[vm.sp-1-distance]
}

func (vm *VM) frameBase() int {
	if len(vm.frames) == 0 {
		return 0
	}
	frame := vm.frames[len(vm.frames)-1]
	// Base pointer points to where arguments start
	return frame.basePointer
}

// Helper methods
func (vm *VM) binaryOp(op string) error {
	b := vm.pop()
	a := vm.pop()

	// String concatenation with type coercion
	if op == "+" {
		aStr, aIsStr := a.(string)
		bStr, bIsStr := b.(string)

		// String + anything
		if aIsStr || bIsStr {
			if !aIsStr {
				aStr = vm.valueToString(a)
			}
			if !bIsStr {
				bStr = vm.valueToString(b)
			}
			vm.push(aStr + bStr)
			return nil
		}
	}

	// Integer arithmetic
	if aInt, ok := a.(int64); ok {
		if bInt, ok := b.(int64); ok {
			switch op {
			case "+":
				vm.push(aInt + bInt)
			case "-":
				vm.push(aInt - bInt)
			case "*":
				vm.push(aInt * bInt)
			case "/":
				if bInt == 0 {
					return fmt.Errorf("division by zero")
				}
				vm.push(aInt / bInt)
			case "%":
				vm.push(aInt % bInt)
			default:
				return fmt.Errorf("unknown operator: %s", op)
			}
			return nil
		}
	}

	// Float arithmetic
	if aFloat, ok := a.(float64); ok {
		if bFloat, ok := b.(float64); ok {
			switch op {
			case "+":
				vm.push(aFloat + bFloat)
			case "-":
				vm.push(aFloat - bFloat)
			case "*":
				vm.push(aFloat * bFloat)
			case "/":
				vm.push(aFloat / bFloat)
			default:
				return fmt.Errorf("unknown operator: %s", op)
			}
			return nil
		}
	}

	return fmt.Errorf("unsupported operands for %s: %T and %T", op, a, b)
}

func (vm *VM) comparison(op string) error {
	b := vm.pop()
	a := vm.pop()

	aInt, aOk := a.(int64)
	bInt, bOk := b.(int64)

	if aOk && bOk {
		switch op {
		case ">":
			vm.push(aInt > bInt)
		case ">=":
			vm.push(aInt >= bInt)
		case "<":
			vm.push(aInt < bInt)
		case "<=":
			vm.push(aInt <= bInt)
		default:
			return fmt.Errorf("unknown comparison: %s", op)
		}
		return nil
	}

	return fmt.Errorf("unsupported operands for %s", op)
}

func (vm *VM) isTruthy(val interface{}) bool {
	if val == nil {
		return false
	}
	if b, ok := val.(bool); ok {
		return b
	}
	if i, ok := val.(int64); ok {
		return i != 0
	}
	return true
}

func (vm *VM) isEqual(a, b interface{}) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	switch aVal := a.(type) {
	case int64:
		if bVal, ok := b.(int64); ok {
			return aVal == bVal
		}
	case float64:
		if bVal, ok := b.(float64); ok {
			return aVal == bVal
		}
	case string:
		if bVal, ok := b.(string); ok {
			return aVal == bVal
		}
	case bool:
		if bVal, ok := b.(bool); ok {
			return aVal == bVal
		}
	}

	return false
}

func (vm *VM) valueToString(val interface{}) string {
	if val == nil {
		return "nil"
	}
	switch v := val.(type) {
	case int64:
		return fmt.Sprintf("%d", v)
	case float64:
		return fmt.Sprintf("%g", v)
	case string:
		return v
	case bool:
		return fmt.Sprintf("%t", v)
	case *interpreter.Integer:
		return fmt.Sprintf("%d", v.Value)
	case *interpreter.String:
		return v.Value
	case *interpreter.Boolean:
		return fmt.Sprintf("%t", v.Value)
	default:
		return fmt.Sprintf("%v", val)
	}
}
