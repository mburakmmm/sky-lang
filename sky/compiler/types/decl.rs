// Type Declarations - Bildirim Tipleri
// Sky dilinin tip bildirim sistemini tanımlar

use crate::compiler::diag::Span;

/// Tip bildirimleri
#[derive(Debug, Clone, PartialEq)]
pub enum TypeDecl {
    Var,    // Dinamik tip (Any)
    Int,    // Tamsayı
    Float,  // Ondalık sayı
    Bool,   // Boolean
    String, // Metin
    List,   // Liste
    Map,    // Sözlük
}

impl TypeDecl {
    /// Keyword'den tip bildirimi oluştur
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
    
    /// Tip bildiriminin dinamik olup olmadığını kontrol et
    pub fn is_dynamic(&self) -> bool {
        matches!(self, Self::Var)
    }
    
    /// Tip bildiriminin primitive olup olmadığını kontrol et
    pub fn is_primitive(&self) -> bool {
        matches!(self, Self::Int | Self::Float | Self::Bool | Self::String)
    }
    
    /// Tip bildiriminin collection olup olmadığını kontrol et
    pub fn is_collection(&self) -> bool {
        matches!(self, Self::List | Self::Map)
    }
}

impl std::fmt::Display for TypeDecl {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        let name = match self {
            TypeDecl::Var => "var",
            TypeDecl::Int => "int",
            TypeDecl::Float => "float",
            TypeDecl::Bool => "bool",
            TypeDecl::String => "string",
            TypeDecl::List => "list",
            TypeDecl::Map => "map",
        };
        write!(f, "{}", name)
    }
}

/// Fonksiyon parametresi
#[derive(Debug, Clone)]
pub struct Param {
    pub name: String,
    pub ty: TypeDecl,
    pub span: Span,
}

impl Param {
    pub fn new(name: String, ty: TypeDecl, span: Span) -> Self {
        Self { name, ty, span }
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
    fn test_type_properties() {
        assert!(TypeDecl::Var.is_dynamic());
        assert!(!TypeDecl::Int.is_dynamic());
        
        assert!(TypeDecl::Int.is_primitive());
        assert!(TypeDecl::String.is_primitive());
        assert!(!TypeDecl::List.is_primitive());
        
        assert!(TypeDecl::List.is_collection());
        assert!(TypeDecl::Map.is_collection());
        assert!(!TypeDecl::Int.is_collection());
    }

    #[test]
    fn test_type_display() {
        assert_eq!(format!("{}", TypeDecl::Int), "int");
        assert_eq!(format!("{}", TypeDecl::Var), "var");
        assert_eq!(format!("{}", TypeDecl::List), "list");
    }
}
