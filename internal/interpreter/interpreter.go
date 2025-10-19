package interpreter

import (
	"fmt"
	"os"

	"github.com/mburakmmm/sky-lang/internal/ast"
)

// Interpreter AST'yi yorumlar ve çalıştırır
type Interpreter struct {
	env        *Environment
	output     *os.File
	trampoline *TrampolineStack // Custom call stack for recursion
}

// New yeni bir interpreter oluşturur
func New() *Interpreter {
	env := NewEnvironment(nil)
	trampoline := NewTrampolineStack(10000) // 10K depth limit

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

	return &Interpreter{
		env:        env,
		output:     os.Stdout,
		trampoline: trampoline,
	}
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

	if list, ok := iterable.(*List); ok {
		// Yeni scope oluştur
		forEnv := NewEnvironment(i.env)
		oldEnv := i.env
		i.env = forEnv
		defer func() { i.env = oldEnv }()

		for _, elem := range list.Elements {
			i.env.Set(stmt.Iterator.Value, elem)
			_, err := i.evalBlockStatement(stmt.Body, i.env)
			if err != nil {
				return nil, err
			}
		}
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

	fn, ok := function.(*Function)
	if !ok {
		return nil, &RuntimeError{Message: fmt.Sprintf("%s is not a function", expr.Function.String())}
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
