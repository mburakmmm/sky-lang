// Symbol Information - Sembol Bilgileri
// Sembol türleri ve slot yönetimi

use crate::compiler::types::TypeDecl;
use crate::compiler::diag::Span;

/// Sembol türleri
#[derive(Debug, Clone, PartialEq)]
pub enum SymbolKind {
    Variable,
    Function,
    Parameter,
}

/// Sembol slot türleri
#[derive(Debug, Clone, PartialEq)]
pub enum Slot {
    Local(u32),   // Yerel değişken slot'u
    Global(u32),  // Global değişken slot'u
}

impl Slot {
    /// Slot'un local olup olmadığını kontrol et
    pub fn is_local(&self) -> bool {
        matches!(self, Self::Local(_))
    }

    /// Slot'un global olup olmadığını kontrol et
    pub fn is_global(&self) -> bool {
        matches!(self, Self::Global(_))
    }

    /// Slot numarasını al
    pub fn index(&self) -> u32 {
        match self {
            Self::Local(idx) => *idx,
            Self::Global(idx) => *idx,
        }
    }
}

/// Sembol bilgisi
#[derive(Debug, Clone)]
pub struct SymbolInfo {
    pub name: String,
    pub kind: SymbolKind,
    pub ty: Option<TypeDecl>,
    pub slot: Slot,
    pub span: Span,
    pub is_public: bool, // Visibility: _ ile başlayanlar private
}

impl SymbolInfo {
    pub fn new(name: String, kind: SymbolKind, ty: Option<TypeDecl>, slot: Slot, span: Span) -> Self {
        let is_public = !name.starts_with('_');
        Self {
            name,
            kind,
            ty,
            slot,
            span,
            is_public,
        }
    }

    /// Sembolün değişken olup olmadığını kontrol et
    pub fn is_variable(&self) -> bool {
        matches!(self.kind, SymbolKind::Variable)
    }

    /// Sembolün fonksiyon olup olmadığını kontrol et
    pub fn is_function(&self) -> bool {
        matches!(self.kind, SymbolKind::Function)
    }

    /// Sembolün parametre olup olmadığını kontrol et
    pub fn is_parameter(&self) -> bool {
        matches!(self.kind, SymbolKind::Parameter)
    }
    
    /// Sembolün public olup olmadığını kontrol et
    pub fn is_public(&self) -> bool {
        self.is_public
    }
    
    /// Sembolün private olup olmadığını kontrol et
    pub fn is_private(&self) -> bool {
        !self.is_public
    }

    /// Sembolün local olup olmadığını kontrol et
    pub fn is_local(&self) -> bool {
        self.slot.is_local()
    }

    /// Sembolün global olup olmadığını kontrol et
    pub fn is_global(&self) -> bool {
        self.slot.is_global()
    }

    /// Sembolün tip bilgisini al
    pub fn get_type(&self) -> Option<TypeDecl> {
        self.ty.clone()
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    fn create_test_symbol() -> SymbolInfo {
        SymbolInfo::new(
            "test_var".to_string(),
            SymbolKind::Variable,
            Some(TypeDecl::Int),
            Slot::Local(0),
            Span::new(0, 0, 0),
        )
    }

    #[test]
    fn test_slot_properties() {
        let local_slot = Slot::Local(5);
        let global_slot = Slot::Global(10);

        assert!(local_slot.is_local());
        assert!(!local_slot.is_global());
        assert_eq!(local_slot.index(), 5);

        assert!(!global_slot.is_local());
        assert!(global_slot.is_global());
        assert_eq!(global_slot.index(), 10);
    }

    #[test]
    fn test_symbol_properties() {
        let symbol = create_test_symbol();

        assert!(symbol.is_variable());
        assert!(!symbol.is_function());
        assert!(!symbol.is_parameter());

        assert!(symbol.is_local());
        assert!(!symbol.is_global());

        assert_eq!(symbol.get_type(), Some(TypeDecl::Int));
    }

    #[test]
    fn test_symbol_kinds() {
        let var_symbol = SymbolInfo::new(
            "var".to_string(),
            SymbolKind::Variable,
            Some(TypeDecl::String),
            Slot::Local(0),
            Span::new(0, 0, 0),
        );

        let func_symbol = SymbolInfo::new(
            "func".to_string(),
            SymbolKind::Function,
            None,
            Slot::Global(0),
            Span::new(0, 0, 0),
        );

        let param_symbol = SymbolInfo::new(
            "param".to_string(),
            SymbolKind::Parameter,
            Some(TypeDecl::Int),
            Slot::Local(0),
            Span::new(0, 0, 0),
        );

        assert!(var_symbol.is_variable());
        assert!(func_symbol.is_function());
        assert!(param_symbol.is_parameter());

        assert!(var_symbol.is_local());
        assert!(func_symbol.is_global());
        assert!(param_symbol.is_local());
    }
}
