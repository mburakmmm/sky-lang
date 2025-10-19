package lexer

import (
	"fmt"
	"strings"
	"unicode"
)

// Lexer SKY kaynak kodunu tokenize eder
type Lexer struct {
	input        string // kaynak kod
	filename     string // dosya adı
	position     int    // mevcut pozisyon (current char)
	readPosition int    // sonraki pozisyon (after current char)
	ch           byte   // mevcut karakter
	line         int    // mevcut satır
	column       int    // mevcut sütun

	// Indentation tracking
	indentStack   []int   // indent level stack
	pendingTokens []Token // bekleyen tokenler (DEDENT için)
	atLineStart   bool    // satır başında mıyız?

	// Parantez tracking (parantez içinde indent önemsiz)
	parenDepth int // (), [], {} derinliği
}

// New yeni bir Lexer oluşturur
func New(input, filename string) *Lexer {
	l := &Lexer{
		input:       input,
		filename:    filename,
		line:        1,
		column:      0,
		indentStack: []int{0}, // başlangıç indent seviyesi 0
		atLineStart: true,
	}
	l.readChar() // ilk karakteri oku
	return l
}

// readChar bir sonraki karakteri okur
func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0 // EOF
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++
	l.column++
}

// peekChar bir sonraki karaktere bakar ama ilerlemez
func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

// peekCharN n karakter ileriyi bakar
func (l *Lexer) peekCharN(n int) byte {
	pos := l.position + n
	if pos >= len(l.input) {
		return 0
	}
	return l.input[pos]
}

// NextToken bir sonraki token'ı döndürür
func (l *Lexer) NextToken() Token {
	// Bekleyen tokenler varsa (DEDENT gibi) önce onları döndür
	if len(l.pendingTokens) > 0 {
		tok := l.pendingTokens[0]
		l.pendingTokens = l.pendingTokens[1:]
		return tok
	}

	// Satır başındaysak ve parantez içinde değilsek indent'i kontrol et
	if l.atLineStart && l.parenDepth == 0 && l.ch != 0 {
		return l.handleIndentation()
	}

	l.skipWhitespace() // boşlukları atla (newline hariç)

	tok := l.makeToken(ILLEGAL, string(l.ch))

	switch l.ch {
	case 0:
		// EOF'ta tüm açık indent'leri kapat
		if len(l.indentStack) > 1 {
			l.indentStack = l.indentStack[:len(l.indentStack)-1]
			tok = l.makeToken(DEDENT, "")
			// Kalan DEDENT'ler için bekleyen tokenler oluştur
			for len(l.indentStack) > 1 {
				l.pendingTokens = append(l.pendingTokens, l.makeToken(DEDENT, ""))
				l.indentStack = l.indentStack[:len(l.indentStack)-1]
			}
			return tok
		}
		tok = l.makeToken(EOF, "")

	case '\n':
		tok = l.makeToken(NEWLINE, "\\n")
		l.readChar()
		l.line++
		l.column = 0
		l.atLineStart = true

	case '#':
		tok = l.scanComment()

	case '"', '\'':
		tok = l.scanString()

	case '+':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = l.makeToken(PLUSEQ, string(ch)+string(l.ch))
			l.readChar()
		} else {
			tok = l.makeToken(PLUS, string(l.ch))
			l.readChar()
		}

	case '-':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = l.makeToken(MINUSEQ, string(ch)+string(l.ch))
			l.readChar()
		} else {
			tok = l.makeToken(MINUS, string(l.ch))
			l.readChar()
		}

	case '*':
		if l.peekChar() == '*' {
			l.readChar()
			tok = l.makeToken(POWER, "**")
			l.readChar()
		} else if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = l.makeToken(STAREQ, string(ch)+string(l.ch))
			l.readChar()
		} else {
			tok = l.makeToken(STAR, string(l.ch))
			l.readChar()
		}

	case '/':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = l.makeToken(SLASHEQ, string(ch)+string(l.ch))
			l.readChar()
		} else {
			tok = l.makeToken(SLASH, string(l.ch))
			l.readChar()
		}

	case '%':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = l.makeToken(PERCENTEQ, string(ch)+string(l.ch))
			l.readChar()
		} else {
			tok = l.makeToken(PERCENT, string(l.ch))
			l.readChar()
		}

	case '=':
		if l.peekChar() == '=' {
			l.readChar()
			tok = l.makeToken(EQ, "==")
			l.readChar()
		} else if l.peekChar() == '>' {
			l.readChar()
			tok = l.makeToken(ARROW, "=>")
			l.readChar()
		} else {
			tok = l.makeToken(ASSIGN, string(l.ch))
			l.readChar()
		}

	case '!':
		if l.peekChar() == '=' {
			l.readChar()
			tok = l.makeToken(NE, "!=")
			l.readChar()
		} else {
			tok = l.makeToken(LNOT, string(l.ch))
			l.readChar()
		}

	case '<':
		if l.peekChar() == '=' {
			l.readChar()
			tok = l.makeToken(LE, "<=")
			l.readChar()
		} else {
			tok = l.makeToken(LT, string(l.ch))
			l.readChar()
		}

	case '>':
		if l.peekChar() == '=' {
			l.readChar()
			tok = l.makeToken(GE, ">=")
			l.readChar()
		} else {
			tok = l.makeToken(GT, string(l.ch))
			l.readChar()
		}

	case '&':
		if l.peekChar() == '&' {
			l.readChar()
			tok = l.makeToken(LAND, "&&")
			l.readChar()
		} else {
			tok = l.makeToken(ILLEGAL, string(l.ch))
			l.readChar()
		}

	case '(':
		tok = l.makeToken(LPAREN, string(l.ch))
		l.parenDepth++
		l.readChar()

	case ')':
		tok = l.makeToken(RPAREN, string(l.ch))
		l.parenDepth--
		l.readChar()

	case '[':
		tok = l.makeToken(LBRACK, string(l.ch))
		l.parenDepth++
		l.readChar()

	case ']':
		tok = l.makeToken(RBRACK, string(l.ch))
		l.parenDepth--
		l.readChar()

	case '{':
		tok = l.makeToken(LBRACE, string(l.ch))
		l.parenDepth++
		l.readChar()

	case '}':
		tok = l.makeToken(RBRACE, string(l.ch))
		l.parenDepth--
		l.readChar()

	case '.':
		// Check for ... (ellipsis)
		if l.peekChar() == '.' && l.peekCharN(2) == '.' {
			l.readChar()
			l.readChar()
			tok = l.makeToken(ELLIPSIS, "...")
			l.readChar()
		} else {
			tok = l.makeToken(DOT, string(l.ch))
			l.readChar()
		}

	case ',':
		tok = l.makeToken(COMMA, string(l.ch))
		l.readChar()

	case ':':
		tok = l.makeToken(COLON, string(l.ch))
		l.readChar()

	case '@':
		tok = l.makeToken(AT, string(l.ch))
		l.readChar()

	case '?':
		tok = l.makeToken(QUESTION, string(l.ch))
		l.readChar()

	case '|':
		if l.peekChar() == '|' {
			l.readChar()
			tok = l.makeToken(LOR, "||")
			l.readChar()
		} else {
			tok = l.makeToken(PIPE, string(l.ch))
			l.readChar()
		}

	default:
		if isLetter(l.ch) {
			return l.scanIdentifier()
		} else if isDigit(l.ch) {
			return l.scanNumber()
		} else {
			tok = l.makeToken(ILLEGAL, string(l.ch))
			l.readChar()
		}
	}

	return tok
}

// handleIndentation satır başındaki indentation'ı işler
func (l *Lexer) handleIndentation() Token {
	l.atLineStart = false

	// Boş satırları ve sadece yorum içeren satırları atla
	if l.ch == '\n' || l.ch == '#' {
		return l.NextToken()
	}

	// Indent seviyesini hesapla
	indentLevel := 0
	startPos := l.position

	for l.ch == ' ' || l.ch == '\t' {
		if l.ch == ' ' {
			indentLevel++
		} else {
			indentLevel += 4 // tab = 4 space
		}
		l.readChar()
	}

	// Boş satır ise indent önemsiz
	if l.ch == '\n' || l.ch == '#' || l.ch == 0 {
		return l.NextToken()
	}

	currentIndent := l.indentStack[len(l.indentStack)-1]

	if indentLevel > currentIndent {
		// Indent arttı
		l.indentStack = append(l.indentStack, indentLevel)
		tok := l.makeToken(INDENT, "")
		tok.Column = startPos + 1
		return tok
	} else if indentLevel < currentIndent {
		// Indent azaldı - DEDENT token'ları oluştur
		dedentCount := 0
		for len(l.indentStack) > 1 && l.indentStack[len(l.indentStack)-1] > indentLevel {
			l.indentStack = l.indentStack[:len(l.indentStack)-1]
			dedentCount++
		}

		// Indent seviyesi stack'te yoksa hata
		if l.indentStack[len(l.indentStack)-1] != indentLevel {
			return l.makeToken(ILLEGAL, fmt.Sprintf("indentation error at line %d", l.line))
		}

		// İlk DEDENT'i döndür, kalanları kuyruğa ekle
		tok := l.makeToken(DEDENT, "")
		tok.Column = startPos + 1
		for i := 1; i < dedentCount; i++ {
			l.pendingTokens = append(l.pendingTokens, l.makeToken(DEDENT, ""))
		}
		return tok
	}

	// Aynı indent seviyesi - normal token'a devam et
	return l.NextToken()
}

// scanIdentifier identifier veya keyword tarar
func (l *Lexer) scanIdentifier() Token {
	start := l.position
	startCol := l.column

	for isLetter(l.ch) || isDigit(l.ch) || l.ch == '_' {
		l.readChar()
	}

	literal := l.input[start:l.position]
	tokenType := LookupIdent(literal)

	tok := NewToken(tokenType, literal, l.line, startCol)
	tok.File = l.filename
	return tok
}

// scanNumber sayı tarar (integer veya float)
func (l *Lexer) scanNumber() Token {
	start := l.position
	startCol := l.column
	tokenType := INT

	// Hex, binary, octal sayılar
	if l.ch == '0' {
		if l.peekChar() == 'x' || l.peekChar() == 'X' {
			l.readChar() // '0'
			l.readChar() // 'x'
			for isHexDigit(l.ch) {
				l.readChar()
			}
			literal := l.input[start:l.position]
			return NewToken(INT, literal, l.line, startCol)
		} else if l.peekChar() == 'b' || l.peekChar() == 'B' {
			l.readChar() // '0'
			l.readChar() // 'b'
			for l.ch == '0' || l.ch == '1' {
				l.readChar()
			}
			literal := l.input[start:l.position]
			return NewToken(INT, literal, l.line, startCol)
		} else if l.peekChar() == 'o' || l.peekChar() == 'O' {
			l.readChar() // '0'
			l.readChar() // 'o'
			for l.ch >= '0' && l.ch <= '7' {
				l.readChar()
			}
			literal := l.input[start:l.position]
			return NewToken(INT, literal, l.line, startCol)
		}
	}

	// Normal sayı
	for isDigit(l.ch) {
		l.readChar()
	}

	// Float kontrolü
	if l.ch == '.' && isDigit(l.peekChar()) {
		tokenType = FLOAT
		l.readChar() // '.'
		for isDigit(l.ch) {
			l.readChar()
		}
	}

	// Exponent
	if l.ch == 'e' || l.ch == 'E' {
		tokenType = FLOAT
		l.readChar()
		if l.ch == '+' || l.ch == '-' {
			l.readChar()
		}
		for isDigit(l.ch) {
			l.readChar()
		}
	}

	literal := l.input[start:l.position]
	tok := NewToken(tokenType, literal, l.line, startCol)
	tok.File = l.filename
	return tok
}

// scanString string literal tarar
func (l *Lexer) scanString() Token {
	startCol := l.column
	quote := l.ch
	l.readChar() // opening quote

	var sb strings.Builder

	for l.ch != quote && l.ch != 0 && l.ch != '\n' {
		if l.ch == '\\' {
			l.readChar()
			switch l.ch {
			case 'n':
				sb.WriteByte('\n')
			case 't':
				sb.WriteByte('\t')
			case 'r':
				sb.WriteByte('\r')
			case '\\':
				sb.WriteByte('\\')
			case '"':
				sb.WriteByte('"')
			case '\'':
				sb.WriteByte('\'')
			case '0':
				sb.WriteByte(0)
			default:
				sb.WriteByte(l.ch)
			}
			l.readChar()
		} else {
			sb.WriteByte(l.ch)
			l.readChar()
		}
	}

	if l.ch != quote {
		tok := NewToken(ILLEGAL, "unterminated string", l.line, startCol)
		tok.File = l.filename
		return tok
	}

	l.readChar() // closing quote

	tok := NewToken(STRING, sb.String(), l.line, startCol)
	tok.File = l.filename
	return tok
}

// scanComment yorum tarar
func (l *Lexer) scanComment() Token {
	start := l.position
	startCol := l.column

	for l.ch != '\n' && l.ch != 0 {
		l.readChar()
	}

	literal := l.input[start:l.position]
	tok := NewToken(COMMENT, literal, l.line, startCol)
	tok.File = l.filename
	return tok
}

// skipWhitespace boşlukları atlar (newline hariç)
func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\r' {
		l.readChar()
	}
}

// makeToken yeni bir token oluşturur
func (l *Lexer) makeToken(typ TokenType, literal string) Token {
	tok := NewToken(typ, literal, l.line, l.column)
	tok.File = l.filename
	return tok
}

// Helper functions

func isLetter(ch byte) bool {
	return unicode.IsLetter(rune(ch)) || ch == '_'
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func isHexDigit(ch byte) bool {
	return isDigit(ch) || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F')
}

// Tokenize tüm kaynak kodu tokenize eder ve token listesi döndürür
func Tokenize(input, filename string) []Token {
	l := New(input, filename)
	tokens := []Token{}

	for {
		tok := l.NextToken()
		tokens = append(tokens, tok)
		if tok.Type == EOF {
			break
		}
	}

	return tokens
}
