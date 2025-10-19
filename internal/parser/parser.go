package parser

import (
	"fmt"
	"strconv"

	"github.com/mburakmmm/sky-lang/internal/ast"
	"github.com/mburakmmm/sky-lang/internal/lexer"
)

// Operator öncelikleri
const (
	_ int = iota
	LOWEST
	ASSIGN      // =, +=, -=, etc.
	LOR         // ||
	LAND        // &&
	EQUALS      // ==, !=
	LESSGREATER // >, <, >=, <=
	SUM         // +, -
	PRODUCT     // *, /, %
	POWER       // **
	PREFIX      // -X, !X
	CALL        // myFunction(X)
	INDEX       // array[index], obj.member
)

var precedences = map[lexer.TokenType]int{
	lexer.ASSIGN:    ASSIGN,
	lexer.PLUSEQ:    ASSIGN,
	lexer.MINUSEQ:   ASSIGN,
	lexer.STAREQ:    ASSIGN,
	lexer.SLASHEQ:   ASSIGN,
	lexer.PERCENTEQ: ASSIGN,
	lexer.LOR:       LOR,
	lexer.LAND:      LAND,
	lexer.EQ:        EQUALS,
	lexer.NE:        EQUALS,
	lexer.LT:        LESSGREATER,
	lexer.LE:        LESSGREATER,
	lexer.GT:        LESSGREATER,
	lexer.GE:        LESSGREATER,
	lexer.PLUS:      SUM,
	lexer.MINUS:     SUM,
	lexer.SLASH:     PRODUCT,
	lexer.STAR:      PRODUCT,
	lexer.PERCENT:   PRODUCT,
	lexer.POWER:     POWER,
	lexer.LPAREN:    CALL,
	lexer.LBRACK:    INDEX,
	lexer.DOT:       INDEX,
}

// Parser sözdizimi analizörü
type Parser struct {
	l      *lexer.Lexer
	errors []string

	curToken  lexer.Token
	peekToken lexer.Token

	prefixParseFns map[lexer.TokenType]prefixParseFn
	infixParseFns  map[lexer.TokenType]infixParseFn
}

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

// New yeni bir parser oluşturur
func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}

	// Prefix parse fonksiyonları
	p.prefixParseFns = make(map[lexer.TokenType]prefixParseFn)
	p.registerPrefix(lexer.IDENT, p.parseIdentifier)
	p.registerPrefix(lexer.INT, p.parseIntegerLiteral)
	p.registerPrefix(lexer.FLOAT, p.parseFloatLiteral)
	p.registerPrefix(lexer.STRING, p.parseStringLiteral)
	p.registerPrefix(lexer.TRUE, p.parseBooleanLiteral)
	p.registerPrefix(lexer.FALSE, p.parseBooleanLiteral)
	p.registerPrefix(lexer.LNOT, p.parsePrefixExpression)
	p.registerPrefix(lexer.MINUS, p.parsePrefixExpression)
	p.registerPrefix(lexer.PLUS, p.parsePrefixExpression)
	p.registerPrefix(lexer.LPAREN, p.parseGroupedExpression)
	p.registerPrefix(lexer.LBRACK, p.parseListLiteral)
	p.registerPrefix(lexer.LBRACE, p.parseDictLiteral)
	p.registerPrefix(lexer.AWAIT, p.parseAwaitExpression)
	p.registerPrefix(lexer.YIELD, p.parseYieldExpression)
	p.registerPrefix(lexer.SELF, p.parseIdentifier)
	p.registerPrefix(lexer.SUPER, p.parseIdentifier)
	p.registerPrefix(lexer.MATCH, p.parseMatchExpression)

	// Infix parse fonksiyonları
	p.infixParseFns = make(map[lexer.TokenType]infixParseFn)
	p.registerInfix(lexer.PLUS, p.parseInfixExpression)
	p.registerInfix(lexer.MINUS, p.parseInfixExpression)
	p.registerInfix(lexer.SLASH, p.parseInfixExpression)
	p.registerInfix(lexer.STAR, p.parseInfixExpression)
	p.registerInfix(lexer.PERCENT, p.parseInfixExpression)
	p.registerInfix(lexer.POWER, p.parseInfixExpression)
	p.registerInfix(lexer.EQ, p.parseInfixExpression)
	p.registerInfix(lexer.NE, p.parseInfixExpression)
	p.registerInfix(lexer.LT, p.parseInfixExpression)
	p.registerInfix(lexer.LE, p.parseInfixExpression)
	p.registerInfix(lexer.GT, p.parseInfixExpression)
	p.registerInfix(lexer.GE, p.parseInfixExpression)
	p.registerInfix(lexer.LAND, p.parseInfixExpression)
	p.registerInfix(lexer.LOR, p.parseInfixExpression)
	p.registerInfix(lexer.ASSIGN, p.parseInfixExpression)
	p.registerInfix(lexer.PLUSEQ, p.parseInfixExpression)
	p.registerInfix(lexer.MINUSEQ, p.parseInfixExpression)
	p.registerInfix(lexer.STAREQ, p.parseInfixExpression)
	p.registerInfix(lexer.SLASHEQ, p.parseInfixExpression)
	p.registerInfix(lexer.PERCENTEQ, p.parseInfixExpression)
	p.registerInfix(lexer.LPAREN, p.parseCallExpression)
	p.registerInfix(lexer.LBRACK, p.parseIndexExpression)
	p.registerInfix(lexer.DOT, p.parseMemberExpression)

	// İlk iki token'ı oku
	p.nextToken()
	p.nextToken()

	return p
}

// Errors parse hatalarını döndürür
func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) peekError(t lexer.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead at %s",
		t, p.peekToken.Type, p.peekToken.Position())
	p.errors = append(p.errors, msg)
}

func (p *Parser) addError(msg string) {
	p.errors = append(p.errors, fmt.Sprintf("%s at %s", msg, p.curToken.Position()))
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()

	// Yorumları atla
	for p.peekToken.Type == lexer.COMMENT {
		p.peekToken = p.l.NextToken()
	}
}

func (p *Parser) curTokenIs(t lexer.TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t lexer.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) expectPeek(t lexer.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}
	p.peekError(t)
	return false
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) registerPrefix(tokenType lexer.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType lexer.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

// ParseProgram kaynak kodu parse eder
func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for !p.curTokenIs(lexer.EOF) {
		// NEWLINE'ları atla
		if p.curTokenIs(lexer.NEWLINE) {
			p.nextToken()
			continue
		}

		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}

	return program
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case lexer.LET:
		return p.parseLetStatement()
	case lexer.CONST:
		return p.parseConstStatement()
	case lexer.RETURN:
		return p.parseReturnStatement()
	case lexer.BREAK:
		return p.parseBreakStatement()
	case lexer.CONTINUE:
		return p.parseContinueStatement()
	case lexer.FUNCTION, lexer.ASYNC:
		return p.parseFunctionStatement()
	case lexer.IF:
		return p.parseIfStatement()
	case lexer.WHILE:
		return p.parseWhileStatement()
	case lexer.FOR:
		return p.parseForStatement()
	case lexer.CLASS:
		return p.parseClassStatement()
	case lexer.IMPORT:
		return p.parseImportStatement()
	case lexer.UNSAFE:
		return p.parseUnsafeStatement()
	case lexer.ENUM:
		return p.parseEnumStatement()
	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseLetStatement() *ast.LetStatement {
	stmt := &ast.LetStatement{Token: p.curToken}

	if !p.expectPeek(lexer.IDENT) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// Tip anotasyonu
	if p.peekTokenIs(lexer.COLON) {
		p.nextToken() // :
		p.nextToken() // tip
		stmt.Type = p.parseTypeAnnotation()
	}

	if !p.expectPeek(lexer.ASSIGN) {
		return nil
	}

	p.nextToken()
	stmt.Value = p.parseExpression(LOWEST)

	// NEWLINE'a kadar ilerle
	for !p.curTokenIs(lexer.NEWLINE) && !p.curTokenIs(lexer.EOF) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseConstStatement() *ast.ConstStatement {
	stmt := &ast.ConstStatement{Token: p.curToken}

	if !p.expectPeek(lexer.IDENT) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// Tip anotasyonu
	if p.peekTokenIs(lexer.COLON) {
		p.nextToken() // :
		p.nextToken() // tip
		stmt.Type = p.parseTypeAnnotation()
	}

	if !p.expectPeek(lexer.ASSIGN) {
		return nil
	}

	p.nextToken()
	stmt.Value = p.parseExpression(LOWEST)

	for !p.curTokenIs(lexer.NEWLINE) && !p.curTokenIs(lexer.EOF) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.curToken}

	p.nextToken()

	// Return değeri opsiyonel
	if !p.curTokenIs(lexer.NEWLINE) && !p.curTokenIs(lexer.EOF) {
		stmt.ReturnValue = p.parseExpression(LOWEST)
	}

	for !p.curTokenIs(lexer.NEWLINE) && !p.curTokenIs(lexer.EOF) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseBreakStatement() *ast.BreakStatement {
	stmt := &ast.BreakStatement{Token: p.curToken}

	// Skip to end of line
	for !p.curTokenIs(lexer.NEWLINE) && !p.curTokenIs(lexer.EOF) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseContinueStatement() *ast.ContinueStatement {
	stmt := &ast.ContinueStatement{Token: p.curToken}

	// Skip to end of line
	for !p.curTokenIs(lexer.NEWLINE) && !p.curTokenIs(lexer.EOF) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.curToken}
	stmt.Expression = p.parseExpression(LOWEST)

	// NEWLINE opsiyonel
	if p.peekTokenIs(lexer.NEWLINE) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	leftExp := prefix()

	for !p.peekTokenIs(lexer.NEWLINE) && !p.peekTokenIs(lexer.EOF) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}

		p.nextToken()
		leftExp = infix(leftExp)
	}

	return leftExp
}

func (p *Parser) noPrefixParseFnError(t lexer.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found", t)
	p.errors = append(p.errors, msg)
}

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Token: p.curToken}

	var value int64
	var err error

	// Farklı sayı formatları
	if len(p.curToken.Literal) > 2 {
		switch p.curToken.Literal[:2] {
		case "0x", "0X":
			value, err = strconv.ParseInt(p.curToken.Literal[2:], 16, 64)
		case "0b", "0B":
			value, err = strconv.ParseInt(p.curToken.Literal[2:], 2, 64)
		case "0o", "0O":
			value, err = strconv.ParseInt(p.curToken.Literal[2:], 8, 64)
		default:
			value, err = strconv.ParseInt(p.curToken.Literal, 10, 64)
		}
	} else {
		value, err = strconv.ParseInt(p.curToken.Literal, 10, 64)
	}

	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	lit.Value = value
	return lit
}

func (p *Parser) parseFloatLiteral() ast.Expression {
	lit := &ast.FloatLiteral{Token: p.curToken}

	value, err := strconv.ParseFloat(p.curToken.Literal, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as float", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	lit.Value = value
	return lit
}

func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseBooleanLiteral() ast.Expression {
	return &ast.BooleanLiteral{Token: p.curToken, Value: p.curTokenIs(lexer.TRUE)}
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}

	p.nextToken()
	expression.Right = p.parseExpression(PREFIX)

	return expression
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}

	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)

	return expression
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()

	exp := p.parseExpression(LOWEST)

	if !p.expectPeek(lexer.RPAREN) {
		return nil
	}

	return exp
}

func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	exp := &ast.CallExpression{Token: p.curToken, Function: function}
	exp.Arguments = p.parseExpressionList(lexer.RPAREN)
	return exp
}

func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	exp := &ast.IndexExpression{Token: p.curToken, Left: left}

	p.nextToken()
	exp.Index = p.parseExpression(LOWEST)

	if !p.expectPeek(lexer.RBRACK) {
		return nil
	}

	return exp
}

func (p *Parser) parseMemberExpression(object ast.Expression) ast.Expression {
	exp := &ast.MemberExpression{Token: p.curToken, Object: object}

	if !p.expectPeek(lexer.IDENT) {
		return nil
	}

	exp.Member = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	return exp
}

func (p *Parser) parseListLiteral() ast.Expression {
	list := &ast.ListLiteral{Token: p.curToken}
	list.Elements = p.parseExpressionList(lexer.RBRACK)
	return list
}

func (p *Parser) parseDictLiteral() ast.Expression {
	dict := &ast.DictLiteral{Token: p.curToken}
	dict.Pairs = make(map[ast.Expression]ast.Expression)

	for !p.peekTokenIs(lexer.RBRACE) && !p.peekTokenIs(lexer.EOF) {
		p.nextToken()
		key := p.parseExpression(LOWEST)

		if !p.expectPeek(lexer.COLON) {
			return nil
		}

		p.nextToken()
		value := p.parseExpression(LOWEST)

		dict.Pairs[key] = value

		if !p.peekTokenIs(lexer.RBRACE) && !p.expectPeek(lexer.COMMA) {
			return nil
		}
	}

	if !p.expectPeek(lexer.RBRACE) {
		return nil
	}

	return dict
}

func (p *Parser) parseExpressionList(end lexer.TokenType) []ast.Expression {
	list := []ast.Expression{}

	if p.peekTokenIs(end) {
		p.nextToken()
		return list
	}

	p.nextToken()
	list = append(list, p.parseExpression(LOWEST))

	for p.peekTokenIs(lexer.COMMA) {
		p.nextToken()
		p.nextToken()
		list = append(list, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(end) {
		return nil
	}

	return list
}

func (p *Parser) parseAwaitExpression() ast.Expression {
	exp := &ast.AwaitExpression{Token: p.curToken}

	p.nextToken()
	exp.Expression = p.parseExpression(LOWEST)

	return exp
}

func (p *Parser) parseYieldExpression() ast.Expression {
	exp := &ast.YieldExpression{Token: p.curToken}

	if !p.peekTokenIs(lexer.NEWLINE) && !p.peekTokenIs(lexer.EOF) {
		p.nextToken()
		exp.Value = p.parseExpression(LOWEST)
	}

	return exp
}

// parseTypeAnnotation tip anotasyonunu parse eder
func (p *Parser) parseTypeAnnotation() ast.TypeAnnotation {
	switch p.curToken.Type {
	case lexer.STAR:
		// Pointer type: *T
		p.nextToken()
		pointeeType := p.parseTypeAnnotation()
		return &ast.PointerType{
			Token:       p.curToken,
			PointeeType: pointeeType,
		}
	case lexer.IDENT:
		// Basic types: int, float, string, bool, any
		return &ast.BasicType{
			Token: p.curToken,
			Name:  p.curToken.Literal,
		}
	case lexer.LBRACK:
		// List type: [T]
		p.nextToken()
		elemType := p.parseTypeAnnotation()
		if !p.expectPeek(lexer.RBRACK) {
			return nil
		}
		return &ast.ListType{
			Token:       p.curToken,
			ElementType: elemType,
		}
	case lexer.LBRACE:
		// Dict type: {K: V}
		p.nextToken()
		keyType := p.parseTypeAnnotation()
		if !p.expectPeek(lexer.COLON) {
			return nil
		}
		p.nextToken()
		valueType := p.parseTypeAnnotation()
		if !p.expectPeek(lexer.RBRACE) {
			return nil
		}
		return &ast.DictType{
			Token:     p.curToken,
			KeyType:   keyType,
			ValueType: valueType,
		}
	default:
		p.addError(fmt.Sprintf("unexpected type: %s", p.curToken.Literal))
		return nil
	}
}

// parseEnumStatement parses enum declaration
func (p *Parser) parseEnumStatement() *ast.EnumStatement {
	stmt := &ast.EnumStatement{Token: p.curToken}

	if !p.expectPeek(lexer.IDENT) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	stmt.Variants = []*ast.EnumVariant{}

	// Skip NEWLINE
	for p.peekTokenIs(lexer.NEWLINE) {
		p.nextToken()
	}

	// Skip INDENT
	if p.peekTokenIs(lexer.INDENT) {
		p.nextToken()
		p.nextToken() // Move to first variant
	}

	// Parse variants
	for !p.curTokenIs(lexer.END) && !p.curTokenIs(lexer.DEDENT) && !p.curTokenIs(lexer.EOF) {
		if p.curTokenIs(lexer.NEWLINE) {
			p.nextToken()
			continue
		}

		if p.curTokenIs(lexer.DEDENT) {
			break
		}

		if !p.curTokenIs(lexer.IDENT) {
			break
		}

		variant := &ast.EnumVariant{
			Name: &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal},
		}

		// Check for payload
		if p.peekTokenIs(lexer.LPAREN) {
			p.nextToken() // (
			p.nextToken() // first type

			variant.Payload = []ast.TypeAnnotation{}

			if !p.curTokenIs(lexer.RPAREN) {
				variant.Payload = append(variant.Payload, p.parseTypeAnnotation())

				for p.peekTokenIs(lexer.COMMA) {
					p.nextToken() // ,
					p.nextToken() // type
					variant.Payload = append(variant.Payload, p.parseTypeAnnotation())
				}

				if !p.expectPeek(lexer.RPAREN) {
					return nil
				}
			}
		}

		stmt.Variants = append(stmt.Variants, variant)
		p.nextToken()
	}

	// Move past DEDENT if present
	if p.curTokenIs(lexer.DEDENT) {
		p.nextToken()
	}

	// Expect END
	if !p.curTokenIs(lexer.END) {
		p.addError("expected 'end' to close enum")
		return stmt
	}

	return stmt
}

// parseMatchExpression parses match expression
func (p *Parser) parseMatchExpression() ast.Expression {
	expr := &ast.MatchExpression{Token: p.curToken}

	p.nextToken()
	expr.Value = p.parseExpression(LOWEST)

	// Skip NEWLINE
	for p.peekTokenIs(lexer.NEWLINE) {
		p.nextToken()
	}

	// Skip INDENT
	if p.peekTokenIs(lexer.INDENT) {
		p.nextToken()
		p.nextToken() // Move to first pattern
	}

	// Parse match arms
	expr.Arms = []*ast.MatchArm{}

	for !p.curTokenIs(lexer.END) && !p.curTokenIs(lexer.DEDENT) && !p.curTokenIs(lexer.EOF) {
		if p.curTokenIs(lexer.NEWLINE) {
			p.nextToken()
			continue
		}

		arm := &ast.MatchArm{}

		// Parse pattern
		arm.Pattern = p.parseExpression(LOWEST)

		// Expect ARROW =>
		if !p.expectPeek(lexer.ARROW) {
			p.addError("expected => in match arm")
			return expr
		}

		// Parse body - single inline expression
		p.nextToken()

		body := &ast.BlockStatement{
			Token:      p.curToken,
			Statements: []ast.Statement{},
		}

		exprStmt := &ast.ExpressionStatement{
			Token:      p.curToken,
			Expression: p.parseExpression(LOWEST),
		}
		body.Statements = append(body.Statements, exprStmt)
		arm.Body = body

		expr.Arms = append(expr.Arms, arm)
		p.nextToken()
	}

	return expr
}
