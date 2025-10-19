package interpreter

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/mburakmmm/sky-lang/internal/ast"
	"github.com/mburakmmm/sky-lang/internal/lexer"
	"github.com/mburakmmm/sky-lang/internal/parser"
)

// Interpreter AST'yi yorumlar ve çalıştırır
type Interpreter struct {
	env            *Environment
	output         *os.File
	trampoline     *TrampolineStack        // Custom call stack for recursion
	moduleCache    map[string]*Environment // Cached loaded modules
	currentDir     string                  // Current working directory for relative imports
	sourceFile     string                  // Source file path for relative imports
	recursionDepth int                     // Track recursion depth
}

// New yeni bir interpreter oluşturur
func New() *Interpreter {
	env := NewEnvironment(nil)
	trampoline := NewTrampolineStack(10000) // 10K depth limit

	// Get current working directory
	currentDir, _ := os.Getwd()

	// Built-in fonksiyonları ekle
	env.Set("print", &Function{
		Name: "print",
		Body: func(env *Environment) (Value, error) {
			// Tüm argümanları yazdır
			args, _ := env.Get("__args__")
			if list, ok := args.(*List); ok {
				for i, arg := range list.Elements {
					if i > 0 {
						fmt.Print(" ")
					}
					fmt.Print(arg.String())
				}
				fmt.Println()
			}
			return &Nil{}, nil
		},
	})

	env.Set("len", &Function{
		Name: "len",
		Body: func(env *Environment) (Value, error) {
			args, _ := env.Get("__args__")
			if list, ok := args.(*List); ok && len(list.Elements) > 0 {
				arg := list.Elements[0]
				switch v := arg.(type) {
				case *String:
					return &Integer{Value: int64(len(v.Value))}, nil
				case *List:
					return &Integer{Value: int64(len(v.Elements))}, nil
				case *Dict:
					return &Integer{Value: int64(len(v.Pairs))}, nil
				}
			}
			return &Integer{Value: 0}, nil
		},
	})

	env.Set("range", &Function{
		Name: "range",
		Body: func(env *Environment) (Value, error) {
			args, _ := env.Get("__args__")
			if list, ok := args.(*List); ok && len(list.Elements) > 0 {
				if n, ok := list.Elements[0].(*Integer); ok {
					elements := make([]Value, n.Value)
					for i := int64(0); i < n.Value; i++ {
						elements[i] = &Integer{Value: i}
					}
					return &List{Elements: elements}, nil
				}
			}
			return &List{Elements: []Value{}}, nil
		},
	})

	// TYPE CONVERSION FUNCTIONS
	addTypeConversionFunctions(env)

	// NUMERIC FUNCTIONS
	addNumericFunctions(env)

	// UTILITY FUNCTIONS
	addUtilityFunctions(env)

	// FUNCTIONAL PROGRAMMING
	addFunctionalFunctions(env)

	// STRING METHODS
	addStringMethods(env)

	// LIST METHODS
	addListMethods(env)

	// DICT METHODS
	addDictMethods(env)

	// NATIVE STDLIB (Go functions)
	addNativeStdlib(env)

	return &Interpreter{
		env:         env,
		output:      os.Stdout,
		trampoline:  trampoline,
		moduleCache: make(map[string]*Environment),
		currentDir:  currentDir,
	}
}

// SetSourceFile sets the source file path for imports
func (i *Interpreter) SetSourceFile(path string) {
	i.sourceFile = path
}

// Eval programı çalıştırır
func (i *Interpreter) Eval(program *ast.Program) error {
	// main fonksiyonunu ara
	var mainFunc *ast.FunctionStatement

	// Önce tüm fonksiyonları tanımla
	for _, stmt := range program.Statements {
		if funcStmt, ok := stmt.(*ast.FunctionStatement); ok {
			err := i.evalFunctionStatement(funcStmt)
			if err != nil {
				return err
			}

			if funcStmt.Name.Value == "main" {
				mainFunc = funcStmt
			}
		} else {
			// Global statement'ları çalıştır
			_, err := i.evalStatement(stmt)
			if err != nil {
				return err
			}
		}
	}

	// main fonksiyonunu çağır
	if mainFunc != nil {
		mainFn, _ := i.env.Get("main")
		if fn, ok := mainFn.(*Function); ok {
			// main'i çağır
			newEnv := NewEnvironment(i.env)
			newEnv.Set("__args__", &List{Elements: []Value{}})
			_, err := fn.Body(newEnv)
			return err
		}
	}

	return nil
}

func (i *Interpreter) evalStatement(stmt ast.Statement) (Value, error) {
	switch s := stmt.(type) {
	case *ast.LetStatement:
		return i.evalLetStatement(s)
	case *ast.ConstStatement:
		return i.evalConstStatement(s)
	case *ast.ReturnStatement:
		return i.evalReturnStatement(s)
	case *ast.BreakStatement:
		return nil, &BreakSignal{}
	case *ast.ContinueStatement:
		return nil, &ContinueSignal{}
	case *ast.ExpressionStatement:
		return i.evalExpression(s.Expression)
	case *ast.FunctionStatement:
		return nil, i.evalFunctionStatement(s)
	case *ast.IfStatement:
		return i.evalIfStatement(s)
	case *ast.WhileStatement:
		return i.evalWhileStatement(s)
	case *ast.ForStatement:
		return i.evalForStatement(s)
	case *ast.ClassStatement:
		return nil, i.evalClassStatement(s)
	case *ast.ImportStatement:
		return nil, i.evalImportStatement(s)
	case *ast.UnsafeStatement:
		return i.evalUnsafeStatement(s)
	case *ast.EnumStatement:
		return i.evalEnumStatement(s)
	case *ast.BlockStatement:
		return i.evalBlockStatement(s, i.env)
	default:
		return nil, &RuntimeError{Message: fmt.Sprintf("unknown statement type: %T", stmt)}
	}
}

func (i *Interpreter) evalLetStatement(stmt *ast.LetStatement) (Value, error) {
	value, err := i.evalExpression(stmt.Value)
	if err != nil {
		return nil, err
	}
	i.env.Set(stmt.Name.Value, value)
	return value, nil
}

func (i *Interpreter) evalConstStatement(stmt *ast.ConstStatement) (Value, error) {
	value, err := i.evalExpression(stmt.Value)
	if err != nil {
		return nil, err
	}
	i.env.Set(stmt.Name.Value, value)
	return value, nil
}

func (i *Interpreter) evalReturnStatement(stmt *ast.ReturnStatement) (Value, error) {
	if stmt.ReturnValue != nil {
		return i.evalExpression(stmt.ReturnValue)
	}
	return &Nil{}, nil
}

func (i *Interpreter) evalFunctionStatement(stmt *ast.FunctionStatement) error {
	// Parametrelerin isimlerini al
	params := make([]string, len(stmt.Parameters))
	for idx, param := range stmt.Parameters {
		params[idx] = param.Name.Value
	}

	funcName := stmt.Name.Value
	capturedStmt := stmt // Capture for closure

	// Fonksiyonu closure olarak sakla
	capturedEnv := i.env
	fn := &Function{
		Name:       funcName,
		Parameters: params,
		Env:        capturedEnv,
		Async:      stmt.Async, // Store async flag
		Body: func(callEnv *Environment) (Value, error) {
			// Check memoization cache first
			args, _ := callEnv.Get("__args__")
			if argList, ok := args.(*List); ok {
				if cached, found := i.trampoline.GetCached(funcName, argList.Elements); found {
					return cached, nil
				}
			}

			// Yeni environment oluştur (fonksiyon tanımlandığı env'i parent olarak al)
			fnEnv := NewEnvironment(capturedEnv)

			// Parametreleri bind et
			if argList, ok := args.(*List); ok {
				for idx, paramName := range params {
					if idx < len(argList.Elements) {
						fnEnv.Set(paramName, argList.Elements[idx])
					} else {
						fnEnv.Set(paramName, &Nil{})
					}
				}
			}

			// Check recursion depth
			if i.recursionDepth >= 1000 {
				return nil, &RuntimeError{Message: fmt.Sprintf("maximum recursion depth exceeded (1000) in function '%s'", funcName)}
			}
			i.recursionDepth++
			defer func() {
				i.recursionDepth--
			}()

			// Fonksiyon body'sini çalıştır
			oldEnv := i.env
			i.env = fnEnv
			defer func() { i.env = oldEnv }()

			result, err := i.evalBlockStatement(capturedStmt.Body, fnEnv)

			// Cache successful results
			if err == nil && result != nil {
				if argList, ok := args.(*List); ok {
					i.trampoline.SetCached(funcName, argList.Elements, result)
				}
			}

			return result, err
		},
	}

	i.env.Set(funcName, fn)
	return nil
}

func (i *Interpreter) evalIfStatement(stmt *ast.IfStatement) (Value, error) {
	condition, err := i.evalExpression(stmt.Condition)
	if err != nil {
		return nil, err
	}

	if condition.IsTruthy() {
		return i.evalBlockStatement(stmt.Consequence, i.env)
	}

	// Elif dallarını kontrol et
	for _, elif := range stmt.Elif {
		cond, err := i.evalExpression(elif.Condition)
		if err != nil {
			return nil, err
		}
		if cond.IsTruthy() {
			return i.evalBlockStatement(elif.Consequence, i.env)
		}
	}

	// Else dalı
	if stmt.Alternative != nil {
		return i.evalBlockStatement(stmt.Alternative, i.env)
	}

	return &Nil{}, nil
}

func (i *Interpreter) evalWhileStatement(stmt *ast.WhileStatement) (Value, error) {
	for {
		condition, err := i.evalExpression(stmt.Condition)
		if err != nil {
			return nil, err
		}

		if !condition.IsTruthy() {
			break
		}

		_, err = i.evalBlockStatement(stmt.Body, i.env)
		if err != nil {
			// Handle break/continue signals
			if _, isBreak := err.(*BreakSignal); isBreak {
				break // Exit the loop
			}
			if _, isContinue := err.(*ContinueSignal); isContinue {
				continue // Continue to next iteration
			}
			// Real error
			return nil, err
		}
	}

	return &Nil{}, nil
}

func (i *Interpreter) evalForStatement(stmt *ast.ForStatement) (Value, error) {
	iterable, err := i.evalExpression(stmt.Iterable)
	if err != nil {
		return nil, err
	}

	// Create new scope
	forEnv := NewEnvironment(i.env)
	oldEnv := i.env
	i.env = forEnv
	defer func() { i.env = oldEnv }()

	// Handle different iterable types
	switch iter := iterable.(type) {
	case *List:
		// Iterate over list elements
		for _, elem := range iter.Elements {
			i.env.Set(stmt.Iterator.Value, elem)
			_, err := i.evalBlockStatement(stmt.Body, i.env)
			if err != nil {
				if _, isBreak := err.(*BreakSignal); isBreak {
					break
				}
				if _, isContinue := err.(*ContinueSignal); isContinue {
					continue
				}
				return nil, err
			}
		}

	case *Dict:
		// Iterate over dict keys
		for key := range iter.Pairs {
			i.env.Set(stmt.Iterator.Value, &String{Value: key})
			_, err := i.evalBlockStatement(stmt.Body, i.env)
			if err != nil {
				if _, isBreak := err.(*BreakSignal); isBreak {
					break
				}
				if _, isContinue := err.(*ContinueSignal); isContinue {
					continue
				}
				return nil, err
			}
		}

	case *String:
		// Iterate over string characters
		for _, ch := range iter.Value {
			i.env.Set(stmt.Iterator.Value, &String{Value: string(ch)})
			_, err := i.evalBlockStatement(stmt.Body, i.env)
			if err != nil {
				if _, isBreak := err.(*BreakSignal); isBreak {
					break
				}
				if _, isContinue := err.(*ContinueSignal); isContinue {
					continue
				}
				return nil, err
			}
		}

	case *Instance:
		// Check for __iter__ method (iterator protocol)
		if iterMethod, hasIter := iter.Get("__iter__"); hasIter {
			if fn, ok := iterMethod.(*Function); ok {
				// Call __iter__ to get iterator
				callEnv := NewEnvironment(fn.Env)
				callEnv.Set("__args__", &List{Elements: []Value{}})
				callEnv.Set("self", iter)

				iterator, err := fn.Body(callEnv)
				if err != nil {
					return nil, err
				}

				// Use iterator protocol
				if iterInstance, ok := iterator.(*Instance); ok {
					for {
						// Call __next__
						if nextMethod, hasNext := iterInstance.Get("__next__"); hasNext {
							if nextFn, ok := nextMethod.(*Function); ok {
								nextEnv := NewEnvironment(nextFn.Env)
								nextEnv.Set("__args__", &List{Elements: []Value{}})
								nextEnv.Set("self", iterInstance)

								value, err := nextFn.Body(nextEnv)
								if err != nil {
									// StopIteration signal
									break
								}

								i.env.Set(stmt.Iterator.Value, value)
								_, err = i.evalBlockStatement(stmt.Body, i.env)
								if err != nil {
									if _, isBreak := err.(*BreakSignal); isBreak {
										break
									}
									if _, isContinue := err.(*ContinueSignal); isContinue {
										continue
									}
									return nil, err
								}
							}
						} else {
							break
						}
					}
				}
			}
		} else {
			return nil, &RuntimeError{Message: fmt.Sprintf("%T is not iterable", iterable)}
		}

	default:
		return nil, &RuntimeError{Message: fmt.Sprintf("%T is not iterable", iterable)}
	}

	return &Nil{}, nil
}

func (i *Interpreter) evalBlockStatement(block *ast.BlockStatement, env *Environment) (Value, error) {
	var result Value = &Nil{}

	for _, stmt := range block.Statements {
		val, err := i.evalStatement(stmt)
		if err != nil {
			return nil, err
		}

		// Return statement'ı handle et
		if _, isReturn := stmt.(*ast.ReturnStatement); isReturn {
			return val, nil
		}

		result = val
	}

	return result, nil
}

func (i *Interpreter) evalExpression(expr ast.Expression) (Value, error) {
	if expr == nil {
		return &Nil{}, nil
	}

	switch e := expr.(type) {
	case *ast.IntegerLiteral:
		return &Integer{Value: e.Value}, nil

	case *ast.FloatLiteral:
		return &Float{Value: e.Value}, nil

	case *ast.StringLiteral:
		return &String{Value: e.Value}, nil

	case *ast.BooleanLiteral:
		return &Boolean{Value: e.Value}, nil

	case *ast.Identifier:
		val, ok := i.env.Get(e.Value)
		if !ok {
			return nil, &RuntimeError{Message: fmt.Sprintf("undefined: %s", e.Value)}
		}
		return val, nil

	case *ast.ListLiteral:
		elements := make([]Value, len(e.Elements))
		for idx, elem := range e.Elements {
			val, err := i.evalExpression(elem)
			if err != nil {
				return nil, err
			}
			elements[idx] = val
		}
		return &List{Elements: elements}, nil

	case *ast.DictLiteral:
		pairs := make(map[string]Value)
		for keyExpr, valueExpr := range e.Pairs {
			keyVal, err := i.evalExpression(keyExpr)
			if err != nil {
				return nil, err
			}
			valueVal, err := i.evalExpression(valueExpr)
			if err != nil {
				return nil, err
			}
			pairs[keyVal.String()] = valueVal
		}
		return &Dict{Pairs: pairs}, nil

	case *ast.PrefixExpression:
		return i.evalPrefixExpression(e)

	case *ast.InfixExpression:
		return i.evalInfixExpression(e)

	case *ast.CallExpression:
		return i.evalCallExpression(e)

	case *ast.IndexExpression:
		return i.evalIndexExpression(e)

	case *ast.AwaitExpression:
		return i.evalAwaitExpression(e)

	case *ast.YieldExpression:
		return i.evalYieldExpression(e)

	case *ast.MemberExpression:
		return i.evalMemberExpression(e)

	case *ast.MatchExpression:
		return i.evalMatchExpression(e)

	default:
		return nil, &RuntimeError{Message: fmt.Sprintf("unknown expression type: %T", expr)}
	}
}

func (i *Interpreter) evalPrefixExpression(expr *ast.PrefixExpression) (Value, error) {
	right, err := i.evalExpression(expr.Right)
	if err != nil {
		return nil, err
	}

	switch expr.Operator {
	case "!":
		return &Boolean{Value: !right.IsTruthy()}, nil
	case "-":
		if intVal, ok := right.(*Integer); ok {
			return &Integer{Value: -intVal.Value}, nil
		}
		if floatVal, ok := right.(*Float); ok {
			return &Float{Value: -floatVal.Value}, nil
		}
		return nil, &RuntimeError{Message: "operator - can only be applied to numbers"}
	case "+":
		return right, nil
	default:
		return nil, &RuntimeError{Message: fmt.Sprintf("unknown prefix operator: %s", expr.Operator)}
	}
}

func (i *Interpreter) evalInfixExpression(expr *ast.InfixExpression) (Value, error) {
	// Assignment operatörleri
	if expr.Operator == "=" || expr.Operator == "+=" || expr.Operator == "-=" ||
		expr.Operator == "*=" || expr.Operator == "/=" || expr.Operator == "%=" {

		// Handle member assignment (obj.field = value)
		if memberExpr, ok := expr.Left.(*ast.MemberExpression); ok {
			object, err := i.evalExpression(memberExpr.Object)
			if err != nil {
				return nil, err
			}

			if instance, ok := object.(*Instance); ok {
				rightVal, err := i.evalExpression(expr.Right)
				if err != nil {
					return nil, err
				}

				memberName := memberExpr.Member.Value

				if expr.Operator != "=" {
					// Compound assignment
					leftVal, found := instance.Get(memberName)
					if !found {
						return nil, &RuntimeError{Message: fmt.Sprintf("undefined property: %s", memberName)}
					}

					var result Value
					switch expr.Operator {
					case "+=":
						result, err = i.evalBinaryOp(leftVal, rightVal, "+")
					case "-=":
						result, err = i.evalBinaryOp(leftVal, rightVal, "-")
					case "*=":
						result, err = i.evalBinaryOp(leftVal, rightVal, "*")
					case "/=":
						result, err = i.evalBinaryOp(leftVal, rightVal, "/")
					case "%=":
						result, err = i.evalBinaryOp(leftVal, rightVal, "%")
					}

					if err != nil {
						return nil, err
					}
					rightVal = result
				}

				instance.Set(memberName, rightVal)
				return rightVal, nil
			}

			return nil, &RuntimeError{Message: "can only assign to instance members"}
		}

		// Handle identifier assignment
		if ident, ok := expr.Left.(*ast.Identifier); ok {
			rightVal, err := i.evalExpression(expr.Right)
			if err != nil {
				return nil, err
			}

			if expr.Operator != "=" {
				// Compound assignment
				leftVal, ok := i.env.Get(ident.Value)
				if !ok {
					return nil, &RuntimeError{Message: fmt.Sprintf("undefined: %s", ident.Value)}
				}

				// Operasyonu yap
				var result Value
				switch expr.Operator {
				case "+=":
					result, err = i.evalBinaryOp(leftVal, rightVal, "+")
				case "-=":
					result, err = i.evalBinaryOp(leftVal, rightVal, "-")
				case "*=":
					result, err = i.evalBinaryOp(leftVal, rightVal, "*")
				case "/=":
					result, err = i.evalBinaryOp(leftVal, rightVal, "/")
				case "%=":
					result, err = i.evalBinaryOp(leftVal, rightVal, "%")
				}

				if err != nil {
					return nil, err
				}
				rightVal = result
			}

			err = i.env.Update(ident.Value, rightVal)
			if err != nil {
				// Eğer değişken yoksa, yeni bir tane oluştur (let gibi)
				i.env.Set(ident.Value, rightVal)
			}
			return rightVal, nil
		}

		return nil, &RuntimeError{Message: "left side of assignment must be an identifier"}
	}

	// Diğer operatörler
	left, err := i.evalExpression(expr.Left)
	if err != nil {
		return nil, err
	}

	right, err := i.evalExpression(expr.Right)
	if err != nil {
		return nil, err
	}

	return i.evalBinaryOp(left, right, expr.Operator)
}

func (i *Interpreter) evalBinaryOp(left, right Value, op string) (Value, error) {
	// String concatenation with type coercion
	if op == "+" {
		strL, okL := left.(*String)
		strR, okR := right.(*String)

		// String + String
		if okL && okR {
			return &String{Value: strL.Value + strR.Value}, nil
		}

		// String + Integer/Float/Boolean (auto-convert to string)
		if okL {
			var rightStr string
			switch v := right.(type) {
			case *Integer:
				rightStr = fmt.Sprintf("%d", v.Value)
			case *Float:
				rightStr = fmt.Sprintf("%g", v.Value)
			case *Boolean:
				rightStr = fmt.Sprintf("%v", v.Value)
			case *String:
				rightStr = v.Value
			default:
				rightStr = fmt.Sprintf("%v", v)
			}
			return &String{Value: strL.Value + rightStr}, nil
		}

		// Integer/Float/Boolean + String (auto-convert to string)
		if okR {
			var leftStr string
			switch v := left.(type) {
			case *Integer:
				leftStr = fmt.Sprintf("%d", v.Value)
			case *Float:
				leftStr = fmt.Sprintf("%g", v.Value)
			case *Boolean:
				leftStr = fmt.Sprintf("%v", v.Value)
			case *String:
				leftStr = v.Value
			default:
				leftStr = fmt.Sprintf("%v", v)
			}
			return &String{Value: leftStr + strR.Value}, nil
		}
	}

	// Arithmetic operations
	if intL, okL := left.(*Integer); okL {
		if intR, okR := right.(*Integer); okR {
			switch op {
			case "+":
				return &Integer{Value: intL.Value + intR.Value}, nil
			case "-":
				return &Integer{Value: intL.Value - intR.Value}, nil
			case "*":
				return &Integer{Value: intL.Value * intR.Value}, nil
			case "/":
				if intR.Value == 0 {
					return nil, &RuntimeError{Message: "division by zero"}
				}
				return &Integer{Value: intL.Value / intR.Value}, nil
			case "%":
				return &Integer{Value: intL.Value % intR.Value}, nil
			case "==":
				return &Boolean{Value: intL.Value == intR.Value}, nil
			case "!=":
				return &Boolean{Value: intL.Value != intR.Value}, nil
			case "<":
				return &Boolean{Value: intL.Value < intR.Value}, nil
			case "<=":
				return &Boolean{Value: intL.Value <= intR.Value}, nil
			case ">":
				return &Boolean{Value: intL.Value > intR.Value}, nil
			case ">=":
				return &Boolean{Value: intL.Value >= intR.Value}, nil
			}
		}
	}

	// Float operations
	if floatL, okL := left.(*Float); okL {
		if floatR, okR := right.(*Float); okR {
			switch op {
			case "+":
				return &Float{Value: floatL.Value + floatR.Value}, nil
			case "-":
				return &Float{Value: floatL.Value - floatR.Value}, nil
			case "*":
				return &Float{Value: floatL.Value * floatR.Value}, nil
			case "/":
				return &Float{Value: floatL.Value / floatR.Value}, nil
			}
		}
	}

	// String concatenation
	if strL, okL := left.(*String); okL {
		if strR, okR := right.(*String); okR {
			if op == "+" {
				return &String{Value: strL.Value + strR.Value}, nil
			}
		}
	}

	// Boolean operations
	if op == "&&" {
		return &Boolean{Value: left.IsTruthy() && right.IsTruthy()}, nil
	}
	if op == "||" {
		return &Boolean{Value: left.IsTruthy() || right.IsTruthy()}, nil
	}

	return nil, &RuntimeError{Message: fmt.Sprintf("unsupported operation: %T %s %T", left, op, right)}
}

func (i *Interpreter) evalCallExpression(expr *ast.CallExpression) (Value, error) {
	function, err := i.evalExpression(expr.Function)
	if err != nil {
		return nil, err
	}

	// Check if it's a class (instantiation)
	if class, ok := function.(*Class); ok {
		// Create new instance
		instance := &Instance{
			Class:  class,
			Fields: make(map[string]Value),
		}

		// Evaluate arguments once
		args := make([]Value, len(expr.Arguments))
		for idx, arg := range expr.Arguments {
			val, err := i.evalExpression(arg)
			if err != nil {
				return nil, err
			}
			args[idx] = val
		}

		// Call constructor hierarchy (superclass first, then subclass)
		constructorChain := []*Function{}
		currentClass := class
		for currentClass != nil {
			if constructor, hasConstructor := currentClass.Methods["init"]; hasConstructor {
				constructorChain = append([]*Function{constructor}, constructorChain...)
			}
			currentClass = currentClass.SuperClass
		}

		// Call constructors in order (base class first)
		for _, constructor := range constructorChain {
			callEnv := NewEnvironment(constructor.Env)
			callEnv.Set("__args__", &List{Elements: args})
			callEnv.Set("self", instance)

			_, err := constructor.Body(callEnv)
			if err != nil {
				return nil, err
			}
		}

		return instance, nil
	}

	// Regular function call
	fn, ok := function.(*Function)
	if !ok {
		return nil, &RuntimeError{Message: fmt.Sprintf("%s is not a function or class", expr.Function.String())}
	}

	// Argümanları değerlendir
	args := make([]Value, len(expr.Arguments))
	for idx, arg := range expr.Arguments {
		val, err := i.evalExpression(arg)
		if err != nil {
			return nil, err
		}
		args[idx] = val
	}

	// If async function, return a Promise
	if fn.Async {
		return NewPromise(func() (Value, error) {
			callEnv := NewEnvironment(fn.Env)
			callEnv.Set("__args__", &List{Elements: args})
			return fn.Body(callEnv)
		}), nil
	}

	// Synchronous function: execute immediately
	callEnv := NewEnvironment(fn.Env)
	callEnv.Set("__args__", &List{Elements: args})

	return fn.Body(callEnv)
}

func (i *Interpreter) evalIndexExpression(expr *ast.IndexExpression) (Value, error) {
	left, err := i.evalExpression(expr.Left)
	if err != nil {
		return nil, err
	}

	index, err := i.evalExpression(expr.Index)
	if err != nil {
		return nil, err
	}

	if list, ok := left.(*List); ok {
		if intIdx, ok := index.(*Integer); ok {
			if intIdx.Value < 0 || intIdx.Value >= int64(len(list.Elements)) {
				return nil, &RuntimeError{Message: "list index out of range"}
			}
			return list.Elements[intIdx.Value], nil
		}
	}

	if dict, ok := left.(*Dict); ok {
		key := index.String()
		if val, ok := dict.Pairs[key]; ok {
			return val, nil
		}
		return &Nil{}, nil
	}

	return nil, &RuntimeError{Message: "index operation not supported"}
}

// evalAwaitExpression waits for a promise to resolve and returns its value
func (i *Interpreter) evalAwaitExpression(expr *ast.AwaitExpression) (Value, error) {
	value, err := i.evalExpression(expr.Expression)
	if err != nil {
		return nil, err
	}

	// If it's a promise, await it
	if promise, ok := value.(*Promise); ok {
		return promise.Await()
	}

	// If not a promise, return as-is (await on non-promise is identity)
	return value, nil
}

// evalYieldExpression yields a value (for coroutines/generators)
func (i *Interpreter) evalYieldExpression(expr *ast.YieldExpression) (Value, error) {
	if expr.Value == nil {
		return &Nil{}, nil
	}

	value, err := i.evalExpression(expr.Value)
	if err != nil {
		return nil, err
	}

	// For now, yield just returns the value
	// In a full implementation, this would suspend the coroutine
	// and pass control back to the caller
	return value, nil
}

// evalClassStatement evaluates a class definition
func (i *Interpreter) evalClassStatement(stmt *ast.ClassStatement) error {
	className := stmt.Name.Value

	// Handle superclass
	var superClass *Class
	if stmt.SuperClass != nil {
		superVal, ok := i.env.Get(stmt.SuperClass.Value)
		if !ok {
			return &RuntimeError{Message: fmt.Sprintf("undefined superclass: %s", stmt.SuperClass.Value)}
		}
		var okClass bool
		superClass, okClass = superVal.(*Class)
		if !okClass {
			return &RuntimeError{Message: fmt.Sprintf("%s is not a class", stmt.SuperClass.Value)}
		}
	}

	// Create class environment
	classEnv := NewEnvironment(i.env)

	// Create class
	class := &Class{
		Name:       className,
		SuperClass: superClass,
		Methods:    make(map[string]*Function),
		Env:        classEnv,
	}

	// Process class body (methods)
	oldEnv := i.env
	i.env = classEnv
	defer func() { i.env = oldEnv }()

	for _, member := range stmt.Body {
		switch m := member.(type) {
		case *ast.FunctionStatement:
			// Add method to class
			funcName := m.Name.Value
			params := make([]string, len(m.Parameters))
			for idx, param := range m.Parameters {
				params[idx] = param.Name.Value
			}

			capturedStmt := m
			capturedEnv := classEnv

			method := &Function{
				Name:       funcName,
				Parameters: params,
				Async:      m.Async,
				Env:        capturedEnv,
				Body: func(callEnv *Environment) (Value, error) {
					// Execute method body
					fnEnv := NewEnvironment(capturedEnv)

					// Get arguments
					args, _ := callEnv.Get("__args__")
					if argList, ok := args.(*List); ok {
						for idx, paramName := range params {
							if idx < len(argList.Elements) {
								fnEnv.Set(paramName, argList.Elements[idx])
							} else {
								fnEnv.Set(paramName, &Nil{})
							}
						}
					}

					// Copy self and super if they exist
					if self, ok := callEnv.Get("self"); ok {
						fnEnv.Set("self", self)
					}
					if super, ok := callEnv.Get("super"); ok {
						fnEnv.Set("super", super)
					}

					oldMethodEnv := i.env
					i.env = fnEnv
					defer func() { i.env = oldMethodEnv }()

					return i.evalBlockStatement(capturedStmt.Body, fnEnv)
				},
			}

			class.Methods[funcName] = method

		default:
			// For now, only methods are supported in class body
			return &RuntimeError{Message: fmt.Sprintf("unsupported class member: %T", member)}
		}
	}

	// Register class in environment
	i.env = oldEnv
	i.env.Set(className, class)

	return nil
}

// evalMemberExpression evaluates member access (obj.member)
func (i *Interpreter) evalMemberExpression(expr *ast.MemberExpression) (Value, error) {
	object, err := i.evalExpression(expr.Object)
	if err != nil {
		return nil, err
	}

	memberName := expr.Member.Value

	// Handle instance member access
	if instance, ok := object.(*Instance); ok {
		if value, found := instance.Get(memberName); found {
			return value, nil
		}
		return nil, &RuntimeError{Message: fmt.Sprintf("undefined property: %s", memberName)}
	}

	// Handle dict member access (for backwards compatibility)
	if dict, ok := object.(*Dict); ok {
		if value, found := dict.Pairs[memberName]; found {
			return value, nil
		}
		return &Nil{}, nil
	}

	// Handle String methods
	if str, ok := object.(*String); ok {
		methodName := "str_" + memberName
		if method, found := i.env.Get(methodName); found {
			if fn, ok := method.(*Function); ok {
				// Return bound method
				return &Function{
					Name:       memberName,
					Parameters: fn.Parameters,
					Env:        fn.Env,
					Body: func(callEnv *Environment) (Value, error) {
						// Inject string as first argument
						args, _ := callEnv.Get("__args__")
						if argList, ok := args.(*List); ok {
							newArgs := append([]Value{str}, argList.Elements...)
							callEnv.Set("__args__", &List{Elements: newArgs})
						} else {
							callEnv.Set("__args__", &List{Elements: []Value{str}})
						}
						return fn.Body(callEnv)
					},
				}, nil
			}
		}
	}

	// Handle List methods
	if list, ok := object.(*List); ok {
		methodName := "list_" + memberName
		if method, found := i.env.Get(methodName); found {
			if fn, ok := method.(*Function); ok {
				// Return bound method
				return &Function{
					Name:       memberName,
					Parameters: fn.Parameters,
					Env:        fn.Env,
					Body: func(callEnv *Environment) (Value, error) {
						// Inject list as first argument
						args, _ := callEnv.Get("__args__")
						if argList, ok := args.(*List); ok {
							newArgs := append([]Value{list}, argList.Elements...)
							callEnv.Set("__args__", &List{Elements: newArgs})
						} else {
							callEnv.Set("__args__", &List{Elements: []Value{list}})
						}
						return fn.Body(callEnv)
					},
				}, nil
			}
		}
	}

	// Handle Dict methods
	if dict, ok := object.(*Dict); ok {
		methodName := "dict_" + memberName
		if method, found := i.env.Get(methodName); found {
			if fn, ok := method.(*Function); ok {
				// Return bound method
				return &Function{
					Name:       memberName,
					Parameters: fn.Parameters,
					Env:        fn.Env,
					Body: func(callEnv *Environment) (Value, error) {
						// Inject dict as first argument
						args, _ := callEnv.Get("__args__")
						if argList, ok := args.(*List); ok {
							newArgs := append([]Value{dict}, argList.Elements...)
							callEnv.Set("__args__", &List{Elements: newArgs})
						} else {
							callEnv.Set("__args__", &List{Elements: []Value{dict}})
						}
						return fn.Body(callEnv)
					},
				}, nil
			}
		}
	}

	return nil, &RuntimeError{Message: fmt.Sprintf("cannot access member of %T", object)}
}

// addTypeConversionFunctions adds int(), float(), bool() converters
func addTypeConversionFunctions(env *Environment) {
	// int(x, [base])
	env.Set("int", &Function{
		Name: "int",
		Body: func(callEnv *Environment) (Value, error) {
			args, _ := callEnv.Get("__args__")
			if list, ok := args.(*List); ok && len(list.Elements) > 0 {
				arg := list.Elements[0]

				switch v := arg.(type) {
				case *Integer:
					return v, nil
				case *Float:
					return &Integer{Value: int64(v.Value)}, nil
				case *String:
					base := 10
					if len(list.Elements) >= 2 {
						if baseVal, ok := list.Elements[1].(*Integer); ok {
							base = int(baseVal.Value)
						}
					}
					val, err := strconv.ParseInt(v.Value, base, 64)
					if err != nil {
						return &Nil{}, &RuntimeError{Message: fmt.Sprintf("invalid literal for int(): %s", v.Value)}
					}
					return &Integer{Value: val}, nil
				case *Boolean:
					if v.Value {
						return &Integer{Value: 1}, nil
					}
					return &Integer{Value: 0}, nil
				default:
					return &Integer{Value: 0}, nil
				}
			}
			return &Integer{Value: 0}, nil
		},
	})

	// float(x)
	env.Set("float", &Function{
		Name: "float",
		Body: func(callEnv *Environment) (Value, error) {
			args, _ := callEnv.Get("__args__")
			if list, ok := args.(*List); ok && len(list.Elements) > 0 {
				arg := list.Elements[0]

				switch v := arg.(type) {
				case *Float:
					return v, nil
				case *Integer:
					return &Float{Value: float64(v.Value)}, nil
				case *String:
					val, err := strconv.ParseFloat(v.Value, 64)
					if err != nil {
						return &Nil{}, &RuntimeError{Message: fmt.Sprintf("invalid literal for float(): %s", v.Value)}
					}
					return &Float{Value: val}, nil
				case *Boolean:
					if v.Value {
						return &Float{Value: 1.0}, nil
					}
					return &Float{Value: 0.0}, nil
				default:
					return &Float{Value: 0.0}, nil
				}
			}
			return &Float{Value: 0.0}, nil
		},
	})

	// bool(x)
	env.Set("bool", &Function{
		Name: "bool",
		Body: func(callEnv *Environment) (Value, error) {
			args, _ := callEnv.Get("__args__")
			if list, ok := args.(*List); ok && len(list.Elements) > 0 {
				arg := list.Elements[0]
				return &Boolean{Value: arg.IsTruthy()}, nil
			}
			return &Boolean{Value: false}, nil
		},
	})

	// str(x)
	env.Set("str", &Function{
		Name: "str",
		Body: func(callEnv *Environment) (Value, error) {
			args, _ := callEnv.Get("__args__")
			if list, ok := args.(*List); ok && len(list.Elements) > 0 {
				arg := list.Elements[0]
				return &String{Value: arg.String()}, nil
			}
			return &String{Value: ""}, nil
		},
	})

	// list(iterable)
	env.Set("list", &Function{
		Name: "list",
		Body: func(callEnv *Environment) (Value, error) {
			args, _ := callEnv.Get("__args__")
			if list, ok := args.(*List); ok && len(list.Elements) > 0 {
				arg := list.Elements[0]

				switch v := arg.(type) {
				case *List:
					// Copy list
					copied := make([]Value, len(v.Elements))
					copy(copied, v.Elements)
					return &List{Elements: copied}, nil
				case *String:
					// String to list of characters
					chars := make([]Value, len(v.Value))
					for i, ch := range v.Value {
						chars[i] = &String{Value: string(ch)}
					}
					return &List{Elements: chars}, nil
				case *Dict:
					// Dict to list of keys
					keys := make([]Value, 0, len(v.Pairs))
					for k := range v.Pairs {
						keys = append(keys, &String{Value: k})
					}
					return &List{Elements: keys}, nil
				default:
					return &List{Elements: []Value{arg}}, nil
				}
			}
			return &List{Elements: []Value{}}, nil
		},
	})

	// dict(pairs or **kwargs)
	env.Set("dict", &Function{
		Name: "dict",
		Body: func(callEnv *Environment) (Value, error) {
			args, _ := callEnv.Get("__args__")
			if list, ok := args.(*List); ok && len(list.Elements) > 0 {
				// If first arg is already a dict, copy it
				if d, ok := list.Elements[0].(*Dict); ok {
					pairs := make(map[string]Value)
					for k, v := range d.Pairs {
						pairs[k] = v
					}
					return &Dict{Pairs: pairs}, nil
				}

				// If it's a list of [key, value] pairs
				if pairsList, ok := list.Elements[0].(*List); ok {
					pairs := make(map[string]Value)
					for _, pairVal := range pairsList.Elements {
						if pair, ok := pairVal.(*List); ok && len(pair.Elements) >= 2 {
							key := pair.Elements[0].String()
							val := pair.Elements[1]
							pairs[key] = val
						}
					}
					return &Dict{Pairs: pairs}, nil
				}
			}
			return &Dict{Pairs: make(map[string]Value)}, nil
		},
	})
}

// addNumericFunctions adds math utility functions
func addNumericFunctions(env *Environment) {
	// abs(x)
	env.Set("abs", &Function{
		Name: "abs",
		Body: func(callEnv *Environment) (Value, error) {
			args, _ := callEnv.Get("__args__")
			if list, ok := args.(*List); ok && len(list.Elements) > 0 {
				arg := list.Elements[0]

				switch v := arg.(type) {
				case *Integer:
					if v.Value < 0 {
						return &Integer{Value: -v.Value}, nil
					}
					return v, nil
				case *Float:
					return &Float{Value: math.Abs(v.Value)}, nil
				default:
					return &Nil{}, &RuntimeError{Message: "abs() requires numeric argument"}
				}
			}
			return &Nil{}, &RuntimeError{Message: "abs() requires argument"}
		},
	})

	// min(a, b, ...)
	env.Set("min", &Function{
		Name: "min",
		Body: func(callEnv *Environment) (Value, error) {
			args, _ := callEnv.Get("__args__")
			if list, ok := args.(*List); ok && len(list.Elements) > 0 {
				minVal := list.Elements[0]

				for _, arg := range list.Elements[1:] {
					switch m := minVal.(type) {
					case *Integer:
						if v, ok := arg.(*Integer); ok {
							if v.Value < m.Value {
								minVal = v
							}
						}
					case *Float:
						if v, ok := arg.(*Float); ok {
							if v.Value < m.Value {
								minVal = v
							}
						}
					}
				}
				return minVal, nil
			}
			return &Nil{}, &RuntimeError{Message: "min() requires at least one argument"}
		},
	})

	// max(a, b, ...)
	env.Set("max", &Function{
		Name: "max",
		Body: func(callEnv *Environment) (Value, error) {
			args, _ := callEnv.Get("__args__")
			if list, ok := args.(*List); ok && len(list.Elements) > 0 {
				maxVal := list.Elements[0]

				for _, arg := range list.Elements[1:] {
					switch m := maxVal.(type) {
					case *Integer:
						if v, ok := arg.(*Integer); ok {
							if v.Value > m.Value {
								maxVal = v
							}
						}
					case *Float:
						if v, ok := arg.(*Float); ok {
							if v.Value > m.Value {
								maxVal = v
							}
						}
					}
				}
				return maxVal, nil
			}
			return &Nil{}, &RuntimeError{Message: "max() requires at least one argument"}
		},
	})

	// round(x, [digits])
	env.Set("round", &Function{
		Name: "round",
		Body: func(callEnv *Environment) (Value, error) {
			args, _ := callEnv.Get("__args__")
			if list, ok := args.(*List); ok && len(list.Elements) > 0 {
				if f, ok := list.Elements[0].(*Float); ok {
					digits := 0
					if len(list.Elements) >= 2 {
						if d, ok := list.Elements[1].(*Integer); ok {
							digits = int(d.Value)
						}
					}

					multiplier := math.Pow(10, float64(digits))
					rounded := math.Round(f.Value*multiplier) / multiplier
					return &Float{Value: rounded}, nil
				}
				if i, ok := list.Elements[0].(*Integer); ok {
					return i, nil
				}
			}
			return &Nil{}, &RuntimeError{Message: "round() requires numeric argument"}
		},
	})

	// pow(x, y)
	env.Set("pow", &Function{
		Name: "pow",
		Body: func(callEnv *Environment) (Value, error) {
			args, _ := callEnv.Get("__args__")
			if list, ok := args.(*List); ok && len(list.Elements) >= 2 {
				var x, y float64

				switch v := list.Elements[0].(type) {
				case *Integer:
					x = float64(v.Value)
				case *Float:
					x = v.Value
				default:
					return &Nil{}, &RuntimeError{Message: "pow() requires numeric arguments"}
				}

				switch v := list.Elements[1].(type) {
				case *Integer:
					y = float64(v.Value)
				case *Float:
					y = v.Value
				default:
					return &Nil{}, &RuntimeError{Message: "pow() requires numeric arguments"}
				}

				result := math.Pow(x, y)

				// Return int if both inputs were int and result is whole number
				if _, ok1 := list.Elements[0].(*Integer); ok1 {
					if _, ok2 := list.Elements[1].(*Integer); ok2 {
						if result == math.Floor(result) {
							return &Integer{Value: int64(result)}, nil
						}
					}
				}

				return &Float{Value: result}, nil
			}
			return &Nil{}, &RuntimeError{Message: "pow() requires two arguments"}
		},
	})

	// sqrt(x)
	env.Set("sqrt", &Function{
		Name: "sqrt",
		Body: func(callEnv *Environment) (Value, error) {
			args, _ := callEnv.Get("__args__")
			if list, ok := args.(*List); ok && len(list.Elements) > 0 {
				var x float64

				switch v := list.Elements[0].(type) {
				case *Integer:
					x = float64(v.Value)
				case *Float:
					x = v.Value
				default:
					return &Nil{}, &RuntimeError{Message: "sqrt() requires numeric argument"}
				}

				if x < 0 {
					return &Nil{}, &RuntimeError{Message: "sqrt() of negative number"}
				}

				return &Float{Value: math.Sqrt(x)}, nil
			}
			return &Nil{}, &RuntimeError{Message: "sqrt() requires argument"}
		},
	})

	// floor(x)
	env.Set("floor", &Function{
		Name: "floor",
		Body: func(callEnv *Environment) (Value, error) {
			args, _ := callEnv.Get("__args__")
			if list, ok := args.(*List); ok && len(list.Elements) > 0 {
				if f, ok := list.Elements[0].(*Float); ok {
					return &Integer{Value: int64(math.Floor(f.Value))}, nil
				}
				if i, ok := list.Elements[0].(*Integer); ok {
					return i, nil
				}
			}
			return &Nil{}, &RuntimeError{Message: "floor() requires numeric argument"}
		},
	})

	// ceil(x)
	env.Set("ceil", &Function{
		Name: "ceil",
		Body: func(callEnv *Environment) (Value, error) {
			args, _ := callEnv.Get("__args__")
			if list, ok := args.(*List); ok && len(list.Elements) > 0 {
				if f, ok := list.Elements[0].(*Float); ok {
					return &Integer{Value: int64(math.Ceil(f.Value))}, nil
				}
				if i, ok := list.Elements[0].(*Integer); ok {
					return i, nil
				}
			}
			return &Nil{}, &RuntimeError{Message: "ceil() requires numeric argument"}
		},
	})

	// sum(iterable)
	env.Set("sum", &Function{
		Name: "sum",
		Body: func(callEnv *Environment) (Value, error) {
			args, _ := callEnv.Get("__args__")
			if list, ok := args.(*List); ok && len(list.Elements) > 0 {
				if items, ok := list.Elements[0].(*List); ok {
					var intSum int64
					var floatSum float64
					hasFloat := false

					for _, item := range items.Elements {
						switch v := item.(type) {
						case *Integer:
							if hasFloat {
								floatSum += float64(v.Value)
							} else {
								intSum += v.Value
							}
						case *Float:
							if !hasFloat {
								floatSum = float64(intSum)
								hasFloat = true
							}
							floatSum += v.Value
						}
					}

					if hasFloat {
						return &Float{Value: floatSum}, nil
					}
					return &Integer{Value: intSum}, nil
				}
			}
			return &Integer{Value: 0}, nil
		},
	})
}

// addUtilityFunctions adds utility functions (input, type, isinstance)
func addUtilityFunctions(env *Environment) {
	// input(prompt)
	env.Set("input", &Function{
		Name: "input",
		Body: func(callEnv *Environment) (Value, error) {
			args, _ := callEnv.Get("__args__")

			// Print prompt if provided
			if list, ok := args.(*List); ok && len(list.Elements) > 0 {
				if prompt, ok := list.Elements[0].(*String); ok {
					fmt.Print(prompt.Value)
				}
			}

			// Read from stdin
			reader := bufio.NewReader(os.Stdin)
			line, err := reader.ReadString('\n')
			if err != nil {
				return &String{Value: ""}, nil
			}

			// Remove trailing newline
			line = strings.TrimSuffix(line, "\n")
			line = strings.TrimSuffix(line, "\r")

			return &String{Value: line}, nil
		},
	})

	// type(x)
	env.Set("type", &Function{
		Name: "type",
		Body: func(callEnv *Environment) (Value, error) {
			args, _ := callEnv.Get("__args__")
			if list, ok := args.(*List); ok && len(list.Elements) > 0 {
				arg := list.Elements[0]

				switch arg.(type) {
				case *Integer:
					return &String{Value: "int"}, nil
				case *Float:
					return &String{Value: "float"}, nil
				case *String:
					return &String{Value: "string"}, nil
				case *Boolean:
					return &String{Value: "bool"}, nil
				case *List:
					return &String{Value: "list"}, nil
				case *Dict:
					return &String{Value: "dict"}, nil
				case *Function:
					return &String{Value: "function"}, nil
				case *Class:
					return &String{Value: "class"}, nil
				case *Instance:
					return &String{Value: "instance"}, nil
				case *Promise:
					return &String{Value: "promise"}, nil
				case *Nil:
					return &String{Value: "nil"}, nil
				default:
					return &String{Value: "unknown"}, nil
				}
			}
			return &String{Value: "nil"}, nil
		},
	})

	// isinstance(obj, type)
	env.Set("isinstance", &Function{
		Name: "isinstance",
		Body: func(callEnv *Environment) (Value, error) {
			args, _ := callEnv.Get("__args__")
			if list, ok := args.(*List); ok && len(list.Elements) >= 2 {
				obj := list.Elements[0]
				if typeName, ok := list.Elements[1].(*String); ok {
					objType := ""

					switch obj.(type) {
					case *Integer:
						objType = "int"
					case *Float:
						objType = "float"
					case *String:
						objType = "string"
					case *Boolean:
						objType = "bool"
					case *List:
						objType = "list"
					case *Dict:
						objType = "dict"
					case *Function:
						objType = "function"
					case *Class:
						objType = "class"
					case *Instance:
						objType = "instance"
					}

					return &Boolean{Value: objType == typeName.Value}, nil
				}
			}
			return &Boolean{Value: false}, nil
		},
	})
}

// addFunctionalFunctions adds functional programming functions
func addFunctionalFunctions(env *Environment) {
	// map(function, iterable)
	env.Set("map", &Function{
		Name: "map",
		Body: func(callEnv *Environment) (Value, error) {
			args, _ := callEnv.Get("__args__")
			if list, ok := args.(*List); ok && len(list.Elements) >= 2 {
				if fn, ok := list.Elements[0].(*Function); ok {
					if items, ok := list.Elements[1].(*List); ok {
						results := make([]Value, len(items.Elements))

						for i, item := range items.Elements {
							// Call function with item
							fnEnv := NewEnvironment(fn.Env)
							fnEnv.Set("__args__", &List{Elements: []Value{item}})

							result, err := fn.Body(fnEnv)
							if err != nil {
								return &Nil{}, err
							}
							results[i] = result
						}

						return &List{Elements: results}, nil
					}
				}
			}
			return &Nil{}, &RuntimeError{Message: "map() requires function and iterable"}
		},
	})

	// filter(function, iterable)
	env.Set("filter", &Function{
		Name: "filter",
		Body: func(callEnv *Environment) (Value, error) {
			args, _ := callEnv.Get("__args__")
			if list, ok := args.(*List); ok && len(list.Elements) >= 2 {
				if fn, ok := list.Elements[0].(*Function); ok {
					if items, ok := list.Elements[1].(*List); ok {
						results := make([]Value, 0, len(items.Elements))

						for _, item := range items.Elements {
							// Call function with item
							fnEnv := NewEnvironment(fn.Env)
							fnEnv.Set("__args__", &List{Elements: []Value{item}})

							result, err := fn.Body(fnEnv)
							if err != nil {
								return &Nil{}, err
							}

							// Include item if result is truthy
							if result.IsTruthy() {
								results = append(results, item)
							}
						}

						return &List{Elements: results}, nil
					}
				}
			}
			return &Nil{}, &RuntimeError{Message: "filter() requires function and iterable"}
		},
	})

	// any(iterable)
	env.Set("any", &Function{
		Name: "any",
		Body: func(callEnv *Environment) (Value, error) {
			args, _ := callEnv.Get("__args__")
			if list, ok := args.(*List); ok && len(list.Elements) > 0 {
				if items, ok := list.Elements[0].(*List); ok {
					for _, item := range items.Elements {
						if item.IsTruthy() {
							return &Boolean{Value: true}, nil
						}
					}
					return &Boolean{Value: false}, nil
				}
			}
			return &Boolean{Value: false}, nil
		},
	})

	// all(iterable)
	env.Set("all", &Function{
		Name: "all",
		Body: func(callEnv *Environment) (Value, error) {
			args, _ := callEnv.Get("__args__")
			if list, ok := args.(*List); ok && len(list.Elements) > 0 {
				if items, ok := list.Elements[0].(*List); ok {
					for _, item := range items.Elements {
						if !item.IsTruthy() {
							return &Boolean{Value: false}, nil
						}
					}
					return &Boolean{Value: true}, nil
				}
			}
			return &Boolean{Value: true}, nil
		},
	})
}

// addStringMethods adds all Python-style string methods
func addStringMethods(env *Environment) {
	// str.upper()
	env.Set("str_upper", &Function{
		Name: "str_upper",
		Body: func(callEnv *Environment) (Value, error) {
			args, _ := callEnv.Get("__args__")
			if list, ok := args.(*List); ok && len(list.Elements) > 0 {
				if s, ok := list.Elements[0].(*String); ok {
					return &String{Value: strings.ToUpper(s.Value)}, nil
				}
			}
			return &Nil{}, &RuntimeError{Message: "upper() requires string argument"}
		},
	})

	// str.lower()
	env.Set("str_lower", &Function{
		Name: "str_lower",
		Body: func(callEnv *Environment) (Value, error) {
			args, _ := callEnv.Get("__args__")
			if list, ok := args.(*List); ok && len(list.Elements) > 0 {
				if s, ok := list.Elements[0].(*String); ok {
					return &String{Value: strings.ToLower(s.Value)}, nil
				}
			}
			return &Nil{}, &RuntimeError{Message: "lower() requires string argument"}
		},
	})

	// str.capitalize()
	env.Set("str_capitalize", &Function{
		Name: "str_capitalize",
		Body: func(callEnv *Environment) (Value, error) {
			args, _ := callEnv.Get("__args__")
			if list, ok := args.(*List); ok && len(list.Elements) > 0 {
				if s, ok := list.Elements[0].(*String); ok {
					if len(s.Value) == 0 {
						return s, nil
					}
					return &String{Value: strings.ToUpper(s.Value[:1]) + strings.ToLower(s.Value[1:])}, nil
				}
			}
			return &Nil{}, &RuntimeError{Message: "capitalize() requires string argument"}
		},
	})

	// str.split(sep)
	env.Set("str_split", &Function{
		Name: "str_split",
		Body: func(callEnv *Environment) (Value, error) {
			args, _ := callEnv.Get("__args__")
			if list, ok := args.(*List); ok && len(list.Elements) >= 1 {
				if s, ok := list.Elements[0].(*String); ok {
					sep := " "
					if len(list.Elements) >= 2 {
						if sepStr, ok := list.Elements[1].(*String); ok {
							sep = sepStr.Value
						}
					}
					parts := strings.Split(s.Value, sep)
					elements := make([]Value, len(parts))
					for i, part := range parts {
						elements[i] = &String{Value: part}
					}
					return &List{Elements: elements}, nil
				}
			}
			return &Nil{}, &RuntimeError{Message: "split() requires string argument"}
		},
	})

	// str.join(list)
	env.Set("str_join", &Function{
		Name: "str_join",
		Body: func(callEnv *Environment) (Value, error) {
			args, _ := callEnv.Get("__args__")
			if list, ok := args.(*List); ok && len(list.Elements) >= 2 {
				if sep, ok := list.Elements[0].(*String); ok {
					if items, ok := list.Elements[1].(*List); ok {
						parts := make([]string, len(items.Elements))
						for i, item := range items.Elements {
							parts[i] = item.String()
						}
						return &String{Value: strings.Join(parts, sep.Value)}, nil
					}
				}
			}
			return &Nil{}, &RuntimeError{Message: "join() requires string separator and list"}
		},
	})

	// str.replace(old, new)
	env.Set("str_replace", &Function{
		Name: "str_replace",
		Body: func(callEnv *Environment) (Value, error) {
			args, _ := callEnv.Get("__args__")
			if list, ok := args.(*List); ok && len(list.Elements) >= 3 {
				if s, ok := list.Elements[0].(*String); ok {
					if old, ok := list.Elements[1].(*String); ok {
						if new, ok := list.Elements[2].(*String); ok {
							return &String{Value: strings.ReplaceAll(s.Value, old.Value, new.Value)}, nil
						}
					}
				}
			}
			return &Nil{}, &RuntimeError{Message: "replace() requires string, old, new"}
		},
	})

	// str.strip()
	env.Set("str_strip", &Function{
		Name: "str_strip",
		Body: func(callEnv *Environment) (Value, error) {
			args, _ := callEnv.Get("__args__")
			if list, ok := args.(*List); ok && len(list.Elements) > 0 {
				if s, ok := list.Elements[0].(*String); ok {
					return &String{Value: strings.TrimSpace(s.Value)}, nil
				}
			}
			return &Nil{}, &RuntimeError{Message: "strip() requires string argument"}
		},
	})

	// str.startswith(prefix)
	env.Set("str_startswith", &Function{
		Name: "str_startswith",
		Body: func(callEnv *Environment) (Value, error) {
			args, _ := callEnv.Get("__args__")
			if list, ok := args.(*List); ok && len(list.Elements) >= 2 {
				if s, ok := list.Elements[0].(*String); ok {
					if prefix, ok := list.Elements[1].(*String); ok {
						return &Boolean{Value: strings.HasPrefix(s.Value, prefix.Value)}, nil
					}
				}
			}
			return &Nil{}, &RuntimeError{Message: "startswith() requires string and prefix"}
		},
	})

	// str.endswith(suffix)
	env.Set("str_endswith", &Function{
		Name: "str_endswith",
		Body: func(callEnv *Environment) (Value, error) {
			args, _ := callEnv.Get("__args__")
			if list, ok := args.(*List); ok && len(list.Elements) >= 2 {
				if s, ok := list.Elements[0].(*String); ok {
					if suffix, ok := list.Elements[1].(*String); ok {
						return &Boolean{Value: strings.HasSuffix(s.Value, suffix.Value)}, nil
					}
				}
			}
			return &Nil{}, &RuntimeError{Message: "endswith() requires string and suffix"}
		},
	})

	// str.find(sub)
	env.Set("str_find", &Function{
		Name: "str_find",
		Body: func(callEnv *Environment) (Value, error) {
			args, _ := callEnv.Get("__args__")
			if list, ok := args.(*List); ok && len(list.Elements) >= 2 {
				if s, ok := list.Elements[0].(*String); ok {
					if sub, ok := list.Elements[1].(*String); ok {
						idx := strings.Index(s.Value, sub.Value)
						return &Integer{Value: int64(idx)}, nil
					}
				}
			}
			return &Integer{Value: -1}, nil
		},
	})

	// str.count(sub)
	env.Set("str_count", &Function{
		Name: "str_count",
		Body: func(callEnv *Environment) (Value, error) {
			args, _ := callEnv.Get("__args__")
			if list, ok := args.(*List); ok && len(list.Elements) >= 2 {
				if s, ok := list.Elements[0].(*String); ok {
					if sub, ok := list.Elements[1].(*String); ok {
						count := strings.Count(s.Value, sub.Value)
						return &Integer{Value: int64(count)}, nil
					}
				}
			}
			return &Integer{Value: 0}, nil
		},
	})
}

// addListMethods adds all Python-style list methods
func addListMethods(env *Environment) {
	// list.append(item)
	env.Set("list_append", &Function{
		Name: "list_append",
		Body: func(callEnv *Environment) (Value, error) {
			args, _ := callEnv.Get("__args__")
			if list, ok := args.(*List); ok && len(list.Elements) >= 2 {
				if targetList, ok := list.Elements[0].(*List); ok {
					item := list.Elements[1]
					targetList.Elements = append(targetList.Elements, item)
					return &Nil{}, nil
				}
			}
			return &Nil{}, &RuntimeError{Message: "append() requires list and item"}
		},
	})

	// list.pop([index])
	env.Set("list_pop", &Function{
		Name: "list_pop",
		Body: func(callEnv *Environment) (Value, error) {
			args, _ := callEnv.Get("__args__")
			if list, ok := args.(*List); ok && len(list.Elements) >= 1 {
				if targetList, ok := list.Elements[0].(*List); ok {
					if len(targetList.Elements) == 0 {
						return &Nil{}, &RuntimeError{Message: "pop from empty list"}
					}

					idx := len(targetList.Elements) - 1
					if len(list.Elements) >= 2 {
						if idxVal, ok := list.Elements[1].(*Integer); ok {
							idx = int(idxVal.Value)
						}
					}

					if idx < 0 || idx >= len(targetList.Elements) {
						return &Nil{}, &RuntimeError{Message: "pop index out of range"}
					}

					item := targetList.Elements[idx]
					targetList.Elements = append(targetList.Elements[:idx], targetList.Elements[idx+1:]...)
					return item, nil
				}
			}
			return &Nil{}, &RuntimeError{Message: "pop() requires list"}
		},
	})

	// list.insert(index, item)
	env.Set("list_insert", &Function{
		Name: "list_insert",
		Body: func(callEnv *Environment) (Value, error) {
			args, _ := callEnv.Get("__args__")
			if list, ok := args.(*List); ok && len(list.Elements) >= 3 {
				if targetList, ok := list.Elements[0].(*List); ok {
					if idx, ok := list.Elements[1].(*Integer); ok {
						item := list.Elements[2]
						i := int(idx.Value)

						if i < 0 {
							i = 0
						}
						if i > len(targetList.Elements) {
							i = len(targetList.Elements)
						}

						targetList.Elements = append(targetList.Elements[:i], append([]Value{item}, targetList.Elements[i:]...)...)
						return &Nil{}, nil
					}
				}
			}
			return &Nil{}, &RuntimeError{Message: "insert() requires list, index, item"}
		},
	})

	// list.remove(item)
	env.Set("list_remove", &Function{
		Name: "list_remove",
		Body: func(callEnv *Environment) (Value, error) {
			args, _ := callEnv.Get("__args__")
			if list, ok := args.(*List); ok && len(list.Elements) >= 2 {
				if targetList, ok := list.Elements[0].(*List); ok {
					item := list.Elements[1]

					for i, elem := range targetList.Elements {
						if elem.String() == item.String() {
							targetList.Elements = append(targetList.Elements[:i], targetList.Elements[i+1:]...)
							return &Nil{}, nil
						}
					}
					return &Nil{}, &RuntimeError{Message: "item not in list"}
				}
			}
			return &Nil{}, &RuntimeError{Message: "remove() requires list and item"}
		},
	})

	// list.clear()
	env.Set("list_clear", &Function{
		Name: "list_clear",
		Body: func(callEnv *Environment) (Value, error) {
			args, _ := callEnv.Get("__args__")
			if list, ok := args.(*List); ok && len(list.Elements) >= 1 {
				if targetList, ok := list.Elements[0].(*List); ok {
					targetList.Elements = []Value{}
					return &Nil{}, nil
				}
			}
			return &Nil{}, &RuntimeError{Message: "clear() requires list"}
		},
	})

	// list.index(item)
	env.Set("list_index", &Function{
		Name: "list_index",
		Body: func(callEnv *Environment) (Value, error) {
			args, _ := callEnv.Get("__args__")
			if list, ok := args.(*List); ok && len(list.Elements) >= 2 {
				if targetList, ok := list.Elements[0].(*List); ok {
					item := list.Elements[1]

					for i, elem := range targetList.Elements {
						if elem.String() == item.String() {
							return &Integer{Value: int64(i)}, nil
						}
					}
					return &Integer{Value: -1}, nil
				}
			}
			return &Integer{Value: -1}, nil
		},
	})

	// list.count(item)
	env.Set("list_count", &Function{
		Name: "list_count",
		Body: func(callEnv *Environment) (Value, error) {
			args, _ := callEnv.Get("__args__")
			if list, ok := args.(*List); ok && len(list.Elements) >= 2 {
				if targetList, ok := list.Elements[0].(*List); ok {
					item := list.Elements[1]
					count := int64(0)

					for _, elem := range targetList.Elements {
						if elem.String() == item.String() {
							count++
						}
					}
					return &Integer{Value: count}, nil
				}
			}
			return &Integer{Value: 0}, nil
		},
	})

	// list.reverse()
	env.Set("list_reverse", &Function{
		Name: "list_reverse",
		Body: func(callEnv *Environment) (Value, error) {
			args, _ := callEnv.Get("__args__")
			if list, ok := args.(*List); ok && len(list.Elements) >= 1 {
				if targetList, ok := list.Elements[0].(*List); ok {
					n := len(targetList.Elements)
					for i := 0; i < n/2; i++ {
						targetList.Elements[i], targetList.Elements[n-1-i] = targetList.Elements[n-1-i], targetList.Elements[i]
					}
					return &Nil{}, nil
				}
			}
			return &Nil{}, &RuntimeError{Message: "reverse() requires list"}
		},
	})

	// list.copy()
	env.Set("list_copy", &Function{
		Name: "list_copy",
		Body: func(callEnv *Environment) (Value, error) {
			args, _ := callEnv.Get("__args__")
			if list, ok := args.(*List); ok && len(list.Elements) >= 1 {
				if targetList, ok := list.Elements[0].(*List); ok {
					copied := make([]Value, len(targetList.Elements))
					copy(copied, targetList.Elements)
					return &List{Elements: copied}, nil
				}
			}
			return &Nil{}, &RuntimeError{Message: "copy() requires list"}
		},
	})

	// list.extend(other)
	env.Set("list_extend", &Function{
		Name: "list_extend",
		Body: func(callEnv *Environment) (Value, error) {
			args, _ := callEnv.Get("__args__")
			if list, ok := args.(*List); ok && len(list.Elements) >= 2 {
				if targetList, ok := list.Elements[0].(*List); ok {
					if otherList, ok := list.Elements[1].(*List); ok {
						targetList.Elements = append(targetList.Elements, otherList.Elements...)
						return &Nil{}, nil
					}
				}
			}
			return &Nil{}, &RuntimeError{Message: "extend() requires two lists"}
		},
	})
}

// addDictMethods adds all Python-style dict methods
func addDictMethods(env *Environment) {
	// dict.keys()
	env.Set("dict_keys", &Function{
		Name: "dict_keys",
		Body: func(callEnv *Environment) (Value, error) {
			args, _ := callEnv.Get("__args__")
			if list, ok := args.(*List); ok && len(list.Elements) >= 1 {
				if d, ok := list.Elements[0].(*Dict); ok {
					keys := make([]Value, 0, len(d.Pairs))
					for k := range d.Pairs {
						keys = append(keys, &String{Value: k})
					}
					return &List{Elements: keys}, nil
				}
			}
			return &Nil{}, &RuntimeError{Message: "keys() requires dict"}
		},
	})

	// dict.values()
	env.Set("dict_values", &Function{
		Name: "dict_values",
		Body: func(callEnv *Environment) (Value, error) {
			args, _ := callEnv.Get("__args__")
			if list, ok := args.(*List); ok && len(list.Elements) >= 1 {
				if d, ok := list.Elements[0].(*Dict); ok {
					values := make([]Value, 0, len(d.Pairs))
					for _, v := range d.Pairs {
						values = append(values, v)
					}
					return &List{Elements: values}, nil
				}
			}
			return &Nil{}, &RuntimeError{Message: "values() requires dict"}
		},
	})

	// dict.get(key, default)
	env.Set("dict_get", &Function{
		Name: "dict_get",
		Body: func(callEnv *Environment) (Value, error) {
			args, _ := callEnv.Get("__args__")
			if list, ok := args.(*List); ok && len(list.Elements) >= 2 {
				if d, ok := list.Elements[0].(*Dict); ok {
					if key, ok := list.Elements[1].(*String); ok {
						if val, exists := d.Pairs[key.Value]; exists {
							return val, nil
						}
						// Return default if provided
						if len(list.Elements) >= 3 {
							return list.Elements[2], nil
						}
						return &Nil{}, nil
					}
				}
			}
			return &Nil{}, &RuntimeError{Message: "get() requires dict and key"}
		},
	})

	// dict.pop(key, [default])
	env.Set("dict_pop", &Function{
		Name: "dict_pop",
		Body: func(callEnv *Environment) (Value, error) {
			args, _ := callEnv.Get("__args__")
			if list, ok := args.(*List); ok && len(list.Elements) >= 2 {
				if d, ok := list.Elements[0].(*Dict); ok {
					if key, ok := list.Elements[1].(*String); ok {
						if val, exists := d.Pairs[key.Value]; exists {
							delete(d.Pairs, key.Value)
							return val, nil
						}
						// Return default if provided
						if len(list.Elements) >= 3 {
							return list.Elements[2], nil
						}
						return &Nil{}, &RuntimeError{Message: "key not found"}
					}
				}
			}
			return &Nil{}, &RuntimeError{Message: "pop() requires dict and key"}
		},
	})

	// dict.clear()
	env.Set("dict_clear", &Function{
		Name: "dict_clear",
		Body: func(callEnv *Environment) (Value, error) {
			args, _ := callEnv.Get("__args__")
			if list, ok := args.(*List); ok && len(list.Elements) >= 1 {
				if d, ok := list.Elements[0].(*Dict); ok {
					d.Pairs = make(map[string]Value)
					return &Nil{}, nil
				}
			}
			return &Nil{}, &RuntimeError{Message: "clear() requires dict"}
		},
	})

	// dict.update(other)
	env.Set("dict_update", &Function{
		Name: "dict_update",
		Body: func(callEnv *Environment) (Value, error) {
			args, _ := callEnv.Get("__args__")
			if list, ok := args.(*List); ok && len(list.Elements) >= 2 {
				if d, ok := list.Elements[0].(*Dict); ok {
					if other, ok := list.Elements[1].(*Dict); ok {
						for k, v := range other.Pairs {
							d.Pairs[k] = v
						}
						return &Nil{}, nil
					}
				}
			}
			return &Nil{}, &RuntimeError{Message: "update() requires two dicts"}
		},
	})
}

// evalImportStatement handles module imports
func (i *Interpreter) evalImportStatement(stmt *ast.ImportStatement) error {
	// Build module path
	modulePath := ""
	for idx, segment := range stmt.Path {
		if idx > 0 {
			modulePath += "/"
		}
		modulePath += segment
	}

	// Check if already loaded
	if moduleEnv, cached := i.moduleCache[modulePath]; cached {
		// Module already loaded, import symbols
		return i.importSymbolsFromModule(moduleEnv, stmt.Alias)
	}

	// Load module file
	moduleFilePath := i.resolveModulePath(modulePath)

	// Read and parse module
	content, err := os.ReadFile(moduleFilePath)
	if err != nil {
		return &RuntimeError{Message: fmt.Sprintf("cannot load module %s: %v", modulePath, err)}
	}

	// Parse module
	l := lexer.New(string(content), moduleFilePath)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		return &RuntimeError{Message: fmt.Sprintf("parse errors in module %s: %v", modulePath, p.Errors())}
	}

	// Create module environment
	moduleEnv := NewEnvironment(nil)
	oldEnv := i.env
	i.env = moduleEnv

	// Execute module
	for _, modStmt := range program.Statements {
		_, err := i.evalStatement(modStmt)
		if err != nil {
			i.env = oldEnv // Restore before returning
			return &RuntimeError{Message: fmt.Sprintf("error in module %s: %v", modulePath, err)}
		}
	}

	// Restore environment BEFORE importing symbols
	i.env = oldEnv

	// Cache module
	i.moduleCache[modulePath] = moduleEnv

	// Import symbols (now i.env is the original environment)
	return i.importSymbolsFromModule(moduleEnv, stmt.Alias)
}

// resolveModulePath resolves module path to file path
func (i *Interpreter) resolveModulePath(modulePath string) string {
	moduleFile := modulePath + ".sky"

	// Try relative to source file directory first
	if i.sourceFile != "" {
		sourceDir := ""
		lastSlash := -1
		for idx, ch := range i.sourceFile {
			if ch == '/' || ch == '\\' {
				lastSlash = idx
			}
		}
		if lastSlash >= 0 {
			sourceDir = i.sourceFile[:lastSlash]
		}

		if sourceDir != "" {
			relToSource := sourceDir + "/" + moduleFile
			if _, err := os.Stat(relToSource); err == nil {
				return relToSource
			}
		}
	}

	// Try relative to current directory
	relPath := i.currentDir + "/" + moduleFile
	if _, err := os.Stat(relPath); err == nil {
		return relPath
	}

	// Try as absolute path
	if _, err := os.Stat(moduleFile); err == nil {
		return moduleFile
	}

	// Default to relative path
	return relPath
}

// importSymbolsFromModule imports exported symbols from module
func (i *Interpreter) importSymbolsFromModule(moduleEnv *Environment, alias *ast.Identifier) error {
	symbols := moduleEnv.GetAll()

	if alias != nil {
		// Import with alias (import foo as f)
		// Create a namespace object
		namespace := &Dict{Pairs: make(map[string]Value)}
		for name, value := range symbols {
			// Only import public symbols (not starting with _)
			if len(name) > 0 && name[0] != '_' {
				namespace.Pairs[name] = value
			}
		}
		i.env.Set(alias.Value, namespace)
	} else {
		// Direct import (import foo)
		for name, value := range symbols {
			// Only import public symbols
			if len(name) > 0 && name[0] != '_' {
				i.env.Set(name, value)
			}
		}
	}
	return nil
}

// evalUnsafeStatement executes an unsafe block
// In unsafe blocks, certain safety checks are disabled
func (i *Interpreter) evalUnsafeStatement(stmt *ast.UnsafeStatement) (Value, error) {
	// For now, unsafe blocks just execute normally
	// In a production implementation, this would:
	// - Disable GC for the duration of the block
	// - Allow raw pointer operations
	// - Disable bounds checking
	// - Allow direct memory access

	return i.evalBlockStatement(stmt.Body, i.env)
}
