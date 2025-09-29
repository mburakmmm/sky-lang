// AST Definitions
// Sky dilinin Abstract Syntax Tree yapılarını tanımlar

use crate::compiler::diag::Span;

/// AST yapısı
#[derive(Debug, Clone)]
pub struct Ast {
    pub statements: Vec<Stmt>,
}

/// Statement türleri
#[derive(Debug, Clone)]
pub enum Stmt {
    VarDecl {
        ty: TypeDecl,
        name: String,
        value: Expr,
        span: Span,
    },
    Func {
        kind: FuncKind,
        name: String,
        params: Vec<Param>,
        body: Vec<Stmt>,
        span: Span,
    },
    If {
        condition: Expr,
        then_branch: Vec<Stmt>,
        elif_branches: Vec<(Expr, Vec<Stmt>)>,
        else_branch: Option<Vec<Stmt>>,
        span: Span,
    },
    For {
        variable: String,
        iterable: Expr,
        body: Vec<Stmt>,
        span: Span,
    },
    While {
        condition: Expr,
        body: Vec<Stmt>,
        span: Span,
    },
    Return {
        value: Option<Expr>,
        span: Span,
    },
    Break {
        span: Span,
    },
    Continue {
        span: Span,
    },
    Import {
        module: String,
        span: Span,
    },
    ExprStmt {
        expr: Expr,
        span: Span,
    },
}

/// Expression türleri
#[derive(Debug, Clone)]
pub enum Expr {
    Lit(Literal),
    Ident(String, Span),
    Call {
        callee: Box<Expr>,
        args: Vec<Expr>,
        span: Span,
    },
    Attr {
        object: Box<Expr>,
        attr: String,
        span: Span,
    },
    Index {
        object: Box<Expr>,
        index: Box<Expr>,
        span: Span,
    },
    Unary {
        op: UnaryOp,
        expr: Box<Expr>,
        span: Span,
    },
    Binary {
        left: Box<Expr>,
        op: BinaryOp,
        right: Box<Expr>,
        span: Span,
    },
    Await {
        expr: Box<Expr>,
        span: Span,
    },
    Yield {
        expr: Option<Box<Expr>>,
        span: Span,
    },
    Interpolated {
        parts: Vec<InterpPart>,
        span: Span,
    },
    Range {
        start: Box<Expr>,
        end: Box<Expr>,
        span: Span,
    },
    Ternary {
        condition: Box<Expr>,
        true_expr: Box<Expr>,
        false_expr: Box<Expr>,
        span: Span,
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

/// String interpolation parçaları
#[derive(Debug, Clone)]
pub enum InterpPart {
    Text(String),
    Expr(Box<Expr>),
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
    Assign,
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
    pub span: Span,
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

impl BinaryOp {
    /// Operatör önceliğini döndür (yüksek sayı = yüksek öncelik)
    pub fn precedence(&self) -> u8 {
        match self {
            Self::Assign => 1, // Assignment has precedence 1, not 0
            Self::Or => 1,
            Self::And => 2,
            Self::Eq | Self::Ne => 3,
            Self::Lt | Self::Le | Self::Gt | Self::Ge => 4,
            Self::Add | Self::Sub => 5,
            Self::Mul | Self::Div | Self::Mod => 6,
        }
    }
    
    /// Operatörün sol-associative olup olmadığını döndür
    pub fn is_left_associative(&self) -> bool {
        match self {
            Self::Assign => false, // Assignment is right-associative
            Self::Or | Self::And | Self::Eq | Self::Ne | 
            Self::Lt | Self::Le | Self::Gt | Self::Ge |
            Self::Add | Self::Sub | Self::Mul | Self::Div | Self::Mod => true,
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_type_decl_parsing() {
        assert_eq!(TypeDecl::from_keyword("int"), Some(TypeDecl::Int));
        assert_eq!(TypeDecl::from_keyword("var"), Some(TypeDecl::Var));
        assert_eq!(TypeDecl::from_keyword("invalid"), None);
    }

    #[test]
    fn test_operator_precedence() {
        assert!(BinaryOp::Mul.precedence() > BinaryOp::Add.precedence());
        assert!(BinaryOp::Add.precedence() > BinaryOp::Eq.precedence());
        assert!(BinaryOp::Eq.precedence() > BinaryOp::Or.precedence());
    }

    #[test]
    fn test_operator_associativity() {
        assert!(BinaryOp::Add.is_left_associative());
        assert!(BinaryOp::Mul.is_left_associative());
        assert!(BinaryOp::Eq.is_left_associative());
    }
}
