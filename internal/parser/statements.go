package parser

import (
	"github.com/mburakmmm/sky-lang/internal/ast"
	"github.com/mburakmmm/sky-lang/internal/lexer"
)

// parseFunctionStatement fonksiyon tanımlamasını parse eder
func (p *Parser) parseFunctionStatement() *ast.FunctionStatement {
	stmt := &ast.FunctionStatement{Token: p.curToken}

	// Parse decorators (@decorator)
	decorators := []*ast.Decorator{}
	for p.curTokenIs(lexer.AT) {
		decorator := p.parseDecorator()
		if decorator != nil {
			decorators = append(decorators, decorator)
		}
		// Expect NEWLINE after decorator
		if p.peekTokenIs(lexer.NEWLINE) {
			p.nextToken()
		}
		p.nextToken() // Move to next token (could be another @ or function)
	}
	stmt.Decorators = decorators

	// async kontrol
	if p.curTokenIs(lexer.ASYNC) {
		stmt.Async = true
		if !p.expectPeek(lexer.FUNCTION) {
			return nil
		}
	}

	// coop kontrol (coroutine/generator)
	if p.curTokenIs(lexer.COOP) {
		stmt.Coop = true
		if !p.expectPeek(lexer.FUNCTION) {
			return nil
		}
	}

	// Fonksiyon adı
	if !p.expectPeek(lexer.IDENT) {
		return nil
	}
	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// Parametreler (opsiyonel)
	if p.peekTokenIs(lexer.LPAREN) {
		p.nextToken()
		stmt.Parameters = p.parseFunctionParameters()
	}

	// Return type (opsiyonel)
	if p.peekTokenIs(lexer.COLON) {
		p.nextToken() // :
		p.nextToken() // tip
		stmt.ReturnType = p.parseTypeAnnotation()
	}

	// NEWLINE'a kadar ilerle
	if !p.expectPeek(lexer.NEWLINE) {
		return nil
	}

	// Fonksiyon body (INDENT ... DEDENT)
	if !p.expectPeek(lexer.INDENT) {
		return nil
	}

	stmt.Body = p.parseBlockStatement()

	// END token'ını bekle
	if p.peekTokenIs(lexer.END) {
		p.nextToken()
	}

	return stmt
}

// parseDecorator decorator'ı parse eder
func (p *Parser) parseDecorator() *ast.Decorator {
	decorator := &ast.Decorator{Token: p.curToken}

	// Decorator name
	if !p.expectPeek(lexer.IDENT) {
		return nil
	}
	decorator.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// Optional arguments
	if p.peekTokenIs(lexer.LPAREN) {
		p.nextToken() // (
		p.nextToken() // first arg or )

		args := []ast.Expression{}
		if !p.curTokenIs(lexer.RPAREN) {
			args = append(args, p.parseExpression(LOWEST))

			for p.peekTokenIs(lexer.COMMA) {
				p.nextToken() // ,
				p.nextToken() // next arg
				args = append(args, p.parseExpression(LOWEST))
			}

			if !p.expectPeek(lexer.RPAREN) {
				return nil
			}
		}
		decorator.Args = args
	}

	return decorator
}

// parseFunctionParameters fonksiyon parametrelerini parse eder
func (p *Parser) parseFunctionParameters() []*ast.FunctionParameter {
	params := []*ast.FunctionParameter{}

	if p.peekTokenIs(lexer.RPAREN) {
		p.nextToken()
		return params
	}

	p.nextToken()

	// Check for varargs (...)
	if p.curTokenIs(lexer.ELLIPSIS) {
		// Varargs parameter
		if !p.expectPeek(lexer.IDENT) {
			return nil
		}

		param := &ast.FunctionParameter{
			Token:    p.curToken,
			Name:     &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal},
			Variadic: true,
		}

		// Tip anotasyonu
		if p.peekTokenIs(lexer.COLON) {
			p.nextToken() // :
			p.nextToken() // tip
			param.Type = p.parseTypeAnnotation()
		}

		params = append(params, param)

		// Varargs must be last parameter
		if !p.expectPeek(lexer.RPAREN) {
			return nil
		}
		return params
	}

	// İlk parametre (normal)
	param := &ast.FunctionParameter{
		Token: p.curToken,
		Name:  &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal},
	}

	// Tip anotasyonu
	if p.peekTokenIs(lexer.COLON) {
		p.nextToken() // :
		p.nextToken() // tip
		param.Type = p.parseTypeAnnotation()
	}

	// Varsayılan değer
	if p.peekTokenIs(lexer.ASSIGN) {
		p.nextToken() // =
		p.nextToken() // değer
		param.DefaultValue = p.parseExpression(LOWEST)
	}

	params = append(params, param)

	// Diğer parametreler
	for p.peekTokenIs(lexer.COMMA) {
		p.nextToken() // ,
		p.nextToken() // parametre adı veya ...

		// Check for varargs
		if p.curTokenIs(lexer.ELLIPSIS) {
			if !p.expectPeek(lexer.IDENT) {
				return nil
			}

			param := &ast.FunctionParameter{
				Token:    p.curToken,
				Name:     &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal},
				Variadic: true,
			}

			// Tip anotasyonu
			if p.peekTokenIs(lexer.COLON) {
				p.nextToken()
				p.nextToken()
				param.Type = p.parseTypeAnnotation()
			}

			params = append(params, param)

			// Varargs must be last
			if !p.expectPeek(lexer.RPAREN) {
				return nil
			}
			return params
		}

		param := &ast.FunctionParameter{
			Token: p.curToken,
			Name:  &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal},
		}

		if p.peekTokenIs(lexer.COLON) {
			p.nextToken()
			p.nextToken()
			param.Type = p.parseTypeAnnotation()
		}

		if p.peekTokenIs(lexer.ASSIGN) {
			p.nextToken()
			p.nextToken()
			param.DefaultValue = p.parseExpression(LOWEST)
		}

		params = append(params, param)
	}

	if !p.expectPeek(lexer.RPAREN) {
		return nil
	}

	return params
}

// parseTryStatement try-catch-finally statement'ı parse eder
func (p *Parser) parseTryStatement() *ast.TryStatement {
	stmt := &ast.TryStatement{Token: p.curToken}

	// NEWLINE ve INDENT
	if !p.expectPeek(lexer.NEWLINE) {
		return nil
	}
	if !p.expectPeek(lexer.INDENT) {
		return nil
	}

	// Try block
	stmt.TryBlock = p.parseBlockStatement()

	// CATCH clause (optional)
	if p.peekTokenIs(lexer.CATCH) {
		p.nextToken() // catch

		// Error variable (optional)
		var errorVar *ast.Identifier
		if p.peekTokenIs(lexer.IDENT) {
			p.nextToken()
			errorVar = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		}

		if !p.expectPeek(lexer.NEWLINE) {
			return nil
		}
		if !p.expectPeek(lexer.INDENT) {
			return nil
		}

		catchBody := p.parseBlockStatement()
		stmt.CatchClause = &ast.CatchClause{
			ErrorVar: errorVar,
			Body:     catchBody,
		}
	}

	// FINALLY clause (optional)
	if p.peekTokenIs(lexer.FINALLY) {
		p.nextToken() // finally

		if !p.expectPeek(lexer.NEWLINE) {
			return nil
		}
		if !p.expectPeek(lexer.INDENT) {
			return nil
		}

		stmt.Finally = p.parseBlockStatement()
	}

	// END token
	if p.peekTokenIs(lexer.END) {
		p.nextToken()
	}

	return stmt
}

// parseThrowStatement throw statement'ı parse eder
func (p *Parser) parseThrowStatement() *ast.ThrowStatement {
	stmt := &ast.ThrowStatement{Token: p.curToken}

	p.nextToken()
	stmt.Value = p.parseExpression(LOWEST)

	return stmt
}

// parseBlockStatement blok statement'ı parse eder
func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.curToken}
	block.Statements = []ast.Statement{}

	p.nextToken()

	for !p.curTokenIs(lexer.DEDENT) && !p.curTokenIs(lexer.EOF) && !p.curTokenIs(lexer.END) {
		// NEWLINE'ları atla
		if p.curTokenIs(lexer.NEWLINE) {
			p.nextToken()
			continue
		}

		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}

	return block
}

// parseIfStatement if-elif-else statement'ı parse eder
func (p *Parser) parseIfStatement() *ast.IfStatement {
	stmt := &ast.IfStatement{Token: p.curToken}

	// Condition
	p.nextToken()
	stmt.Condition = p.parseExpression(LOWEST)

	if !p.expectPeek(lexer.NEWLINE) {
		return nil
	}

	// Body
	if !p.expectPeek(lexer.INDENT) {
		return nil
	}

	stmt.Consequence = p.parseBlockStatement()

	// Elif dalları
	for p.peekTokenIs(lexer.ELIF) {
		p.nextToken() // elif
		elifClause := &ast.ElifClause{Token: p.curToken}

		p.nextToken()
		elifClause.Condition = p.parseExpression(LOWEST)

		if !p.expectPeek(lexer.NEWLINE) {
			return nil
		}

		if !p.expectPeek(lexer.INDENT) {
			return nil
		}

		elifClause.Consequence = p.parseBlockStatement()
		stmt.Elif = append(stmt.Elif, elifClause)
	}

	// Else dalı
	if p.peekTokenIs(lexer.ELSE) {
		p.nextToken() // else

		if !p.expectPeek(lexer.NEWLINE) {
			return nil
		}

		if !p.expectPeek(lexer.INDENT) {
			return nil
		}

		stmt.Alternative = p.parseBlockStatement()
	}

	// END token
	if p.peekTokenIs(lexer.END) {
		p.nextToken()
	}

	return stmt
}

// parseWhileStatement while döngüsünü parse eder
func (p *Parser) parseWhileStatement() *ast.WhileStatement {
	stmt := &ast.WhileStatement{Token: p.curToken}

	// Condition
	p.nextToken()
	stmt.Condition = p.parseExpression(LOWEST)

	if !p.expectPeek(lexer.NEWLINE) {
		return nil
	}

	// Body
	if !p.expectPeek(lexer.INDENT) {
		return nil
	}

	stmt.Body = p.parseBlockStatement()

	// END token
	if p.peekTokenIs(lexer.END) {
		p.nextToken()
	}

	return stmt
}

// parseForStatement for döngüsünü parse eder
func (p *Parser) parseForStatement() *ast.ForStatement {
	stmt := &ast.ForStatement{Token: p.curToken}

	// Iterator
	if !p.expectPeek(lexer.IDENT) {
		return nil
	}
	stmt.Iterator = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// IN keyword
	if !p.expectPeek(lexer.IN) {
		return nil
	}

	// Iterable
	p.nextToken()
	stmt.Iterable = p.parseExpression(LOWEST)

	if !p.expectPeek(lexer.NEWLINE) {
		return nil
	}

	// Body
	if !p.expectPeek(lexer.INDENT) {
		return nil
	}

	stmt.Body = p.parseBlockStatement()

	// END token
	if p.peekTokenIs(lexer.END) {
		p.nextToken()
	}

	return stmt
}

// parseClassStatement class tanımlamasını parse eder
func (p *Parser) parseClassStatement() *ast.ClassStatement {
	stmt := &ast.ClassStatement{Token: p.curToken}

	// Class adı
	if !p.expectPeek(lexer.IDENT) {
		return nil
	}
	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// SuperClass (opsiyonel)
	if p.peekTokenIs(lexer.LPAREN) {
		p.nextToken() // (
		if !p.expectPeek(lexer.IDENT) {
			return nil
		}
		stmt.SuperClass = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		if !p.expectPeek(lexer.RPAREN) {
			return nil
		}
	}

	if !p.expectPeek(lexer.NEWLINE) {
		return nil
	}

	// Class body
	if !p.expectPeek(lexer.INDENT) {
		return nil
	}

	p.nextToken()

	stmt.Body = []ast.Statement{}
	for !p.curTokenIs(lexer.DEDENT) && !p.curTokenIs(lexer.EOF) && !p.curTokenIs(lexer.END) {
		if p.curTokenIs(lexer.NEWLINE) {
			p.nextToken()
			continue
		}

		member := p.parseStatement()
		if member != nil {
			stmt.Body = append(stmt.Body, member)
		}
		p.nextToken()
	}

	// END token
	if p.peekTokenIs(lexer.END) {
		p.nextToken()
	}

	return stmt
}

// parseImportStatement import ifadesini parse eder
func (p *Parser) parseImportStatement() *ast.ImportStatement {
	stmt := &ast.ImportStatement{Token: p.curToken}

	// Import path
	if !p.expectPeek(lexer.IDENT) {
		return nil
	}

	stmt.Path = []string{p.curToken.Literal}

	// Dotted path (e.g., module.submodule)
	for p.peekTokenIs(lexer.DOT) {
		p.nextToken() // .
		if !p.expectPeek(lexer.IDENT) {
			return nil
		}
		stmt.Path = append(stmt.Path, p.curToken.Literal)
	}

	// AS alias (opsiyonel)
	if p.peekTokenIs(lexer.AS) {
		p.nextToken() // as
		if !p.expectPeek(lexer.IDENT) {
			return nil
		}
		stmt.Alias = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	}

	// NEWLINE'a kadar ilerle
	for !p.curTokenIs(lexer.NEWLINE) && !p.curTokenIs(lexer.EOF) {
		p.nextToken()
	}

	return stmt
}

// parseUnsafeStatement unsafe bloğunu parse eder
func (p *Parser) parseUnsafeStatement() *ast.UnsafeStatement {
	stmt := &ast.UnsafeStatement{Token: p.curToken}

	if !p.expectPeek(lexer.NEWLINE) {
		return nil
	}

	// Body
	if !p.expectPeek(lexer.INDENT) {
		return nil
	}

	stmt.Body = p.parseBlockStatement()

	// END token
	if p.peekTokenIs(lexer.END) {
		p.nextToken()
	}

	return stmt
}
