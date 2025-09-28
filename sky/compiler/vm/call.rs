// Call Frame - Çağrı Çerçevesi
// Fonksiyon çağrıları için frame yönetimi

use super::Value;
use crate::compiler::bytecode::Chunk;

/// Çağrı çerçevesi
#[derive(Debug, Clone, PartialEq)]
pub struct CallFrame {
    pub chunk: Chunk,
    pub ip: usize,           // Instruction pointer
    pub stack_start: usize,  // Stack'teki başlangıç pozisyonu
    pub locals: Vec<Value>,  // Yerel değişkenler
    pub function: Option<FunctionInfo>,
}

/// Fonksiyon bilgisi
#[derive(Debug, Clone, PartialEq)]
pub struct FunctionInfo {
    pub name: String,
    pub arity: u8,
    pub is_native: bool,
}

impl CallFrame {
    pub fn new(chunk: Chunk, stack_start: usize) -> Self {
        Self {
            chunk,
            ip: 0,
            stack_start,
            locals: Vec::new(),
            function: None,
        }
    }

    pub fn new_function(chunk: Chunk, stack_start: usize, name: String, arity: u8) -> Self {
        Self {
            chunk,
            ip: 0,
            stack_start,
            locals: Vec::new(),
            function: Some(FunctionInfo {
                name,
                arity,
                is_native: false,
            }),
        }
    }

    /// Parametreli fonksiyon için yeni frame oluştur
    pub fn new_function_with_params(chunk: Chunk, stack_start: usize, name: String, arity: u8, param_count: usize) -> Self {
        let mut locals = Vec::with_capacity(param_count);
        // Parametreler için null değerlerle başlat
        for _ in 0..param_count {
            locals.push(Value::Null);
        }
        
        Self {
            chunk,
            ip: 0,
            stack_start,
            locals,
            function: Some(FunctionInfo {
                name,
                arity,
                is_native: false,
            }),
        }
    }

    pub fn new_native(name: String, arity: u8) -> Self {
        Self {
            chunk: Chunk::new(),
            ip: 0,
            stack_start: 0,
            locals: Vec::new(),
            function: Some(FunctionInfo {
                name,
                arity,
                is_native: true,
            }),
        }
    }

    /// Mevcut instruction'ı al
    pub fn current_instruction(&self) -> Option<u8> {
        self.chunk.get_byte(self.ip)
    }

    /// IP'yi ilerlet
    pub fn advance_ip(&mut self) {
        self.ip += 1;
    }

    /// IP'yi belirli bir değere ayarla
    pub fn set_ip(&mut self, ip: usize) {
        self.ip = ip;
    }

    /// IP'yi relative olarak değiştir
    pub fn jump_ip(&mut self, offset: isize) {
        self.ip = (self.ip as isize + offset) as usize;
    }

    /// Fonksiyon olup olmadığını kontrol et
    pub fn is_function(&self) -> bool {
        self.function.is_some()
    }

    /// Native fonksiyon olup olmadığını kontrol et
    pub fn is_native(&self) -> bool {
        self.function.as_ref().map_or(false, |f| f.is_native)
    }

    /// Fonksiyon adını al
    pub fn function_name(&self) -> Option<&str> {
        self.function.as_ref().map(|f| f.name.as_str())
    }

    /// Fonksiyon arity'sini al
    pub fn function_arity(&self) -> Option<u8> {
        self.function.as_ref().map(|f| f.arity)
    }

    /// Yerel değişken ekle
    pub fn add_local(&mut self, value: Value) {
        self.locals.push(value);
    }

    /// Yerel değişken al
    pub fn get_local(&self, index: usize) -> Option<&Value> {
        self.locals.get(index)
    }

    /// Yerel değişken ayarla
    pub fn set_local(&mut self, index: usize, value: Value) -> bool {
        // Eğer index mevcut locals'tan büyükse, gerekli kadar null ile genişlet
        while self.locals.len() <= index {
            self.locals.push(Value::Null);
        }
        
        if let Some(local) = self.locals.get_mut(index) {
            *local = value;
            true
        } else {
            false
        }
    }

    /// Yerel değişken sayısını al
    pub fn local_count(&self) -> usize {
        self.locals.len()
    }

    /// Stack pozisyonunu al
    pub fn stack_position(&self, stack_len: usize) -> usize {
        stack_len - self.stack_start
    }

    /// Frame'in tamamlanıp tamamlanmadığını kontrol et
    pub fn is_finished(&self) -> bool {
        self.ip >= self.chunk.len()
    }

    /// Frame'in durumunu string olarak döndür (debugging için)
    pub fn status(&self) -> String {
        if let Some(func) = &self.function {
            format!("{}({}) at ip={}", func.name, func.arity, self.ip)
        } else {
            format!("<anonymous> at ip={}", self.ip)
        }
    }

    /// Frame state'i save et (coroutine için)
    pub fn save_state(&mut self) -> Result<(), super::RuntimeError> {
        // Stack snapshot, local variables, IP position save et
        // Gerçek implementasyonda:
        // - Stack state'i snapshot al
        // - Local variables'ları kaydet
        // - Instruction pointer'ı kaydet
        // - Frame context'i kaydet
        Ok(())
    }

    /// Frame state'i restore et (coroutine için)
    pub fn restore_state(&mut self) -> Result<(), super::RuntimeError> {
        // Stack snapshot, local variables, IP position restore et
        // Gerçek implementasyonda:
        // - Stack state'i restore et
        // - Local variables'ları restore et
        // - Instruction pointer'ı restore et
        // - Frame context'i restore et
        Ok(())
    }
}

impl FunctionInfo {
    pub fn new(name: String, arity: u8) -> Self {
        Self {
            name,
            arity,
            is_native: false,
        }
    }

    pub fn new_native(name: String, arity: u8) -> Self {
        Self {
            name,
            arity,
            is_native: true,
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::compiler::bytecode::Chunk;

    #[test]
    fn test_call_frame_creation() {
        let chunk = Chunk::new();
        let frame = CallFrame::new(chunk, 0);
        
        assert_eq!(frame.ip, 0);
        assert_eq!(frame.stack_start, 0);
        assert!(!frame.is_function());
        assert!(!frame.is_native());
    }

    #[test]
    fn test_function_frame_creation() {
        let chunk = Chunk::new();
        let frame = CallFrame::new_function(chunk, 10, "test".to_string(), 2);
        
        assert!(frame.is_function());
        assert!(!frame.is_native());
        assert_eq!(frame.function_name(), Some("test"));
        assert_eq!(frame.function_arity(), Some(2));
    }

    #[test]
    fn test_native_frame_creation() {
        let frame = CallFrame::new_native("native_test".to_string(), 1);
        
        assert!(frame.is_function());
        assert!(frame.is_native());
        assert_eq!(frame.function_name(), Some("native_test"));
        assert_eq!(frame.function_arity(), Some(1));
    }

    #[test]
    fn test_ip_operations() {
        let chunk = Chunk::new();
        let mut frame = CallFrame::new(chunk, 0);
        
        assert_eq!(frame.ip, 0);
        
        frame.advance_ip();
        assert_eq!(frame.ip, 1);
        
        frame.set_ip(10);
        assert_eq!(frame.ip, 10);
        
        frame.jump_ip(5);
        assert_eq!(frame.ip, 15);
        
        frame.jump_ip(-3);
        assert_eq!(frame.ip, 12);
    }

    #[test]
    fn test_local_variables() {
        let chunk = Chunk::new();
        let mut frame = CallFrame::new(chunk, 0);
        
        assert_eq!(frame.local_count(), 0);
        
        frame.add_local(Value::Int(42));
        frame.add_local(Value::String("hello".to_string()));
        
        assert_eq!(frame.local_count(), 2);
        assert_eq!(frame.get_local(0), Some(&Value::Int(42)));
        assert_eq!(frame.get_local(1), Some(&Value::String("hello".to_string())));
        assert_eq!(frame.get_local(2), None);
        
        frame.set_local(0, Value::Float(3.14));
        assert_eq!(frame.get_local(0), Some(&Value::Float(3.14)));
        
        assert!(!frame.set_local(5, Value::Bool(true)));
    }

    #[test]
    fn test_stack_position() {
        let chunk = Chunk::new();
        let frame = CallFrame::new(chunk, 10);
        
        assert_eq!(frame.stack_position(15), 5);
        assert_eq!(frame.stack_position(10), 0);
    }

    #[test]
    fn test_frame_status() {
        let chunk = Chunk::new();
        let frame = CallFrame::new_function(chunk, 0, "test_func".to_string(), 3);
        
        let status = frame.status();
        assert!(status.contains("test_func"));
        assert!(status.contains("3"));
        assert!(status.contains("ip=0"));
    }
}
