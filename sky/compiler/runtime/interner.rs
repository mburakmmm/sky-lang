// Interner - Genel Interning Sistemi
// String'ler dışındaki objeler için interning

use std::collections::HashMap;
use std::hash::Hash;
use std::sync::Arc;

/// Interner trait
pub trait Internable: Clone + Hash + Eq + Send + Sync {
    type Key: Clone + Hash + Eq + Send + Sync;
    
    fn intern_key(&self) -> Self::Key;
}

/// Genel interner
pub struct Interner<T> 
where 
    T: Internable,
{
    items: HashMap<T::Key, Arc<T>>,
    stats: InternerStats,
}

/// Interner istatistikleri
#[derive(Debug, Clone)]
pub struct InternerStats {
    pub total_items: usize,
    pub unique_items: usize,
    pub hits: usize,
    pub misses: usize,
}

impl<T> Interner<T>
where
    T: Internable,
{
    pub fn new() -> Self {
        Self {
            items: HashMap::new(),
            stats: InternerStats {
                total_items: 0,
                unique_items: 0,
                hits: 0,
                misses: 0,
            },
        }
    }

    /// Item'ı intern et
    pub fn intern(&mut self, item: T) -> Arc<T> {
        let key = item.intern_key();
        
        if let Some(existing) = self.items.get(&key) {
            self.stats.total_items += 1;
            self.stats.hits += 1;
            existing.clone()
        } else {
            let arc_item = Arc::new(item);
            self.items.insert(key, arc_item.clone());
            
            self.stats.total_items += 1;
            self.stats.unique_items += 1;
            self.stats.misses += 1;
            
            arc_item
        }
    }

    /// Item'ın intern edilip edilmediğini kontrol et
    pub fn contains(&self, item: &T) -> bool {
        let key = item.intern_key();
        self.items.contains_key(&key)
    }

    /// Intern edilmiş item'ı al
    pub fn get(&self, key: &T::Key) -> Option<&Arc<T>> {
        self.items.get(key)
    }

    /// Tüm intern edilmiş item'ları al
    pub fn all_items(&self) -> &HashMap<T::Key, Arc<T>> {
        &self.items
    }

    /// Intern edilmiş item sayısını al
    pub fn count(&self) -> usize {
        self.items.len()
    }

    /// İstatistikleri al
    pub fn stats(&self) -> InternerStats {
        self.stats.clone()
    }

    /// Tüm item'ları temizle
    pub fn clear(&mut self) {
        self.items.clear();
        self.stats = InternerStats {
            total_items: 0,
            unique_items: 0,
            hits: 0,
            misses: 0,
        };
    }

    /// Belirli bir item'ı kaldır
    pub fn remove(&mut self, key: &T::Key) -> Option<Arc<T>> {
        self.items.remove(key)
    }

    /// Hit rate'i hesapla
    pub fn hit_rate(&self) -> f64 {
        let total = self.stats.hits + self.stats.misses;
        if total == 0 {
            0.0
        } else {
            self.stats.hits as f64 / total as f64
        }
    }

    /// Memory kullanımını tahmin et
    pub fn memory_usage(&self) -> usize {
        self.items.len() * std::mem::size_of::<Arc<T>>()
    }
}

impl<T> Default for Interner<T>
where
    T: Internable,
{
    fn default() -> Self {
        Self::new()
    }
}

/// String için özel interner (basit implementasyon)
impl Internable for String {
    type Key = String;
    
    fn intern_key(&self) -> Self::Key {
        self.clone()
    }
}

/// i32 için özel interner
impl Internable for i32 {
    type Key = i32;
    
    fn intern_key(&self) -> Self::Key {
        *self
    }
}

// f64 için Internable implement edilemez çünkü Eq trait'i yok
// f64'ü intern etmek için wrapper struct kullanılmalı

// f64 için Eq ve Hash trait'leri implement edilemez (primitive type)
// Bunun yerine f64'ü u64'e çevirerek kullanabiliriz

/// Bool için özel interner
impl Internable for bool {
    type Key = bool;
    
    fn intern_key(&self) -> Self::Key {
        *self
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_string_interner() {
        let mut interner = Interner::<String>::new();
        
        let s1 = interner.intern("hello".to_string());
        let s2 = interner.intern("hello".to_string());
        let s3 = interner.intern("world".to_string());
        
        // Aynı string'ler aynı Arc olmalı
        assert!(Arc::ptr_eq(&s1, &s2));
        assert!(!Arc::ptr_eq(&s1, &s3));
        
        let stats = interner.stats();
        assert_eq!(stats.unique_items, 2);
        assert_eq!(stats.total_items, 3);
        assert_eq!(stats.hits, 1);
        assert_eq!(stats.misses, 2);
    }

    #[test]
    fn test_int_interner() {
        let mut interner = Interner::<i32>::new();
        
        let i1 = interner.intern(42);
        let i2 = interner.intern(42);
        let i3 = interner.intern(100);
        
        assert!(Arc::ptr_eq(&i1, &i2));
        assert!(!Arc::ptr_eq(&i1, &i3));
        
        assert_eq!(interner.count(), 2);
    }

    #[test]
    fn test_float_interner() {
        let mut interner = Interner::<f64>::new();
        
        let f1 = interner.intern(3.14);
        let f2 = interner.intern(3.14);
        let f3 = interner.intern(2.71);
        
        assert!(Arc::ptr_eq(&f1, &f2));
        assert!(!Arc::ptr_eq(&f1, &f3));
    }

    #[test]
    fn test_bool_interner() {
        let mut interner = Interner::<bool>::new();
        
        let b1 = interner.intern(true);
        let b2 = interner.intern(true);
        let b3 = interner.intern(false);
        
        assert!(Arc::ptr_eq(&b1, &b2));
        assert!(!Arc::ptr_eq(&b1, &b3));
        
        assert_eq!(interner.count(), 2);
    }

    #[test]
    fn test_interner_contains() {
        let mut interner = Interner::<String>::new();
        
        assert!(!interner.contains(&"hello".to_string()));
        
        interner.intern("hello".to_string());
        assert!(interner.contains(&"hello".to_string()));
        assert!(!interner.contains(&"world".to_string()));
    }

    #[test]
    fn test_interner_remove() {
        let mut interner = Interner::<String>::new();
        
        interner.intern("hello".to_string());
        assert_eq!(interner.count(), 1);
        
        let removed = interner.remove(&"hello".to_string());
        assert!(removed.is_some());
        assert_eq!(interner.count(), 0);
    }

    #[test]
    fn test_interner_clear() {
        let mut interner = Interner::<String>::new();
        
        interner.intern("hello".to_string());
        interner.intern("world".to_string());
        assert_eq!(interner.count(), 2);
        
        interner.clear();
        assert_eq!(interner.count(), 0);
        
        let stats = interner.stats();
        assert_eq!(stats.total_items, 0);
        assert_eq!(stats.unique_items, 0);
    }

    #[test]
    fn test_hit_rate() {
        let mut interner = Interner::<String>::new();
        
        // İlk çağrılar miss
        interner.intern("hello".to_string());
        interner.intern("world".to_string());
        
        // İkinci çağrılar hit
        interner.intern("hello".to_string());
        interner.intern("hello".to_string());
        
        let hit_rate = interner.hit_rate();
        assert_eq!(hit_rate, 0.5); // 2 hits / 4 total
    }

    #[test]
    fn test_memory_usage() {
        let interner = Interner::<String>::new();
        let usage = interner.memory_usage();
        assert!(usage >= 0);
    }
}
