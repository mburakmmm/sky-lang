package interpreter

import "github.com/mburakmmm/sky-lang/internal/ast"

// Generator represents a generator/coroutine instance
type Generator struct {
	Function *Function           // The generator function
	Env      *Environment        // Generator's environment
	State    int                 // Current state (0 = not started, -1 = finished)
	Values   []Value             // Yielded values queue
	Done     bool                // Is generator finished?
	interp   *Interpreter        // Interpreter reference
	stmt     *ast.BlockStatement // Function body
}

func (g *Generator) Kind() ValueKind { return GeneratorValue }
func (g *Generator) String() string  { return "<generator>" }
func (g *Generator) IsTruthy() bool  { return !g.Done }

// Next advances the generator and returns the next yielded value
func (g *Generator) Next() (Value, error) {
	if g.Done {
		return &Nil{}, &RuntimeError{Message: "generator exhausted"}
	}

	// For now, simple implementation: just execute the function once
	// In a full implementation, this would resume from last yield point

	// Execute generator body if not started
	if g.State == 0 {
		g.State = 1

		// Execute and collect yields
		oldEnv := g.interp.env
		g.interp.env = g.Env
		defer func() { g.interp.env = oldEnv }()

		_, err := g.interp.evalBlockStatement(g.stmt, g.Env)
		if err != nil {
			g.Done = true
			return nil, err
		}

		g.Done = true
	}

	// Return nil when done
	if len(g.Values) == 0 {
		return &Nil{}, nil
	}

	// Return next value
	value := g.Values[0]
	g.Values = g.Values[1:]

	if len(g.Values) == 0 {
		g.Done = true
	}

	return value, nil
}

// YieldValue adds a value to the generator's yield queue
func (g *Generator) YieldValue(value Value) {
	g.Values = append(g.Values, value)
}
