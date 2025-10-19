package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/mburakmmm/sky-lang/internal/ast"
	"github.com/mburakmmm/sky-lang/internal/lexer"
	"github.com/mburakmmm/sky-lang/internal/parser"
)

func dumpCommand(args []string) {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: sky dump --tokens <file>")
		fmt.Fprintln(os.Stderr, "       sky dump --ast <file>")
		os.Exit(1)
	}

	flag := args[0]
	filename := args[1]

	switch flag {
	case "--tokens", "-t":
		dumpTokens(filename)
	case "--ast", "-a":
		dumpAST(filename)
	case "--json":
		if len(args) < 3 {
			fmt.Fprintln(os.Stderr, "Usage: sky dump --json <--tokens|--ast> <file>")
			os.Exit(1)
		}
		subflag := args[1]
		filename := args[2]
		switch subflag {
		case "--tokens":
			dumpTokensJSON(filename)
		case "--ast":
			dumpASTJSON(filename)
		default:
			fmt.Fprintf(os.Stderr, "Unknown flag: %s\n", subflag)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "Unknown flag: %s\n", flag)
		fmt.Fprintln(os.Stderr, "Usage: sky dump --tokens <file>")
		fmt.Fprintln(os.Stderr, "       sky dump --ast <file>")
		os.Exit(1)
	}
}

func dumpTokens(filename string) {
	content, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("=== Tokens for %s ===\n\n", filename)

	l := lexer.New(string(content), filename)

	for {
		tok := l.NextToken()

		// Token gösterimi
		displayToken(tok)

		if tok.Type == lexer.EOF {
			break
		}
	}

	fmt.Println("\n=== End of tokens ===")
}

func displayToken(tok lexer.Token) {
	// Özel token tiplerinin görsel gösterimi
	var display string

	switch tok.Type {
	case lexer.NEWLINE:
		display = "⏎"
	case lexer.INDENT:
		display = "→"
	case lexer.DEDENT:
		display = "←"
	case lexer.EOF:
		display = "EOF"
	case lexer.STRING:
		display = fmt.Sprintf("\"%s\"", tok.Literal)
	case lexer.COMMENT:
		// Yorumları kısalt
		comment := tok.Literal
		if len(comment) > 50 {
			comment = comment[:50] + "..."
		}
		display = comment
	default:
		display = tok.Literal
		if display == "" {
			display = tok.Type.String()
		}
	}

	// Konum bilgisi
	pos := fmt.Sprintf("%3d:%-3d", tok.Line, tok.Column)

	// Tip bilgisi
	typeStr := fmt.Sprintf("%-12s", tok.Type.String())

	// Kategori
	var category string
	if tok.Type.IsKeyword() {
		category = "KEYWORD  "
	} else if tok.Type.IsOperator() {
		category = "OPERATOR "
	} else if tok.Type.IsLiteral() {
		category = "LITERAL  "
	} else {
		category = "SPECIAL  "
	}

	fmt.Printf("%s | %s | %s | %s\n", pos, category, typeStr, display)
}

func dumpTokensJSON(filename string) {
	content, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	tokens := lexer.Tokenize(string(content), filename)

	// JSON çıktısı için token'ları basit yapıya çevir
	type SimpleToken struct {
		Type    string `json:"type"`
		Literal string `json:"literal"`
		Line    int    `json:"line"`
		Column  int    `json:"column"`
	}

	simpleTokens := make([]SimpleToken, len(tokens))
	for i, tok := range tokens {
		simpleTokens[i] = SimpleToken{
			Type:    tok.Type.String(),
			Literal: tok.Literal,
			Line:    tok.Line,
			Column:  tok.Column,
		}
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(simpleTokens); err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", err)
		os.Exit(1)
	}
}

func dumpAST(filename string) {
	content, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("=== AST for %s ===\n\n", filename)

	l := lexer.New(string(content), filename)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		fmt.Println("Parser errors:")
		for _, err := range p.Errors() {
			fmt.Printf("  - %s\n", err)
		}
		fmt.Println()
	}

	printAST(program, 0)
	fmt.Println("\n=== End of AST ===")
}

func printAST(node ast.Node, indent int) {
	if node == nil {
		return
	}

	prefix := strings.Repeat("  ", indent)

	switch n := node.(type) {
	case *ast.Program:
		fmt.Printf("%sProgram:\n", prefix)
		for _, stmt := range n.Statements {
			printAST(stmt, indent+1)
		}

	case *ast.LetStatement:
		fmt.Printf("%sLetStatement: %s = ", prefix, n.Name.Value)
		if n.Type != nil {
			fmt.Printf(": %s ", n.Type.String())
		}
		printAST(n.Value, 0)
		fmt.Println()

	case *ast.ConstStatement:
		fmt.Printf("%sConstStatement: %s = ", prefix, n.Name.Value)
		if n.Type != nil {
			fmt.Printf(": %s ", n.Type.String())
		}
		printAST(n.Value, 0)
		fmt.Println()

	case *ast.ReturnStatement:
		fmt.Printf("%sReturnStatement: ", prefix)
		if n.ReturnValue != nil {
			printAST(n.ReturnValue, 0)
		}
		fmt.Println()

	case *ast.ExpressionStatement:
		fmt.Printf("%sExpressionStatement: ", prefix)
		printAST(n.Expression, 0)
		fmt.Println()

	case *ast.FunctionStatement:
		fmt.Printf("%sFunctionStatement: %s(", prefix, n.Name.Value)
		for i, param := range n.Parameters {
			if i > 0 {
				fmt.Print(", ")
			}
			fmt.Print(param.Name.Value)
			if param.Type != nil {
				fmt.Printf(": %s", param.Type.String())
			}
		}
		fmt.Print(")")
		if n.ReturnType != nil {
			fmt.Printf(": %s", n.ReturnType.String())
		}
		fmt.Println()
		printAST(n.Body, indent+1)

	case *ast.BlockStatement:
		for _, stmt := range n.Statements {
			printAST(stmt, indent)
		}

	case *ast.IfStatement:
		fmt.Printf("%sIfStatement:\n", prefix)
		fmt.Printf("%s  Condition: ", prefix)
		printAST(n.Condition, 0)
		fmt.Println()
		fmt.Printf("%s  Consequence:\n", prefix)
		printAST(n.Consequence, indent+2)
		if len(n.Elif) > 0 {
			for _, elif := range n.Elif {
				fmt.Printf("%s  Elif Condition: ", prefix)
				printAST(elif.Condition, 0)
				fmt.Println()
				printAST(elif.Consequence, indent+2)
			}
		}
		if n.Alternative != nil {
			fmt.Printf("%s  Alternative:\n", prefix)
			printAST(n.Alternative, indent+2)
		}

	case *ast.WhileStatement:
		fmt.Printf("%sWhileStatement:\n", prefix)
		fmt.Printf("%s  Condition: ", prefix)
		printAST(n.Condition, 0)
		fmt.Println()
		printAST(n.Body, indent+1)

	case *ast.ForStatement:
		fmt.Printf("%sForStatement: %s in ", prefix, n.Iterator.Value)
		printAST(n.Iterable, 0)
		fmt.Println()
		printAST(n.Body, indent+1)

	case *ast.Identifier:
		fmt.Print(n.Value)

	case *ast.IntegerLiteral:
		fmt.Printf("%d", n.Value)

	case *ast.FloatLiteral:
		fmt.Printf("%f", n.Value)

	case *ast.StringLiteral:
		fmt.Printf("\"%s\"", n.Value)

	case *ast.BooleanLiteral:
		fmt.Printf("%t", n.Value)

	case *ast.PrefixExpression:
		fmt.Printf("(%s", n.Operator)
		printAST(n.Right, 0)
		fmt.Print(")")

	case *ast.InfixExpression:
		fmt.Print("(")
		printAST(n.Left, 0)
		fmt.Printf(" %s ", n.Operator)
		printAST(n.Right, 0)
		fmt.Print(")")

	case *ast.CallExpression:
		printAST(n.Function, 0)
		fmt.Print("(")
		for i, arg := range n.Arguments {
			if i > 0 {
				fmt.Print(", ")
			}
			printAST(arg, 0)
		}
		fmt.Print(")")

	case *ast.ListLiteral:
		fmt.Print("[")
		for i, elem := range n.Elements {
			if i > 0 {
				fmt.Print(", ")
			}
			printAST(elem, 0)
		}
		fmt.Print("]")

	case *ast.IndexExpression:
		printAST(n.Left, 0)
		fmt.Print("[")
		printAST(n.Index, 0)
		fmt.Print("]")

	case *ast.MemberExpression:
		printAST(n.Object, 0)
		fmt.Print(".")
		printAST(n.Member, 0)

	case *ast.AwaitExpression:
		fmt.Print("await ")
		printAST(n.Expression, 0)

	case *ast.YieldExpression:
		fmt.Print("yield")
		if n.Value != nil {
			fmt.Print(" ")
			printAST(n.Value, 0)
		}

	default:
		fmt.Printf("%s<Unhandled: %T>", prefix, node)
	}
}

func dumpASTJSON(filename string) {
	fmt.Println("{")
	fmt.Printf("  \"error\": \"AST parser not yet implemented (Sprint 2)\",\n")
	fmt.Printf("  \"file\": %q\n", filename)
	fmt.Println("}")
}
