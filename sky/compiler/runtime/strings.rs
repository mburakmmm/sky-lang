// String Interning - String Havuzu
// Aynı string'lerin tek kopyasını tutar

use std::collections::HashMap;
use super::gc::{Gc, Heap};

/// String interner
pub struct StringInterner {
    strings: HashMap<String, Gc<crate::compiler::vm::value::Value>>,
    stats: StringInternerStats,
}

/// String interner istatistikleri
#[derive(Debug, Clone)]
struct StringInternerStats {
    total_strings: usize,
    total_bytes: usize,
    unique_strings: usize,
}

impl StringInterner {
    pub fn new() -> Self {
        Self {
            strings: HashMap::new(),
            stats: StringInternerStats {
                total_strings: 0,
                total_bytes: 0,
                unique_strings: 0,
            },
        }
    }

    /// String'i intern et
    pub fn intern(&mut self, string: &str, heap: &mut Heap) -> Gc<crate::compiler::vm::value::Value> {
        if let Some(gc_string) = self.strings.get(string) {
            self.stats.total_strings += 1;
            gc_string.clone()
        } else {
            let gc_string = heap.allocate(crate::compiler::vm::value::Value::String(string.to_string()));
            self.strings.insert(string.to_string(), gc_string.clone());
            
            self.stats.total_strings += 1;
            self.stats.total_bytes += string.len();
            self.stats.unique_strings += 1;
            
            gc_string
        }
    }

    /// String'in intern edilip edilmediğini kontrol et
    pub fn contains(&self, string: &str) -> bool {
        self.strings.contains_key(string)
    }

    /// Intern edilmiş string'i al
    pub fn get(&self, string: &str) -> Option<&Gc<crate::compiler::vm::value::Value>> {
        self.strings.get(string)
    }

    /// Tüm intern edilmiş string'leri al
    pub fn all_strings(&self) -> &HashMap<String, Gc<crate::compiler::vm::value::Value>> {
        &self.strings
    }

    /// Intern edilmiş string sayısını al
    pub fn count(&self) -> usize {
        self.strings.len()
    }

    /// İstatistikleri al
    pub fn stats(&self) -> super::StringStats {
        super::StringStats {
            total_strings: self.stats.total_strings,
            total_bytes: self.stats.total_bytes,
            unique_strings: self.stats.unique_strings,
        }
    }

    /// Tüm string'leri temizle
    pub fn clear(&mut self) {
        self.strings.clear();
        self.stats = StringInternerStats {
            total_strings: 0,
            total_bytes: 0,
            unique_strings: 0,
        };
    }

    /// Belirli bir string'i kaldır
    pub fn remove(&mut self, string: &str) -> Option<Gc<crate::compiler::vm::value::Value>> {
        if let Some(gc_string) = self.strings.remove(string) {
            self.stats.unique_strings -= 1;
            self.stats.total_bytes -= string.len();
            Some(gc_string)
        } else {
            None
        }
    }

    /// String pool'undaki memory kullanımını hesapla
    pub fn memory_usage(&self) -> usize {
        self.strings.values().map(|s| match &**s {
            crate::compiler::vm::value::Value::String(string) => string.len(),
            _ => 0,
        }).sum()
    }

    /// En uzun string'i bul
    pub fn longest_string(&self) -> Option<&str> {
        self.strings.keys().max_by_key(|s| s.len()).map(|s| s.as_str())
    }

    /// En kısa string'i bul
    pub fn shortest_string(&self) -> Option<&str> {
        self.strings.keys().min_by_key(|s| s.len()).map(|s| s.as_str())
    }

    /// Ortalama string uzunluğunu hesapla
    pub fn average_length(&self) -> f64 {
        if self.strings.is_empty() {
            0.0
        } else {
            let total_length: usize = self.strings.keys().map(|s| s.len()).sum();
            total_length as f64 / self.strings.len() as f64
        }
    }

    /// Belirli bir uzunluktaki string'leri say
    pub fn count_by_length(&self, length: usize) -> usize {
        self.strings.keys().filter(|s| s.len() == length).count()
    }

    /// String'leri uzunluğa göre grupla
    pub fn group_by_length(&self) -> HashMap<usize, usize> {
        let mut groups = HashMap::new();
        for string in self.strings.keys() {
            let count = groups.entry(string.len()).or_insert(0);
            *count += 1;
        }
        groups
    }
}

impl Default for StringInterner {
    fn default() -> Self {
        Self::new()
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use super::super::gc::Heap;

    #[test]
    fn test_string_interner_creation() {
        let interner = StringInterner::new();
        assert_eq!(interner.count(), 0);
        assert!(interner.is_empty());
    }

    #[test]
    fn test_string_intern() {
        let mut interner = StringInterner::new();
        let mut heap = Heap::new();
        
        let s1 = interner.intern("hello", &mut heap);
        let s2 = interner.intern("hello", &mut heap);
        
        // Aynı string'ler aynı referans olmalı
        assert_eq!(s1.as_ptr(), s2.as_ptr());
        assert_eq!(interner.count(), 1);
    }

    #[test]
    fn test_string_intern_different_strings() {
        let mut interner = StringInterner::new();
        let mut heap = Heap::new();
        
        let s1 = interner.intern("hello", &mut heap);
        let s2 = interner.intern("world", &mut heap);
        
        // Farklı string'ler farklı referans olmalı
        assert_ne!(s1.as_ptr(), s2.as_ptr());
        assert_eq!(interner.count(), 2);
    }

    #[test]
    fn test_string_interner_stats() {
        let mut interner = StringInterner::new();
        let mut heap = Heap::new();
        
        interner.intern("hello", &mut heap);
        interner.intern("hello", &mut heap); // Duplicate
        interner.intern("world", &mut heap);
        
        let stats = interner.stats();
        assert_eq!(stats.unique_strings, 2);
        assert_eq!(stats.total_strings, 3);
        assert_eq!(stats.total_bytes, 10); // "hello" + "world"
    }

    #[test]
    fn test_string_contains() {
        let mut interner = StringInterner::new();
        let mut heap = Heap::new();
        
        assert!(!interner.contains("hello"));
        
        interner.intern("hello", &mut heap);
        assert!(interner.contains("hello"));
        assert!(!interner.contains("world"));
    }

    #[test]
    fn test_string_remove() {
        let mut interner = StringInterner::new();
        let mut heap = Heap::new();
        
        interner.intern("hello", &mut heap);
        assert_eq!(interner.count(), 1);
        
        let removed = interner.remove("hello");
        assert!(removed.is_some());
        assert_eq!(interner.count(), 0);
        
        let not_found = interner.remove("world");
        assert!(not_found.is_none());
    }

    #[test]
    fn test_string_interner_clear() {
        let mut interner = StringInterner::new();
        let mut heap = Heap::new();
        
        interner.intern("hello", &mut heap);
        interner.intern("world", &mut heap);
        assert_eq!(interner.count(), 2);
        
        interner.clear();
        assert_eq!(interner.count(), 0);
    }

    #[test]
    fn test_string_length_stats() {
        let mut interner = StringInterner::new();
        let mut heap = Heap::new();
        
        interner.intern("a", &mut heap);
        interner.intern("hello", &mut heap);
        interner.intern("world", &mut heap);
        interner.intern("test", &mut heap);
        
        assert_eq!(interner.shortest_string(), Some("a"));
        assert_eq!(interner.longest_string(), Some("hello"));
        assert_eq!(interner.average_length(), 4.25); // (1+5+5+4)/4
        assert_eq!(interner.count_by_length(5), 2);
    }
}
