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
    List,   // Liste (generic olmayan)
    Map,    // Sözlük
    ListParam(PrimitiveType), // Parametreli liste: list[int], list[string] vb.
}

/// Primitive tip türleri (list parametreleri için)
#[derive(Debug, Clone, Copy, PartialEq)]
pub enum PrimitiveType {
    Int,
    Float,
    Bool,
    String,
}

impl PrimitiveType {
    /// Keyword'den primitive tip oluştur
    pub fn from_keyword(keyword: &str) -> Option<Self> {
        match keyword {
            "int" => Some(Self::Int),
            "float" => Some(Self::Float),
            "bool" => Some(Self::Bool),
            "string" => Some(Self::String),
            _ => None,
        }
    }
    
    /// Primitive tipi TypeDecl'e çevir
    pub fn to_type_decl(&self) -> TypeDecl {
        match self {
            PrimitiveType::Int => TypeDecl::Int,
            PrimitiveType::Float => TypeDecl::Float,
            PrimitiveType::Bool => TypeDecl::Bool,
            PrimitiveType::String => TypeDecl::String,
        }
    }
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
        matches!(self, Self::List | Self::Map | Self::ListParam(_))
    }
    
    /// Tip bildiriminin parametreli liste olup olmadığını kontrol et
    pub fn is_parameterized_list(&self) -> bool {
        matches!(self, Self::ListParam(_))
    }
    
    /// Tip bildirimini bytecode için u8'e çevir
    pub fn to_bytecode_type(&self) -> u8 {
        match self {
            Self::Var => 0,
            Self::Int => 1,
            Self::Float => 2,
            Self::Bool => 3,
            Self::String => 4,
            Self::List => 5,
            Self::Map => 6,
            Self::ListParam(_) => 7, // Parametreli liste
        }
    }
    
    /// Bytecode tip kodundan TypeDecl oluştur
    pub fn from_bytecode_type(code: u8) -> Option<Self> {
        match code {
            0 => Some(Self::Var),
            1 => Some(Self::Int),
            2 => Some(Self::Float),
            3 => Some(Self::Bool),
            4 => Some(Self::String),
            5 => Some(Self::List),
            6 => Some(Self::Map),
            7 => Some(Self::ListParam(PrimitiveType::String)), // Default parametreli liste
            _ => None,
        }
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
            TypeDecl::ListParam(param) => {
                let param_name = match param {
                    PrimitiveType::Int => "int",
                    PrimitiveType::Float => "float",
                    PrimitiveType::Bool => "bool",
                    PrimitiveType::String => "string",
                };
                return write!(f, "list[{}]", param_name);
            }
        };
        write!(f, "{}", name)
    }
}

impl std::fmt::Display for PrimitiveType {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        let name = match self {
            PrimitiveType::Int => "int",
            PrimitiveType::Float => "float",
            PrimitiveType::Bool => "bool",
            PrimitiveType::String => "string",
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
