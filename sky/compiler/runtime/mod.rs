// Runtime - GC & Bellek Yardımcıları
// Mark-sweep GC, string interning, basit allocator

pub mod gc;
pub mod alloc;
pub mod strings;
pub mod interner;

use crate::compiler::vm::value::Value;

// Re-exports
pub use gc::{Gc, GcRootSet, Heap};
pub use strings::StringInterner;
pub use alloc::Allocator;

/// Runtime yöneticisi
pub struct Runtime {
    pub heap: Heap,
    pub roots: GcRootSet,
    pub strings: StringInterner,
    pub allocator: Allocator,
}

impl Runtime {
    pub fn new() -> Self {
        Self {
            heap: Heap::new(),
            roots: GcRootSet::new(),
            strings: StringInterner::new(),
            allocator: Allocator::new(),
        }
    }

    /// Garbage collection çalıştır
    pub fn gc_collect(&mut self) -> usize {
        self.heap.collect(&self.roots)
    }

    /// String intern et
    pub fn intern_string(&mut self, string: &str) -> Gc<crate::compiler::vm::value::Value> {
        self.strings.intern(string, &mut self.heap)
    }

    /// Yeni değer allocate et
    pub fn allocate_value(&mut self, value: Value) -> Gc<Value> {
        self.heap.allocate(value)
    }

    /// Root ekle
    pub fn add_root(&mut self, value: &Gc<Value>) {
        self.roots.add(value);
    }

    /// Root kaldır
    pub fn remove_root(&mut self, value: &Gc<Value>) {
        self.roots.remove(value);
    }

    /// Heap istatistiklerini al
    pub fn heap_stats(&self) -> HeapStats {
        self.heap.stats()
    }

    /// String pool istatistiklerini al
    pub fn string_stats(&self) -> StringStats {
        self.strings.stats()
    }
}

/// Heap istatistikleri
#[derive(Debug, Clone)]
pub struct HeapStats {
    pub total_bytes: usize,
    pub used_bytes: usize,
    pub free_bytes: usize,
    pub object_count: usize,
    pub gc_count: usize,
}

/// String pool istatistikleri
#[derive(Debug, Clone)]
pub struct StringStats {
    pub total_strings: usize,
    pub total_bytes: usize,
    pub unique_strings: usize,
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_runtime_creation() {
        let runtime = Runtime::new();
        let stats = runtime.heap_stats();
        assert_eq!(stats.total_bytes, 0);
        assert_eq!(stats.object_count, 0);
    }

    #[test]
    fn test_string_intern() {
        let mut runtime = Runtime::new();
        let s1 = runtime.intern_string("hello");
        let s2 = runtime.intern_string("hello");
        let s3 = runtime.intern_string("world");
        
        // Aynı string'ler aynı referans olmalı
        assert_eq!(s1.as_ptr(), s2.as_ptr());
        assert_ne!(s1.as_ptr(), s3.as_ptr());
    }

    #[test]
    fn test_value_allocation() {
        let mut runtime = Runtime::new();
        let value = runtime.allocate_value(Value::Int(42));
        assert_eq!(**value, Value::Int(42));
    }

    #[test]
    fn test_gc_collection() {
        let mut runtime = Runtime::new();
        let value = runtime.allocate_value(Value::Int(42));
        runtime.add_root(&value);
        
        let collected = runtime.gc_collect();
        assert_eq!(collected, 0); // Root olan değer toplanmamalı
        
        runtime.remove_root(&value);
        let collected = runtime.gc_collect();
        assert!(collected > 0); // Artık root olmayan değer toplanmalı
    }
}
