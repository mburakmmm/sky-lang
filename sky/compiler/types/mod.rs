// Types - Bildirim Tipleri & Runtime Türleri
// Bildirilen tip ile çalışma zamanı değerlerinin uyuşmasını kontrol edecek metadata

pub mod decl;
pub mod runtime;

use crate::compiler::diag::{Diagnostic, codes};

// Re-exports
pub use decl::{TypeDecl, Param};
pub use runtime::{ValueKind, type_check};

/// Tip uyumluluğu kontrolü
pub fn is_compatible(declared: &TypeDecl, actual: &ValueKind) -> bool {
    match declared {
        TypeDecl::Var => true, // var her türü kabul eder
        TypeDecl::Int => matches!(actual, ValueKind::Int),
        TypeDecl::Float => matches!(actual, ValueKind::Float),
        TypeDecl::Bool => matches!(actual, ValueKind::Bool),
        TypeDecl::String => matches!(actual, ValueKind::String),
        TypeDecl::List => matches!(actual, ValueKind::List),
        TypeDecl::Map => matches!(actual, ValueKind::Map),
    }
}

/// Tip uyumsuzluğu hatası oluştur
pub fn type_mismatch_error(
    declared: &TypeDecl,
    actual: &ValueKind,
    span: crate::compiler::diag::Span,
) -> Diagnostic {
    Diagnostic::error(
        codes::TYPE_MISMATCH,
        &format!("Expected {}, found {}", declared.to_string(), actual.to_string()),
        span,
    )
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_type_compatibility() {
        assert!(is_compatible(&TypeDecl::Var, &ValueKind::Int));
        assert!(is_compatible(&TypeDecl::Int, &ValueKind::Int));
        assert!(is_compatible(&TypeDecl::String, &ValueKind::String));
        
        assert!(!is_compatible(&TypeDecl::Int, &ValueKind::String));
        assert!(!is_compatible(&TypeDecl::Bool, &ValueKind::Float));
    }
}
