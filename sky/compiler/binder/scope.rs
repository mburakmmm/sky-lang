// Scope Management - Kapsam Yönetimi
// Scope stack ve sembol çözümleme mantığı

use std::collections::HashMap;
use super::symbols::SymbolInfo;

/// Tek bir scope
#[derive(Debug, Clone)]
pub struct Scope {
    symbols: HashMap<String, SymbolInfo>,
}

impl Scope {
    pub fn new() -> Self {
        Self {
            symbols: HashMap::new(),
        }
    }

    pub fn insert(&mut self, name: String, symbol: SymbolInfo) {
        self.symbols.insert(name, symbol);
    }

    pub fn get(&self, name: &str) -> Option<&SymbolInfo> {
        self.symbols.get(name)
    }

    pub fn contains(&self, name: &str) -> bool {
        self.symbols.contains_key(name)
    }
}

/// Scope stack - iç içe kapsamları yönetir
#[derive(Debug, Clone)]
pub struct ScopeStack {
    scopes: Vec<Scope>,
}

impl ScopeStack {
    pub fn new() -> Self {
        Self {
            scopes: Vec::new(),
        }
    }

    pub fn push(&mut self, scope: Scope) {
        self.scopes.push(scope);
    }

    pub fn pop(&mut self) -> Option<Scope> {
        self.scopes.pop()
    }

    pub fn current(&self) -> &Scope {
        self.scopes.last().expect("Scope stack is empty")
    }

    pub fn current_mut(&mut self) -> &mut Scope {
        self.scopes.last_mut().expect("Scope stack is empty")
    }

    pub fn depth(&self) -> usize {
        self.scopes.len()
    }

    /// En yakın scope'tan başlayarak sembolü çözümle
    pub fn resolve(&self, name: &str) -> Option<&SymbolInfo> {
        for scope in self.scopes.iter().rev() {
            if let Some(symbol) = scope.get(name) {
                return Some(symbol);
            }
        }
        None
    }

    /// Tüm scope'ları döndür (debugging için)
    pub fn all_scopes(&self) -> Vec<Scope> {
        self.scopes.clone()
    }

    /// Scope stack'in boş olup olmadığını kontrol et
    pub fn is_empty(&self) -> bool {
        self.scopes.is_empty()
    }

    /// Global scope'ta olup olmadığını kontrol et (sadece 1 scope varsa global)
    pub fn is_global_scope(&self) -> bool {
        self.scopes.len() == 1
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use super::symbols::{SymbolInfo, SymbolKind, Slot};

    fn create_test_symbol(name: &str) -> SymbolInfo {
        SymbolInfo {
            name: name.to_string(),
            kind: SymbolKind::Variable,
            ty: Some(crate::compiler::types::TypeDecl::Int),
            slot: Slot::Local(0),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        }
    }

    #[test]
    fn test_scope_basic_operations() {
        let mut scope = Scope::new();
        let symbol = create_test_symbol("x");
        
        assert!(!scope.contains("x"));
        assert_eq!(scope.get("x"), None);
        
        scope.insert("x".to_string(), symbol.clone());
        
        assert!(scope.contains("x"));
        assert_eq!(scope.get("x"), Some(&symbol));
    }

    #[test]
    fn test_scope_stack_operations() {
        let mut stack = ScopeStack::new();
        
        assert!(stack.is_empty());
        
        let mut scope1 = Scope::new();
        scope1.insert("x".to_string(), create_test_symbol("x"));
        stack.push(scope1);
        
        assert!(!stack.is_empty());
        assert_eq!(stack.depth(), 1);
        
        let mut scope2 = Scope::new();
        scope2.insert("y".to_string(), create_test_symbol("y"));
        stack.push(scope2);
        
        assert_eq!(stack.depth(), 2);
        
        let popped = stack.pop();
        assert!(popped.is_some());
        assert_eq!(stack.depth(), 1);
    }

    #[test]
    fn test_symbol_resolution() {
        let mut stack = ScopeStack::new();
        
        // Global scope
        let mut global = Scope::new();
        global.insert("global_var".to_string(), create_test_symbol("global_var"));
        stack.push(global);
        
        // Local scope
        let mut local = Scope::new();
        local.insert("local_var".to_string(), create_test_symbol("local_var"));
        local.insert("global_var".to_string(), create_test_symbol("shadowed_global_var"));
        stack.push(local);
        
        // Local değişken öncelikli olmalı
        let resolved = stack.resolve("global_var");
        assert!(resolved.is_some());
        assert_eq!(resolved.unwrap().name, "shadowed_global_var");
        
        // Sadece local'de olan değişken
        let resolved = stack.resolve("local_var");
        assert!(resolved.is_some());
        assert_eq!(resolved.unwrap().name, "local_var");
        
        // Olmayan değişken
        let resolved = stack.resolve("nonexistent");
        assert!(resolved.is_none());
    }

    #[test]
    fn test_nested_scopes() {
        let mut stack = ScopeStack::new();
        
        // En dış scope
        let mut outer = Scope::new();
        outer.insert("x".to_string(), create_test_symbol("outer_x"));
        stack.push(outer);
        
        // Orta scope
        let mut middle = Scope::new();
        middle.insert("y".to_string(), create_test_symbol("middle_y"));
        stack.push(middle);
        
        // En iç scope
        let mut inner = Scope::new();
        inner.insert("z".to_string(), create_test_symbol("inner_z"));
        inner.insert("x".to_string(), create_test_symbol("inner_x"));
        stack.push(inner);
        
        // Çözümleme testleri
        assert_eq!(stack.resolve("x").unwrap().name, "inner_x");
        assert_eq!(stack.resolve("y").unwrap().name, "middle_y");
        assert_eq!(stack.resolve("z").unwrap().name, "inner_z");
        
        // Scope'ları kapat
        stack.pop(); // inner
        assert_eq!(stack.resolve("x").unwrap().name, "outer_x");
        assert_eq!(stack.resolve("y").unwrap().name, "middle_y");
        assert!(stack.resolve("z").is_none());
        
        stack.pop(); // middle
        assert_eq!(stack.resolve("x").unwrap().name, "outer_x");
        assert!(stack.resolve("y").is_none());
        assert!(stack.resolve("z").is_none());
    }
}
