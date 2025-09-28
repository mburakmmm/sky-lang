use crate::compiler::{
    ast::{Ast, BinaryOp, Expr, Stmt, UnaryOp, FuncKind},
    diag::{Diagnostic, Span},
    lexer::{Token, TokenKind},
};

pub struct Parser {
    tokens: Vec<Token>,
    current: usize,
}

impl Parser {
    pub fn new(tokens: Vec<Token>) -> Self {
        Self { tokens, current: 0 }
    }

    pub fn parse(&mut self) -> Result<Ast, Diagnostic> {
        let mut statements = Vec::new();

        while !self.is_at_end() {
            if let Some(stmt) = self.declaration()? {
                statements.push(stmt);
            }
        }

        Ok(Ast { statements })
    }

    fn is_at_end(&self) -> bool {
        self.current >= self.tokens.len() || self.peek().kind == TokenKind::Eof
    }

    fn peek(&self) -> &Token {
        &self.tokens[self.current]
    }

    fn advance(&mut self) -> &Token {
        if !self.is_at_end() {
            self.current += 1;
        }
        self.previous()
    }

    fn previous(&self) -> &Token {
        &self.tokens[self.current - 1]
    }

    fn check(&self, kind: TokenKind) -> bool {
        if self.is_at_end() {
            false
        } else {
            self.peek().kind == kind
        }
    }

    fn match_kinds(&mut self, kinds: &[TokenKind]) -> bool {
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
            Err(Diagnostic::error(self.peek().span, message.to_string()))
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

    fn is_type_declaration(&self) -> bool {
        matches!(self.peek().kind,
            TokenKind::Var | TokenKind::IntType | TokenKind::FloatType |
            TokenKind::BoolType | TokenKind::StringType | TokenKind::ListType |
            TokenKind::MapType
        )
    }

    pub fn parse_program(&mut self) -> Result<Ast, Diagnostic> {
        let mut statements = Vec::new();

        while !self.is_at_end() {
            // Newline token'larını skip et
            if self.match_kinds(&[TokenKind::Newline]) {
                continue;
            }
            
            match self.declaration() {
                Ok(Some(stmt)) => {
                    statements.push(stmt);
                }
                Ok(None) => {
                    // Empty declaration, continue
                }
                Err(e) => {
                    self.synchronize();
                }
            }
        }
        Ok(Ast { statements })
    }

    fn declaration(&mut self) -> Result<Option<Stmt>, Diagnostic> {
        if self.check(TokenKind::KwFunction) {
            self.function_declaration()
        } else if self.check(TokenKind::KwAsync) {
            self.async_function_declaration()
        } else if self.check(TokenKind::KwCoop) {
            self.coop_function_declaration()
        } else if self.check(TokenKind::KwImport) {
            self.import_declaration()
        } else if self.is_type_declaration() {
            self.variable_declaration()
        } else {
            self.statement()
        }
    }

    fn function_declaration(&mut self) -> Result<Option<Stmt>, Diagnostic> {
        self.consume(TokenKind::KwFunction, "Expect 'function'.")?;
        
        let name = self.consume(TokenKind::Ident, "Expect function name.")?;
        let name_str = name.literal.clone().unwrap_or_default();

        self.consume(TokenKind::LParen, "Expect '(' after function name.")?;
        
        let mut params = Vec::new();
        if !self.check(TokenKind::RParen) {
            loop {
                let param_name = self.consume(TokenKind::Ident, "Expect parameter name.")?;
                let param_name_str = param_name.literal.clone().unwrap_or_default();
                
                self.consume(TokenKind::Colon, "Expect ':' after parameter name.")?;
                
                let param_type = self.parse_type()?;
                
                params.push((param_name_str, param_type));
                
                if !self.match_kinds(&[TokenKind::Comma]) {
                    break;
                }
            }
        }
        
        self.consume(TokenKind::RParen, "Expect ')' after parameters.")?;
        
        let body = self.block()?;
        
        Ok(Some(Stmt::Func {
            kind: FuncKind::Normal,
            name: name_str,
            params,
            body,
        }))
    }

    fn async_function_declaration(&mut self) -> Result<Option<Stmt>, Diagnostic> {
        self.consume(TokenKind::KwAsync, "Expect 'async'.")?;
        self.consume(TokenKind::KwFunction, "Expect 'function'.")?;
        
        let name = self.consume(TokenKind::Ident, "Expect function name.")?;
        let name_str = name.literal.clone().unwrap_or_default();

        self.consume(TokenKind::LParen, "Expect '(' after function name.")?;
        
        let mut params = Vec::new();
        if !self.check(TokenKind::RParen) {
            loop {
                let param_name = self.consume(TokenKind::Ident, "Expect parameter name.")?;
                let param_name_str = param_name.literal.clone().unwrap_or_default();
                
                self.consume(TokenKind::Colon, "Expect ':' after parameter name.")?;
                
                let param_type = self.parse_type()?;
                
                params.push((param_name_str, param_type));
                
                if !self.match_kinds(&[TokenKind::Comma]) {
                    break;
                }
            }
        }
        
        self.consume(TokenKind::RParen, "Expect ')' after parameters.")?;
        
        let body = self.block()?;
        
        Ok(Some(Stmt::Func {
            kind: FuncKind::Async,
            name: name_str,
            params,
            body,
        }))
    }

    fn coop_function_declaration(&mut self) -> Result<Option<Stmt>, Diagnostic> {
        self.consume(TokenKind::KwCoop, "Expect 'coop'.")?;
        self.consume(TokenKind::KwFunction, "Expect 'function'.")?;
        
        let name = self.consume(TokenKind::Ident, "Expect function name.")?;
        let name_str = name.literal.clone().unwrap_or_default();

        self.consume(TokenKind::LParen, "Expect '(' after function name.")?;
        
        let mut params = Vec::new();
        if !self.check(TokenKind::RParen) {
            loop {
                let param_name = self.consume(TokenKind::Ident, "Expect parameter name.")?;
                let param_name_str = param_name.literal.clone().unwrap_or_default();
                
                self.consume(TokenKind::Colon, "Expect ':' after parameter name.")?;
                
                let param_type = self.parse_type()?;
                
                params.push((param_name_str, param_type));
                
                if !self.match_kinds(&[TokenKind::Comma]) {
                    break;
                }
            }
        }
        
        self.consume(TokenKind::RParen, "Expect ')' after parameters.")?;
        
        let body = self.block()?;
        
        Ok(Some(Stmt::Func {
            kind: FuncKind::Coop,
            name: name_str,
            params,
            body,
        }))
    }

    fn import_declaration(&mut self) -> Result<Option<Stmt>, Diagnostic> {
        self.consume(TokenKind::KwImport, "Expect 'import'.")?;
        
        let module_name = self.consume(TokenKind::Ident, "Expect module name.")?;
        let module_name_str = module_name.literal.clone().unwrap_or_default();
        
        Ok(Some(Stmt::Import {
            module: module_name_str,
        }))
    }

    fn statement(&mut self) -> Result<Option<Stmt>, Diagnostic> {
        if self.match_kinds(&[TokenKind::KwVar, TokenKind::KwInt, TokenKind::KwFloat, 
                             TokenKind::KwBool, TokenKind::KwString, TokenKind::KwList, TokenKind::KwMap]) {
            self.variable_declaration()
        } else if self.match_kinds(&[TokenKind::KwIf]) {
            self.if_statement()
        } else if self.match_kinds(&[TokenKind::KwFor]) {
            self.for_statement()
        } else if self.match_kinds(&[TokenKind::KwWhile]) {
            self.while_statement()
        } else if self.match_kinds(&[TokenKind::KwReturn]) {
            self.return_statement()
        } else if self.match_kinds(&[TokenKind::KwBreak]) {
            self.break_statement()
        } else if self.match_kinds(&[TokenKind::KwContinue]) {
            self.continue_statement()
        } else {
            self.expression_statement()
        }
    }

    fn variable_declaration(&mut self) -> Result<Option<Stmt>, Diagnostic> {
        let type_token = self.previous();
        let var_type = match type_token.kind {
            TokenKind::KwVar => "var".to_string(),
            TokenKind::KwInt => "int".to_string(),
            TokenKind::KwFloat => "float".to_string(),
            TokenKind::KwBool => "bool".to_string(),
            TokenKind::KwString => "string".to_string(),
            TokenKind::KwList => "list".to_string(),
            TokenKind::KwMap => "map".to_string(),
            _ => return Err(Diagnostic::error(type_token.span, "Invalid type declaration.".to_string())),
        };

        let name = self.consume(TokenKind::Ident, "Expect variable name.")?;
        let name_str = name.literal.clone().unwrap_or_default();
        
        self.consume(TokenKind::Equal, "Expect '=' after variable name.")?;
        
        let initializer = self.expression()?;
        
        Ok(Some(Stmt::VarDecl {
            ty: var_type,
            name: name_str,
            initializer,
        }))
    }

    fn if_statement(&mut self) -> Result<Option<Stmt>, Diagnostic> {
        self.consume(TokenKind::LParen, "Expect '(' after 'if'.")?;
        let condition = self.expression()?;
        self.consume(TokenKind::RParen, "Expect ')' after if condition.")?;
        
        let then_branch = self.block()?;
        
        let mut elif_branches = Vec::new();
        while self.check(TokenKind::KwElif) {
            self.advance();
            self.consume(TokenKind::LParen, "Expect '(' after 'elif'.")?;
            let elif_condition = self.expression()?;
            self.consume(TokenKind::RParen, "Expect ')' after elif condition.")?;
            let elif_branch = self.block()?;
            elif_branches.push((elif_condition, elif_branch));
        }
        
        let else_branch = if self.check(TokenKind::KwElse) {
            self.advance();
            Some(self.block()?)
        } else {
            None
        };
        
        Ok(Some(Stmt::If {
            condition,
            then_branch,
            elif_branches,
            else_branch,
        }))
    }

    fn for_statement(&mut self) -> Result<Option<Stmt>, Diagnostic> {
        self.consume(TokenKind::LParen, "Expect '(' after 'for'.")?;
        
        let var_name = self.consume(TokenKind::Ident, "Expect variable name.")?;
        let var_name_str = var_name.literal.clone().unwrap_or_default();
        
        self.consume(TokenKind::Colon, "Expect ':' after variable name.")?;
        
        let var_type = self.parse_type()?;
        
        self.consume(TokenKind::KwIn, "Expect 'in' after variable declaration.")?;
        
        let iterable = self.expression()?;
        self.consume(TokenKind::RParen, "Expect ')' after for loop header.")?;
        
        let body = self.block()?;
        
        Ok(Some(Stmt::For {
            var_name: var_name_str,
            var_type,
            iterable,
            body,
        }))
    }

    fn while_statement(&mut self) -> Result<Option<Stmt>, Diagnostic> {
        self.consume(TokenKind::LParen, "Expect '(' after 'while'.")?;
        let condition = self.expression()?;
        self.consume(TokenKind::RParen, "Expect ')' after while condition.")?;
        
        let body = self.block()?;
        
        Ok(Some(Stmt::While {
            condition,
            body,
        }))
    }

    fn return_statement(&mut self) -> Result<Option<Stmt>, Diagnostic> {
        let value = if !self.check(TokenKind::Semicolon) {
            Some(self.expression()?)
        } else {
            None
        };
        
        Ok(Some(Stmt::Return { value }))
    }

    fn break_statement(&mut self) -> Result<Option<Stmt>, Diagnostic> {
        Ok(Some(Stmt::Break))
    }

    fn continue_statement(&mut self) -> Result<Option<Stmt>, Diagnostic> {
        Ok(Some(Stmt::Continue))
    }

    fn expression_statement(&mut self) -> Result<Option<Stmt>, Diagnostic> {
        let expr = self.expression()?;
        Ok(Some(Stmt::ExprStmt { expr }))
    }

    fn block(&mut self) -> Result<Vec<Stmt>, Diagnostic> {
        let mut statements = Vec::new();

        // Skip INDENT token
        if self.check(TokenKind::Indent) {
            self.advance();
        }

        while !self.check(TokenKind::Dedent) && !self.is_at_end() {
            // Skip newlines and comments
            if self.check(TokenKind::Newline) || self.check(TokenKind::Comment) {
                self.advance();
                continue;
            }

            if let Some(stmt) = self.declaration()? {
                statements.push(stmt);
            }
        }

        // Skip DEDENT token
        if self.check(TokenKind::Dedent) {
            self.advance();
        }

        Ok(statements)
    }

    fn expression(&mut self) -> Result<Expr, Diagnostic> {
        self.assignment()
    }

    fn assignment(&mut self) -> Result<Expr, Diagnostic> {
        let expr = self.or()?;

        if self.match_kinds(&[TokenKind::Equal]) {
            let value = self.assignment()?;
            return Ok(Expr::Binary {
                left: Box::new(expr),
                op: BinaryOp::Assign,
                right: Box::new(value),
            });
        }

        Ok(expr)
    }

    fn or(&mut self) -> Result<Expr, Diagnostic> {
        let mut expr = self.and()?;

        while self.match_kinds(&[TokenKind::KwOr]) {
            let operator = BinaryOp::Or;
            let right = self.and()?;
            expr = Expr::Binary {
                left: Box::new(expr),
                op: operator,
                right: Box::new(right),
            };
        }

        Ok(expr)
    }

    fn and(&mut self) -> Result<Expr, Diagnostic> {
        let mut expr = self.equality()?;

        while self.match_kinds(&[TokenKind::KwAnd]) {
            let operator = BinaryOp::And;
            let right = self.equality()?;
            expr = Expr::Binary {
                left: Box::new(expr),
                op: operator,
                right: Box::new(right),
            };
        }

        Ok(expr)
    }

    fn equality(&mut self) -> Result<Expr, Diagnostic> {
        let mut expr = self.comparison()?;

        while self.match_kinds(&[TokenKind::EqualEqual, TokenKind::BangEqual]) {
            let operator = match self.previous().kind {
                TokenKind::EqualEqual => BinaryOp::Eq,
                TokenKind::BangEqual => BinaryOp::Ne,
                _ => unreachable!(),
            };
            let right = self.comparison()?;
            expr = Expr::Binary {
                left: Box::new(expr),
                op: operator,
                right: Box::new(right),
            };
        }

        Ok(expr)
    }

    fn comparison(&mut self) -> Result<Expr, Diagnostic> {
        let mut expr = self.term()?;

        while self.match_kinds(&[TokenKind::Greater, TokenKind::GreaterEqual, TokenKind::Less, TokenKind::LessEqual]) {
            let operator = match self.previous().kind {
                TokenKind::Greater => BinaryOp::Gt,
                TokenKind::GreaterEqual => BinaryOp::Ge,
                TokenKind::Less => BinaryOp::Lt,
                TokenKind::LessEqual => BinaryOp::Le,
                _ => unreachable!(),
            };
            let right = self.term()?;
            expr = Expr::Binary {
                left: Box::new(expr),
                op: operator,
                right: Box::new(right),
            };
        }

        Ok(expr)
    }

    fn term(&mut self) -> Result<Expr, Diagnostic> {
        let mut expr = self.factor()?;

        while self.match_kinds(&[TokenKind::Plus, TokenKind::Minus]) {
            let operator = match self.previous().kind {
                TokenKind::Plus => BinaryOp::Add,
                TokenKind::Minus => BinaryOp::Sub,
                _ => unreachable!(),
            };
            let right = self.factor()?;
            expr = Expr::Binary {
                left: Box::new(expr),
                op: operator,
                right: Box::new(right),
            };
        }

        Ok(expr)
    }

    fn factor(&mut self) -> Result<Expr, Diagnostic> {
        let mut expr = self.unary()?;

        while self.match_kinds(&[TokenKind::Star, TokenKind::Slash, TokenKind::Percent]) {
            let operator = match self.previous().kind {
                TokenKind::Star => BinaryOp::Mul,
                TokenKind::Slash => BinaryOp::Div,
                TokenKind::Percent => BinaryOp::Mod,
                _ => unreachable!(),
            };
            let right = self.unary()?;
            expr = Expr::Binary {
                left: Box::new(expr),
                op: operator,
                right: Box::new(right),
            };
        }

        Ok(expr)
    }

    fn unary(&mut self) -> Result<Expr, Diagnostic> {
        if self.match_kinds(&[TokenKind::Bang, TokenKind::Minus]) {
            let operator = match self.previous().kind {
                TokenKind::Bang => UnaryOp::Not,
                TokenKind::Minus => UnaryOp::Neg,
                _ => unreachable!(),
            };
            let right = self.unary()?;
            return Ok(Expr::Unary {
                op: operator,
                expr: Box::new(right),
            });
        }

        self.call()
    }

    fn call(&mut self) -> Result<Expr, Diagnostic> {
        let mut expr = self.primary()?;

        loop {
            if self.match_kinds(&[TokenKind::LParen]) {
                expr = self.finish_call(expr)?;
            } else if self.match_kinds(&[TokenKind::Dot]) {
                let name = self.consume(TokenKind::Ident, "Expect property name after '.'.")?;
                let name_str = name.literal.clone().unwrap_or_default();
                expr = Expr::Attr {
                    object: Box::new(expr),
                    attr: name_str,
                };
            } else if self.match_kinds(&[TokenKind::LBracket]) {
                let index = self.expression()?;
                self.consume(TokenKind::RBracket, "Expect ']' after index.")?;
                expr = Expr::Index {
                    object: Box::new(expr),
                    index: Box::new(index),
                };
            } else {
                break;
            }
        }

        Ok(expr)
    }

    fn finish_call(&mut self, callee: Expr) -> Result<Expr, Diagnostic> {
        let mut arguments = Vec::new();
        
        if !self.check(TokenKind::RParen) {
            loop {
                arguments.push(self.expression()?);
                if !self.match_kinds(&[TokenKind::Comma]) {
                    break;
                }
            }
        }

        self.consume(TokenKind::RParen, "Expect ')' after arguments.")?;

        Ok(Expr::Call {
            callee: Box::new(callee),
            arguments,
        })
    }

    fn primary(&mut self) -> Result<Expr, Diagnostic> {
        if self.match_kinds(&[TokenKind::KwFalse]) {
            return Ok(Expr::Literal { value: "false".to_string() });
        }
        if self.match_kinds(&[TokenKind::KwTrue]) {
            return Ok(Expr::Literal { value: "true".to_string() });
        }
        if self.match_kinds(&[TokenKind::KwNull]) {
            return Ok(Expr::Literal { value: "null".to_string() });
        }

        if self.match_kinds(&[TokenKind::Number, TokenKind::String]) {
            let value = self.previous().literal.clone().unwrap_or_default();
            return Ok(Expr::Literal { value });
        }

        if self.match_kinds(&[TokenKind::Ident]) {
            let name = self.previous().literal.clone().unwrap_or_default();
            return Ok(Expr::Ident { name });
        }

        if self.match_kinds(&[TokenKind::LParen]) {
            let expr = self.expression()?;
            self.consume(TokenKind::RParen, "Expect ')' after expression.")?;
            return Ok(Expr::Grouping { expr: Box::new(expr) });
        }

        if self.match_kinds(&[TokenKind::LBracket]) {
            return self.list();
        }

        if self.match_kinds(&[TokenKind::LBrace]) {
            return self.map();
        }

        Err(Diagnostic::error(self.peek().span, "Expect expression.".to_string()))
    }

    fn list(&mut self) -> Result<Expr, Diagnostic> {
        let mut elements = Vec::new();
        
        if !self.check(TokenKind::RBracket) {
            loop {
                elements.push(self.expression()?);
                if !self.match_kinds(&[TokenKind::Comma]) {
                    break;
                }
            }
        }

        self.consume(TokenKind::RBracket, "Expect ']' after list elements.")?;
        Ok(Expr::List { elements })
    }

    fn map(&mut self) -> Result<Expr, Diagnostic> {
        let mut pairs = Vec::new();
        
        if !self.check(TokenKind::RBrace) {
            loop {
                let key = self.expression()?;
                self.consume(TokenKind::Colon, "Expect ':' after map key.")?;
                let value = self.expression()?;
                pairs.push((key, value));
                
                if !self.match_kinds(&[TokenKind::Comma]) {
                    break;
                }
            }
        }

        self.consume(TokenKind::RBrace, "Expect '}' after map pairs.")?;
        Ok(Expr::Map { pairs })
    }

    fn parse_type(&mut self) -> Result<String, Diagnostic> {
        if self.match_kinds(&[TokenKind::KwVar]) {
            Ok("var".to_string())
        } else if self.match_kinds(&[TokenKind::KwInt]) {
            Ok("int".to_string())
        } else if self.match_kinds(&[TokenKind::KwFloat]) {
            Ok("float".to_string())
        } else if self.match_kinds(&[TokenKind::KwBool]) {
            Ok("bool".to_string())
        } else if self.match_kinds(&[TokenKind::KwString]) {
            Ok("string".to_string())
        } else if self.match_kinds(&[TokenKind::KwList]) {
            Ok("list".to_string())
        } else if self.match_kinds(&[TokenKind::KwMap]) {
            Ok("map".to_string())
        } else {
            Err(Diagnostic::error(self.peek().span, "Invalid type.".to_string()))
        }
    }
}
