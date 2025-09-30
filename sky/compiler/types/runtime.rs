// Runtime Types - Çalışma Zamanı Türleri
// VM'de çalışan değerlerin türlerini tanımlar

use super::decl::TypeDecl;
use crate::compiler::diag::{Diagnostic, codes};

/// Runtime değer türleri
#[derive(Debug, Clone, PartialEq)]
pub enum ValueKind {
    Int,
    Float,
    Bool,
    String,
    List,
    Map,
    Function,
    NativeFn,
    Future,
    Coroutine,
    Range,
    Null,
}

impl ValueKind {
    /// Değer türünün primitive olup olmadığını kontrol et
    pub fn is_primitive(&self) -> bool {
        matches!(self, Self::Int | Self::Float | Self::Bool | Self::String)
    }
    
    /// Değer türünün collection olup olmadığını kontrol et
    pub fn is_collection(&self) -> bool {
        matches!(self, Self::List | Self::Map)
    }
    
    /// Değer türünün callable olup olmadığını kontrol et
    pub fn is_callable(&self) -> bool {
        matches!(self, Self::Function | Self::NativeFn)
    }
    
    /// Değer türünün async olup olmadığını kontrol et
    pub fn is_async(&self) -> bool {
        matches!(self, Self::Future | Self::Coroutine)
    }
}

impl std::fmt::Display for ValueKind {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        let name = match self {
            ValueKind::Int => "int",
            ValueKind::Float => "float",
            ValueKind::Bool => "bool",
            ValueKind::String => "string",
            ValueKind::List => "list",
            ValueKind::Map => "map",
            ValueKind::Function => "function",
            ValueKind::NativeFn => "native function",
            ValueKind::Future => "future",
            ValueKind::Coroutine => "coroutine",
            ValueKind::Range => "range",
            ValueKind::Null => "null",
        };
        write!(f, "{}", name)
    }
}

/// Tip uyumluluğu kontrolü
pub fn type_check(declared: &TypeDecl, actual: &ValueKind) -> Result<(), Diagnostic> {
    if is_compatible(declared, actual) {
        Ok(())
    } else {
        Err(Diagnostic::error(
            codes::TYPE_MISMATCH,
            &format!("Expected {}, found {}", declared.to_string(), actual.to_string()),
            crate::compiler::diag::Span::new(0, 0, 0), // Runtime span
        ))
    }
}

/// Tip uyumluluğu kontrolü (bool döndürür)
pub fn is_compatible(declared: &TypeDecl, actual: &ValueKind) -> bool {
    match declared {
        TypeDecl::Var => true, // var her türü kabul eder
        TypeDecl::Int => matches!(actual, ValueKind::Int),
        TypeDecl::Float => matches!(actual, ValueKind::Float | ValueKind::Int), // Int'i Float'a otomatik dönüştür
        TypeDecl::Bool => matches!(actual, ValueKind::Bool),
        TypeDecl::String => matches!(actual, ValueKind::String),
        TypeDecl::List => matches!(actual, ValueKind::List),
        TypeDecl::Map => matches!(actual, ValueKind::Map),
    }
}

/// Tip dönüştürme fonksiyonları
pub mod conversions {
    use super::*;

    /// Int'den Float'a dönüştürme
    pub fn int_to_float(value: i64) -> f64 {
        value as f64
    }

    /// Float'dan Int'e dönüştürme (truncate)
    pub fn float_to_int(value: f64) -> i64 {
        value as i64
    }

    /// String'den Int'e dönüştürme
    pub fn string_to_int(value: &str) -> Result<i64, String> {
        value.parse::<i64>().map_err(|_| "Invalid integer".to_string())
    }

    /// String'den Float'a dönüştürme
    pub fn string_to_float(value: &str) -> Result<f64, String> {
        value.parse::<f64>().map_err(|_| "Invalid float".to_string())
    }

    /// String'den Bool'a dönüştürme
    pub fn string_to_bool(value: &str) -> bool {
        value == "true"
    }

    /// Herhangi bir değeri String'e dönüştürme
    pub fn to_string(value: &ValueKind) -> String {
        match value {
            ValueKind::Int => "int".to_string(),
            ValueKind::Float => "float".to_string(),
            ValueKind::Bool => "bool".to_string(),
            ValueKind::String => "string".to_string(),
            ValueKind::List => "list".to_string(),
            ValueKind::Map => "map".to_string(),
            ValueKind::Function => "function".to_string(),
            ValueKind::NativeFn => "native function".to_string(),
            ValueKind::Future => "future".to_string(),
            ValueKind::Coroutine => "coroutine".to_string(),
            ValueKind::Range => "range".to_string(),
            ValueKind::Null => "null".to_string(),
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_value_kind_properties() {
        assert!(ValueKind::Int.is_primitive());
        assert!(ValueKind::String.is_primitive());
        assert!(!ValueKind::List.is_primitive());
        
        assert!(ValueKind::List.is_collection());
        assert!(ValueKind::Map.is_collection());
        assert!(!ValueKind::Int.is_collection());
        
        assert!(ValueKind::Function.is_callable());
        assert!(ValueKind::NativeFn.is_callable());
        assert!(!ValueKind::Int.is_callable());
        
        assert!(ValueKind::Future.is_async());
        assert!(ValueKind::Coroutine.is_async());
        assert!(!ValueKind::Function.is_async());
    }

    #[test]
    fn test_type_compatibility() {
        assert!(is_compatible(&TypeDecl::Var, &ValueKind::Int));
        assert!(is_compatible(&TypeDecl::Int, &ValueKind::Int));
        assert!(is_compatible(&TypeDecl::String, &ValueKind::String));
        
        assert!(!is_compatible(&TypeDecl::Int, &ValueKind::String));
        assert!(!is_compatible(&TypeDecl::Bool, &ValueKind::Float));
    }

    #[test]
    fn test_type_check_success() {
        assert!(type_check(&TypeDecl::Int, &ValueKind::Int).is_ok());
        assert!(type_check(&TypeDecl::Var, &ValueKind::String).is_ok());
    }

    #[test]
    fn test_type_check_failure() {
        let result = type_check(&TypeDecl::Int, &ValueKind::String);
        assert!(result.is_err());
        if let Err(diag) = result {
            assert_eq!(diag.code, codes::TYPE_MISMATCH);
        }
    }

    #[test]
    fn test_conversions() {
        assert_eq!(conversions::int_to_float(42), 42.0);
        assert_eq!(conversions::float_to_int(42.7), 42);
        assert_eq!(conversions::string_to_int("123"), Ok(123));
        assert_eq!(conversions::string_to_float("3.14"), Ok(3.14));
        assert_eq!(conversions::string_to_bool("true"), true);
        assert_eq!(conversions::string_to_bool("false"), false);
    }

    #[test]
    fn test_value_kind_display() {
        assert_eq!(format!("{}", ValueKind::Int), "int");
        assert_eq!(format!("{}", ValueKind::String), "string");
        assert_eq!(format!("{}", ValueKind::List), "list");
    }
}
