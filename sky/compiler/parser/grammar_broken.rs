// Grammar Parser - Pratt Parser Implementation
// Token'ları AST'ye dönüştürür, await/yield bağlam kurallarını kontrol eder

use super::ast::*;
use crate::compiler::lexer::{token::Token, token::TokenKind};
use crate::compiler::diag::{Diagnostic, Span, codes};

// Ast struct'ı ast.rs'ye taşındı

pub struct Parser {
    tokens: Vec<Token>,
    current: usize,
    in_async_context: bool,
    in_coop_context: bool,
    source: String,
}

impl Parser {
    pub fn new(tokens: Vec<Token>) -> Self {
        Self {
            tokens,
            current: 0,
            in_async_context: false,
            in_coop_context: false,
            source: String::new(), // Placeholder - gerçek source eklenmeli
        }
    }
    
    pub fn new_with_source(tokens: Vec<Token>, source: String) -> Self {
        Self {
            tokens,
            current: 0,
            in_async_context: false,
            in_coop_context: false,
            source,
        }
    }

    fn is_at_end(&self) -> bool {
        self.peek().kind == TokenKind::Eof
    }

    fn peek(&self) -> &Token {
        &self.tokens[self.current]
    }

    fn peek_ahead(&self, distance: usize) -> Option<&Token> {
        let index = self.current + distance;
        if index < self.tokens.len() {
            Some(&self.tokens[index])
        } else {
            None
        }
    }

    fn previous(&self) -> &Token {
        &self.tokens[self.current - 1]
    }

    fn advance(&mut self) -> &Token {
        if !self.is_at_end() {
            self.current += 1;
        }
        self.previous()
    }

    fn check(&self, kind: TokenKind) -> bool {
        if self.is_at_end() {
            false
        } else {
            self.peek().kind == kind
        }
    }

    fn match_token(&mut self, kinds: &[TokenKind]) -> bool {
        for kind in kinds {
            if self.check(kind.clone()) {
                self.advance();
                return true;
            }
        }
        false
    }

    fn consume(&mut self, kind: TokenKind, message: &str) -> Result<&Token, Diagnostic> {
        if self.check(kind) {
            Ok(self.advance())
        } else {
            Err(Diagnostic::error("E0000", message, self.peek().span))
        }
    }

    fn synchronize(&mut self) {
        self.advance();

        while !self.is_at_end() {
            if self.previous().kind == TokenKind::Newline {
                return;
            }

            match self.peek().kind {
                TokenKind::Function | TokenKind::Async | TokenKind::Coop |
                TokenKind::Var | TokenKind::If | TokenKind::For | TokenKind::While |
                TokenKind::Return | TokenKind::Import => {
                    return;
                }
                TokenKind::Indent | TokenKind::Dedent => {
                    self.advance();
                }
                _ => {
                    self.advance();
                }
            }
        }
    }

    pub fn parse_program(&mut self) -> Result<Ast, Diagnostic> {
        let mut statements = Vec::new();

        while !self.is_at_end() {
            
            // Newline token'larını skip et
            if self.match_token(&[TokenKind::Newline]) {
                continue;
            }
            
            match self.declaration() {
                Ok(stmt) => {
                    statements.push(stmt);
                }
                Err(e) => {
                    self.synchronize();
                }
            }
        }
        Ok(Ast { statements })
    }

    fn declaration(&mut self) -> Result<Stmt, Diagnostic> {
        if self.match_token(&[TokenKind::Function]) {
            self.function_declaration(FuncKind::Normal)
        } else if self.match_token(&[TokenKind::Async]) {
            self.consume(TokenKind::Function, "Expected 'function' after 'async'")?;
            self.function_declaration(FuncKind::Async)
        } else if self.match_token(&[TokenKind::Coop]) {
            self.consume(TokenKind::Function, "Expected 'function' after 'coop'")?;
            self.function_declaration(FuncKind::Coop)
        } else if self.match_token(&[TokenKind::Import]) {
            self.import_declaration()
        } else if self.is_type_declaration() {
            self.var_declaration()
        } else {
            self.statement()
        }
    }

    fn is_type_declaration(&self) -> bool {
        let is_type = matches!(self.peek().kind,
            TokenKind::Var | TokenKind::IntType | TokenKind::FloatType |
            TokenKind::BoolType | TokenKind::StringType | TokenKind::ListType |
            TokenKind::MapType
        );
        is_type
    }

    fn function_declaration(&mut self, kind: FuncKind) -> Result<Stmt, Diagnostic> {
        // Function name
        let source_clone = self.source.clone();
        let name_token = self.consume(TokenKind::Identifier, "Expected function name")?;
        let (name, name_span) = (name_token.text(&source_clone).to_string(), name_token.span);
        
        // Separate scope for next consume
        {
            let _ = self.consume(TokenKind::LeftParen, "Expected '(' after function name")?;
        }

        // Parameters
        let mut params = Vec::new();
        if !self.check(TokenKind::RightParen) {
            loop {
                let source_clone = self.source.clone();
                let param_name = self.consume(TokenKind::Identifier, "Expected parameter name")?;
                let param_name_text = param_name.text(&source_clone).to_string();
                let param_span = param_name.span;
                
                self.consume(TokenKind::Colon, "Expected ':' after parameter name")?;
                let source_clone2 = self.source.clone();
                let type_token = if self.match_token(&[TokenKind::IntType]) {
                    self.previous()
                } else if self.match_token(&[TokenKind::FloatType]) {
                    self.previous()
                } else if self.match_token(&[TokenKind::BoolType]) {
                    self.previous()
                } else if self.match_token(&[TokenKind::StringType]) {
                    self.previous()
                } else if self.match_token(&[TokenKind::ListType]) {
                    self.previous()
                } else if self.match_token(&[TokenKind::MapType]) {
                    self.previous()
                } else if self.match_token(&[TokenKind::Var]) {
                    self.previous()
                } else {
                    return Err(Diagnostic::error("E0000", "Expected parameter type", self.peek().span));
                };
                let type_text = type_token.text(&source_clone2).to_string();
                
                let ty = TypeDecl::from_keyword(&type_text)
                    .ok_or_else(|| Diagnostic::error("E0000", "Invalid type", type_token.span))?;

                params.push(Param {
                    name: param_name_text,
                    ty,
                    span: param_span,
                });

                if !self.match_token(&[TokenKind::Comma]) {
                    break;
                }
            }
        }

        self.consume(TokenKind::RightParen, "Expected ')' after parameters")?;

        // Fonksiyon gövdesi - context'i önce kaydet
        let old_async = {
            let temp = self.in_async_context;
            temp
        };
        let old_coop = {
            let temp = self.in_coop_context;
            temp
        };
        
        // Set context flags
        {
            match kind {
                FuncKind::Async => self.in_async_context = true,
                FuncKind::Coop => self.in_coop_context = true,
                _ => {}
            }
        }

        let body = self.block()?;

        // Reset context flags
        {
            self.in_async_context = old_async;
            self.in_coop_context = old_coop;
        }

        Ok(Stmt::Func {
            kind,
            name,
            params,
            body,
            span: name_span,
        })
    }

    fn import_declaration(&mut self) -> Result<Stmt, Diagnostic> {
        let module_token = self.consume(TokenKind::Identifier, "Expected module name")?;
        let module = module_token.text("").to_string();

        Ok(Stmt::Import {
            module,
            span: module_token.span,
        })
    }

    fn var_declaration(&mut self) -> Result<Stmt, Diagnostic> {
        let source_clone = self.source.clone();
        let type_token = self.advance();
        let (ty, type_span) = {
            let ty = TypeDecl::from_keyword(type_token.text(&source_clone))
                .ok_or_else(|| Diagnostic::error("E0000", "Invalid type", type_token.span))?;
            (ty, type_token.span)
        };

        let name_token = self.consume(TokenKind::Identifier, "Expected variable name")?;
        let name = name_token.text(&source_clone).to_string();

        self.consume(TokenKind::Equal, "Expected '=' after variable name")?;

        let value = self.expression()?;

        Ok(Stmt::VarDecl {
            ty,
            name,
            value,
            span: type_span,
        })
    }

    fn statement(&mut self) -> Result<Stmt, Diagnostic> {
        // Indent, Dedent ve Comment token'larını atla
        if self.check(TokenKind::Indent) {
            self.advance();
            return self.statement();
        }
        if self.check(TokenKind::Dedent) {
            self.advance();
            return self.statement();
        }
        if self.check(TokenKind::Comment) {
            self.advance();
            return self.statement();
        }
        
        // Elif ve Else token'larını skip et - bunlar sadece if statement içinde kullanılır
        if self.check(TokenKind::Elif) || self.check(TokenKind::Else) {
            self.advance();
            return self.statement();
        }
        
        if self.match_token(&[TokenKind::If]) {
            self.if_statement()
        } else if self.match_token(&[TokenKind::For]) {
            self.for_statement()
        } else if self.match_token(&[TokenKind::While]) {
            self.while_statement()
        } else if self.match_token(&[TokenKind::Return]) {
            self.return_statement()
        } else if self.match_token(&[TokenKind::Break]) {
            Ok(Stmt::Break { span: self.previous().span })
        } else if self.match_token(&[TokenKind::Continue]) {
            Ok(Stmt::Continue { span: self.previous().span })
        } else if self.is_type_declaration() {
            self.var_declaration()
        } else {
            self.expression_statement()
        }
    }

    fn block(&mut self) -> Result<Vec<Stmt>, Diagnostic> {
        let mut statements = Vec::new();

        // Indent token'ını atla
        if self.check(TokenKind::Indent) {
            self.advance();
        }

        // Panic recovery mekanizması - infinite loop koruması
        let mut loop_count = 0;
        const MAX_LOOPS: usize = 50; // Daha düşük limit
        let mut last_token = self.peek().kind.clone();

        while !self.check(TokenKind::Dedent) && !self.is_at_end() {
            loop_count += 1;
            
            // Panic recovery: Aynı token'da takılırsa panic et
            if loop_count > MAX_LOOPS {
                // Panic recovery: Token'ı zorla advance et ve loop'u kır
                if !self.is_at_end() {
                    self.advance();
                }
                break; // Loop'u kır
            }
            
            // Token değişikliği kontrolü
            if self.peek().kind != last_token {
                loop_count = 0; // Token değişti, loop count'u sıfırla
            }
            last_token = self.peek().kind.clone();
            
            
            // Newline token'larını atla
            if self.check(TokenKind::Newline) {
                self.advance();
                continue;
            }
            
            // Comment token'larını atla (boş block'lar için)
            if self.check(TokenKind::Comment) {
                self.advance();
                continue;
            }
            
            // Infinite loop'u önlemek için beklenmeyen token'ları atla
            if self.check(TokenKind::Greater) || self.check(TokenKind::Less) || 
               self.check(TokenKind::Equal) || self.check(TokenKind::Plus) ||
               self.check(TokenKind::Minus) || self.check(TokenKind::Star) ||
               self.check(TokenKind::Slash) || self.check(TokenKind::Percent) {
                self.advance();
                continue;
            }
            
            match self.statement() {
                Ok(stmt) => {
                    statements.push(stmt);
                },
                Err(_) => {
                    // Statement parsing başarısız, panic recovery
                    // Token'ı zorla advance et
                    if !self.is_at_end() {
                        self.advance();
                    }
                    
                    // Eğer hala aynı token'daysa veya beklenmeyen token varsa, loop'u kır
                    if !self.is_at_end() && (self.peek().kind == last_token || 
                        self.check(TokenKind::Greater) || self.check(TokenKind::Less)) {
                        break;
                    }
                }
            }
        }

        // Infinite loop koruması
        if loop_count >= MAX_LOOPS {
        }

        // Dedent token'ını atla
        if self.check(TokenKind::Dedent) {
            self.advance();
        }

        Ok(statements)
    }

    fn if_statement(&mut self) -> Result<Stmt, Diagnostic> {
        let condition = self.expression()?;
        self.consume(TokenKind::Colon, "Expected ':' after if condition")?;
        let then_branch = self.block()?;

        let mut elif_branches = Vec::new();
        while self.match_token(&[TokenKind::Elif]) {
            let elif_condition = self.expression()?;
            self.consume(TokenKind::Colon, "Expected ':' after elif condition")?;
            let elif_branch = self.block()?;
            elif_branches.push((elif_condition, elif_branch));
        }

        let else_branch = if self.match_token(&[TokenKind::Else]) {
            self.consume(TokenKind::Colon, "Expected ':' after else")?;
            Some(self.block()?)
        } else {
            None
        };

        let condition_span = condition.span();
        Ok(Stmt::If {
            condition,
            then_branch,
            elif_branches,
            else_branch,
            span: condition_span,
        })
    }

    fn for_statement(&mut self) -> Result<Stmt, Diagnostic> {
        let source = self.source.clone();
        let var_token = self.consume(TokenKind::Identifier, "Expected variable name")?;
        let (variable, var_span) = (var_token.text(&source).to_string(), var_token.span);

        self.consume(TokenKind::Colon, "Expected ':' after variable name")?;
        self.consume(TokenKind::Var, "Expected 'var' for loop variable type")?;
        self.consume(TokenKind::Identifier, "Expected 'in' keyword")?; // 'in' keyword check

        let iterable = self.expression()?;
        
        // Indent token'ını atla
        if self.check(TokenKind::Indent) {
            self.advance();
        }
        
        let mut statements = Vec::new();
        while !self.check(TokenKind::Dedent) && !self.is_at_end() {
            // Newline token'larını atla
            if self.check(TokenKind::Newline) {
                self.advance();
                continue;
            }
            
            // Indent token'ını atla
            if self.check(TokenKind::Indent) {
                self.advance();
                continue;
            }
            
            // Statement'ı parse et
            if let Ok(stmt) = self.statement() {
                statements.push(stmt);
            } else {
                self.synchronize();
            }
        }
        
        // Dedent token'ını atla
        if self.check(TokenKind::Dedent) {
            self.advance();
        }
        

        Ok(Stmt::For {
            variable,
            iterable,
            body: statements,
            span: var_span,
        })
    }

    fn while_statement(&mut self) -> Result<Stmt, Diagnostic> {
        let condition = self.expression()?;
        
        // Indent token'ını atla
        if self.check(TokenKind::Indent) {
            self.advance();
        }
        
        let mut statements = Vec::new();
        while !self.check(TokenKind::Dedent) && !self.is_at_end() {
            // Newline token'larını atla
            if self.check(TokenKind::Newline) {
                self.advance();
                continue;
            }
            
            // Indent token'ını atla
            if self.check(TokenKind::Indent) {
                self.advance();
                continue;
            }
            
            // Statement'ı parse et
            if let Ok(stmt) = self.statement() {
                statements.push(stmt);
            } else {
                self.synchronize();
            }
        }
        
        // Dedent token'ını atla
        if self.check(TokenKind::Dedent) {
            self.advance();
        }
        

        let condition_span = condition.span();
        Ok(Stmt::While {
            condition,
            body: statements,
            span: condition_span,
        })
    }

    fn return_statement(&mut self) -> Result<Stmt, Diagnostic> {
        let value = if !self.check(TokenKind::Newline) {
            Some(self.expression()?)
        } else {
            None
        };

        Ok(Stmt::Return {
            value,
            span: self.previous().span,
        })
    }



    fn expression_statement(&mut self) -> Result<Stmt, Diagnostic> {
        let expr = self.expression()?;
        let expr_span = expr.span();
        Ok(Stmt::ExprStmt {
            expr,
            span: expr_span,
        })
    }

    fn expression(&mut self) -> Result<Expr, Diagnostic> {
        self.parse_precedence(0)
    }

    fn parse_precedence(&mut self, precedence: u8) -> Result<Expr, Diagnostic> {
        let mut expr = self.parse_unary()?;
        //          precedence, self.peek().kind, self.get_precedence());

        while precedence < self.get_precedence() {
            //          precedence, precedence, self.get_precedence());
            let op = self.parse_binary_op()?;
            let mut right_precedence = op.precedence();
            if !op.is_left_associative() {
                right_precedence += 1;
            }
            let right = self.parse_precedence(right_precedence)?;
            let expr_span = expr.span();
            expr = Expr::Binary {
                left: Box::new(expr),
                op,
                right: Box::new(right),
                span: expr_span,
            };
        }

        Ok(expr)
    }

    fn parse_unary(&mut self) -> Result<Expr, Diagnostic> {
        if self.match_token(&[TokenKind::Minus]) {
            let expr = self.parse_unary()?;
            let expr_span = expr.span();
            Ok(Expr::Unary {
                op: UnaryOp::Neg,
                expr: Box::new(expr),
                span: expr_span,
            })
        } else if self.match_token(&[TokenKind::Bang, TokenKind::Not]) {
            let op = self.previous().kind.clone();
            let expr = self.parse_unary()?;
            let expr_span = expr.span();
            Ok(Expr::Unary {
                op: match op {
                    TokenKind::Bang | TokenKind::Not => UnaryOp::Not,
                    _ => unreachable!(),
                },
                expr: Box::new(expr),
                span: expr_span,
            })
        } else if self.match_token(&[TokenKind::Await]) {
            if !self.in_async_context {
                return Err(Diagnostic::error(
                    codes::AWAIT_OUTSIDE_ASYNC,
                    "'await' can only be used in async functions",
                    self.previous().span
                ));
            }
            let expr = self.parse_unary()?;
            Ok(Expr::Await {
                expr: Box::new(expr),
                span: self.previous().span,
            })
        } else if self.match_token(&[TokenKind::Yield]) {
            if !self.in_coop_context {
                return Err(Diagnostic::error(
                    codes::YIELD_OUTSIDE_COOP,
                    "'yield' can only be used in coop functions",
                    self.previous().span
                ));
            }
            let expr = if !self.check(TokenKind::Newline) {
                Some(Box::new(self.expression()?))
            } else {
                None
            };
            Ok(Expr::Yield {
                expr,
                span: self.previous().span,
            })
        } else {
            self.parse_primary_with_calls()
        }
    }

    fn parse_primary_with_calls(&mut self) -> Result<Expr, Diagnostic> {
        if self.match_token(&[TokenKind::True]) {
            Ok(Expr::Lit(Literal::Bool(true)))
        } else if self.match_token(&[TokenKind::False]) {
            Ok(Expr::Lit(Literal::Bool(false)))
        } else if self.match_token(&[TokenKind::Null]) {
            Ok(Expr::Lit(Literal::Null))
        } else if self.match_token(&[TokenKind::Int]) {
            let source_clone = self.source.clone();
            let token = self.previous();
            let value = token.text(&source_clone).parse::<i64>()
                .map_err(|_| Diagnostic::error("E0000", "Invalid integer", token.span))?;
            Ok(Expr::Lit(Literal::Int(value)))
        } else if self.match_token(&[TokenKind::IntType]) {
            // IntType token'ı bir expression değil, type declaration
            return Err(Diagnostic::error("E0000", "Type declaration not allowed in expression", self.previous().span));
        } else if self.match_token(&[TokenKind::Float]) {
            let source_clone = self.source.clone();
            let token = self.previous();
            let value = token.text(&source_clone).parse::<f64>()
                .map_err(|_| Diagnostic::error("E0000", "Invalid float", token.span))?;
            Ok(Expr::Lit(Literal::Float(value)))
        } else if self.match_token(&[TokenKind::String]) {
            let source_clone = self.source.clone();
            let token = self.previous();
            let text = token.text(&source_clone);
            // String literal'ın tırnaklarını çıkar
            let value = if text.len() >= 2 && text.starts_with('"') && text.ends_with('"') {
                text[1..text.len()-1].to_string()
            } else {
                text.to_string()
            };
            Ok(Expr::Lit(Literal::String(value)))
        } else if self.match_token(&[TokenKind::LeftParen]) {
            let expr = self.expression()?;
            self.consume(TokenKind::RightParen, "Expected ')' after expression")?;
            Ok(expr)
        } else if self.match_token(&[TokenKind::LeftBracket]) {
            self.parse_list()
        } else if self.match_token(&[TokenKind::LeftBrace]) {
            self.parse_map()
        } else if self.match_token(&[TokenKind::Identifier]) {
            let source_clone = self.source.clone();
            let token = self.previous();
            let name = token.text(&source_clone).to_string();
            let span = token.span;
            
            // Fonksiyon çağrısı kontrol et
            if self.match_token(&[TokenKind::LeftParen]) {
                let mut args = Vec::new();
                
                if !self.check(TokenKind::RightParen) {
                    loop {
                        args.push(self.expression()?);
                        if !self.match_token(&[TokenKind::Comma]) {
                            break;
                        }
                    }
                }
                
                self.consume(TokenKind::RightParen, "Expected ')' after arguments")?;
                
                Ok(Expr::Call {
                    callee: Box::new(Expr::Ident(name, span)),
                    args,
                    span,
                })
            } else {
                Ok(Expr::Ident(name, span))
            }
        } else {
            Err(Diagnostic::error("E0000", "Expected expression", self.peek().span))
        }
    }

    fn parse_list(&mut self) -> Result<Expr, Diagnostic> {
        let mut elements = Vec::new();

        if !self.check(TokenKind::RightBracket) {
            loop {
                elements.push(self.expression()?);
                if !self.match_token(&[TokenKind::Comma]) {
                    break;
                }
            }
        }

        self.consume(TokenKind::RightBracket, "Expected ']' after list elements")?;
        Ok(Expr::Lit(Literal::List(elements)))
    }

    fn parse_map(&mut self) -> Result<Expr, Diagnostic> {
        let mut pairs = Vec::new();

        if !self.check(TokenKind::RightBrace) {
            loop {
                let source_clone = self.source.clone();
                let key_token = self.consume(TokenKind::String, "Expected string key")?;
                let key = key_token.text(&source_clone).to_string(); // String parsing
                self.consume(TokenKind::Colon, "Expected ':' after key")?;
                let value = self.expression()?;
                pairs.push((key, value));

                if !self.match_token(&[TokenKind::Comma]) {
                    break;
                }
            }
        }

        self.consume(TokenKind::RightBrace, "Expected '}' after map pairs")?;
        Ok(Expr::Lit(Literal::Map(pairs)))
    }

    fn get_precedence(&self) -> u8 {
        if self.is_at_end() {
            return 0;
        }

        let precedence = match self.peek().kind {
            TokenKind::Equal => 1, // Assignment precedence is 1
            TokenKind::Or => BinaryOp::Or.precedence(),
            TokenKind::And => BinaryOp::And.precedence(),
            TokenKind::EqualEqual => BinaryOp::Eq.precedence(),
            TokenKind::BangEqual => BinaryOp::Ne.precedence(),
            TokenKind::Less => BinaryOp::Lt.precedence(),
            TokenKind::LessEqual => BinaryOp::Le.precedence(),
            TokenKind::Greater => BinaryOp::Gt.precedence(),
            TokenKind::GreaterEqual => BinaryOp::Ge.precedence(),
            TokenKind::Plus => BinaryOp::Add.precedence(),
            TokenKind::Minus => BinaryOp::Sub.precedence(),
            TokenKind::Star => BinaryOp::Mul.precedence(),
            TokenKind::Slash => BinaryOp::Div.precedence(),
            TokenKind::Percent => BinaryOp::Mod.precedence(),
            _ => 0,
        };
        precedence
    }

    fn parse_binary_op(&mut self) -> Result<BinaryOp, Diagnostic> {
        let token = self.advance();
        match token.kind {
            TokenKind::Plus => Ok(BinaryOp::Add),
            TokenKind::Minus => Ok(BinaryOp::Sub),
            TokenKind::Star => Ok(BinaryOp::Mul),
            TokenKind::Slash => Ok(BinaryOp::Div),
            TokenKind::Percent => Ok(BinaryOp::Mod),
            TokenKind::EqualEqual => Ok(BinaryOp::Eq),
            TokenKind::BangEqual => Ok(BinaryOp::Ne),
            TokenKind::Less => Ok(BinaryOp::Lt),
            TokenKind::LessEqual => Ok(BinaryOp::Le),
            TokenKind::Greater => Ok(BinaryOp::Gt),
            TokenKind::GreaterEqual => Ok(BinaryOp::Ge),
            TokenKind::Equal => Ok(BinaryOp::Assign),
            TokenKind::And => Ok(BinaryOp::And),
            TokenKind::Or => Ok(BinaryOp::Or),
            _ => Err(Diagnostic::error("E0000", "Expected binary operator", token.span)),
        }
    }
}

// Helper trait for expressions to get span
trait ExprSpan {
    fn span(&self) -> Span;
}

impl ExprSpan for Expr {
    fn span(&self) -> Span {
        match self {
            Expr::Lit(_) => Span::new(0, 0, 0), // Literal span (token'dan alınabilir)
            Expr::Ident(_, span) => *span,
            Expr::Call { span, .. } => *span,
            Expr::Attr { span, .. } => *span,
            Expr::Index { span, .. } => *span,
            Expr::Unary { span, .. } => *span,
            Expr::Binary { span, .. } => *span,
            Expr::Await { span, .. } => *span,
            Expr::Yield { span, .. } => *span,
            Expr::Interpolated { span, .. } => *span,
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::compiler::lexer::lex;

    #[test]
    fn test_arithmetic_parsing() {
        let tokens = lex("1 + 2 * 3").unwrap();
        let mut parser = Parser::new(tokens);
        let expr = parser.expression().unwrap();
        
        // Should parse as 1 + (2 * 3) due to precedence
        if let Expr::Binary { left, op, right, .. } = expr {
            assert_eq!(*op, BinaryOp::Add);
            if let Expr::Binary { op: right_op, .. } = *right {
                assert_eq!(right_op, BinaryOp::Mul);
            } else {
                panic!("Expected multiplication in right operand");
            }
        } else {
            panic!("Expected binary expression");
        }
    }

    #[test]
    fn test_await_context_check() {
        let tokens = lex("await sleep(1)").unwrap();
        let mut parser = Parser::new(tokens);
        let result = parser.expression();
        assert!(result.is_err());
        if let Err(diag) = result {
            assert_eq!(diag.code, codes::AWAIT_OUTSIDE_ASYNC);
        }
    }

    fn parse_unary(&mut self) -> Result<Expr, Diagnostic> {
        if self.match_token(&[TokenKind::Minus, TokenKind::Not]) {
            let op = self.previous().kind;
            let right = self.parse_unary()?;
            let span = self.previous().span;
            Ok(Expr::Unary {
                op: match op {
                    TokenKind::Minus => UnaryOp::Neg,
                    TokenKind::Not => UnaryOp::Not,
                    _ => unreachable!(),
                },
                right: Box::new(right),
                span,
            })
        } else {
            self.parse_primary_with_calls()
        }
    }

    fn parse_primary_with_calls(&mut self) -> Result<Expr, Diagnostic> {
        if self.match_token(&[TokenKind::True]) {
            Ok(Expr::Lit(Literal::Bool(true)))
        } else if self.match_token(&[TokenKind::False]) {
            Ok(Expr::Lit(Literal::Bool(false)))
        } else if self.match_token(&[TokenKind::Null]) {
            Ok(Expr::Lit(Literal::Null))
        } else if self.match_token(&[TokenKind::Int]) {
            let source_clone = self.source.clone();
            let token = self.previous();
            let value = token.text(&source_clone).parse::<i64>()
                .map_err(|_| Diagnostic::error("E0000", "Invalid integer", token.span))?;
            Ok(Expr::Lit(Literal::Int(value)))
        } else if self.match_token(&[TokenKind::Float]) {
            let source_clone = self.source.clone();
            let token = self.previous();
            let value = token.text(&source_clone).parse::<f64>()
                .map_err(|_| Diagnostic::error("E0000", "Invalid float", token.span))?;
            Ok(Expr::Lit(Literal::Float(value)))
        } else if self.match_token(&[TokenKind::String]) {
            let source_clone = self.source.clone();
            let token = self.previous();
            let value = token.text(&source_clone).to_string();
            Ok(Expr::Lit(Literal::String(value)))
        } else if self.match_token(&[TokenKind::LeftParen]) {
            let expr = self.expression()?;
            self.consume(TokenKind::RightParen, "Expected ')' after expression")?;
            Ok(expr)
        } else if self.match_token(&[TokenKind::Identifier]) {
            let source_clone = self.source.clone();
            let token = self.previous();
            let name = token.text(&source_clone).to_string();
            let span = token.span;
            
            // Fonksiyon çağrısı kontrol et
            if self.match_token(&[TokenKind::LeftParen]) {
                let mut args = Vec::new();
                
                if !self.check(TokenKind::RightParen) {
                    loop {
                        args.push(self.expression()?);
                        if !self.match_token(&[TokenKind::Comma]) {
                            break;
                        }
                    }
                }
                
                self.consume(TokenKind::RightParen, "Expected ')' after arguments")?;
                
                Ok(Expr::Call {
                    callee: Box::new(Expr::Ident(name, span)),
                    args,
                    span,
                })
            } else {
                Ok(Expr::Ident(name, span))
            }
        } else {
            Err(Diagnostic::error("E0000", "Expected expression", self.peek().span))
        }
    }

    fn parse_binary_op(&mut self) -> Result<BinaryOp, Diagnostic> {
        let token = self.advance();
        match token.kind {
            TokenKind::Plus => Ok(BinaryOp::Add),
            TokenKind::Minus => Ok(BinaryOp::Sub),
            TokenKind::Star => Ok(BinaryOp::Mul),
            TokenKind::Slash => Ok(BinaryOp::Div),
            TokenKind::Percent => Ok(BinaryOp::Mod),
            TokenKind::EqualEqual => Ok(BinaryOp::Eq),
            TokenKind::BangEqual => Ok(BinaryOp::Ne),
            TokenKind::Less => Ok(BinaryOp::Lt),
            TokenKind::LessEqual => Ok(BinaryOp::Le),
            TokenKind::Greater => Ok(BinaryOp::Gt),
            TokenKind::GreaterEqual => Ok(BinaryOp::Ge),
            TokenKind::Equal => Ok(BinaryOp::Assign),
            TokenKind::And => Ok(BinaryOp::And),
            TokenKind::Or => Ok(BinaryOp::Or),
            _ => Err(Diagnostic::error("E0000", "Expected binary operator", token.span)),
        }
    }

    fn get_precedence(&self) -> u8 {
        let precedence = match self.peek().kind {
            TokenKind::Equal => 1, // Assignment precedence is 1
            TokenKind::Or => 1,
            TokenKind::And => 2,
            TokenKind::EqualEqual | TokenKind::BangEqual => 3,
            TokenKind::Greater | TokenKind::GreaterEqual | TokenKind::Less | TokenKind::LessEqual => 4,
            TokenKind::Plus | TokenKind::Minus => 5,
            TokenKind::Star | TokenKind::Slash | TokenKind::Percent => 6,
            _ => 0,
        };
        precedence
    }

    fn peek(&self) -> &Token {
        &self.tokens[self.current]
    }

    fn previous(&self) -> &Token {
        &self.tokens[self.current - 1]
    }

    fn is_at_end(&self) -> bool {
        self.peek().kind == TokenKind::Eof
    }

    fn advance(&mut self) -> &Token {
        if !self.is_at_end() {
            self.current += 1;
        }
        self.previous()
    }

    fn check(&self, kind: TokenKind) -> bool {
        if self.is_at_end() {
            false
        } else {
            self.peek().kind == kind
        }
    }

    fn match_token(&mut self, kinds: &[TokenKind]) -> bool {
        for kind in kinds {
            if self.check(*kind) {
                self.advance();
                return true;
            }
        }
        false
    }

    fn consume(&mut self, kind: TokenKind, message: &str) -> Result<&Token, Diagnostic> {
        if self.check(kind) {
            Ok(self.advance())
        } else {
            Err(Diagnostic::error("E0000", message, self.peek().span))
        }
    }

    fn synchronize(&mut self) {
        // Hata sonrası senkronizasyon - token'ları skip et
        while !self.is_at_end() {
            if self.previous().kind == TokenKind::Newline {
                return;
            }
            
            match self.peek().kind {
                TokenKind::Function | TokenKind::Async | TokenKind::Coop |
                TokenKind::Import | TokenKind::Var | TokenKind::IntType |
                TokenKind::FloatType | TokenKind::BoolType | TokenKind::StringType |
                TokenKind::ListType | TokenKind::MapType => {
                    return;
                }
                _ => {
                    self.advance();
                }
            }
        }
    }

    #[test]
    fn test_yield_context_check() {
        let tokens = lex("yield 1").unwrap();
        let mut parser = Parser::new(tokens);
        let result = parser.expression();
        assert!(result.is_err());
        if let Err(diag) = result {
            assert_eq!(diag.code, codes::YIELD_OUTSIDE_COOP);
        }
    }
}
