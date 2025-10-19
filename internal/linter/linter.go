package linter

import (
	"fmt"
	"strings"

	"github.com/mburakmmm/sky-lang/internal/ast"
	"github.com/mburakmmm/sky-lang/internal/lexer"
	"github.com/mburakmmm/sky-lang/internal/parser"
)

// Issue represents a lint issue
type Issue struct {
	File     string
	Line     int
	Column   int
	Severity string // "error", "warning", "info"
	Rule     string
	Message  string
}

// Linter checks code for issues
type Linter struct {
	issues      []Issue
	filename    string
	inUnsafe    bool
	inFunction  bool
	usedVars    map[string]bool
	definedVars map[string]lexer.Token
}

// NewLinter creates a new linter
func NewLinter() *Linter {
	return &Linter{
		issues:      []Issue{},
		usedVars:    make(map[string]bool),
		definedVars: make(map[string]lexer.Token),
	}
}

// Lint checks a program for issues
func (l *Linter) Lint(program *ast.Program, filename string) []Issue {
	l.filename = filename
	l.issues = []Issue{}

	for _, stmt := range program.Statements {
		l.checkStatement(stmt)
	}

	// Check for unused variables
	for name, token := range l.definedVars {
		if !l.usedVars[name] && !strings.HasPrefix(name, "_") {
			l.addIssue(token.Line, token.Column, "warning", "unused-var",
				fmt.Sprintf("variable '%s' is defined but never used", name))
		}
	}

	return l.issues
}

func (l *Linter) addIssue(line, col int, severity, rule, message string) {
	l.issues = append(l.issues, Issue{
		File:     l.filename,
		Line:     line,
		Column:   col,
		Severity: severity,
		Rule:     rule,
		Message:  message,
	})
}

func (l *Linter) checkStatement(stmt ast.Statement) {
	switch s := stmt.(type) {
	case *ast.LetStatement:
		l.checkLetStatement(s)
	case *ast.ConstStatement:
		l.checkConstStatement(s)
	case *ast.FunctionStatement:
		l.checkFunctionStatement(s)
	case *ast.ReturnStatement:
		if !l.inFunction {
			l.addIssue(s.Token.Line, s.Token.Column, "error", "return-outside-function",
				"return statement outside of function")
		}
		if s.ReturnValue != nil {
			l.checkExpression(s.ReturnValue)
		}
	case *ast.UnsafeStatement:
		l.checkUnsafeStatement(s)
	case *ast.IfStatement:
		l.checkExpression(s.Condition)
		l.checkBlock(s.Consequence)
		for _, elif := range s.Elif {
			l.checkExpression(elif.Condition)
			l.checkBlock(elif.Consequence)
		}
		if s.Alternative != nil {
			l.checkBlock(s.Alternative)
		}
	case *ast.WhileStatement:
		l.checkExpression(s.Condition)
		l.checkBlock(s.Body)
	case *ast.ForStatement:
		l.checkExpression(s.Iterable)
		l.checkBlock(s.Body)
	case *ast.ExpressionStatement:
		l.checkExpression(s.Expression)
	}
}

func (l *Linter) checkLetStatement(stmt *ast.LetStatement) {
	name := stmt.Name.Value

	// Check shadowing
	if _, exists := l.definedVars[name]; exists {
		l.addIssue(stmt.Token.Line, stmt.Token.Column, "warning", "shadowing",
			fmt.Sprintf("variable '%s' shadows previous declaration", name))
	}

	l.definedVars[name] = stmt.Token
	l.checkExpression(stmt.Value)
}

func (l *Linter) checkConstStatement(stmt *ast.ConstStatement) {
	name := stmt.Name.Value
	l.definedVars[name] = stmt.Token
	l.checkExpression(stmt.Value)
}

func (l *Linter) checkFunctionStatement(stmt *ast.FunctionStatement) {
	wasInFunction := l.inFunction
	l.inFunction = true
	defer func() { l.inFunction = wasInFunction }()

	l.checkBlock(stmt.Body)
}

func (l *Linter) checkUnsafeStatement(stmt *ast.UnsafeStatement) {
	wasInUnsafe := l.inUnsafe
	l.inUnsafe = true
	defer func() { l.inUnsafe = wasInUnsafe }()

	l.addIssue(stmt.Token.Line, stmt.Token.Column, "warning", "unsafe-block",
		"unsafe block bypasses safety checks")

	l.checkBlock(stmt.Body)
}

func (l *Linter) checkBlock(block *ast.BlockStatement) {
	for _, stmt := range block.Statements {
		l.checkStatement(stmt)
	}
}

func (l *Linter) checkExpression(expr ast.Expression) {
	switch e := expr.(type) {
	case *ast.Identifier:
		l.usedVars[e.Value] = true
	case *ast.InfixExpression:
		l.checkExpression(e.Left)
		l.checkExpression(e.Right)

		// Check for dangerous operations
		if e.Operator == "/" {
			// Check division by zero
			if lit, ok := e.Right.(*ast.IntegerLiteral); ok {
				if lit.Value == 0 {
					l.addIssue(e.Token.Line, e.Token.Column, "error", "division-by-zero",
						"division by zero")
				}
			}
		}
	case *ast.PrefixExpression:
		l.checkExpression(e.Right)
	case *ast.CallExpression:
		l.checkExpression(e.Function)
		for _, arg := range e.Arguments {
			l.checkExpression(arg)
		}
	case *ast.IndexExpression:
		l.checkExpression(e.Left)
		l.checkExpression(e.Index)
	case *ast.MemberExpression:
		l.checkExpression(e.Object)
	case *ast.ListLiteral:
		for _, elem := range e.Elements {
			l.checkExpression(elem)
		}
	case *ast.DictLiteral:
		for key, val := range e.Pairs {
			l.checkExpression(key)
			l.checkExpression(val)
		}
	case *ast.AwaitExpression:
		l.checkExpression(e.Expression)
	case *ast.YieldExpression:
		if e.Value != nil {
			l.checkExpression(e.Value)
		}
	}
}

// LintFile lints a source file
func LintFile(filename, source string) ([]Issue, error) {
	l := lexer.New(source, filename)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		return nil, fmt.Errorf("parse errors: %v", p.Errors())
	}

	linter := NewLinter()
	return linter.Lint(program, filename), nil
}
