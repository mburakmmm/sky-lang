package interpreter

import (
	"fmt"

	"github.com/mburakmmm/sky-lang/internal/ast"
)

// EnumType stores enum type metadata
type EnumType struct {
	Name     string
	Variants map[string]*VariantInfo
}

func (et *EnumType) Kind() ValueKind { return FunctionValue } // Use FunctionValue for now
func (et *EnumType) String() string  { return "enum " + et.Name }
func (et *EnumType) IsTruthy() bool  { return true }

// VariantInfo stores variant metadata
type VariantInfo struct {
	Name         string
	PayloadCount int
}

// EnumInstance represents a runtime enum value
type EnumInstance struct {
	TypeName string
	Variant  string
	Payload  []Value
}

func (ei *EnumInstance) Kind() ValueKind { return InstanceValue }
func (ei *EnumInstance) String() string  { return ei.TypeName + "::" + ei.Variant }
func (ei *EnumInstance) IsTruthy() bool  { return true }

// evalEnumStatement evaluates an enum declaration
func (i *Interpreter) evalEnumStatement(stmt *ast.EnumStatement) (Value, error) {
	enumType := &EnumType{
		Name:     stmt.Name.Value,
		Variants: make(map[string]*VariantInfo),
	}

	// Register all variants
	for _, variant := range stmt.Variants {
		enumType.Variants[variant.Name.Value] = &VariantInfo{
			Name:         variant.Name.Value,
			PayloadCount: len(variant.Payload),
		}

		// Create constructor function for each variant
		variantName := variant.Name.Value
		payloadCount := len(variant.Payload)

		constructor := &Function{
			Name:       variantName,
			Parameters: []string{},
			Env:        i.env,
			Body: func(callEnv *Environment) (Value, error) {
				// Get arguments
				args, _ := callEnv.Get("__args__")
				argList, _ := args.(*List)

				if len(argList.Elements) != payloadCount {
					return nil, &RuntimeError{
						Message: fmt.Sprintf("%s expects %d arguments, got %d",
							variantName, payloadCount, len(argList.Elements)),
					}
				}

				// Create enum instance
				payload := make([]Value, len(argList.Elements))
				copy(payload, argList.Elements)

				instance := &EnumInstance{
					TypeName: stmt.Name.Value,
					Variant:  variantName,
					Payload:  payload,
				}

				return instance, nil
			},
		}

		i.env.Set(variantName, constructor)
	}

	// Store enum type
	i.env.Set(stmt.Name.Value, enumType)

	return &Nil{}, nil
}

// evalMatchExpression evaluates a match expression
func (i *Interpreter) evalMatchExpression(expr *ast.MatchExpression) (Value, error) {
	// Evaluate the value to match
	matchValue, err := i.evalExpression(expr.Value)
	if err != nil {
		return nil, err
	}

	// Try each match arm
	for _, arm := range expr.Arms {
		matched, bindings, err := i.matchPattern(arm.Pattern, matchValue)
		if err != nil {
			return nil, err
		}

		if matched {
			// Create new scope with pattern bindings
			matchEnv := NewEnvironment(i.env)

			// Bind pattern variables
			for name, value := range bindings {
				matchEnv.Set(name, value)
			}

			// Execute arm body
			oldEnv := i.env
			i.env = matchEnv

			// Evaluate the expression in the arm body
			if len(arm.Body.Statements) > 0 {
				if exprStmt, ok := arm.Body.Statements[0].(*ast.ExpressionStatement); ok {
					result, err := i.evalExpression(exprStmt.Expression)
					i.env = oldEnv
					if err != nil {
						return nil, err
					}
					return result, nil
				}
			}

			// Fallback: eval block
			result, err := i.evalBlockStatement(arm.Body, matchEnv)
			i.env = oldEnv
			return result, err
		}
	}

	return nil, &RuntimeError{Message: "non-exhaustive match: no pattern matched"}
}

// matchPattern checks if a pattern matches a value
func (i *Interpreter) matchPattern(pattern ast.Expression, value Value) (bool, map[string]Value, error) {
	bindings := make(map[string]Value)

	switch p := pattern.(type) {
	case *ast.Identifier:
		// Wildcard or variable binding
		if p.Value == "_" {
			return true, bindings, nil
		}
		// Bind identifier to value
		bindings[p.Value] = value
		return true, bindings, nil

	case *ast.IntegerLiteral:
		// Match integer literal
		if intVal, ok := value.(*Integer); ok {
			return intVal.Value == p.Value, bindings, nil
		}
		return false, nil, nil

	case *ast.StringLiteral:
		// Match string literal
		if strVal, ok := value.(*String); ok {
			return strVal.Value == p.Value, bindings, nil
		}
		return false, nil, nil

	case *ast.BooleanLiteral:
		// Match boolean literal
		if boolVal, ok := value.(*Boolean); ok {
			return boolVal.Value == p.Value, bindings, nil
		}
		return false, nil, nil

	case *ast.CallExpression:
		// Enum variant pattern: VariantName(args...)
		if ident, ok := p.Function.(*ast.Identifier); ok {
			if enumVal, ok := value.(*EnumInstance); ok {
				if enumVal.Variant == ident.Value {
					// Match arguments
					if len(p.Arguments) != len(enumVal.Payload) {
						return false, nil, nil
					}

					for idx, arg := range p.Arguments {
						matched, argBindings, err := i.matchPattern(arg, enumVal.Payload[idx])
						if err != nil {
							return false, nil, err
						}
						if !matched {
							return false, nil, nil
						}
						// Merge bindings
						for k, v := range argBindings {
							bindings[k] = v
						}
					}
					return true, bindings, nil
				}
			}
		}
		return false, nil, nil

	case *ast.ListLiteral:
		// List pattern
		if listVal, ok := value.(*List); ok {
			if len(p.Elements) == len(listVal.Elements) {
				for idx, elem := range p.Elements {
					matched, elemBindings, err := i.matchPattern(elem, listVal.Elements[idx])
					if err != nil {
						return false, nil, err
					}
					if !matched {
						return false, nil, nil
					}
					for k, v := range elemBindings {
						bindings[k] = v
					}
				}
				return true, bindings, nil
			}
		}
		return false, nil, nil

	default:
		return false, nil, &RuntimeError{Message: fmt.Sprintf("unsupported pattern type: %T", pattern)}
	}
}
