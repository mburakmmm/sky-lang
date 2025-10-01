// Parser - AST ve Grammar Kuralları
// Token'lardan AST üret, await/yield bağlam kuralları, tip zorunluluğu denetimi

pub mod ast;
pub mod grammar;

use grammar::Parser;
pub use ast::Ast;
use crate::compiler::lexer::token::Token;
use crate::compiler::diag::Diagnostic;

/// Ana parse fonksiyonu
pub fn parse(tokens: Vec<Token>) -> Result<Ast, Diagnostic> {
    // Source string'i boş olarak geç - token'lar zaten span'leri içeriyor
    let mut parser = Parser::new_with_source(tokens, String::new());
    parser.parse_program()
}

/// Source string ile parse fonksiyonu
pub fn parse_with_source(tokens: Vec<Token>, source: String) -> Result<Ast, Diagnostic> {
    let mut parser = Parser::new_with_source(tokens, source);
    let result = parser.parse_program();
    result
}

/// Statement türleri
#[derive(Debug, Clone)]
pub enum Stmt {
    VarDecl {
        ty: TypeDecl,
        name: String,
        value: Expr,
        span: crate::compiler::diag::Span,
    },
    Func {
        kind: FuncKind,
        name: String,
        params: Vec<Param>,
        body: Vec<Stmt>,
        span: crate::compiler::diag::Span,
    },
    If {
        condition: Expr,
        then_branch: Vec<Stmt>,
        elif_branches: Vec<(Expr, Vec<Stmt>)>,
        else_branch: Option<Vec<Stmt>>,
        span: crate::compiler::diag::Span,
    },
    For {
        variable: String,
        iterable: Expr,
        body: Vec<Stmt>,
        span: crate::compiler::diag::Span,
    },
    While {
        condition: Expr,
        body: Vec<Stmt>,
        span: crate::compiler::diag::Span,
    },
    Return {
        value: Option<Expr>,
        span: crate::compiler::diag::Span,
    },
    Break {
        span: crate::compiler::diag::Span,
    },
    Continue {
        span: crate::compiler::diag::Span,
    },
    Import {
        module: String,
        span: crate::compiler::diag::Span,
    },
    ExprStmt {
        expr: Expr,
        span: crate::compiler::diag::Span,
    },
}

/// Expression türleri
#[derive(Debug, Clone)]
pub enum Expr {
    Lit(Literal),
    Ident(String, crate::compiler::diag::Span),
    Call {
        callee: Box<Expr>,
        args: Vec<Expr>,
        span: crate::compiler::diag::Span,
    },
    Attr {
        object: Box<Expr>,
        attr: String,
        span: crate::compiler::diag::Span,
    },
    Index {
        object: Box<Expr>,
        index: Box<Expr>,
        span: crate::compiler::diag::Span,
    },
    Unary {
        op: UnaryOp,
        expr: Box<Expr>,
        span: crate::compiler::diag::Span,
    },
    Binary {
        left: Box<Expr>,
        op: BinaryOp,
        right: Box<Expr>,
        span: crate::compiler::diag::Span,
    },
    Await {
        expr: Box<Expr>,
        span: crate::compiler::diag::Span,
    },
    Yield {
        expr: Option<Box<Expr>>,
        span: crate::compiler::diag::Span,
    },
}

/// Literal değerler
#[derive(Debug, Clone)]
pub enum Literal {
    Int(i64),
    Float(f64),
    String(String),
    Bool(bool),
    Null,
    List(Vec<Expr>),
    Map(Vec<(String, Expr)>),
}

/// Unary operatörler
#[derive(Debug, Clone, PartialEq)]
pub enum UnaryOp {
    Neg,
    Not,
}

/// Binary operatörler
#[derive(Debug, Clone, PartialEq)]
pub enum BinaryOp {
    Add,
    Sub,
    Mul,
    Div,
    Mod,
    Eq,
    Ne,
    Lt,
    Le,
    Gt,
    Ge,
    And,
    Or,
}

/// Fonksiyon türleri
#[derive(Debug, Clone, PartialEq)]
pub enum FuncKind {
    Normal,
    Async,
    Coop,
}

/// Tip bildirimleri
#[derive(Debug, Clone, PartialEq)]
pub enum TypeDecl {
    Var,
    Int,
    Float,
    Bool,
    String,
    List,
    Map,
}

/// Fonksiyon parametresi
#[derive(Debug, Clone)]
pub struct Param {
    pub name: String,
    pub ty: TypeDecl,
    pub span: crate::compiler::diag::Span,
}

impl TypeDecl {
    pub fn from_keyword(keyword: &str) -> Option<Self> {
        match keyword {
            "var" => Some(Self::Var),
            "int" => Some(Self::Int),
            "float" => Some(Self::Float),
            "bool" => Some(Self::Bool),
            "string" => Some(Self::String),
            "list" => Some(Self::List),
            "map" => Some(Self::Map),
            _ => None,
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::compiler::lexer::{lex, token::TokenKind};
    use crate::compiler::diag::Span;

    #[test]
    fn test_var_decl_parsing() {
        let tokens = lex("int x = 42").unwrap();
        let ast = parse(tokens).unwrap();
        
        assert_eq!(ast.statements.len(), 1);
        if let Stmt::VarDecl { ty, name, .. } = &ast.statements[0] {
            assert_eq!(*ty, TypeDecl::Int);
            assert_eq!(name, "x");
        } else {
            panic!("Expected VarDecl");
        }
    }

    #[test]
    fn test_function_parsing() {
        let tokens = lex("function test(x: int)\n  return x").unwrap();
        let ast = parse(tokens).unwrap();
        
        assert_eq!(ast.statements.len(), 1);
        if let Stmt::Func { name, params, .. } = &ast.statements[0] {
            assert_eq!(name, "test");
            assert_eq!(params.len(), 1);
            assert_eq!(params[0].name, "x");
            assert_eq!(params[0].ty, TypeDecl::Int);
        } else {
            panic!("Expected Func");
        }
    }

    #[test]
    fn test_missing_type_annotation() {
        let tokens = lex("x = 42").unwrap();
        let result = parse(tokens);
        assert!(result.is_err());
        if let Err(diag) = result {
            assert_eq!(diag.code, codes::MISSING_TYPE_ANNOTATION);
        }
    }

    #[test]
    fn test_await_outside_async() {
        let tokens = lex("function test()\n  await sleep(1)").unwrap();
        let result = parse(tokens);
        assert!(result.is_err());
        if let Err(diag) = result {
            assert_eq!(diag.code, codes::AWAIT_OUTSIDE_ASYNC);
        }
    }

    #[test]
    fn test_yield_outside_coop() {
        let tokens = lex("function test()\n  yield 1").unwrap();
        let result = parse(tokens);
        assert!(result.is_err());
        if let Err(diag) = result {
            assert_eq!(diag.code, codes::YIELD_OUTSIDE_COOP);
        }
    }
}
