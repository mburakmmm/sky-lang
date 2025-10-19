package lexer

import "fmt"

// TokenType token tiplerini temsil eder
type TokenType int

const (
	// Special tokens
	ILLEGAL TokenType = iota
	EOF
	COMMENT

	// Literals
	IDENT  // main, x, y, myFunction
	INT    // 123, 0xFF, 0b1010
	FLOAT  // 3.14, 1.0e10
	STRING // "hello", 'world'

	// Keywords
	FUNCTION // function
	END      // end
	CLASS    // class
	LET      // let
	CONST    // const
	IF       // if
	ELIF     // elif
	ELSE     // else
	FOR      // for
	WHILE    // while
	RETURN   // return
	BREAK    // break
	CONTINUE // continue
	ASYNC    // async
	AWAIT    // await
	COOP     // coop
	YIELD    // yield
	UNSAFE   // unsafe
	SELF     // self
	SUPER    // super
	IMPORT   // import
	AS       // as
	IN       // in
	TRUE     // true
	FALSE    // false
	AND      // and
	OR       // or
	NOT      // not
	ENUM     // enum
	MATCH    // match
	TRY      // try
	CATCH    // catch
	FINALLY  // finally
	THROW    // throw

	// Operators
	PLUS    // +
	MINUS   // -
	STAR    // *
	SLASH   // /
	PERCENT // %
	POWER   // **

	EQ // ==
	NE // !=
	LT // <
	LE // <=
	GT // >
	GE // >=

	LAND // &&
	LOR  // ||
	LNOT // !

	ASSIGN    // =
	PLUSEQ    // +=
	MINUSEQ   // -=
	STAREQ    // *=
	SLASHEQ   // /=
	PERCENTEQ // %=

	// Delimiters
	LPAREN   // (
	RPAREN   // )
	LBRACK   // [
	RBRACK   // ]
	LBRACE   // {
	RBRACE   // }
	DOT      // .
	COMMA    // ,
	COLON    // :
	ARROW    // =>
	ELLIPSIS // ...
	AT       // @
	QUESTION // ?
	PIPE     // |

	// Special indentation tokens
	NEWLINE // \n (significant newlines)
	INDENT  // increase in indentation
	DEDENT  // decrease in indentation
)

var keywords = map[string]TokenType{
	"function": FUNCTION,
	"end":      END,
	"class":    CLASS,
	"let":      LET,
	"const":    CONST,
	"if":       IF,
	"elif":     ELIF,
	"else":     ELSE,
	"for":      FOR,
	"while":    WHILE,
	"return":   RETURN,
	"break":    BREAK,
	"continue": CONTINUE,
	"async":    ASYNC,
	"await":    AWAIT,
	"coop":     COOP,
	"yield":    YIELD,
	"unsafe":   UNSAFE,
	"self":     SELF,
	"super":    SUPER,
	"import":   IMPORT,
	"as":       AS,
	"in":       IN,
	"true":     TRUE,
	"false":    FALSE,
	"and":      AND,
	"or":       OR,
	"not":      NOT,
	"enum":     ENUM,
	"match":    MATCH,
	"try":      TRY,
	"catch":    CATCH,
	"finally":  FINALLY,
	"throw":    THROW,
}

// LookupIdent identifier'ın keyword olup olmadığını kontrol eder
func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}

// Token kaynak koddaki bir token'ı temsil eder
type Token struct {
	Type    TokenType // token tipi
	Literal string    // literal değer
	Line    int       // satır numarası (1'den başlar)
	Column  int       // sütun numarası (1'den başlar)
	File    string    // dosya adı (opsiyonel)
}

// NewToken yeni bir token oluşturur
func NewToken(typ TokenType, literal string, line, column int) Token {
	return Token{
		Type:    typ,
		Literal: literal,
		Line:    line,
		Column:  column,
	}
}

// String token'ı string olarak döndürür (debugging için)
func (t Token) String() string {
	return fmt.Sprintf("Token{Type: %s, Literal: %q, Line: %d, Column: %d}",
		t.Type.String(), t.Literal, t.Line, t.Column)
}

// String TokenType'ı string olarak döndürür
func (tt TokenType) String() string {
	names := map[TokenType]string{
		ILLEGAL: "ILLEGAL",
		EOF:     "EOF",
		COMMENT: "COMMENT",

		IDENT:  "IDENT",
		INT:    "INT",
		FLOAT:  "FLOAT",
		STRING: "STRING",

		FUNCTION: "FUNCTION",
		END:      "END",
		CLASS:    "CLASS",
		LET:      "LET",
		CONST:    "CONST",
		IF:       "IF",
		ELIF:     "ELIF",
		ELSE:     "ELSE",
		FOR:      "FOR",
		WHILE:    "WHILE",
		RETURN:   "RETURN",
		ASYNC:    "ASYNC",
		AWAIT:    "AWAIT",
		COOP:     "COOP",
		YIELD:    "YIELD",
		UNSAFE:   "UNSAFE",
		SELF:     "SELF",
		SUPER:    "SUPER",
		IMPORT:   "IMPORT",
		AS:       "AS",
		IN:       "IN",
		TRUE:     "TRUE",
		FALSE:    "FALSE",
		AND:      "AND",
		OR:       "OR",
		NOT:      "NOT",
		ENUM:     "ENUM",
		MATCH:    "MATCH",
		TRY:      "TRY",
		CATCH:    "CATCH",
		FINALLY:  "FINALLY",
		THROW:    "THROW",
		BREAK:    "BREAK",
		CONTINUE: "CONTINUE",

		PLUS:    "PLUS",
		MINUS:   "MINUS",
		STAR:    "STAR",
		SLASH:   "SLASH",
		PERCENT: "PERCENT",
		POWER:   "POWER",

		EQ: "EQ",
		NE: "NE",
		LT: "LT",
		LE: "LE",
		GT: "GT",
		GE: "GE",

		LAND: "LAND",
		LOR:  "LOR",
		LNOT: "LNOT",

		ASSIGN:    "ASSIGN",
		PLUSEQ:    "PLUSEQ",
		MINUSEQ:   "MINUSEQ",
		STAREQ:    "STAREQ",
		SLASHEQ:   "SLASHEQ",
		PERCENTEQ: "PERCENTEQ",

		LPAREN:   "LPAREN",
		RPAREN:   "RPAREN",
		LBRACK:   "LBRACK",
		RBRACK:   "RBRACK",
		LBRACE:   "LBRACE",
		RBRACE:   "RBRACE",
		DOT:      "DOT",
		COMMA:    "COMMA",
		COLON:    "COLON",
		ARROW:    "ARROW",
		ELLIPSIS: "ELLIPSIS",
		AT:       "AT",
		QUESTION: "QUESTION",
		PIPE:     "PIPE",

		NEWLINE: "NEWLINE",
		INDENT:  "INDENT",
		DEDENT:  "DEDENT",
	}

	if name, ok := names[tt]; ok {
		return name
	}
	return fmt.Sprintf("TokenType(%d)", tt)
}

// IsKeyword token'ın keyword olup olmadığını kontrol eder
func (tt TokenType) IsKeyword() bool {
	return tt >= FUNCTION && tt <= NOT
}

// IsOperator token'ın operator olup olmadığını kontrol eder
func (tt TokenType) IsOperator() bool {
	return tt >= PLUS && tt <= PERCENTEQ
}

// IsLiteral token'ın literal olup olmadığını kontrol eder
func (tt TokenType) IsLiteral() bool {
	return tt >= IDENT && tt <= STRING
}

// Position token'ın pozisyon bilgisini döndürür
func (t Token) Position() string {
	if t.File != "" {
		return fmt.Sprintf("%s:%d:%d", t.File, t.Line, t.Column)
	}
	return fmt.Sprintf("%d:%d", t.Line, t.Column)
}
