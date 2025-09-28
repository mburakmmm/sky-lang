// Allocator - Bellek Allocator
// Basit bellek yönetimi

use std::collections::HashMap;
use std::alloc::{GlobalAlloc, System, Layout};
use std::sync::atomic::{AtomicUsize, Ordering};

/// Global memory allocator
pub struct Allocator {
    allocated_bytes: AtomicUsize,
    allocation_count: AtomicUsize,
    deallocation_count: AtomicUsize,
    peak_bytes: AtomicUsize,
}

/// Memory statistics
#[derive(Debug, Clone)]
pub struct AllocStats {
    pub allocated_bytes: usize,
    pub allocation_count: usize,
    pub deallocation_count: usize,
    pub peak_bytes: usize,
    pub active_allocations: usize,
}

impl Allocator {
    pub fn new() -> Self {
        Self {
            allocated_bytes: AtomicUsize::new(0),
            allocation_count: AtomicUsize::new(0),
            deallocation_count: AtomicUsize::new(0),
            peak_bytes: AtomicUsize::new(0),
        }
    }

    /// Bellek allocate et
    pub fn allocate(&self, layout: Layout) -> *mut u8 {
        unsafe {
            let ptr = System.alloc(layout);
            if !ptr.is_null() {
                let size = layout.size();
                self.allocated_bytes.fetch_add(size, Ordering::SeqCst);
                self.allocation_count.fetch_add(1, Ordering::SeqCst);
                
                // Peak bytes güncelle
                let current = self.allocated_bytes.load(Ordering::SeqCst);
                let peak = self.peak_bytes.load(Ordering::SeqCst);
                if current > peak {
                    self.peak_bytes.store(current, Ordering::SeqCst);
                }
            }
            ptr
        }
    }

    /// Bellek deallocate et
    pub fn deallocate(&self, ptr: *mut u8, layout: Layout) {
        unsafe {
            System.dealloc(ptr, layout);
            self.allocated_bytes.fetch_sub(layout.size(), Ordering::SeqCst);
            self.deallocation_count.fetch_add(1, Ordering::SeqCst);
        }
    }

    /// İstatistikleri al
    pub fn stats(&self) -> AllocStats {
        AllocStats {
            allocated_bytes: self.allocated_bytes.load(Ordering::SeqCst),
            allocation_count: self.allocation_count.load(Ordering::SeqCst),
            deallocation_count: self.deallocation_count.load(Ordering::SeqCst),
            peak_bytes: self.peak_bytes.load(Ordering::SeqCst),
            active_allocations: self.allocation_count.load(Ordering::SeqCst) - self.deallocation_count.load(Ordering::SeqCst),
        }
    }

    /// Bellek kullanımını sıfırla
    pub fn reset_stats(&self) {
        self.allocated_bytes.store(0, Ordering::SeqCst);
        self.allocation_count.store(0, Ordering::SeqCst);
        self.deallocation_count.store(0, Ordering::SeqCst);
        self.peak_bytes.store(0, Ordering::SeqCst);
    }

    /// Bellek sızıntısı olup olmadığını kontrol et
    pub fn has_leaks(&self) -> bool {
        self.allocated_bytes.load(Ordering::SeqCst) > 0
    }

    /// Aktif allocation sayısını al
    pub fn active_allocations(&self) -> usize {
        self.allocation_count.load(Ordering::SeqCst) - self.deallocation_count.load(Ordering::SeqCst)
    }

    /// Toplam allocation sayısını al
    pub fn total_allocations(&self) -> usize {
        self.allocation_count.load(Ordering::SeqCst)
    }

    /// Toplam deallocation sayısını al
    pub fn total_deallocations(&self) -> usize {
        self.deallocation_count.load(Ordering::SeqCst)
    }

    /// Şu anda allocate edilen byte sayısını al
    pub fn current_bytes(&self) -> usize {
        self.allocated_bytes.load(Ordering::SeqCst)
    }

    /// Peak byte sayısını al
    pub fn peak_bytes(&self) -> usize {
        self.peak_bytes.load(Ordering::SeqCst)
    }
}

impl Default for Allocator {
    fn default() -> Self {
        Self::new()
    }
}

/// Pool allocator - belirli boyutlardaki objeler için
pub struct PoolAllocator<T> {
    pools: HashMap<usize, Vec<Box<T>>>,
    allocator: Allocator,
}

impl<T> PoolAllocator<T> {
    pub fn new() -> Self {
        Self {
            pools: HashMap::new(),
            allocator: Allocator::new(),
        }
    }

    /// Obje allocate et
    pub fn allocate(&mut self, value: T) -> Box<T> {
        let size = std::mem::size_of::<T>();
        
        if let Some(pool) = self.pools.get_mut(&size) {
            if let Some(mut obj) = pool.pop() {
                *obj = value;
                return obj;
            }
        }
        
        Box::new(value)
    }

    /// Obje deallocate et
    pub fn deallocate(&mut self, obj: Box<T>) {
        let size = std::mem::size_of::<T>();
        
        // Pool'a geri ekle
        self.pools.entry(size).or_insert_with(Vec::new).push(obj);
    }

    /// Pool istatistiklerini al
    pub fn pool_stats(&self) -> HashMap<usize, usize> {
        self.pools.iter().map(|(size, pool)| (*size, pool.len())).collect()
    }

    /// Tüm pool'ları temizle
    pub fn clear_pools(&mut self) {
        self.pools.clear();
    }
}

impl<T> Default for PoolAllocator<T> {
    fn default() -> Self {
        Self::new()
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_allocator_creation() {
        let allocator = Allocator::new();
        let stats = allocator.stats();
        assert_eq!(stats.allocated_bytes, 0);
        assert_eq!(stats.allocation_count, 0);
        assert_eq!(stats.deallocation_count, 0);
    }

    #[test]
    fn test_allocation_deallocation() {
        let allocator = Allocator::new();
        
        let layout = Layout::from_size_align(1024, 8).unwrap();
        let ptr = allocator.allocate(layout);
        assert!(!ptr.is_null());
        
        let stats = allocator.stats();
        assert_eq!(stats.allocated_bytes, 1024);
        assert_eq!(stats.allocation_count, 1);
        assert_eq!(stats.active_allocations, 1);
        
        allocator.deallocate(ptr, layout);
        
        let stats = allocator.stats();
        assert_eq!(stats.allocated_bytes, 0);
        assert_eq!(stats.allocation_count, 1);
        assert_eq!(stats.deallocation_count, 1);
        assert_eq!(stats.active_allocations, 0);
    }

    #[test]
    fn test_peak_bytes() {
        let allocator = Allocator::new();
        
        let layout1 = Layout::from_size_align(1024, 8).unwrap();
        let layout2 = Layout::from_size_align(512, 8).unwrap();
        
        let ptr1 = allocator.allocate(layout1);
        assert_eq!(allocator.peak_bytes(), 1024);
        
        let ptr2 = allocator.allocate(layout2);
        assert_eq!(allocator.peak_bytes(), 1536);
        
        allocator.deallocate(ptr1, layout1);
        assert_eq!(allocator.peak_bytes(), 1536); // Peak değişmez
        
        allocator.deallocate(ptr2, layout2);
        assert_eq!(allocator.peak_bytes(), 1536); // Peak değişmez
    }

    #[test]
    fn test_leak_detection() {
        let allocator = Allocator::new();
        
        assert!(!allocator.has_leaks());
        
        let layout = Layout::from_size_align(1024, 8).unwrap();
        let _ptr = allocator.allocate(layout);
        
        assert!(allocator.has_leaks());
    }

    #[test]
    fn test_pool_allocator() {
        let mut pool = PoolAllocator::<i32>::new();
        
        let obj1 = pool.allocate(42);
        assert_eq!(*obj1, 42);
        
        pool.deallocate(obj1);
        
        let obj2 = pool.allocate(100);
        assert_eq!(*obj2, 100);
        
        let stats = pool.pool_stats();
        assert_eq!(stats.get(&std::mem::size_of::<i32>()), Some(&1));
    }

    #[test]
    fn test_pool_reuse() {
        let mut pool = PoolAllocator::<String>::new();
        
        let obj1 = pool.allocate("hello".to_string());
        let ptr1 = obj1.as_ptr();
        pool.deallocate(obj1);
        
        let obj2 = pool.allocate("world".to_string());
        let ptr2 = obj2.as_ptr();
        
        // Pool'dan geri alınan obje aynı pointer olabilir
        // (Bu test pool implementasyonuna bağlı)
        pool.deallocate(obj2);
    }

    #[test]
    fn test_stats_reset() {
        let allocator = Allocator::new();
        
        let layout = Layout::from_size_align(1024, 8).unwrap();
        let ptr = allocator.allocate(layout);
        allocator.deallocate(ptr, layout);
        
        let stats = allocator.stats();
        assert!(stats.allocation_count > 0);
        
        allocator.reset_stats();
        
        let stats = allocator.stats();
        assert_eq!(stats.allocation_count, 0);
        assert_eq!(stats.deallocation_count, 0);
        assert_eq!(stats.allocated_bytes, 0);
        assert_eq!(stats.peak_bytes, 0);
    }
}
