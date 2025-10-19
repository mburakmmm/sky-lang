package interpreter

import "github.com/mburakmmm/sky-lang/internal/ast"

// YieldSignal represents a yield from a generator
type YieldSignal struct {
	Value Value
}

func (y *YieldSignal) Error() string {
	return "yield"
}

// Generator represents a generator/coroutine instance
// Simplified implementation: collects yields during execution with iteration limit
type Generator struct {
	Values    []Value // All yielded values
	Index     int     // Current index
	Done      bool    // Is generator finished?
	maxYields int     // Maximum yields to prevent infinite loops
}

func (g *Generator) Kind() ValueKind { return GeneratorValue }
func (g *Generator) String() string  { return "<generator>" }
func (g *Generator) IsTruthy() bool  { return !g.Done }

// NewGenerator creates a generator by executing the function and collecting yields
func NewGenerator(interp *Interpreter, env *Environment, body *ast.BlockStatement) (*Generator, error) {
	gen := &Generator{
		Values:    []Value{},
		Index:     0,
		Done:      false,
		maxYields: 10000, // Prevent infinite loops
	}

	// Create yield collector
	collector := &YieldCollector{
		interp:    interp,
		yields:    &gen.Values,
		maxYields: gen.maxYields,
		count:     0,
	}

	// Execute body and collect all yields
	oldEnv := interp.env
	interp.env = env
	defer func() { interp.env = oldEnv }()

	_, err := collector.executeAndCollectYields(body, env)
	if err != nil {
		// Check if it's a yield limit error
		if runtimeErr, ok := err.(*RuntimeError); ok {
			if runtimeErr.Message == "generator yield limit exceeded" {
				// This is OK, just stop collecting
				return gen, nil
			}
		}
		// Ignore YieldSignal errors (they're expected)
		if _, isYield := err.(*YieldSignal); !isYield {
			if _, isReturn := err.(*ReturnSignal); !isReturn {
				return nil, err
			}
		}
	}

	return gen, nil
}

// Next returns the next yielded value
func (g *Generator) Next() (Value, error) {
	if g.Index >= len(g.Values) {
		g.Done = true
		return &Nil{}, nil
	}

	value := g.Values[g.Index]
	g.Index++

	if g.Index >= len(g.Values) {
		g.Done = true
	}

	return value, nil
}

// YieldCollector collects yielded values during execution
type YieldCollector struct {
	interp    *Interpreter
	yields    *[]Value
	maxYields int
	count     int
}

// executeAndCollectYields executes code and collects all yields
func (yc *YieldCollector) executeAndCollectYields(block *ast.BlockStatement, env *Environment) (Value, error) {
	var result Value = &Nil{}

	for _, stmt := range block.Statements {
		val, err := yc.executeStatement(stmt)

		if err != nil {
			// YieldSignal is not an error, collect the value
			if yieldSig, isYield := err.(*YieldSignal); isYield {
				yc.count++
				if yc.count > yc.maxYields {
					return nil, &RuntimeError{Message: "generator yield limit exceeded"}
				}
				*yc.yields = append(*yc.yields, yieldSig.Value)
				continue // Continue executing
			}
			// ReturnSignal ends generator
			if _, isReturn := err.(*ReturnSignal); isReturn {
				return nil, nil
			}
			return nil, err
		}

		result = val
	}

	return result, nil
}

// executeStatement executes a statement, handling loops properly
func (yc *YieldCollector) executeStatement(stmt ast.Statement) (Value, error) {
	// Handle while loops specially to collect yields
	if whileStmt, ok := stmt.(*ast.WhileStatement); ok {
		return yc.executeWhileLoop(whileStmt)
	}

	// Handle for loops
	if forStmt, ok := stmt.(*ast.ForStatement); ok {
		return yc.executeForLoop(forStmt)
	}

	// Handle if statements
	if ifStmt, ok := stmt.(*ast.IfStatement); ok {
		return yc.executeIfStatement(ifStmt)
	}

	// For other statements, use regular interpreter
	return yc.interp.evalStatement(stmt)
}

// executeWhileLoop executes a while loop and collects yields
func (yc *YieldCollector) executeWhileLoop(stmt *ast.WhileStatement) (Value, error) {
	for {
		// Check yield limit
		if yc.count > yc.maxYields {
			return nil, &RuntimeError{Message: "generator yield limit exceeded"}
		}

		condition, err := yc.interp.evalExpression(stmt.Condition)
		if err != nil {
			return nil, err
		}

		if !condition.IsTruthy() {
			break
		}

		_, err = yc.executeAndCollectYields(stmt.Body, yc.interp.env)
		if err != nil {
			if _, isBreak := err.(*BreakSignal); isBreak {
				break
			}
			if _, isContinue := err.(*ContinueSignal); isContinue {
				continue
			}
			if _, isReturn := err.(*ReturnSignal); isReturn {
				return nil, err
			}
			// Don't return on YieldSignal
			if _, isYield := err.(*YieldSignal); isYield {
				// Already collected by executeAndCollectYields
				continue
			}
			return nil, err
		}
	}

	return &Nil{}, nil
}

// executeForLoop executes a for loop and collects yields
func (yc *YieldCollector) executeForLoop(stmt *ast.ForStatement) (Value, error) {
	iterable, err := yc.interp.evalExpression(stmt.Iterable)
	if err != nil {
		return nil, err
	}

	if list, ok := iterable.(*List); ok {
		for _, elem := range list.Elements {
			yc.interp.env.Set(stmt.Iterator.Value, elem)
			_, err := yc.executeAndCollectYields(stmt.Body, yc.interp.env)
			if err != nil {
				if _, isBreak := err.(*BreakSignal); isBreak {
					break
				}
				if _, isContinue := err.(*ContinueSignal); isContinue {
					continue
				}
				if _, isReturn := err.(*ReturnSignal); isReturn {
					return nil, err
				}
				return nil, err
			}
		}
	}

	return &Nil{}, nil
}

// executeIfStatement executes if statement and collects yields
func (yc *YieldCollector) executeIfStatement(stmt *ast.IfStatement) (Value, error) {
	condition, err := yc.interp.evalExpression(stmt.Condition)
	if err != nil {
		return nil, err
	}

	if condition.IsTruthy() {
		return yc.executeAndCollectYields(stmt.Consequence, yc.interp.env)
	}

	// Elif clauses
	for _, elif := range stmt.Elif {
		cond, err := yc.interp.evalExpression(elif.Condition)
		if err != nil {
			return nil, err
		}
		if cond.IsTruthy() {
			return yc.executeAndCollectYields(elif.Consequence, yc.interp.env)
		}
	}

	// Else clause
	if stmt.Alternative != nil {
		return yc.executeAndCollectYields(stmt.Alternative, yc.interp.env)
	}

	return &Nil{}, nil
}
