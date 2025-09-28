// Garbage Collector - Mark-Sweep GC
// Single-threaded mark-sweep garbage collector

use std::collections::HashSet;
use std::ptr::NonNull;
use crate::compiler::vm::Value;

/// Garbage collected pointer
pub struct Gc<T> {
    ptr: NonNull<GcObject<T>>,
}

/// GC object wrapper
pub struct GcObject<T> {
    pub data: T,
    marked: bool,
    size: usize,
}

/// GC heap
pub struct Heap {
    objects: Vec<GcObject<Value>>,
    free_list: Vec<usize>,
    bytes_allocated: usize,
    next_gc: usize,
    gc_threshold: usize,
    gc_count: usize,
}

/// GC root set
pub struct GcRootSet {
    roots: HashSet<*const GcObject<Value>>,
}

impl<T> Gc<T> {
    pub fn new(data: T) -> Self {
        // Bu basit bir implementasyon - gerçek implementasyon daha karmaşık olacak
        unsafe {
            let ptr = std::alloc::alloc(std::alloc::Layout::new::<GcObject<T>>()) as *mut GcObject<T>;
            std::ptr::write(ptr, GcObject {
                data,
                marked: false,
                size: std::mem::size_of::<GcObject<T>>(),
            });
            Self {
                ptr: NonNull::new_unchecked(ptr),
            }
        }
    }

    pub fn as_ptr(&self) -> *const T {
        unsafe { &self.ptr.as_ref().data as *const T }
    }

    pub fn as_mut_ptr(&mut self) -> *mut T {
        unsafe { &mut self.ptr.as_mut().data as *mut T }
    }
}

impl<T> std::ops::Deref for Gc<T> {
    type Target = T;

    fn deref(&self) -> &Self::Target {
        unsafe { &self.ptr.as_ref().data }
    }
}

impl<T> std::ops::DerefMut for Gc<T> {
    fn deref_mut(&mut self) -> &mut Self::Target {
        unsafe { &mut self.ptr.as_mut().data }
    }
}

impl<T> Clone for Gc<T> {
    fn clone(&self) -> Self {
        Self {
            ptr: self.ptr,
        }
    }
}

impl<T> std::fmt::Debug for Gc<T>
where
    T: std::fmt::Debug,
{
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        f.debug_struct("Gc")
            .field("data", &**self)
            .finish()
    }
}

impl<T> PartialEq for Gc<T>
where
    T: PartialEq,
{
    fn eq(&self, other: &Self) -> bool {
        **self == **other
    }
}

impl Heap {
    pub fn new() -> Self {
        Self {
            objects: Vec::new(),
            free_list: Vec::new(),
            bytes_allocated: 0,
            next_gc: 1024, // 1KB threshold
            gc_threshold: 1024,
            gc_count: 0,
        }
    }

    pub fn allocate(&mut self, value: Value) -> Gc<Value> {
        let size = self.size_of_value(&value);
        
        if self.bytes_allocated + size > self.next_gc {
            self.collect(&GcRootSet::new());
        }

        let gc_object = GcObject {
            data: value,
            marked: false,
            size,
        };

        self.bytes_allocated += size;
        self.objects.push(gc_object);
        
        let index = self.objects.len() - 1;
        
        // Gc wrapper oluştur
        Gc::new(self.objects[index].data.clone())
    }

    pub fn collect(&mut self, roots: &GcRootSet) -> usize {
        let before = self.objects.len();
        
        // Mark phase
        for root in &roots.roots {
            if let Some(index) = self.find_object_index(*root) {
                self.mark_object(index);
            }
        }
        
        // Sweep phase
        self.sweep();
        
        self.gc_count += 1;
        self.next_gc = self.bytes_allocated * 2;
        
        before - self.objects.len()
    }

    fn mark_object(&mut self, index: usize) {
        if index >= self.objects.len() {
            return;
        }
        
        if self.objects[index].marked {
            return; // Already marked
        }
        
        self.objects[index].marked = true;
        
        // Recursively mark referenced objects
        // Bu basit implementasyon - gerçek implementasyon daha karmaşık olacak
    }

    fn sweep(&mut self) {
        let mut i = 0;
        while i < self.objects.len() {
            if self.objects[i].marked {
                self.objects[i].marked = false;
                i += 1;
            } else {
                self.bytes_allocated -= self.objects[i].size;
                self.objects.remove(i);
            }
        }
    }

    fn find_object_index(&self, ptr: *const GcObject<Value>) -> Option<usize> {
        for (i, obj) in self.objects.iter().enumerate() {
            if std::ptr::addr_of!(*obj) == ptr {
                return Some(i);
            }
        }
        None
    }

    fn size_of_value(&self, value: &Value) -> usize {
        match value {
            Value::Int(_) => std::mem::size_of::<i64>(),
            Value::Float(_) => std::mem::size_of::<f64>(),
            Value::Bool(_) => std::mem::size_of::<bool>(),
            Value::String(s) => std::mem::size_of::<String>() + s.len(),
            Value::List(l) => std::mem::size_of::<Vec<Value>>() + l.len() * std::mem::size_of::<Value>(),
            Value::Map(m) => std::mem::size_of::<std::collections::HashMap<String, Value>>() + m.len() * (std::mem::size_of::<String>() + std::mem::size_of::<Value>()),
            _ => std::mem::size_of::<Value>(),
        }
    }

    pub fn stats(&self) -> super::HeapStats {
        super::HeapStats {
            total_bytes: self.bytes_allocated,
            used_bytes: self.bytes_allocated,
            free_bytes: 0, // Basit implementasyon
            object_count: self.objects.len(),
            gc_count: self.gc_count,
        }
    }
}

impl GcRootSet {
    pub fn new() -> Self {
        Self {
            roots: HashSet::new(),
        }
    }

    pub fn add(&mut self, value: &Gc<Value>) {
        self.roots.insert(value.as_ptr() as *const GcObject<Value>);
    }

    pub fn remove(&mut self, value: &Gc<Value>) {
        self.roots.remove(&(value.as_ptr() as *const GcObject<Value>));
    }

    pub fn contains(&self, value: &Gc<Value>) -> bool {
        self.roots.contains(&(value.as_ptr() as *const GcObject<Value>))
    }

    pub fn len(&self) -> usize {
        self.roots.len()
    }

    pub fn is_empty(&self) -> bool {
        self.roots.is_empty()
    }

    pub fn clear(&mut self) {
        self.roots.clear();
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_gc_creation() {
        let gc = Gc::new(42i32);
        assert_eq!(*gc, 42);
    }

    #[test]
    fn test_heap_allocation() {
        let mut heap = Heap::new();
        let value = heap.allocate(Value::Int(42));
        assert_eq!(**value, Value::Int(42));
        assert_eq!(heap.stats().object_count, 1);
    }

    #[test]
    fn test_gc_collection() {
        let mut heap = Heap::new();
        let value = heap.allocate(Value::Int(42));
        
        let mut roots = GcRootSet::new();
        roots.add(&value);
        
        let collected = heap.collect(&roots);
        assert_eq!(collected, 0); // Root olan değer toplanmamalı
        
        roots.remove(&value);
        let collected = heap.collect(&roots);
        assert!(collected > 0); // Artık root olmayan değer toplanmalı
    }

    #[test]
    fn test_root_set() {
        let mut roots = GcRootSet::new();
        let mut heap = Heap::new();
        let value = heap.allocate(Value::Int(42));
        
        assert!(!roots.contains(&value));
        assert!(roots.is_empty());
        
        roots.add(&value);
        assert!(roots.contains(&value));
        assert_eq!(roots.len(), 1);
        
        roots.remove(&value);
        assert!(!roots.contains(&value));
        assert!(roots.is_empty());
    }

    #[test]
    fn test_gc_clone() {
        let gc1 = Gc::new(42i32);
        let gc2 = gc1.clone();
        
        assert_eq!(*gc1, *gc2);
        assert_eq!(gc1.as_ptr(), gc2.as_ptr()); // Same pointer
    }
}
