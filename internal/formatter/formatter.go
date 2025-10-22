package formatter

import (
	"fmt"
	"strings"

	"github.com/mburakmmm/sky-lang/internal/ast"
	"github.com/mburakmmm/sky-lang/internal/lexer"
	"github.com/mburakmmm/sky-lang/internal/parser"
)

// Formatter formats SKY source code
type Formatter struct {
	indentLevel int
	indentStr   string
	output      strings.Builder
}

// NewFormatter creates a new formatter
func NewFormatter() *Formatter {
	return &Formatter{
		indentLevel: 0,
		indentStr:   "  ", // 2 spaces
	}
}

// Format formats a program
func (f *Formatter) Format(program *ast.Program) string {
	f.output.Reset()
	f.indentLevel = 0

	for i, stmt := range program.Statements {
		f.formatStatement(stmt)

		// Add blank line between top-level declarations
		if i < len(program.Statements)-1 {
			nextStmt := program.Statements[i+1]
			if f.needsBlankLine(stmt, nextStmt) {
				f.output.WriteString("\n")
			}
		}
	}

	return f.output.String()
}

// needsBlankLine determines if blank line is needed between statements
func (f *Formatter) needsBlankLine(stmt1, stmt2 ast.Statement) bool {
	// Blank line between functions
	if _, ok := stmt1.(*ast.FunctionStatement); ok {
		return true
	}
	if _, ok := stmt1.(*ast.ClassStatement); ok {
		return true
	}
	return false
}

// formatStatement formats a statement
func (f *Formatter) formatStatement(stmt ast.Statement) {
	switch s := stmt.(type) {
	case *ast.LetStatement:
		f.formatLetStatement(s)
	case *ast.ConstStatement:
		f.formatConstStatement(s)
	case *ast.FunctionStatement:
		f.formatFunctionStatement(s)
	case *ast.ClassStatement:
		f.formatClassStatement(s)
	case *ast.ReturnStatement:
		f.formatReturnStatement(s)
	case *ast.BreakStatement:
		f.writeIndent()
		f.output.WriteString("break\n")
	case *ast.ContinueStatement:
		f.writeIndent()
		f.output.WriteString("continue\n")
	case *ast.IfStatement:
		f.formatIfStatement(s)
	case *ast.WhileStatement:
		f.formatWhileStatement(s)
	case *ast.ForStatement:
		f.formatForStatement(s)
	case *ast.ImportStatement:
		f.formatImportStatement(s)
	case *ast.UnsafeStatement:
		f.formatUnsafeStatement(s)
	case *ast.ExpressionStatement:
		f.writeIndent()
		f.formatExpression(s.Expression)
		f.output.WriteString("\n")
	}
}

func (f *Formatter) formatLetStatement(stmt *ast.LetStatement) {
	f.writeIndent()
	f.output.WriteString("let ")
	f.output.WriteString(stmt.Name.Value)

	if stmt.Type != nil {
		f.output.WriteString(": ")
		f.formatType(stmt.Type)
	}

	f.output.WriteString(" = ")
	f.formatExpression(stmt.Value)
	f.output.WriteString("\n")
}

func (f *Formatter) formatConstStatement(stmt *ast.ConstStatement) {
	f.writeIndent()
	f.output.WriteString("const ")
	f.output.WriteString(stmt.Name.Value)

	if stmt.Type != nil {
		f.output.WriteString(": ")
		f.formatType(stmt.Type)
	}

	f.output.WriteString(" = ")
	f.formatExpression(stmt.Value)
	f.output.WriteString("\n")
}

func (f *Formatter) formatFunctionStatement(stmt *ast.FunctionStatement) {
	f.writeIndent()

	if stmt.Async {
		f.output.WriteString("async ")
	}

	f.output.WriteString("function ")
	f.output.WriteString(stmt.Name.Value)

	// Parameters
	if len(stmt.Parameters) > 0 {
		f.output.WriteString("(")
		for i, param := range stmt.Parameters {
			if i > 0 {
				f.output.WriteString(", ")
			}
			f.output.WriteString(param.Name.Value)
			if param.Type != nil {
				f.output.WriteString(": ")
				f.formatType(param.Type)
			}
		}
		f.output.WriteString(")")
	}

	// Return type
	if stmt.ReturnType != nil {
		f.output.WriteString(": ")
		f.formatType(stmt.ReturnType)
	}

	f.output.WriteString("\n")

	// Body
	f.indentLevel++
	for _, bodyStmt := range stmt.Body.Statements {
		f.formatStatement(bodyStmt)
	}
	f.indentLevel--

	f.writeIndent()
	f.output.WriteString("end\n")
}

func (f *Formatter) formatClassStatement(stmt *ast.ClassStatement) {
	f.writeIndent()
	f.output.WriteString("class ")
	f.output.WriteString(stmt.Name.Value)

	if len(stmt.SuperClasses) > 0 {
		f.output.WriteString(" : ")
		for i, superClass := range stmt.SuperClasses {
			if i > 0 {
				f.output.WriteString(", ")
			}
			f.output.WriteString(superClass.Value)
		}
	}

	f.output.WriteString("\n")

	// Body
	f.indentLevel++
	for _, member := range stmt.Body {
		f.formatStatement(member)
	}
	f.indentLevel--

	f.writeIndent()
	f.output.WriteString("end\n")
}

func (f *Formatter) formatIfStatement(stmt *ast.IfStatement) {
	f.writeIndent()
	f.output.WriteString("if ")
	f.formatExpression(stmt.Condition)
	f.output.WriteString("\n")

	f.indentLevel++
	for _, s := range stmt.Consequence.Statements {
		f.formatStatement(s)
	}
	f.indentLevel--

	// Elif clauses
	for _, elif := range stmt.Elif {
		f.writeIndent()
		f.output.WriteString("elif ")
		f.formatExpression(elif.Condition)
		f.output.WriteString("\n")

		f.indentLevel++
		for _, s := range elif.Consequence.Statements {
			f.formatStatement(s)
		}
		f.indentLevel--
	}

	// Else clause
	if stmt.Alternative != nil {
		f.writeIndent()
		f.output.WriteString("else\n")

		f.indentLevel++
		for _, s := range stmt.Alternative.Statements {
			f.formatStatement(s)
		}
		f.indentLevel--
	}

	f.writeIndent()
	f.output.WriteString("end\n")
}

func (f *Formatter) formatWhileStatement(stmt *ast.WhileStatement) {
	f.writeIndent()
	f.output.WriteString("while ")
	f.formatExpression(stmt.Condition)
	f.output.WriteString("\n")

	f.indentLevel++
	for _, s := range stmt.Body.Statements {
		f.formatStatement(s)
	}
	f.indentLevel--

	f.writeIndent()
	f.output.WriteString("end\n")
}

func (f *Formatter) formatForStatement(stmt *ast.ForStatement) {
	f.writeIndent()
	f.output.WriteString("for ")
	f.output.WriteString(stmt.Iterator.Value)
	f.output.WriteString(" in ")
	f.formatExpression(stmt.Iterable)
	f.output.WriteString("\n")

	f.indentLevel++
	for _, s := range stmt.Body.Statements {
		f.formatStatement(s)
	}
	f.indentLevel--

	f.writeIndent()
	f.output.WriteString("end\n")
}

func (f *Formatter) formatReturnStatement(stmt *ast.ReturnStatement) {
	f.writeIndent()
	f.output.WriteString("return")

	if stmt.ReturnValue != nil {
		f.output.WriteString(" ")
		f.formatExpression(stmt.ReturnValue)
	}

	f.output.WriteString("\n")
}

func (f *Formatter) formatImportStatement(stmt *ast.ImportStatement) {
	f.writeIndent()
	f.output.WriteString("import ")
	f.output.WriteString(strings.Join(stmt.Path, "."))

	if stmt.Alias != nil {
		f.output.WriteString(" as ")
		f.output.WriteString(stmt.Alias.Value)
	}

	f.output.WriteString("\n")
}

func (f *Formatter) formatUnsafeStatement(stmt *ast.UnsafeStatement) {
	f.writeIndent()
	f.output.WriteString("unsafe\n")

	f.indentLevel++
	for _, s := range stmt.Body.Statements {
		f.formatStatement(s)
	}
	f.indentLevel--

	f.writeIndent()
	f.output.WriteString("end\n")
}

func (f *Formatter) formatExpression(expr ast.Expression) {
	switch e := expr.(type) {
	case *ast.IntegerLiteral:
		f.output.WriteString(fmt.Sprintf("%d", e.Value))
	case *ast.FloatLiteral:
		f.output.WriteString(fmt.Sprintf("%g", e.Value))
	case *ast.StringLiteral:
		f.output.WriteString(fmt.Sprintf("\"%s\"", e.Value))
	case *ast.BooleanLiteral:
		if e.Value {
			f.output.WriteString("true")
		} else {
			f.output.WriteString("false")
		}
	case *ast.Identifier:
		f.output.WriteString(e.Value)
	case *ast.ListLiteral:
		f.output.WriteString("[")
		for i, elem := range e.Elements {
			if i > 0 {
				f.output.WriteString(", ")
			}
			f.formatExpression(elem)
		}
		f.output.WriteString("]")
	case *ast.DictLiteral:
		f.output.WriteString("{")
		i := 0
		for key, val := range e.Pairs {
			if i > 0 {
				f.output.WriteString(", ")
			}
			f.formatExpression(key)
			f.output.WriteString(": ")
			f.formatExpression(val)
			i++
		}
		f.output.WriteString("}")
	case *ast.PrefixExpression:
		f.output.WriteString(e.Operator)
		f.formatExpression(e.Right)
	case *ast.InfixExpression:
		f.formatExpression(e.Left)
		f.output.WriteString(" ")
		f.output.WriteString(e.Operator)
		f.output.WriteString(" ")
		f.formatExpression(e.Right)
	case *ast.CallExpression:
		f.formatExpression(e.Function)
		f.output.WriteString("(")
		for i, arg := range e.Arguments {
			if i > 0 {
				f.output.WriteString(", ")
			}
			f.formatExpression(arg)
		}
		f.output.WriteString(")")
	case *ast.IndexExpression:
		f.formatExpression(e.Left)
		f.output.WriteString("[")
		f.formatExpression(e.Index)
		f.output.WriteString("]")
	case *ast.MemberExpression:
		f.formatExpression(e.Object)
		f.output.WriteString(".")
		f.output.WriteString(e.Member.Value)
	case *ast.AwaitExpression:
		f.output.WriteString("await ")
		f.formatExpression(e.Expression)
	case *ast.YieldExpression:
		f.output.WriteString("yield")
		if e.Value != nil {
			f.output.WriteString(" ")
			f.formatExpression(e.Value)
		}
	}
}

func (f *Formatter) formatType(typeAnn ast.TypeAnnotation) {
	switch t := typeAnn.(type) {
	case *ast.BasicType:
		f.output.WriteString(t.Name)
	case *ast.ListType:
		f.output.WriteString("[")
		f.formatType(t.ElementType)
		f.output.WriteString("]")
	case *ast.DictType:
		f.output.WriteString("{")
		f.formatType(t.KeyType)
		f.output.WriteString(": ")
		f.formatType(t.ValueType)
		f.output.WriteString("}")
	case *ast.PointerType:
		f.output.WriteString("*")
		f.formatType(t.PointeeType)
	case *ast.FunctionType:
		f.output.WriteString("(")
		for i, param := range t.ParamTypes {
			if i > 0 {
				f.output.WriteString(", ")
			}
			f.formatType(param)
		}
		f.output.WriteString(") => ")
		f.formatType(t.ReturnType)
	}
}

func (f *Formatter) writeIndent() {
	for i := 0; i < f.indentLevel; i++ {
		f.output.WriteString(f.indentStr)
	}
}

// FormatFile formats a source file
func FormatFile(filename, source string) (string, error) {
	l := lexer.New(source, filename)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		return "", fmt.Errorf("parse errors: %v", p.Errors())
	}

	formatter := NewFormatter()
	return formatter.Format(program), nil
}
