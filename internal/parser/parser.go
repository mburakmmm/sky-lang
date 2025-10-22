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
	lexer.ARROW:     ASSIGN, // => has same precedence as assignment
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
	p.registerPrefix(lexer.VOID, p.parseIdentifier)
	p.registerPrefix(lexer.ANY, p.parseIdentifier)
	p.registerPrefix(lexer.MATCH, p.parseMatchExpression)
	p.registerPrefix(lexer.FUNCTION, p.parseLambdaExpression)
	p.registerPrefix(lexer.NEWLINE, p.parseEmptyExpression)
	p.registerPrefix(lexer.RBRACE, p.parseEmptyExpression)
	p.registerPrefix(lexer.COLON, p.parseEmptyExpression)
	p.registerPrefix(lexer.COMMA, p.parseEmptyExpression)
	p.registerPrefix(lexer.RPAREN, p.parseEmptyExpression)
	p.registerPrefix(lexer.INDENT, p.parseEmptyExpression)

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
	p.registerInfix(lexer.ARROW, p.parseArrowExpression)

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
		// NEWLINE, DEDENT, END'leri atla (top-level'da olmamalı)
		if p.curTokenIs(lexer.NEWLINE) || p.curTokenIs(lexer.DEDENT) || p.curTokenIs(lexer.END) {
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
	case lexer.FUNCTION, lexer.ASYNC, lexer.COOP:
		return p.parseFunctionStatement()
	case lexer.ABSTRACT:
		if p.peekTokenIs(lexer.FUNCTION) {
			return p.parseAbstractMethodStatement()
		}
		return p.parseAbstractClassStatement()
	case lexer.STATIC:
		if p.peekTokenIs(lexer.FUNCTION) {
			return p.parseStaticMethodStatement()
		}
		return p.parseStaticPropertyStatement()
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
	case lexer.TRY:
		return p.parseTryStatement()
	case lexer.THROW:
		return p.parseThrowStatement()
	case lexer.AT:
		// Decorator followed by function
		return p.parseFunctionStatement()
	case lexer.COMMENT:
		// COMMENT'leri atla
		return nil
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

func (p *Parser) parseLambdaExpression() ast.Expression {
	// Lambda syntax: function(x) x * x end
	lambda := &ast.LambdaExpression{Token: p.curToken}

	// Parametreler
	if !p.expectPeek(lexer.LPAREN) {
		return nil
	}
	lambda.Parameters = p.parseFunctionParameters()

	// Return type (opsiyonel)
	if p.peekTokenIs(lexer.COLON) {
		p.nextToken() // :
		p.nextToken() // tip
		lambda.ReturnType = p.parseTypeAnnotation()
	}

	// Lambda body - single expression
	p.nextToken() // move to first token of expression

	// Parse expression as return value
	returnExpr := p.parseExpression(LOWEST)
	if returnExpr == nil {
		return nil
	}

	// Lambda body oluştur
	lambda.Body = &ast.BlockStatement{
		Token: p.curToken,
		Statements: []ast.Statement{
			&ast.ReturnStatement{
				Token:       p.curToken,
				ReturnValue: returnExpr,
			},
		},
	}

	// END token'ını bekle
	if !p.expectPeek(lexer.END) {
		return nil
	}

	return lambda
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
	var baseType ast.TypeAnnotation

	switch p.curToken.Type {
	case lexer.STAR:
		// Pointer type: *T
		p.nextToken()
		pointeeType := p.parseTypeAnnotation()
		baseType = &ast.PointerType{
			Token:       p.curToken,
			PointeeType: pointeeType,
		}
	case lexer.FUNCTION:
		// Function type: function(ArgTypes) -> ReturnType
		token := p.curToken
		p.nextToken() // function

		if !p.expectPeek(lexer.LPAREN) {
			return nil
		}
		p.nextToken() // (

		// Parse parameter types
		paramTypes := []ast.TypeAnnotation{}
		if !p.curTokenIs(lexer.RPAREN) {
			paramTypes = append(paramTypes, p.parseTypeAnnotation())
			for p.peekTokenIs(lexer.COMMA) {
				p.nextToken() // ,
				p.nextToken() // next type
				paramTypes = append(paramTypes, p.parseTypeAnnotation())
			}
		}

		if !p.expectPeek(lexer.RPAREN) {
			return nil
		}

		// Parse return type (optional)
		var returnType ast.TypeAnnotation
		if p.peekTokenIs(lexer.RARROW) {
			p.nextToken() // ->
			p.nextToken() // return type
			returnType = p.parseTypeAnnotation()
		}

		baseType = &ast.FunctionType{
			Token:      token,
			ParamTypes: paramTypes,
			ReturnType: returnType,
		}
	case lexer.LPAREN:
		// Check if this is a function type: (ArgTypes) => ReturnType
		// or a grouped expression: (expr)
		token := p.curToken
		p.nextToken() // (

		// Look ahead to see if this is a function type
		// Function type: (type1, type2) => returnType
		// Grouped expr: (expr)
		if p.curTokenIs(lexer.RPAREN) {
			// Empty parameter list: () => returnType
			p.nextToken() // )
			if p.peekTokenIs(lexer.ARROW) {
				// This is a function type: () => returnType
				p.nextToken() // =>
				p.nextToken() // return type
				returnType := p.parseTypeAnnotation()
				return &ast.FunctionType{
					Token:      token,
					ParamTypes: []ast.TypeAnnotation{},
					ReturnType: returnType,
				}
			} else {
				// This is a grouped expression, not a type annotation
				// We're in the wrong context - this shouldn't happen in parseTypeAnnotation
				return nil
			}
		} else {
			// Check if first token is a type (IDENT)
			if p.curTokenIs(lexer.IDENT) {
				// This could be a function type: (type1, type2) => returnType
				// Parse parameter types
				paramTypes := []ast.TypeAnnotation{}
				paramTypes = append(paramTypes, p.parseTypeAnnotation())
				for p.peekTokenIs(lexer.COMMA) {
					p.nextToken() // ,
					p.nextToken() // next type
					paramTypes = append(paramTypes, p.parseTypeAnnotation())
				}

				if !p.expectPeek(lexer.RPAREN) {
					return nil
				}

				// Check if this is followed by =>
				if p.peekTokenIs(lexer.ARROW) {
					// This is a function type
					p.nextToken() // =>
					p.nextToken() // return type
					returnType := p.parseTypeAnnotation()
					return &ast.FunctionType{
						Token:      token,
						ParamTypes: paramTypes,
						ReturnType: returnType,
					}
				} else {
					// This is a grouped expression, not a type annotation
					return nil
				}
			} else {
				// This is a grouped expression, not a type annotation
				return nil
			}
		}
	case lexer.VOID:
		baseType = &ast.BasicType{
			Token: p.curToken,
			Name:  "void",
		}
	case lexer.ANY:
		baseType = &ast.BasicType{
			Token: p.curToken,
			Name:  "any",
		}
	case lexer.IDENT:
		// Basic types or Generic: int, float, List<T>
		typeName := p.curToken.Literal
		token := p.curToken

		// Check for generic type: List<T>
		if p.peekTokenIs(lexer.LT) {
			p.nextToken() // <
			p.nextToken() // first type arg

			typeArgs := []ast.TypeAnnotation{}
			typeArgs = append(typeArgs, p.parseTypeAnnotation())

			for p.peekTokenIs(lexer.COMMA) {
				p.nextToken() // ,
				p.nextToken() // next type arg
				typeArgs = append(typeArgs, p.parseTypeAnnotation())
			}

			if !p.expectPeek(lexer.GT) {
				return nil
			}

			baseType = &ast.GenericType{
				Token:    token,
				BaseName: typeName,
				TypeArgs: typeArgs,
			}
		} else {
			baseType = &ast.BasicType{
				Token: token,
				Name:  typeName,
			}
		}
	case lexer.LBRACK:
		// List type: [T]
		p.nextToken()
		elemType := p.parseTypeAnnotation()
		if !p.expectPeek(lexer.RBRACK) {
			return nil
		}
		baseType = &ast.ListType{
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
		baseType = &ast.DictType{
			Token:     p.curToken,
			KeyType:   keyType,
			ValueType: valueType,
		}
	default:
		p.addError(fmt.Sprintf("unexpected type: %s", p.curToken.Literal))
		return nil
	}

	// Check for optional type: T?
	if p.peekTokenIs(lexer.QUESTION) {
		p.nextToken()
		return &ast.OptionalType{
			Token:    p.curToken,
			BaseType: baseType,
		}
	}

	// Check for union type: T1|T2
	if p.peekTokenIs(lexer.PIPE) {
		types := []ast.TypeAnnotation{baseType}

		for p.peekTokenIs(lexer.PIPE) {
			p.nextToken() // |
			p.nextToken() // next type
			nextType := p.parseTypeAnnotation()
			types = append(types, nextType)
		}

		return &ast.UnionType{
			Token: p.curToken,
			Types: types,
		}
	}

	return baseType
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
		// Skip NEWLINE and COMMENT tokens
		if p.curTokenIs(lexer.NEWLINE) || p.curTokenIs(lexer.COMMENT) {
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

// parseArrowExpression parses arrow expressions (=>)
func (p *Parser) parseArrowExpression(left ast.Expression) ast.Expression {
	expr := &ast.ArrowExpression{
		Token: p.curToken,
		Left:  left,
	}

	precedence := p.curPrecedence()
	p.nextToken()

	expr.Right = p.parseExpression(precedence)
	return expr
}

// parseEmptyExpression parses empty expressions (for NEWLINE, RBRACE)
func (p *Parser) parseEmptyExpression() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: ""}
}

// parseMatchExpression parses match expression
func (p *Parser) parseMatchExpression() ast.Expression {
	expr := &ast.MatchExpression{Token: p.curToken}

	p.nextToken()
	expr.Value = p.parseExpression(LOWEST)

	// Check for brace syntax: match value { pattern => result }
	if p.peekTokenIs(lexer.LBRACE) {
		p.nextToken() // consume {

		expr.Arms = []*ast.MatchArm{}

		for !p.curTokenIs(lexer.RBRACE) && !p.curTokenIs(lexer.EOF) {
			// Skip NEWLINE and COMMENT tokens
			if p.curTokenIs(lexer.NEWLINE) || p.curTokenIs(lexer.COMMENT) {
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

			// Skip NEWLINE after arm
			if p.peekTokenIs(lexer.NEWLINE) {
				p.nextToken()
			}
		}

		if !p.expectPeek(lexer.RBRACE) {
			p.addError("expected } in match expression")
			return expr
		}

		return expr
	}

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
		// Skip NEWLINE and COMMENT tokens
		if p.curTokenIs(lexer.NEWLINE) || p.curTokenIs(lexer.COMMENT) {
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

// parseAbstractClassStatement abstract class tanımlamasını parse eder
func (p *Parser) parseAbstractClassStatement() *ast.AbstractClassStatement {
	stmt := &ast.AbstractClassStatement{Token: p.curToken}

	// Abstract class adı
	if !p.expectPeek(lexer.CLASS) {
		return nil
	}
	if !p.expectPeek(lexer.IDENT) {
		return nil
	}
	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// Multiple inheritance (abstract class Duck : Flying, Swimming)
	if p.peekTokenIs(lexer.COLON) {
		p.nextToken() // :
		p.nextToken() // first superclass

		stmt.SuperClasses = []*ast.Identifier{}
		stmt.SuperClasses = append(stmt.SuperClasses, &ast.Identifier{
			Token: p.curToken,
			Value: p.curToken.Literal,
		})

		// Additional superclasses (comma-separated)
		for p.peekTokenIs(lexer.COMMA) {
			p.nextToken() // ,
			p.nextToken() // next superclass
			stmt.SuperClasses = append(stmt.SuperClasses, &ast.Identifier{
				Token: p.curToken,
				Value: p.curToken.Literal,
			})
		}
	}

	// NEWLINE'a kadar ilerle
	if !p.expectPeek(lexer.NEWLINE) {
		return nil
	}

	// COMMENT'leri atla ve INDENT'ı bul
	for {
		if p.peekTokenIs(lexer.COMMENT) {
			p.nextToken() // COMMENT
			if p.peekTokenIs(lexer.NEWLINE) {
				p.nextToken() // NEWLINE
			}
		} else if p.peekTokenIs(lexer.INDENT) {
			p.nextToken() // INDENT
			break
		} else {
			p.peekError(lexer.INDENT)
			return nil
		}
	}

	stmt.Body = p.parseBlockStatement().Statements

	return stmt
}

// parseAbstractMethodStatement abstract method tanımlamasını parse eder
func (p *Parser) parseAbstractMethodStatement() *ast.AbstractMethodStatement {
	stmt := &ast.AbstractMethodStatement{Token: p.curToken}

	// Abstract function adı
	if !p.expectPeek(lexer.FUNCTION) {
		return nil
	}
	if !p.expectPeek(lexer.IDENT) {
		return nil
	}
	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// Parametreler
	if p.peekTokenIs(lexer.LPAREN) {
		stmt.Parameters = p.parseFunctionParameters()
	}

	// Return type
	if p.peekTokenIs(lexer.COLON) {
		p.nextToken()
		p.nextToken()
		stmt.ReturnType = p.parseTypeAnnotation()
	}

	return stmt
}

// parseStaticMethodStatement static method tanımlamasını parse eder
func (p *Parser) parseStaticMethodStatement() *ast.StaticMethodStatement {
	stmt := &ast.StaticMethodStatement{Token: p.curToken}

	// Static function adı
	if !p.expectPeek(lexer.FUNCTION) {
		return nil
	}
	if !p.expectPeek(lexer.IDENT) {
		return nil
	}
	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// Parametreler
	if p.peekTokenIs(lexer.LPAREN) {
		p.nextToken() // (
		stmt.Parameters = p.parseFunctionParameters()
	}

	// Return type
	if p.peekTokenIs(lexer.COLON) {
		p.nextToken()
		p.nextToken()
		stmt.ReturnType = p.parseTypeAnnotation()
	}

	// NEWLINE'a kadar ilerle
	if !p.expectPeek(lexer.NEWLINE) {
		return nil
	}

	// COMMENT'leri atla ve INDENT'ı bul
	for {
		if p.peekTokenIs(lexer.COMMENT) {
			p.nextToken() // COMMENT
			if p.peekTokenIs(lexer.NEWLINE) {
				p.nextToken() // NEWLINE
			}
		} else if p.peekTokenIs(lexer.INDENT) {
			p.nextToken() // INDENT
			break
		} else {
			p.peekError(lexer.INDENT)
			return nil
		}
	}

	stmt.Body = p.parseBlockStatement()

	return stmt
}

// parseStaticPropertyStatement static property tanımlamasını parse eder
func (p *Parser) parseStaticPropertyStatement() *ast.StaticPropertyStatement {
	stmt := &ast.StaticPropertyStatement{Token: p.curToken}

	// Static let adı
	if !p.expectPeek(lexer.LET) {
		return nil
	}
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

	// Değer
	if p.peekTokenIs(lexer.ASSIGN) {
		p.nextToken() // =
		p.nextToken() // değer
		stmt.Value = p.parseExpression(LOWEST)
	}

	return stmt
}
