package optimizer

import (
	"github.com/mburakmmm/sky-lang/internal/ast"
)

// EscapeAnalyzer analyzes variable escape behavior
type EscapeAnalyzer struct {
	escapes map[string]bool // variable name -> escapes?
}

// NewEscapeAnalyzer creates a new escape analyzer
func NewEscapeAnalyzer() *EscapeAnalyzer {
	return &EscapeAnalyzer{
		escapes: make(map[string]bool),
	}
}

// Analyze performs escape analysis on a program
func (ea *EscapeAnalyzer) Analyze(program *ast.Program) map[string]bool {
	for _, stmt := range program.Statements {
		ea.analyzeStatement(stmt)
	}
	return ea.escapes
}

func (ea *EscapeAnalyzer) analyzeStatement(stmt ast.Statement) {
	switch s := stmt.(type) {
	case *ast.LetStatement:
		// Check if variable escapes
		if ea.valueEscapes(s.Value) {
			ea.escapes[s.Name.Value] = true
		}
	case *ast.FunctionStatement:
		// Analyze function body
		for _, bodyStmt := range s.Body.Statements {
			ea.analyzeStatement(bodyStmt)
		}
	case *ast.ReturnStatement:
		// Returned values escape
		if s.ReturnValue != nil {
			ea.markEscape(s.ReturnValue)
		}
	}
}

func (ea *EscapeAnalyzer) valueEscapes(expr ast.Expression) bool {
	switch e := expr.(type) {
	case *ast.CallExpression:
		// Function calls may cause escape
		return true
	case *ast.ListLiteral:
		// Lists may escape
		return true
	case *ast.DictLiteral:
		// Dicts may escape
		return true
	case *ast.Identifier:
		// Check if identifier escapes
		return ea.escapes[e.Value]
	default:
		return false
	}
}

func (ea *EscapeAnalyzer) markEscape(expr ast.Expression) {
	switch e := expr.(type) {
	case *ast.Identifier:
		ea.escapes[e.Value] = true
	case *ast.ListLiteral:
		for _, elem := range e.Elements {
			ea.markEscape(elem)
		}
	case *ast.DictLiteral:
		for _, val := range e.Pairs {
			ea.markEscape(val)
		}
	}
}

// CanStackAllocate checks if a variable can be stack-allocated
func (ea *EscapeAnalyzer) CanStackAllocate(varName string) bool {
	return !ea.escapes[varName]
}

