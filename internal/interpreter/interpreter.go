package interpreter

import (
	"bufio"
	"crypto/aes"
	"crypto/md5"
	"crypto/sha256"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

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

	// GLOBAL FUNCTIONS
	addGlobalFunctions(env)

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

			// If main is async, it returns a Promise - await it
			if fn.Async {
				promise := NewPromise(func() (Value, error) {
					return fn.Body(newEnv)
				})
				_, err := promise.Await()
				return err
			}

			// Synchronous main
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
	case *ast.AbstractClassStatement:
		return nil, i.evalAbstractClassStatement(s)
	case *ast.AbstractMethodStatement:
		return nil, i.evalAbstractMethodStatement(s)
	case *ast.StaticMethodStatement:
		return nil, i.evalStaticMethodStatement(s)
	case *ast.StaticPropertyStatement:
		return nil, i.evalStaticPropertyStatement(s)
	case *ast.ImportStatement:
		return nil, i.evalImportStatement(s)
	case *ast.UnsafeStatement:
		return i.evalUnsafeStatement(s)
	case *ast.EnumStatement:
		return i.evalEnumStatement(s)
	case *ast.BlockStatement:
		return i.evalBlockStatement(s, i.env)
	case *ast.TryStatement:
		return i.evalTryStatement(s)
	case *ast.ThrowStatement:
		return i.evalThrowStatement(s)
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
		val, err := i.evalExpression(stmt.ReturnValue)
		if err != nil {
			return nil, err
		}
		return nil, &ReturnSignal{Value: val}
	}
	return nil, &ReturnSignal{Value: &Nil{}}
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

			// Yeni environment oluştur
			// callEnv'i parent olarak kullan (parametreler ve self için)
			fnEnv := NewEnvironment(callEnv)

			// Copy 'self' from callEnv to fnEnv for direct access
			if self, ok := callEnv.Get("self"); ok {
				fnEnv.Set("self", self)
			}

			// Parametreleri bind et
			if argList, ok := args.(*List); ok {
				for idx, param := range capturedStmt.Parameters {
					paramName := param.Name.Value

					// Handle varargs parameter
					if param.Variadic {
						// Collect remaining arguments into a list
						varargs := &List{Elements: []Value{}}
						for i := idx; i < len(argList.Elements); i++ {
							varargs.Elements = append(varargs.Elements, argList.Elements[i])
						}
						fnEnv.Set(paramName, varargs)
						break // Varargs is always last
					}

					if idx < len(argList.Elements) {
						fnEnv.Set(paramName, argList.Elements[idx])
					} else if param.DefaultValue != nil {
						// Use default value
						defaultVal, err := i.evalExpression(param.DefaultValue)
						if err != nil {
							return nil, err
						}
						fnEnv.Set(paramName, defaultVal)
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

			// Handle ReturnSignal - extract value and return normally
			if retSignal, isReturn := err.(*ReturnSignal); isReturn {
				result = retSignal.Value
				err = nil
			}

			// Cache successful results
			if err == nil && result != nil {
				if argList, ok := args.(*List); ok {
					i.trampoline.SetCached(funcName, argList.Elements, result)
				}
			}

			return result, err
		},
	}

	// Apply decorators (in reverse order - innermost first)
	decoratedFn := fn
	for j := len(stmt.Decorators) - 1; j >= 0; j-- {
		decorator := stmt.Decorators[j]

		// Get decorator function
		decoratorVal, ok := i.env.Get(decorator.Name.Value)
		if !ok {
			return &RuntimeError{Message: fmt.Sprintf("undefined decorator: %s", decorator.Name.Value)}
		}

		decoratorFn, ok := decoratorVal.(*Function)
		if !ok {
			return &RuntimeError{Message: fmt.Sprintf("%s is not a decorator function", decorator.Name.Value)}
		}

		// Call decorator with function as argument
		callEnv := NewEnvironment(decoratorFn.Env)
		args := []Value{decoratedFn}

		// Add decorator arguments if any
		for _, argExpr := range decorator.Args {
			argVal, err := i.evalExpression(argExpr)
			if err != nil {
				return err
			}
			args = append(args, argVal)
		}

		callEnv.Set("__args__", &List{Elements: args})

		// Execute decorator
		result, err := decoratorFn.Body(callEnv)
		if err != nil {
			return err
		}

		// Decorator should return a function
		if resultFn, ok := result.(*Function); ok {
			decoratedFn = resultFn
		} else {
			return &RuntimeError{Message: fmt.Sprintf("decorator %s did not return a function", decorator.Name.Value)}
		}
	}

	// Handle coroutine/generator functions
	if stmt.Coop {
		// Wrap in generator factory
		generatorFactory := &Function{
			Name:       funcName,
			Parameters: params,
			Env:        capturedEnv,
			Body: func(callEnv *Environment) (Value, error) {
				// Create generator instance
				genEnv := NewEnvironment(capturedEnv)

				// Bind parameters
				args, _ := callEnv.Get("__args__")
				if argList, ok := args.(*List); ok {
					for idx, param := range capturedStmt.Parameters {
						paramName := param.Name.Value
						if idx < len(argList.Elements) {
							genEnv.Set(paramName, argList.Elements[idx])
						} else if param.DefaultValue != nil {
							defaultVal, err := i.evalExpression(param.DefaultValue)
							if err != nil {
								return nil, err
							}
							genEnv.Set(paramName, defaultVal)
						} else {
							genEnv.Set(paramName, &Nil{})
						}
					}
				}

				// Create generator by executing and collecting yields
				gen, err := NewGenerator(i, genEnv, capturedStmt.Body)
				if err != nil {
					return nil, err
				}

				// Add next() method to generator
				nextMethod := &Function{
					Name: "next",
					Body: func(callEnv *Environment) (Value, error) {
						return gen.Next()
					},
				}

				// Return dict with next method
				result := &Dict{Pairs: map[string]Value{
					"next": nextMethod,
				}}

				return result, nil
			},
		}
		i.env.Set(funcName, generatorFactory)
	} else {
		i.env.Set(funcName, decoratedFn)
	}

	return nil
}

func (i *Interpreter) evalIfStatement(stmt *ast.IfStatement) (Value, error) {
	condition, err := i.evalExpression(stmt.Condition)
	if err != nil {
		return nil, err
	}

	if condition.IsTruthy() {
		val, err := i.evalBlockStatement(stmt.Consequence, i.env)
		// Propagate signals
		if _, isReturn := err.(*ReturnSignal); isReturn {
			return nil, err
		}
		if _, isYield := err.(*YieldSignal); isYield {
			return nil, err
		}
		return val, err
	}

	// Elif dallarını kontrol et
	for _, elif := range stmt.Elif {
		cond, err := i.evalExpression(elif.Condition)
		if err != nil {
			return nil, err
		}
		if cond.IsTruthy() {
			val, err := i.evalBlockStatement(elif.Consequence, i.env)
			// Propagate signals
			if _, isReturn := err.(*ReturnSignal); isReturn {
				return nil, err
			}
			if _, isYield := err.(*YieldSignal); isYield {
				return nil, err
			}
			return val, err
		}
	}

	// Else dalı
	if stmt.Alternative != nil {
		val, err := i.evalBlockStatement(stmt.Alternative, i.env)
		// Propagate signals
		if _, isReturn := err.(*ReturnSignal); isReturn {
			return nil, err
		}
		if _, isYield := err.(*YieldSignal); isYield {
			return nil, err
		}
		return val, err
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
			if _, isReturn := err.(*ReturnSignal); isReturn {
				return nil, err // Propagate return signal
			}
			if _, isYield := err.(*YieldSignal); isYield {
				return nil, err // Propagate yield signal
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
				if _, isReturn := err.(*ReturnSignal); isReturn {
					return nil, err // Propagate return signal
				}
				if _, isYield := err.(*YieldSignal); isYield {
					return nil, err // Propagate yield signal
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
				if _, isReturn := err.(*ReturnSignal); isReturn {
					return nil, err // Propagate return signal
				}
				if _, isYield := err.(*YieldSignal); isYield {
					return nil, err // Propagate yield signal
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
				if _, isReturn := err.(*ReturnSignal); isReturn {
					return nil, err // Propagate return signal
				}
				if _, isYield := err.(*YieldSignal); isYield {
					return nil, err // Propagate yield signal
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

		// Check for control flow signals
		if err != nil {
			// Propagate ReturnSignal up
			if _, isReturn := err.(*ReturnSignal); isReturn {
				return nil, err
			}
			// Propagate YieldSignal up (for generators)
			if _, isYield := err.(*YieldSignal); isYield {
				return nil, err
			}
			// Other errors
			return nil, err
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
		// Special handling for 'self'
		if e.Value == "self" {
			// Search in environment chain for 'self'
			env := i.env
			for env != nil {
				if val, ok := env.Get("self"); ok {
					return val, nil
				}
				env = env.parent
			}
			return nil, &RuntimeError{Message: "undefined: self"}
		}

		// Special handling for 'super'
		if e.Value == "super" {
			// Search in environment chain for 'super'
			env := i.env
			for env != nil {
				if val, ok := env.Get("super"); ok {
					return val, nil
				}
				env = env.parent
			}
			return nil, &RuntimeError{Message: "undefined: super"}
		}

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

	case *ast.ArrowExpression:
		return i.evalArrowExpression(e)
	case *ast.MatchExpression:
		return i.evalMatchExpression(e)

	case *ast.LambdaExpression:
		return i.evalLambdaExpression(e)

	default:
		return nil, &RuntimeError{Message: fmt.Sprintf("unknown expression type: %T", expr)}
	}
}

// evalLambdaExpression evaluates a lambda expression
func (i *Interpreter) evalLambdaExpression(expr *ast.LambdaExpression) (Value, error) {
	// Convert lambda to function
	params := make([]string, len(expr.Parameters))
	for idx, param := range expr.Parameters {
		params[idx] = param.Name.Value
	}

	return &Function{
		Name:       "lambda",
		Parameters: params,
		Env:        i.env,
		Body: func(callEnv *Environment) (Value, error) {
			// Create function environment
			fnEnv := NewEnvironment(i.env)

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

			// Execute lambda body
			oldEnv := i.env
			i.env = fnEnv
			defer func() { i.env = oldEnv }()

			result, err := i.evalBlockStatement(expr.Body, fnEnv)
			if err != nil {
				// Handle return signal
				if returnSignal, isReturn := err.(*ReturnSignal); isReturn {
					return returnSignal.Value, nil
				}
				return nil, err
			}
			return result, nil
		},
	}, nil
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
			// Special handling for 'self' identifier
			var object Value
			var err error
			if ident, ok := memberExpr.Object.(*ast.Identifier); ok && ident.Value == "self" {
				// Get 'self' directly from environment
				if selfVal, found := i.env.Get("self"); found {
					object = selfVal
				} else {
					return nil, &RuntimeError{Message: "undefined: self"}
				}
			} else {
				object, err = i.evalExpression(memberExpr.Object)
				if err != nil {
					return nil, err
				}
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

	// String operations
	if strL, okL := left.(*String); okL {
		if strR, okR := right.(*String); okR {
			switch op {
			case "+":
				return &String{Value: strL.Value + strR.Value}, nil
			case "==":
				return &Boolean{Value: strL.Value == strR.Value}, nil
			case "!=":
				return &Boolean{Value: strL.Value != strR.Value}, nil
			}
		}
	}

	// Nil comparison operations
	if _, okL := left.(*Nil); okL {
		if _, okR := right.(*Nil); okR {
			switch op {
			case "==":
				return &Boolean{Value: true}, nil
			case "!=":
				return &Boolean{Value: false}, nil
			}
		}
		// left is nil, right is not
		switch op {
		case "==":
			return &Boolean{Value: false}, nil
		case "!=":
			return &Boolean{Value: true}, nil
		}
	}
	if _, okR := right.(*Nil); okR {
		// right is nil, left is not
		switch op {
		case "==":
			return &Boolean{Value: false}, nil
		case "!=":
			return &Boolean{Value: true}, nil
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

	// Check if it's an instance method call (obj.method())
	if memberExpr, ok := expr.Function.(*ast.MemberExpression); ok {
		instance, err := i.evalExpression(memberExpr.Object)
		if err != nil {
			return nil, err
		}

		if inst, ok := instance.(*Instance); ok {
			methodName := memberExpr.Member.Value

			// Get method from instance (checks class and superclasses)
			if methodVal, found := inst.Get(methodName); found {
				// Check if it's a function
				if method, ok := methodVal.(*Function); ok {
					// Evaluate arguments
					args := make([]Value, len(expr.Arguments))
					for idx, arg := range expr.Arguments {
						val, err := i.evalExpression(arg)
						if err != nil {
							return nil, err
						}
						args[idx] = val
					}

					// Create call environment
					callEnv := NewEnvironment(method.Env)
					callEnv.Set("__args__", &List{Elements: args})

					// Call method (self is already bound in Instance.Get())
					return method.Body(callEnv)
				}
			}

			return nil, &RuntimeError{Message: fmt.Sprintf("undefined method: %s", methodName)}
		}

		// Handle super.method() calls
		if superClass, ok := instance.(*Class); ok {
			methodName := memberExpr.Member.Value

			// Get method from superclass
			if method, found := superClass.Methods[methodName]; found {
				// Evaluate arguments
				args := make([]Value, len(expr.Arguments))
				for idx, arg := range expr.Arguments {
					val, err := i.evalExpression(arg)
					if err != nil {
						return nil, err
					}
					args[idx] = val
				}

				// Get current self from environment
				selfVal, found := i.env.Get("self")
				if !found {
					return nil, &RuntimeError{Message: "undefined: self"}
				}

				// Create call environment with self bound
				callEnv := NewEnvironment(method.Env)
				callEnv.Set("__args__", &List{Elements: args})
				callEnv.Set("self", selfVal)

				// Set super to parent class if exists
				if len(superClass.SuperClasses) > 0 {
					callEnv.Set("super", superClass.SuperClasses[0])
				}

				// Call method
				return method.Body(callEnv)
			}

			return nil, &RuntimeError{Message: fmt.Sprintf("undefined method: %s", methodName)}
		}
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

		// Call constructor hierarchy (multiple inheritance)
		constructorChain := []*Function{}

		// Collect constructors from all superclasses
		for _, superClass := range class.SuperClasses {
			if constructor, hasConstructor := superClass.Methods["init"]; hasConstructor {
				constructorChain = append(constructorChain, constructor)
			}
		}

		// Add current class constructor if exists
		if constructor, hasConstructor := class.Methods["init"]; hasConstructor {
			constructorChain = append(constructorChain, constructor)
		}

		// Call constructors in order (superclasses first, then current class)
		for _, constructor := range constructorChain {
			callEnv := NewEnvironment(constructor.Env)
			callEnv.Set("__args__", &List{Elements: args})
			callEnv.Set("self", instance)

			// Set super to parent class if exists
			if len(class.SuperClasses) > 0 {
				callEnv.Set("super", class.SuperClasses[0]) // Use first superclass
			}

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
		return nil, &YieldSignal{Value: &Nil{}}
	}

	value, err := i.evalExpression(expr.Value)
	if err != nil {
		return nil, err
	}

	// Return YieldSignal to pause execution and return value
	return nil, &YieldSignal{Value: value}
}

// evalClassStatement evaluates a class definition
func (i *Interpreter) evalClassStatement(stmt *ast.ClassStatement) error {
	className := stmt.Name.Value

	// Handle multiple inheritance
	var superClasses []*Class
	for _, superClassIdent := range stmt.SuperClasses {
		superVal, ok := i.env.Get(superClassIdent.Value)
		if !ok {
			return &RuntimeError{Message: fmt.Sprintf("undefined superclass: %s", superClassIdent.Value)}
		}
		// Check if it's a regular class or abstract class
		if superClass, okClass := superVal.(*Class); okClass {
			superClasses = append(superClasses, superClass)
		} else if abstractClass, okAbstract := superVal.(*AbstractClass); okAbstract {
			// Convert abstract class to regular class for inheritance
			convertedClass := &Class{
				Name:         abstractClass.Name,
				SuperClasses: abstractClass.SuperClasses,
				Methods:      abstractClass.Methods,
				Env:          abstractClass.Env,
			}
			superClasses = append(superClasses, convertedClass)
		} else {
			return &RuntimeError{Message: fmt.Sprintf("%s is not a class", superClassIdent.Value)}
		}
	}

	// Create class environment
	classEnv := NewEnvironment(i.env)

	// Create class with multiple inheritance
	class := &Class{
		Name:         className,
		SuperClasses: superClasses, // Multiple inheritance
		Methods:      make(map[string]*Function),
		Env:          classEnv,
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

					result, err := i.evalBlockStatement(capturedStmt.Body, fnEnv)
					if err != nil {
						// Handle return signal
						if returnSignal, isReturn := err.(*ReturnSignal); isReturn {
							return returnSignal.Value, nil
						}
						return nil, err
					}
					return result, nil
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

// evalArrowExpression evaluates arrow expressions (pattern => result)
func (i *Interpreter) evalArrowExpression(expr *ast.ArrowExpression) (Value, error) {
	// For now, just return the right side (simple implementation)
	return i.evalExpression(expr.Right)
}

// evalMemberExpression evaluates member access (obj.member)
func (i *Interpreter) evalMemberExpression(expr *ast.MemberExpression) (Value, error) {
	// Special handling for 'self' and 'super' identifiers
	var object Value
	var err error
	if ident, ok := expr.Object.(*ast.Identifier); ok {
		if ident.Value == "self" {
			// Get 'self' directly from environment
			if selfVal, found := i.env.Get("self"); found {
				object = selfVal
			} else {
				return nil, &RuntimeError{Message: "undefined: self"}
			}
		} else if ident.Value == "super" {
			// Get 'super' directly from environment
			if superVal, found := i.env.Get("super"); found {
				object = superVal
			} else {
				return nil, &RuntimeError{Message: "undefined: super"}
			}
		} else {
			object, err = i.evalExpression(expr.Object)
			if err != nil {
				return nil, err
			}
		}
	} else {
		object, err = i.evalExpression(expr.Object)
		if err != nil {
			return nil, err
		}
	}

	memberName := expr.Member.Value

	// Handle instance member access
	if instance, ok := object.(*Instance); ok {
		if value, found := instance.Get(memberName); found {
			return value, nil
		}
		return nil, &RuntimeError{Message: fmt.Sprintf("undefined property: %s", memberName)}
	}

	// Handle class member access (for super.method())
	if class, ok := object.(*Class); ok {
		if method, found := class.Methods[memberName]; found {
			return method, nil
		}
		return nil, &RuntimeError{Message: fmt.Sprintf("undefined method: %s", memberName)}
	}

	// Handle dict member access (for backwards compatibility)
	if dict, ok := object.(*Dict); ok {
		if value, found := dict.Pairs[memberName]; found {
			return value, nil
		}

		// Check for dict methods (get, set, has_key, etc.)
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

	// Handle Class method access (e.g., TestClass.new)
	if class, ok := object.(*Class); ok {
		if method, found := class.Methods[memberName]; found {
			return method, nil
		}
		// Check in class environment
		if value, found := class.Env.Get(memberName); found {
			return value, nil
		}
		return nil, &RuntimeError{Message: fmt.Sprintf("undefined method: %s", memberName)}
	}

	// Handle AbstractClass method access
	if abstractClass, ok := object.(*AbstractClass); ok {
		if method, found := abstractClass.Methods[memberName]; found {
			return method, nil
		}
		// Check in abstract class environment
		if value, found := abstractClass.Env.Get(memberName); found {
			return value, nil
		}
		return nil, &RuntimeError{Message: fmt.Sprintf("undefined method: %s", memberName)}
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

// addGlobalFunctions adds global utility functions
func addGlobalFunctions(env *Environment) {
	// join(separator, list) - global function
	env.Set("join", &Function{
		Name: "join",
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

	// null constant
	env.Set("null", &Nil{})

	// nil constant (alias for null)
	env.Set("nil", &Nil{})

	// time_now() - returns current time in milliseconds
	env.Set("time_now", &Function{
		Name: "time_now",
		Body: func(callEnv *Environment) (Value, error) {
			return &Integer{Value: int64(time.Now().UnixNano() / 1000000)}, nil
		},
	})

	// FS Module Functions
	env.Set("fs_read_text", &Function{
		Name: "fs_read_text",
		Body: func(callEnv *Environment) (Value, error) {
			args, _ := callEnv.Get("__args__")
			if list, ok := args.(*List); ok && len(list.Elements) >= 1 {
				if filename, ok := list.Elements[0].(*String); ok {
					content, err := os.ReadFile(filename.Value)
					if err != nil {
						return &Nil{}, &RuntimeError{Message: fmt.Sprintf("fs_read_text error: %v", err)}
					}
					return &String{Value: string(content)}, nil
				}
			}
			return &Nil{}, &RuntimeError{Message: "fs_read_text() requires filename string"}
		},
	})

	env.Set("fs_write_text", &Function{
		Name: "fs_write_text",
		Body: func(callEnv *Environment) (Value, error) {
			args, _ := callEnv.Get("__args__")
			if list, ok := args.(*List); ok && len(list.Elements) >= 2 {
				if filename, ok := list.Elements[0].(*String); ok {
					if content, ok := list.Elements[1].(*String); ok {
						err := os.WriteFile(filename.Value, []byte(content.Value), 0644)
						if err != nil {
							return &Nil{}, &RuntimeError{Message: fmt.Sprintf("fs_write_text error: %v", err)}
						}
						return &Boolean{Value: true}, nil
					}
				}
			}
			return &Nil{}, &RuntimeError{Message: "fs_write_text() requires filename and content strings"}
		},
	})

	env.Set("fs_exists", &Function{
		Name: "fs_exists",
		Body: func(callEnv *Environment) (Value, error) {
			args, _ := callEnv.Get("__args__")
			if list, ok := args.(*List); ok && len(list.Elements) >= 1 {
				if filename, ok := list.Elements[0].(*String); ok {
					_, err := os.Stat(filename.Value)
					return &Boolean{Value: err == nil}, nil
				}
			}
			return &Nil{}, &RuntimeError{Message: "fs_exists() requires filename string"}
		},
	})

	env.Set("fs_mkdir", &Function{
		Name: "fs_mkdir",
		Body: func(callEnv *Environment) (Value, error) {
			args, _ := callEnv.Get("__args__")
			if list, ok := args.(*List); ok && len(list.Elements) >= 1 {
				if dirname, ok := list.Elements[0].(*String); ok {
					err := os.MkdirAll(dirname.Value, 0755)
					if err != nil {
						return &Nil{}, &RuntimeError{Message: fmt.Sprintf("fs_mkdir error: %v", err)}
					}
					return &Boolean{Value: true}, nil
				}
			}
			return &Nil{}, &RuntimeError{Message: "fs_mkdir() requires directory name string"}
		},
	})

	env.Set("fs_list_dir", &Function{
		Name: "fs_list_dir",
		Body: func(callEnv *Environment) (Value, error) {
			args, _ := callEnv.Get("__args__")
			if list, ok := args.(*List); ok && len(list.Elements) >= 1 {
				if dirname, ok := list.Elements[0].(*String); ok {
					entries, err := os.ReadDir(dirname.Value)
					if err != nil {
						return &Nil{}, &RuntimeError{Message: fmt.Sprintf("fs_list_dir error: %v", err)}
					}

					files := make([]Value, len(entries))
					for i, entry := range entries {
						files[i] = &String{Value: entry.Name()}
					}
					return &List{Elements: files}, nil
				}
			}
			return &Nil{}, &RuntimeError{Message: "fs_list_dir() requires directory name string"}
		},
	})

	// OS Module Functions
	env.Set("os_platform", &Function{
		Name: "os_platform",
		Body: func(callEnv *Environment) (Value, error) {
			return &String{Value: runtime.GOOS}, nil
		},
	})

	env.Set("os_getcwd", &Function{
		Name: "os_getcwd",
		Body: func(callEnv *Environment) (Value, error) {
			dir, err := os.Getwd()
			if err != nil {
				return &Nil{}, &RuntimeError{Message: fmt.Sprintf("os_getcwd error: %v", err)}
			}
			return &String{Value: dir}, nil
		},
	})

	env.Set("os_getenv", &Function{
		Name: "os_getenv",
		Body: func(callEnv *Environment) (Value, error) {
			args, _ := callEnv.Get("__args__")
			if list, ok := args.(*List); ok && len(list.Elements) >= 1 {
				if key, ok := list.Elements[0].(*String); ok {
					value := os.Getenv(key.Value)
					return &String{Value: value}, nil
				}
			}
			return &Nil{}, &RuntimeError{Message: "os_getenv() requires environment variable name string"}
		},
	})

	env.Set("os_setenv", &Function{
		Name: "os_setenv",
		Body: func(callEnv *Environment) (Value, error) {
			args, _ := callEnv.Get("__args__")
			if list, ok := args.(*List); ok && len(list.Elements) >= 2 {
				if key, ok := list.Elements[0].(*String); ok {
					if value, ok := list.Elements[1].(*String); ok {
						err := os.Setenv(key.Value, value.Value)
						if err != nil {
							return &Nil{}, &RuntimeError{Message: fmt.Sprintf("os_setenv error: %v", err)}
						}
						return &Boolean{Value: true}, nil
					}
				}
			}
			return &Nil{}, &RuntimeError{Message: "os_setenv() requires key and value strings"}
		},
	})

	// Time Module Functions
	env.Set("time_format", &Function{
		Name: "time_format",
		Body: func(callEnv *Environment) (Value, error) {
			args, _ := callEnv.Get("__args__")
			if list, ok := args.(*List); ok && len(list.Elements) >= 2 {
				if timestamp, ok := list.Elements[0].(*Integer); ok {
					if format, ok := list.Elements[1].(*String); ok {
						t := time.Unix(0, timestamp.Value*1000000) // Convert milliseconds to nanoseconds
						formatted := t.Format(format.Value)
						return &String{Value: formatted}, nil
					}
				}
			}
			return &Nil{}, &RuntimeError{Message: "time_format() requires timestamp (int) and format string"}
		},
	})

	env.Set("time_parse", &Function{
		Name: "time_parse",
		Body: func(callEnv *Environment) (Value, error) {
			args, _ := callEnv.Get("__args__")
			if list, ok := args.(*List); ok && len(list.Elements) >= 2 {
				if timeStr, ok := list.Elements[0].(*String); ok {
					if format, ok := list.Elements[1].(*String); ok {
						t, err := time.Parse(format.Value, timeStr.Value)
						if err != nil {
							return &Nil{}, &RuntimeError{Message: fmt.Sprintf("time_parse error: %v", err)}
						}
						return &Integer{Value: t.UnixNano() / 1000000}, nil // Convert to milliseconds
					}
				}
			}
			return &Nil{}, &RuntimeError{Message: "time_parse() requires time string and format string"}
		},
	})

	env.Set("time_add", &Function{
		Name: "time_add",
		Body: func(callEnv *Environment) (Value, error) {
			args, _ := callEnv.Get("__args__")
			if list, ok := args.(*List); ok && len(list.Elements) >= 3 {
				if timestamp, ok := list.Elements[0].(*Integer); ok {
					if duration, ok := list.Elements[1].(*Integer); ok {
						if unit, ok := list.Elements[2].(*String); ok {
							t := time.Unix(0, timestamp.Value*1000000)
							var d time.Duration
							switch unit.Value {
							case "ms":
								d = time.Duration(duration.Value) * time.Millisecond
							case "s":
								d = time.Duration(duration.Value) * time.Second
							case "m":
								d = time.Duration(duration.Value) * time.Minute
							case "h":
								d = time.Duration(duration.Value) * time.Hour
							case "d":
								d = time.Duration(duration.Value) * 24 * time.Hour
							default:
								return &Nil{}, &RuntimeError{Message: "time_add() unit must be ms, s, m, h, or d"}
							}
							newTime := t.Add(d)
							return &Integer{Value: newTime.UnixNano() / 1000000}, nil
						}
					}
				}
			}
			return &Nil{}, &RuntimeError{Message: "time_add() requires timestamp, duration, and unit (ms/s/m/h/d)"}
		},
	})

	env.Set("time_diff", &Function{
		Name: "time_diff",
		Body: func(callEnv *Environment) (Value, error) {
			args, _ := callEnv.Get("__args__")
			if list, ok := args.(*List); ok && len(list.Elements) >= 2 {
				if timestamp1, ok := list.Elements[0].(*Integer); ok {
					if timestamp2, ok := list.Elements[1].(*Integer); ok {
						t1 := time.Unix(0, timestamp1.Value*1000000)
						t2 := time.Unix(0, timestamp2.Value*1000000)
						diff := t2.Sub(t1)
						return &Integer{Value: diff.Milliseconds()}, nil
					}
				}
			}
			return &Nil{}, &RuntimeError{Message: "time_diff() requires two timestamps"}
		},
	})

	// Promise.all() - Paralel async operations
	env.Set("Promise_all", &Function{
		Name: "Promise_all",
		Body: func(callEnv *Environment) (Value, error) {
			args, _ := callEnv.Get("__args__")
			if list, ok := args.(*List); ok && len(list.Elements) >= 1 {
				if promises, ok := list.Elements[0].(*List); ok {
					// Tüm promise'ları paralel olarak çalıştır
					results := make([]Value, len(promises.Elements))
					errors := make([]error, len(promises.Elements))

					// Her promise'ı çalıştır
					for i, promise := range promises.Elements {
						if fn, ok := promise.(*Function); ok {
							// Promise fonksiyonunu çalıştır
							result, err := fn.Body(callEnv)
							results[i] = result
							errors[i] = err
						} else {
							// Promise değilse direkt kullan
							results[i] = promise
							errors[i] = nil
						}
					}

					// Herhangi bir hata var mı kontrol et
					for _, err := range errors {
						if err != nil {
							return &Nil{}, err
						}
					}

					return &List{Elements: results}, nil
				}
			}
			return &Nil{}, &RuntimeError{Message: "Promise_all() requires a list of promises"}
		},
	})

	// Promise.allSettled() - Tüm promise'ları bekler, hata olsa bile
	env.Set("Promise_allSettled", &Function{
		Name: "Promise_allSettled",
		Body: func(callEnv *Environment) (Value, error) {
			args, _ := callEnv.Get("__args__")
			if list, ok := args.(*List); ok && len(list.Elements) >= 1 {
				if promises, ok := list.Elements[0].(*List); ok {
					// Tüm promise'ları paralel olarak çalıştır
					results := make([]Value, len(promises.Elements))

					// Her promise'ı çalıştır
					for i, promise := range promises.Elements {
						if fn, ok := promise.(*Function); ok {
							// Promise fonksiyonunu çalıştır
							result, err := fn.Body(callEnv)
							if err != nil {
								// Hata durumunda error objesi oluştur
								errorObj := &Dict{Pairs: map[string]Value{
									"status": &String{Value: "rejected"},
									"reason": &String{Value: err.Error()},
								}}
								results[i] = errorObj
							} else {
								// Başarı durumunda result objesi oluştur
								successObj := &Dict{Pairs: map[string]Value{
									"status": &String{Value: "fulfilled"},
									"value":  result,
								}}
								results[i] = successObj
							}
						} else {
							// Promise değilse direkt kullan
							successObj := &Dict{Pairs: map[string]Value{
								"status": &String{Value: "fulfilled"},
								"value":  promise,
							}}
							results[i] = successObj
						}
					}

					return &List{Elements: results}, nil
				}
			}
			return &Nil{}, &RuntimeError{Message: "Promise_allSettled() requires a list of promises"}
		},
	})

	// HTTP Module Functions
	env.Set("http_get", &Function{
		Name: "http_get",
		Body: func(callEnv *Environment) (Value, error) {
			args, _ := callEnv.Get("__args__")
			if list, ok := args.(*List); ok && len(list.Elements) >= 1 {
				if url, ok := list.Elements[0].(*String); ok {
					// Simple HTTP GET implementation
					resp, err := http.Get(url.Value)
					if err != nil {
						return &Nil{}, &RuntimeError{Message: fmt.Sprintf("http_get error: %v", err)}
					}
					defer resp.Body.Close()

					body, err := io.ReadAll(resp.Body)
					if err != nil {
						return &Nil{}, &RuntimeError{Message: fmt.Sprintf("http_get read error: %v", err)}
					}

					return &String{Value: string(body)}, nil
				}
			}
			return &Nil{}, &RuntimeError{Message: "http_get() requires URL string"}
		},
	})

	env.Set("http_post", &Function{
		Name: "http_post",
		Body: func(callEnv *Environment) (Value, error) {
			args, _ := callEnv.Get("__args__")
			if list, ok := args.(*List); ok && len(list.Elements) >= 2 {
				if url, ok := list.Elements[0].(*String); ok {
					if data, ok := list.Elements[1].(*String); ok {
						// Simple HTTP POST implementation
						resp, err := http.Post(url.Value, "application/json", strings.NewReader(data.Value))
						if err != nil {
							return &Nil{}, &RuntimeError{Message: fmt.Sprintf("http_post error: %v", err)}
						}
						defer resp.Body.Close()

						body, err := io.ReadAll(resp.Body)
						if err != nil {
							return &Nil{}, &RuntimeError{Message: fmt.Sprintf("http_post read error: %v", err)}
						}

						return &String{Value: string(body)}, nil
					}
				}
			}
			return &Nil{}, &RuntimeError{Message: "http_post() requires URL and data strings"}
		},
	})

	env.Set("http_put", &Function{
		Name: "http_put",
		Body: func(callEnv *Environment) (Value, error) {
			args, _ := callEnv.Get("__args__")
			if list, ok := args.(*List); ok && len(list.Elements) >= 2 {
				if url, ok := list.Elements[0].(*String); ok {
					if data, ok := list.Elements[1].(*String); ok {
						// Simple HTTP PUT implementation
						req, err := http.NewRequest("PUT", url.Value, strings.NewReader(data.Value))
						if err != nil {
							return &Nil{}, &RuntimeError{Message: fmt.Sprintf("http_put error: %v", err)}
						}
						req.Header.Set("Content-Type", "application/json")

						client := &http.Client{}
						resp, err := client.Do(req)
						if err != nil {
							return &Nil{}, &RuntimeError{Message: fmt.Sprintf("http_put error: %v", err)}
						}
						defer resp.Body.Close()

						body, err := io.ReadAll(resp.Body)
						if err != nil {
							return &Nil{}, &RuntimeError{Message: fmt.Sprintf("http_put read error: %v", err)}
						}

						return &String{Value: string(body)}, nil
					}
				}
			}
			return &Nil{}, &RuntimeError{Message: "http_put() requires URL and data strings"}
		},
	})

	env.Set("http_delete", &Function{
		Name: "http_delete",
		Body: func(callEnv *Environment) (Value, error) {
			args, _ := callEnv.Get("__args__")
			if list, ok := args.(*List); ok && len(list.Elements) >= 1 {
				if url, ok := list.Elements[0].(*String); ok {
					// Simple HTTP DELETE implementation
					req, err := http.NewRequest("DELETE", url.Value, nil)
					if err != nil {
						return &Nil{}, &RuntimeError{Message: fmt.Sprintf("http_delete error: %v", err)}
					}

					client := &http.Client{}
					resp, err := client.Do(req)
					if err != nil {
						return &Nil{}, &RuntimeError{Message: fmt.Sprintf("http_delete error: %v", err)}
					}
					defer resp.Body.Close()

					body, err := io.ReadAll(resp.Body)
					if err != nil {
						return &Nil{}, &RuntimeError{Message: fmt.Sprintf("http_delete read error: %v", err)}
					}

					return &String{Value: string(body)}, nil
				}
			}
			return &Nil{}, &RuntimeError{Message: "http_delete() requires URL string"}
		},
	})

	// Crypto Module Functions
	env.Set("crypto_md5", &Function{
		Name: "crypto_md5",
		Body: func(callEnv *Environment) (Value, error) {
			args, _ := callEnv.Get("__args__")
			if list, ok := args.(*List); ok && len(list.Elements) >= 1 {
				if data, ok := list.Elements[0].(*String); ok {
					hash := md5.Sum([]byte(data.Value))
					return &String{Value: fmt.Sprintf("%x", hash)}, nil
				}
			}
			return &Nil{}, &RuntimeError{Message: "crypto_md5() requires data string"}
		},
	})

	env.Set("crypto_sha256", &Function{
		Name: "crypto_sha256",
		Body: func(callEnv *Environment) (Value, error) {
			args, _ := callEnv.Get("__args__")
			if list, ok := args.(*List); ok && len(list.Elements) >= 1 {
				if data, ok := list.Elements[0].(*String); ok {
					hash := sha256.Sum256([]byte(data.Value))
					return &String{Value: fmt.Sprintf("%x", hash)}, nil
				}
			}
			return &Nil{}, &RuntimeError{Message: "crypto_sha256() requires data string"}
		},
	})

	env.Set("crypto_aes_encrypt", &Function{
		Name: "crypto_aes_encrypt",
		Body: func(callEnv *Environment) (Value, error) {
			args, _ := callEnv.Get("__args__")
			if list, ok := args.(*List); ok && len(list.Elements) >= 2 {
				if data, ok := list.Elements[0].(*String); ok {
					if key, ok := list.Elements[1].(*String); ok {
						// Simple AES encryption (for demo purposes)
						block, err := aes.NewCipher([]byte(key.Value))
						if err != nil {
							return &Nil{}, &RuntimeError{Message: fmt.Sprintf("crypto_aes_encrypt error: %v", err)}
						}

						// Pad data to block size
						paddedData := []byte(data.Value)
						blockSize := block.BlockSize()
						remainder := len(paddedData) % blockSize
						if remainder != 0 {
							padding := make([]byte, blockSize-remainder)
							paddedData = append(paddedData, padding...)
						}

						// Encrypt
						encrypted := make([]byte, len(paddedData))
						for i := 0; i < len(paddedData); i += blockSize {
							block.Encrypt(encrypted[i:i+blockSize], paddedData[i:i+blockSize])
						}

						return &String{Value: fmt.Sprintf("%x", encrypted)}, nil
					}
				}
			}
			return &Nil{}, &RuntimeError{Message: "crypto_aes_encrypt() requires data and key strings"}
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
		return i.importSymbolsFromModule(moduleEnv, stmt.Alias, modulePath)
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
	return i.importSymbolsFromModule(moduleEnv, stmt.Alias, modulePath)
}

// resolveModulePath resolves module path to file path
func (i *Interpreter) resolveModulePath(modulePath string) string {
	moduleFile := modulePath + ".sky"

	// Try Wing dependencies first
	dependencyPath := i.currentDir + "/dependencies/" + modulePath + "/src/" + moduleFile
	if _, err := os.Stat(dependencyPath); err == nil {
		return dependencyPath
	}

	// Try Wing dependencies with main.sky
	dependencyMainPath := i.currentDir + "/dependencies/" + modulePath + "/src/main.sky"
	if _, err := os.Stat(dependencyMainPath); err == nil {
		return dependencyMainPath
	}

	// Try relative to source file directory
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
func (i *Interpreter) importSymbolsFromModule(moduleEnv *Environment, alias *ast.Identifier, modulePath string) error {
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
		// Direct import (import foo) - create namespace with module name
		// Get module name from the last segment of the import path
		moduleName := modulePath
		if lastSlash := strings.LastIndex(modulePath, "/"); lastSlash >= 0 {
			moduleName = modulePath[lastSlash+1:]
		}

		// Create namespace object
		namespace := &Dict{Pairs: make(map[string]Value)}
		for name, value := range symbols {
			// Only import public symbols (not starting with _)
			if len(name) > 0 && name[0] != '_' {
				namespace.Pairs[name] = value
			}
		}

		// Set namespace with module name
		i.env.Set(moduleName, namespace)
	}
	return nil
}

// evalTryStatement handles try-catch-finally
func (i *Interpreter) evalTryStatement(stmt *ast.TryStatement) (Value, error) {
	var result Value = &Nil{}
	var tryErr error

	// Execute try block
	result, tryErr = i.evalBlockStatement(stmt.TryBlock, i.env)

	// If there's an error and a catch clause, execute catch
	if tryErr != nil && stmt.CatchClause != nil {
		catchEnv := NewEnvironment(i.env)

		// Bind error to error variable if specified
		if stmt.CatchClause.ErrorVar != nil {
			// Convert error to string value
			errorStr := tryErr.Error()
			catchEnv.Set(stmt.CatchClause.ErrorVar.Value, &String{Value: errorStr})
		}

		// Execute catch block
		oldEnv := i.env
		i.env = catchEnv
		result, tryErr = i.evalBlockStatement(stmt.CatchClause.Body, catchEnv)
		i.env = oldEnv
	}

	// Execute finally block if it exists
	if stmt.Finally != nil {
		_, _ = i.evalBlockStatement(stmt.Finally, i.env)
	}

	return result, tryErr
}

// evalThrowStatement throws an error
func (i *Interpreter) evalThrowStatement(stmt *ast.ThrowStatement) (Value, error) {
	value, err := i.evalExpression(stmt.Value)
	if err != nil {
		return nil, err
	}

	// Create runtime error with the thrown value
	return nil, &RuntimeError{Message: value.String()}
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

// evalAbstractClassStatement evaluates abstract class statements
func (i *Interpreter) evalAbstractClassStatement(stmt *ast.AbstractClassStatement) error {
	className := stmt.Name.Value

	// Create abstract class environment
	classEnv := NewEnvironment(i.env)

	// Create abstract class
	abstractClass := &AbstractClass{
		Name:            className,
		SuperClasses:    []*Class{}, // Will be populated if needed
		Methods:         make(map[string]*Function),
		AbstractMethods: make(map[string]*AbstractMethod),
		Env:             classEnv,
	}

	// Process class body
	for _, stmt := range stmt.Body {
		switch s := stmt.(type) {
		case *ast.FunctionStatement:
			// Regular method
			fn := &Function{
				Name: s.Name.Value,
				Body: func(env *Environment) (Value, error) {
					return i.evalBlockStatement(s.Body, env)
				},
			}
			abstractClass.Methods[s.Name.Value] = fn
		case *ast.AbstractMethodStatement:
			// Abstract method
			abstractMethod := &AbstractMethod{
				Name:       s.Name.Value,
				Parameters: s.Parameters,
				ReturnType: s.ReturnType,
			}
			abstractClass.AbstractMethods[s.Name.Value] = abstractMethod
		}
	}

	// Define abstract class in environment
	i.env.Set(className, abstractClass)

	return nil
}

// evalAbstractMethodStatement evaluates abstract method statements
func (i *Interpreter) evalAbstractMethodStatement(stmt *ast.AbstractMethodStatement) error {
	// Abstract methods are handled within abstract classes
	// This is just a placeholder
	return nil
}

// evalStaticMethodStatement evaluates static method statements
func (i *Interpreter) evalStaticMethodStatement(stmt *ast.StaticMethodStatement) error {
	// Parametrelerin isimlerini al
	params := make([]string, len(stmt.Parameters))
	for idx, param := range stmt.Parameters {
		params[idx] = param.Name.Value
	}

	funcName := stmt.Name.Value
	capturedStmt := stmt // Capture for closure
	capturedEnv := i.env

	// Static methods are stored in the global environment
	fn := &Function{
		Name:       funcName,
		Parameters: params,
		Env:        capturedEnv,
		Body: func(callEnv *Environment) (Value, error) {
			// Yeni environment oluştur
			fnEnv := NewEnvironment(capturedEnv)

			// Parametreleri bind et
			args, _ := callEnv.Get("__args__")
			if argList, ok := args.(*List); ok {
				for idx, param := range capturedStmt.Parameters {
					paramName := param.Name.Value
					if idx < len(argList.Elements) {
						fnEnv.Set(paramName, argList.Elements[idx])
					} else if param.DefaultValue != nil {
						// Use default value
						defaultVal, err := i.evalExpression(param.DefaultValue)
						if err != nil {
							return nil, err
						}
						fnEnv.Set(paramName, defaultVal)
					} else {
						fnEnv.Set(paramName, &Nil{})
					}
				}
			}

			// Fonksiyon body'sini çalıştır
			oldEnv := i.env
			i.env = fnEnv
			defer func() { i.env = oldEnv }()

			result, err := i.evalBlockStatement(capturedStmt.Body, fnEnv)

			// Handle ReturnSignal - extract value and return normally
			if retSignal, isReturn := err.(*ReturnSignal); isReturn {
				result = retSignal.Value
				err = nil
			}

			return result, err
		},
	}

	// Store in global environment
	i.env.Set(funcName, fn)

	return nil
}

// evalStaticPropertyStatement evaluates static property statements
func (i *Interpreter) evalStaticPropertyStatement(stmt *ast.StaticPropertyStatement) error {
	// Static properties are stored in the global environment
	// They can be accessed as ClassName.propertyName

	var value Value = &Nil{}

	if stmt.Value != nil {
		var err error
		value, err = i.evalExpression(stmt.Value)
		if err != nil {
			return err
		}
	}

	// Store in global environment
	i.env.Set(stmt.Name.Value, value)

	return nil
}
