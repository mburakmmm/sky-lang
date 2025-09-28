// Stack - VM Stack Sistemi
// Stack tabanlı VM için stack yönetimi

use super::Value;
use super::RuntimeError;

const STACK_MAX: usize = 256;

/// VM Stack
#[derive(Debug)]
pub struct Stack {
    values: Vec<Value>,
}

impl Stack {
    pub fn new() -> Self {
        Self {
            values: Vec::with_capacity(STACK_MAX),
        }
    }

    /// Stack'e değer ekle
    pub fn push(&mut self, value: Value) -> Result<(), RuntimeError> {
        if self.values.len() >= STACK_MAX {
            return Err(RuntimeError::StackOverflow);
        }
        self.values.push(value);
        Ok(())
    }

    /// Stack'ten değer çıkar
    pub fn pop(&mut self) -> Result<Value, RuntimeError> {
        self.values.pop().ok_or(RuntimeError::StackUnderflow)
    }

    /// Stack'in üstündeki değeri al (çıkarmadan)
    pub fn peek(&self, distance: usize) -> Result<&Value, RuntimeError> {
        if distance >= self.values.len() {
            return Err(RuntimeError::StackUnderflow);
        }
        Ok(&self.values[self.values.len() - 1 - distance])
    }

    /// Stack'in üstündeki değeri değiştir
    pub fn peek_mut(&mut self, distance: usize) -> Result<&mut Value, RuntimeError> {
        if distance >= self.values.len() {
            return Err(RuntimeError::StackUnderflow);
        }
        let len = self.values.len();
        Ok(&mut self.values[len - 1 - distance])
    }

    /// Stack'teki değer sayısını döndür
    pub fn len(&self) -> usize {
        self.values.len()
    }

    /// Stack'in boş olup olmadığını kontrol et
    pub fn is_empty(&self) -> bool {
        self.values.is_empty()
    }

    /// Stack'i temizle
    pub fn clear(&mut self) {
        self.values.clear();
    }

    /// Stack'in kapasitesini döndür
    pub fn capacity(&self) -> usize {
        self.values.capacity()
    }

    /// Stack'in doluluk oranını döndür (0.0 - 1.0)
    pub fn usage_ratio(&self) -> f64 {
        self.values.len() as f64 / self.capacity() as f64
    }

    /// Stack'in son N elemanını al
    pub fn peek_n(&self, n: usize) -> Result<&[Value], RuntimeError> {
        if n > self.values.len() {
            return Err(RuntimeError::StackUnderflow);
        }
        Ok(&self.values[self.values.len() - n..])
    }

    /// Stack'in son N elemanını çıkar
    pub fn pop_n(&mut self, n: usize) -> Result<Vec<Value>, RuntimeError> {
        if n > self.values.len() {
            return Err(RuntimeError::StackUnderflow);
        }
        let start = self.values.len() - n;
        Ok(self.values.split_off(start))
    }

    /// Stack'teki tüm değerleri döndür (debugging için)
    pub fn all_values(&self) -> &[Value] {
        &self.values
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_stack_basic_operations() {
        let mut stack = Stack::new();
        
        assert!(stack.is_empty());
        assert_eq!(stack.len(), 0);
        
        stack.push(Value::Int(42)).unwrap();
        assert!(!stack.is_empty());
        assert_eq!(stack.len(), 1);
        
        let value = stack.pop().unwrap();
        assert_eq!(value, Value::Int(42));
        assert!(stack.is_empty());
    }

    #[test]
    fn test_stack_peek() {
        let mut stack = Stack::new();
        
        stack.push(Value::Int(1)).unwrap();
        stack.push(Value::Int(2)).unwrap();
        stack.push(Value::Int(3)).unwrap();
        
        assert_eq!(*stack.peek(0).unwrap(), Value::Int(3));
        assert_eq!(*stack.peek(1).unwrap(), Value::Int(2));
        assert_eq!(*stack.peek(2).unwrap(), Value::Int(1));
        
        // Stack değişmedi
        assert_eq!(stack.len(), 3);
    }

    #[test]
    fn test_stack_peek_mut() {
        let mut stack = Stack::new();
        
        stack.push(Value::Int(42)).unwrap();
        
        let value = stack.peek_mut(0).unwrap();
        *value = Value::String("hello".to_string());
        
        let popped = stack.pop().unwrap();
        assert_eq!(popped, Value::String("hello".to_string()));
    }

    #[test]
    fn test_stack_overflow() {
        let mut stack = Stack::new();
        
        // Stack'i doldur
        for i in 0..STACK_MAX {
            stack.push(Value::Int(i as i64)).unwrap();
        }
        
        // Overflow
        assert!(stack.push(Value::Int(999)).is_err());
    }

    #[test]
    fn test_stack_underflow() {
        let mut stack = Stack::new();
        
        assert!(stack.pop().is_err());
        assert!(stack.peek(0).is_err());
    }

    #[test]
    fn test_stack_peek_n() {
        let mut stack = Stack::new();
        
        stack.push(Value::Int(1)).unwrap();
        stack.push(Value::Int(2)).unwrap();
        stack.push(Value::Int(3)).unwrap();
        
        let values = stack.peek_n(2).unwrap();
        assert_eq!(values.len(), 2);
        assert_eq!(values[0], Value::Int(2));
        assert_eq!(values[1], Value::Int(3));
    }

    #[test]
    fn test_stack_pop_n() {
        let mut stack = Stack::new();
        
        stack.push(Value::Int(1)).unwrap();
        stack.push(Value::Int(2)).unwrap();
        stack.push(Value::Int(3)).unwrap();
        
        let values = stack.pop_n(2).unwrap();
        assert_eq!(values.len(), 2);
        assert_eq!(values[0], Value::Int(2));
        assert_eq!(values[1], Value::Int(3));
        
        assert_eq!(stack.len(), 1);
        assert_eq!(*stack.peek(0).unwrap(), Value::Int(1));
    }

    #[test]
    fn test_stack_usage_ratio() {
        let mut stack = Stack::new();
        
        assert_eq!(stack.usage_ratio(), 0.0);
        
        stack.push(Value::Int(1)).unwrap();
        let ratio = stack.usage_ratio();
        assert!(ratio > 0.0);
        assert!(ratio < 1.0);
    }
}
