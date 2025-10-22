package docgen

import (
	"fmt"
	"strings"

	"github.com/mburakmmm/sky-lang/internal/ast"
	"github.com/mburakmmm/sky-lang/internal/lexer"
	"github.com/mburakmmm/sky-lang/internal/parser"
)

// DocEntry represents a documentation entry
type DocEntry struct {
	Name        string
	Type        string // "function", "class", "const"
	Signature   string
	Description string
	File        string
	Line        int
}

// DocGenerator generates documentation
type DocGenerator struct {
	entries  []DocEntry
	filename string
	comments []string
}

// NewDocGenerator creates a new doc generator
func NewDocGenerator() *DocGenerator {
	return &DocGenerator{
		entries:  []DocEntry{},
		comments: []string{},
	}
}

// Generate generates documentation for a program
func (d *DocGenerator) Generate(program *ast.Program, filename string) []DocEntry {
	d.filename = filename
	d.entries = []DocEntry{}

	for _, stmt := range program.Statements {
		d.processStatement(stmt)
	}

	return d.entries
}

func (d *DocGenerator) processStatement(stmt ast.Statement) {
	switch s := stmt.(type) {
	case *ast.FunctionStatement:
		d.addFunctionDoc(s)
	case *ast.ClassStatement:
		d.addClassDoc(s)
	case *ast.ConstStatement:
		d.addConstDoc(s)
	}
}

func (d *DocGenerator) addFunctionDoc(stmt *ast.FunctionStatement) {
	sig := d.buildFunctionSignature(stmt)
	desc := d.extractDescription(stmt.Token.Line)

	d.entries = append(d.entries, DocEntry{
		Name:        stmt.Name.Value,
		Type:        "function",
		Signature:   sig,
		Description: desc,
		File:        d.filename,
		Line:        stmt.Token.Line,
	})
}

func (d *DocGenerator) addClassDoc(stmt *ast.ClassStatement) {
	sig := fmt.Sprintf("class %s", stmt.Name.Value)
	if len(stmt.SuperClasses) > 0 {
		sig += " : "
		for i, superClass := range stmt.SuperClasses {
			if i > 0 {
				sig += ", "
			}
			sig += superClass.Value
		}
	}

	desc := d.extractDescription(stmt.Token.Line)

	d.entries = append(d.entries, DocEntry{
		Name:        stmt.Name.Value,
		Type:        "class",
		Signature:   sig,
		Description: desc,
		File:        d.filename,
		Line:        stmt.Token.Line,
	})
}

func (d *DocGenerator) addConstDoc(stmt *ast.ConstStatement) {
	sig := fmt.Sprintf("const %s", stmt.Name.Value)
	if stmt.Type != nil {
		sig += ": " + d.typeToString(stmt.Type)
	}

	desc := d.extractDescription(stmt.Token.Line)

	d.entries = append(d.entries, DocEntry{
		Name:        stmt.Name.Value,
		Type:        "const",
		Signature:   sig,
		Description: desc,
		File:        d.filename,
		Line:        stmt.Token.Line,
	})
}

func (d *DocGenerator) buildFunctionSignature(stmt *ast.FunctionStatement) string {
	var sb strings.Builder

	if stmt.Async {
		sb.WriteString("async ")
	}

	sb.WriteString("function ")
	sb.WriteString(stmt.Name.Value)
	sb.WriteString("(")

	for i, param := range stmt.Parameters {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(param.Name.Value)
		if param.Type != nil {
			sb.WriteString(": ")
			sb.WriteString(d.typeToString(param.Type))
		}
	}

	sb.WriteString(")")

	if stmt.ReturnType != nil {
		sb.WriteString(": ")
		sb.WriteString(d.typeToString(stmt.ReturnType))
	}

	return sb.String()
}

func (d *DocGenerator) typeToString(typeAnn ast.TypeAnnotation) string {
	switch t := typeAnn.(type) {
	case *ast.BasicType:
		return t.Name
	case *ast.ListType:
		return "[" + d.typeToString(t.ElementType) + "]"
	case *ast.DictType:
		return "{" + d.typeToString(t.KeyType) + ": " + d.typeToString(t.ValueType) + "}"
	case *ast.PointerType:
		return "*" + d.typeToString(t.PointeeType)
	case *ast.FunctionType:
		var params []string
		for _, pt := range t.ParamTypes {
			params = append(params, d.typeToString(pt))
		}
		return "(" + strings.Join(params, ", ") + ") => " + d.typeToString(t.ReturnType)
	}
	return "any"
}

func (d *DocGenerator) extractDescription(line int) string {
	// TODO: Extract comments above the line
	return ""
}

// GenerateMarkdown generates markdown documentation
func GenerateMarkdown(entries []DocEntry) string {
	var sb strings.Builder

	sb.WriteString("# API Documentation\n\n")

	// Group by type
	functions := []DocEntry{}
	classes := []DocEntry{}
	constants := []DocEntry{}

	for _, entry := range entries {
		switch entry.Type {
		case "function":
			functions = append(functions, entry)
		case "class":
			classes = append(classes, entry)
		case "const":
			constants = append(constants, entry)
		}
	}

	if len(constants) > 0 {
		sb.WriteString("## Constants\n\n")
		for _, entry := range constants {
			sb.WriteString(fmt.Sprintf("### %s\n\n", entry.Name))
			sb.WriteString(fmt.Sprintf("```sky\n%s\n```\n\n", entry.Signature))
			if entry.Description != "" {
				sb.WriteString(entry.Description + "\n\n")
			}
			sb.WriteString(fmt.Sprintf("*Defined in %s:%d*\n\n", entry.File, entry.Line))
		}
	}

	if len(functions) > 0 {
		sb.WriteString("## Functions\n\n")
		for _, entry := range functions {
			sb.WriteString(fmt.Sprintf("### %s\n\n", entry.Name))
			sb.WriteString(fmt.Sprintf("```sky\n%s\n```\n\n", entry.Signature))
			if entry.Description != "" {
				sb.WriteString(entry.Description + "\n\n")
			}
			sb.WriteString(fmt.Sprintf("*Defined in %s:%d*\n\n", entry.File, entry.Line))
		}
	}

	if len(classes) > 0 {
		sb.WriteString("## Classes\n\n")
		for _, entry := range classes {
			sb.WriteString(fmt.Sprintf("### %s\n\n", entry.Name))
			sb.WriteString(fmt.Sprintf("```sky\n%s\n```\n\n", entry.Signature))
			if entry.Description != "" {
				sb.WriteString(entry.Description + "\n\n")
			}
			sb.WriteString(fmt.Sprintf("*Defined in %s:%d*\n\n", entry.File, entry.Line))
		}
	}

	return sb.String()
}

// GenerateDoc generates documentation from source file
func GenerateDoc(filename, source string) (string, error) {
	l := lexer.New(source, filename)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		return "", fmt.Errorf("parse errors: %v", p.Errors())
	}

	docgen := NewDocGenerator()
	entries := docgen.Generate(program, filename)

	return GenerateMarkdown(entries), nil
}
