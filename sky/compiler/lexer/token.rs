// Token Definitions
// Sky dilinin tüm token türlerini tanımlar

use crate::compiler::diag::Span;

/// Token türleri
#[derive(Debug, Clone, PartialEq)]
pub enum TokenKind {
    // Literals
    Int,
    Float,
    String,
    True,
    False,
    Null,
    
    // String interpolation
    StringBegin,
    StringText,
    StringEnd,
    FStringBegin,
    FStringText,
    FStringEnd,
    InterpStartDartBrace,  // ${
    InterpEndBrace,        // }
    InterpDartIdent,       // $ident
    
    // Identifiers
    Identifier,
    
    // Keywords
    Var,
    IntType,
    FloatType,
    BoolType,
    StringType,
    ListType,
    MapType,
    Function,
    Async,
    Coop,
    Return,
    If,
    Elif,
    Else,
    For,
    While,
    Break,
    Continue,
    Await,
    Yield,
    Import,
    And,
    Or,
    Not,
    
    // Operators
    Plus,
    Minus,
    Star,
    Slash,
    Percent,
    Equal,
    EqualEqual,
    Bang,
    BangEqual,
    Less,
    LessEqual,
    Greater,
    GreaterEqual,
    
    // Punctuation
    Colon,
    Comma,
    Dot,  // .
    LeftParen,
    RightParen,
    LeftBracket,
    RightBracket,
    LeftBrace,
    RightBrace,
    Range,  // ..
    Question,  // ?
    
    // Special
    Newline,
    Indent,
    Dedent,
    Comment,
    Eof,
}

impl TokenKind {
    /// Keyword'den token türü döndür
    pub fn from_keyword(keyword: &str) -> Option<Self> {
        match keyword {
            "var" => Some(Self::Var),
            "int" => {
                Some(Self::IntType)
            },
            "float" => Some(Self::FloatType),
            "bool" => Some(Self::BoolType),
            "string" => Some(Self::StringType),
            "list" => Some(Self::ListType),
            "map" => Some(Self::MapType),
            "function" => {
                Some(Self::Function)
            },
            "async" => Some(Self::Async),
            "coop" => Some(Self::Coop),
            "return" => Some(Self::Return),
            "if" => Some(Self::If),
            "elif" => Some(Self::Elif),
            "else" => Some(Self::Else),
            "for" => Some(Self::For),
            "while" => Some(Self::While),
            "break" => Some(Self::Break),
            "continue" => Some(Self::Continue),
            "await" => Some(Self::Await),
            "yield" => Some(Self::Yield),
            "import" => Some(Self::Import),
            "and" => Some(Self::And),
            "or" => Some(Self::Or),
            "not" => Some(Self::Not),
            "true" => Some(Self::True),
            "false" => Some(Self::False),
            "null" => Some(Self::Null),
            _ => None,
        }
    }
    
    /// Token'ın literal olup olmadığını kontrol et
    pub fn is_literal(&self) -> bool {
        matches!(self, Self::Int | Self::Float | Self::String | Self::True | Self::False | Self::Null)
    }
    
    /// Token'ın keyword olup olmadığını kontrol et
    pub fn is_keyword(&self) -> bool {
        matches!(self, 
            Self::Var | Self::IntType | Self::FloatType | Self::BoolType | 
            Self::StringType | Self::ListType | Self::MapType | Self::Function |
            Self::Async | Self::Coop | Self::Return | Self::If | Self::Elif |
            Self::Else | Self::For | Self::While | Self::Break | Self::Continue |
            Self::Await | Self::Yield | Self::Import | Self::And | Self::Or | Self::Not |
            Self::True | Self::False | Self::Null
        )
    }
    
    /// Token'ın operatör olup olmadığını kontrol et
    pub fn is_operator(&self) -> bool {
        matches!(self,
            Self::Plus | Self::Minus | Self::Star | Self::Slash | Self::Percent |
            Self::Equal | Self::EqualEqual | Self::Bang | Self::BangEqual |
            Self::Less | Self::LessEqual | Self::Greater | Self::GreaterEqual |
            Self::And | Self::Or
        )
    }
}

/// Token yapısı
#[derive(Debug, Clone)]
pub struct Token {
    pub kind: TokenKind,
    pub span: Span,
    pub value: Option<String>, // String değerleri için (StringText, InterpDartIdent, etc.)
    pub indent_value: Option<usize>, // Indent seviyesi için
}

impl Token {
    pub fn new(kind: TokenKind, span: Span) -> Self {
        Self { kind, span, value: None, indent_value: None }
    }
    
    pub fn new_with_value(kind: TokenKind, span: Span, value: String) -> Self {
        Self { kind, span, value: Some(value), indent_value: None }
    }
    
    pub fn new_with_indent(kind: TokenKind, span: Span, indent_value: usize) -> Self {
        Self { kind, span, value: None, indent_value: Some(indent_value) }
    }
    
    /// Token'ın kaynak metindeki değerini döndür
    pub fn text<'a>(&self, source: &'a str) -> &'a str {
        if self.span.start >= source.len() || self.span.end > source.len() {
            return "";
        }
        &source[self.span.start..self.span.end]
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_keyword_parsing() {
        assert_eq!(TokenKind::from_keyword("int"), Some(TokenKind::IntType));
        assert_eq!(TokenKind::from_keyword("function"), Some(TokenKind::Function));
        assert_eq!(TokenKind::from_keyword("invalid"), None);
    }

    #[test]
    fn test_token_properties() {
        assert!(TokenKind::Int.is_literal());
        assert!(TokenKind::IntType.is_keyword());
        assert!(TokenKind::Plus.is_operator());
        
        assert!(!TokenKind::Identifier.is_literal());
        assert!(!TokenKind::Int.is_keyword());
        assert!(!TokenKind::Identifier.is_operator());
    }
}
