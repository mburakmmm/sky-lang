package lexer

import (
	"testing"
)

func TestNextToken(t *testing.T) {
	input := `let x = 5
let y = 10

function add(a, b)
  return a + b
end
`

	tests := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{LET, "let"},
		{IDENT, "x"},
		{ASSIGN, "="},
		{INT, "5"},
		{NEWLINE, "\\n"},

		{LET, "let"},
		{IDENT, "y"},
		{ASSIGN, "="},
		{INT, "10"},
		{NEWLINE, "\\n"},
		{NEWLINE, "\\n"},

		{FUNCTION, "function"},
		{IDENT, "add"},
		{LPAREN, "("},
		{IDENT, "a"},
		{COMMA, ","},
		{IDENT, "b"},
		{RPAREN, ")"},
		{NEWLINE, "\\n"},
		{INDENT, ""},

		{RETURN, "return"},
		{IDENT, "a"},
		{PLUS, "+"},
		{IDENT, "b"},
		{NEWLINE, "\\n"},

		{DEDENT, ""},
		{END, "end"},
		{NEWLINE, "\\n"},
		{EOF, ""},
	}

	l := New(input, "test.sky")

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q (literal=%q)",
				i, tt.expectedType, tok.Type, tok.Literal)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestIndentation(t *testing.T) {
	input := `function test
  let x = 1
  if x > 0
    print("positive")
  end
end`

	tests := []struct {
		expectedType TokenType
	}{
		{FUNCTION},
		{IDENT},
		{NEWLINE},
		{INDENT},
		{LET},
		{IDENT},
		{ASSIGN},
		{INT},
		{NEWLINE},
		{IF},
		{IDENT},
		{GT},
		{INT},
		{NEWLINE},
		{INDENT},
		{IDENT}, // print
		{LPAREN},
		{STRING},
		{RPAREN},
		{NEWLINE},
		{DEDENT},
		{END},
		{NEWLINE},
		{DEDENT},
		{END},
		{EOF},
	}

	l := New(input, "test.sky")

	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}
	}
}

func TestOperators(t *testing.T) {
	input := `+ - * / % **
== != < <= > >=
&& || !
= += -= *= /= %=
=> . , :`

	tests := []TokenType{
		PLUS, MINUS, STAR, SLASH, PERCENT, POWER, NEWLINE,
		EQ, NE, LT, LE, GT, GE, NEWLINE,
		LAND, LOR, LNOT, NEWLINE,
		ASSIGN, PLUSEQ, MINUSEQ, STAREQ, SLASHEQ, PERCENTEQ, NEWLINE,
		ARROW, DOT, COMMA, COLON,
		EOF,
	}

	l := New(input, "test.sky")

	for i, expectedType := range tests {
		tok := l.NextToken()
		if tok.Type != expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, expectedType, tok.Type)
		}
	}
}

func TestKeywords(t *testing.T) {
	input := `function end class let const if elif else for while return
async await coop yield unsafe self super import as in
true false and or not`

	keywords := []TokenType{
		FUNCTION, END, CLASS, LET, CONST, IF, ELIF, ELSE, FOR, WHILE, RETURN, NEWLINE,
		ASYNC, AWAIT, COOP, YIELD, UNSAFE, SELF, SUPER, IMPORT, AS, IN, NEWLINE,
		TRUE, FALSE, AND, OR, NOT,
	}

	l := New(input, "test.sky")

	for i, expected := range keywords {
		tok := l.NextToken()
		if tok.Type != expected {
			t.Fatalf("tests[%d] - wrong token. expected=%q, got=%q",
				i, expected, tok.Type)
		}
	}
}

func TestNumbers(t *testing.T) {
	tests := []struct {
		input        string
		expectedType TokenType
		expectedLit  string
	}{
		{"42", INT, "42"},
		{"3.14", FLOAT, "3.14"},
		{"1.0e10", FLOAT, "1.0e10"},
		{"0xFF", INT, "0xFF"},
		{"0b1010", INT, "0b1010"},
		{"0o777", INT, "0o777"},
		{"2.5e-3", FLOAT, "2.5e-3"},
	}

	for _, tt := range tests {
		l := New(tt.input, "test.sky")
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Errorf("input=%q - wrong type. expected=%q, got=%q",
				tt.input, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLit {
			t.Errorf("input=%q - wrong literal. expected=%q, got=%q",
				tt.input, tt.expectedLit, tok.Literal)
		}
	}
}

func TestStrings(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"hello"`, "hello"},
		{`'world'`, "world"},
		{`"hello\nworld"`, "hello\nworld"},
		{`"tab\there"`, "tab\there"},
		{`"quote\"here"`, "quote\"here"},
	}

	for _, tt := range tests {
		l := New(tt.input, "test.sky")
		tok := l.NextToken()

		if tok.Type != STRING {
			t.Errorf("input=%q - wrong type. expected=STRING, got=%q",
				tt.input, tok.Type)
		}

		if tok.Literal != tt.expected {
			t.Errorf("input=%q - wrong literal. expected=%q, got=%q",
				tt.input, tt.expected, tok.Literal)
		}
	}
}

func TestComments(t *testing.T) {
	input := `# This is a comment
let x = 5 # inline comment`

	l := New(input, "test.sky")

	tok := l.NextToken()
	if tok.Type != COMMENT {
		t.Fatalf("expected COMMENT, got=%q", tok.Type)
	}

	tok = l.NextToken()
	if tok.Type != NEWLINE {
		t.Fatalf("expected NEWLINE, got=%q", tok.Type)
	}

	tok = l.NextToken()
	if tok.Type != LET {
		t.Fatalf("expected LET, got=%q", tok.Type)
	}
}

func TestParenthesesIgnoreIndent(t *testing.T) {
	input := `function test(a,
    b,
    c)
  return a + b + c
end`

	l := New(input, "test.sky")

	// FUNCTION, IDENT(test), LPAREN, IDENT(a), COMMA
	for i := 0; i < 5; i++ {
		l.NextToken()
	}

	// NEWLINE sonrası parantez içinde olduğumuz için INDENT olmamalı
	tok := l.NextToken()
	if tok.Type != NEWLINE {
		t.Fatalf("expected NEWLINE, got %v", tok.Type)
	}

	tok = l.NextToken()
	if tok.Type == INDENT {
		t.Fatalf("should not have INDENT inside parentheses")
	}
}

func TestTokenize(t *testing.T) {
	input := `let x = 42`

	tokens := Tokenize(input, "test.sky")

	if len(tokens) == 0 {
		t.Fatal("expected tokens, got none")
	}

	if tokens[len(tokens)-1].Type != EOF {
		t.Fatal("expected last token to be EOF")
	}
}

func TestPosition(t *testing.T) {
	input := "let x = 5"

	l := New(input, "test.sky")

	tok := l.NextToken() // let
	if tok.Line != 1 || tok.Column != 1 {
		t.Errorf("wrong position for 'let': line=%d, col=%d", tok.Line, tok.Column)
	}

	tok = l.NextToken() // x
	if tok.Line != 1 {
		t.Errorf("wrong line for 'x': %d", tok.Line)
	}
}
