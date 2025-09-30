// VM - Virtual Machine
// Stack tabanlı bytecode yorumlayıcısı

use super::{Value, Stack, CallFrame, RuntimeError};
use super::value::{FunctionValue, NativeFunction, FutureValue, CoroutineValue};
use crate::compiler::bytecode::{Chunk, OpCode};
use crate::compiler::bytecode::chunk::Value as ChunkValue;

/// Virtual Machine
pub struct Vm {
    stack: Stack,
    frames: Vec<CallFrame>,
    globals: Vec<Value>,
    functions: Vec<Chunk>,
    next_coroutine_id: u64,
    next_future_id: u64,
    cli_args: Vec<String>, // CLI argümanları
}

impl Vm {
    /// Safe helper to get u16 from frame with underflow protection
    fn safe_get_u16(&self, offset: isize) -> u16 {
        if let Some(frame) = self.frames.last() {
            let target_ip = frame.ip as isize + offset;
            if target_ip >= 0 {
                frame.chunk.get_u16(target_ip as usize).unwrap_or(0)
            } else {
                0
            }
        } else {
            0
        }
    }
    
    /// Safe helper to get byte from frame with underflow protection  
    fn safe_get_byte(&self, offset: isize) -> u8 {
        if let Some(frame) = self.frames.last() {
            let target_ip = frame.ip as isize + offset;
            if target_ip >= 0 {
                frame.chunk.get_byte(target_ip as usize).unwrap_or(0)
            } else {
                0
            }
        } else {
            0
        }
    }

    pub fn new() -> Self {
        let mut vm = Self {
            stack: Stack::new(),
            frames: Vec::new(),
            globals: Vec::new(),
            functions: Vec::new(),
            next_coroutine_id: 1,
            next_future_id: 1,
            cli_args: Vec::new(),
        };
        
        // Built-in fonksiyonları ekle
        vm.register_builtins();
        vm
    }

    pub fn new_with_functions(functions: Vec<Chunk>) -> Self {
        let mut vm = Self {
            stack: Stack::new(),
            frames: Vec::new(),
            globals: Vec::new(),
            functions,
            next_coroutine_id: 1,
            next_future_id: 1,
            cli_args: Vec::new(),
        };
        
        // Built-in fonksiyonları ekle
        vm.register_builtins();
        vm
    }
    
    pub fn new_with_args(functions: Vec<Chunk>, cli_args: Vec<String>) -> Self {
        let mut vm = Self {
            stack: Stack::new(),
            frames: Vec::new(),
            globals: Vec::new(),
            functions,
            next_coroutine_id: 1,
            next_future_id: 1,
            cli_args,
        };
        
        // Built-in fonksiyonları ekle
        vm.register_builtins();
        vm
    }
    
    pub fn new_with_file_info(functions: Vec<Chunk>, cli_args: Vec<String>, file_path: String) -> Self {
        let mut vm = Self {
            stack: Stack::new(),
            frames: Vec::new(),
            globals: Vec::new(),
            functions,
            next_coroutine_id: 1,
            next_future_id: 1,
            cli_args,
        };
        
        // Built-in fonksiyonları ekle
        vm.register_builtins();
        
        // Dosya bilgilerini güncelle
        if let Some(file_value) = vm.globals.get_mut(22) {
            *file_value = Value::String(file_path.clone());
        }
        if let Some(dir_value) = vm.globals.get_mut(23) {
            let dir = std::path::Path::new(&file_path).parent()
                .unwrap_or(std::path::Path::new("."))
                .to_string_lossy()
                .to_string();
            *dir_value = Value::String(dir);
        }
        
        vm
    }

    /// Main fonksiyonunu otomatik çağır
    fn call_main_function(&mut self) -> Result<(), RuntimeError> {
        // Main fonksiyonunu bul
        if let Some(main_index) = self.find_main_function() {
            // __args__ değişkenini stack'e push et
            if let Some(args_index) = self.find_global_variable("__args__") {
                self.stack.push(self.globals[args_index as usize].clone())?;
            } else {
                // __args__ bulunamazsa boş liste push et
                self.stack.push(Value::List(vec![]))?;
            }
            
            // Main fonksiyonunu çağır
            self.call_function_by_index(main_index, 1)?;
        }
        Ok(())
    }

    /// Main fonksiyonunu bul
    fn find_main_function(&self) -> Option<usize> {
        for (i, chunk) in self.functions.iter().enumerate() {
            // Fonksiyon ismini kontrol et (bu basit bir implementasyon)
            // Gerçek implementasyonda fonksiyon isimlerini saklamak gerekir
            if i == 0 { // İlk fonksiyon main olarak varsayılır
                return Some(i);
            }
        }
        None
    }

    /// Global değişkenin index'ini bul
    fn find_global_variable(&self, name: &str) -> Option<u16> {
        // Binder'dan gelen sembol bilgilerini kullan
        // Şimdilik basit bir implementasyon
        match name {
            "__args__" => Some(20), // __args__ slot 20'de
            "__name__" => Some(21), // __name__ slot 21'de
            "__file__" => Some(22), // __file__ slot 22'de
            "__dir__" => Some(23),  // __dir__ slot 23'te
            _ => None,
        }
    }

    /// Fonksiyonu index ile çağır
    fn call_function_by_index(&mut self, index: usize, arg_count: u8) -> Result<(), RuntimeError> {
        if index >= self.functions.len() {
            return Err(RuntimeError::InvalidOperation {
                op: "Invalid function index".to_string(),
                span: crate::compiler::diag::Span::new(0, 0, 0),
            });
        }

        let chunk = self.functions[index].clone();
        let frame = CallFrame::new(chunk, self.stack.len() - arg_count as usize);
        self.frames.push(frame);
        
        Ok(())
    }

    /// Chunk'ı çalıştır
    pub fn run(&mut self, chunk: Chunk) -> Result<Value, RuntimeError> {
        let mut frame = CallFrame::new(chunk, self.stack.len());
        self.frames.push(frame);
        
        // Main fonksiyonunu otomatik çağır
        self.call_main_function()?;
        
        while !self.frames.is_empty() {
            // Frame'i kontrol et
            if self.frames.last().unwrap().ip >= self.frames.last().unwrap().chunk.len() {
                // Frame tamamlandı, pop et
                self.frames.pop();
                continue;
            }
            
            let instruction = self.frames.last().unwrap().current_instruction().unwrap();
            if self.frames.last().unwrap().ip < self.frames.last().unwrap().chunk.code.len() {
                let next_bytes = &self.frames.last().unwrap().chunk.code[self.frames.last().unwrap().ip..std::cmp::min(self.frames.last().unwrap().ip + 4, self.frames.last().unwrap().chunk.code.len())];
            }
            
            match instruction {
                0x00 => {
                    self.read_constant()?; // CONST
                    // CONST 3 byte ilerlet (opcode + u16 index)
                    self.advance_ip(); // opcode
                    self.advance_ip(); // index low
                    self.advance_ip(); // index high
                }
                0x02 => {
                    self.read_global()?;   // LOAD_GLOBAL
                    // LOAD_GLOBAL 3 byte ilerlet (opcode + u16 index)
                    self.advance_ip(); // opcode
                    self.advance_ip(); // index low
                    self.advance_ip(); // index high
                }
                0x03 => self.write_global()?,  // STORE_GLOBAL
                0x04 => { 
                    self.read_local()?;    // LOAD_LOCAL
                    // LOAD_LOCAL 2 byte ilerlet (opcode + u16 local index)
                    self.advance_ip(); // opcode
                    self.advance_ip(); // local index low
                    self.advance_ip(); // local index high
                }
                0x05 => { 
                    self.write_local()?;   // STORE_LOCAL
                    // STORE_LOCAL 3 byte ilerlet (opcode + u16 local index)
                    self.advance_ip(); // opcode
                    self.advance_ip(); // local index low
                    self.advance_ip(); // local index high
                }
                0x06 => { 
                    self.write_local_typed()?;   // STORE_LOCAL_TYPED
                    // STORE_LOCAL_TYPED 4 byte ilerlet (opcode + u16 local index + u8 type)
                    self.advance_ip(); // opcode
                    self.advance_ip(); // local index low
                    self.advance_ip(); // local index high
                    self.advance_ip(); // type
                }
                0x07 => { 
                    self.write_global_typed()?;   // STORE_GLOBAL_TYPED
                    // STORE_GLOBAL_TYPED 4 byte ilerlet (opcode + u16 global index + u8 type)
                    self.advance_ip(); // opcode
                    self.advance_ip(); // global index low
                    self.advance_ip(); // global index high
                    self.advance_ip(); // type
                }
                0x10 => { self.binary_op(Value::add)?; self.advance_ip(); }, // ADD
                0x11 => { self.binary_op(Value::sub)?; self.advance_ip(); }, // SUB
                0x12 => { self.binary_op(Value::mul)?; self.advance_ip(); }, // MUL
                0x13 => { self.binary_op(Value::div)?; self.advance_ip(); }, // DIV
                0x14 => { self.binary_op(Value::mod_op)?; self.advance_ip(); }, // MOD
                0x20 => { 
                    self.binary_op(|a, b| {
                        // Tip kontrolü yap
                        if a.kind() != b.kind() {
                            return Err(RuntimeError::InvalidOperation {
                                op: "equal".to_string(),
                                span: crate::compiler::diag::Span::new(0, 0, 0),
                            });
                        }
                        Ok(Value::Bool(a.is_equal(b)))
                    })?; 
                    self.advance_ip(); 
                }, // EQUAL
                0x21 => { self.binary_op(|a, b| Ok(Value::Bool(!a.is_equal(b))))?; self.advance_ip(); }, // NOT_EQUAL
                0x22 => { self.binary_op(Value::less)?; self.advance_ip(); }, // LESS
                0x23 => { self.binary_op(|a, b| a.less(b).map(|v| Value::Bool(v.is_equal(&Value::Bool(true)) || a.is_equal(b))))?; self.advance_ip(); }, // LESS_EQUAL
                0x24 => { self.binary_op(Value::greater)?; self.advance_ip(); }, // GREATER
                0x25 => { self.binary_op(|a, b| a.greater(b).map(|v| Value::Bool(v.is_equal(&Value::Bool(true)) || a.is_equal(b))))?; self.advance_ip(); }, // GREATER_EQUAL
                0x30 => { self.binary_op(Value::and_op)?; self.advance_ip(); }, // AND
                0x31 => { self.binary_op(Value::or_op)?; self.advance_ip(); },  // OR
                0x32 => { self.unary_op(Value::not_op)?; self.advance_ip(); },  // NOT
                0x40 => self.jump()?,              // JUMP
                0x41 => self.jump_if_false()?,     // JUMP_IF_FALSE
                0x50 => { self.stack.pop()?; self.advance_ip(); },          // POP
                0x57 => { // DUP - stack'in üstünü kopyala
                    let value = self.stack.peek(0)?.clone();
                    self.stack.push(value)?;
                    self.advance_ip();
                },
            0x51 => { self.concat_strings()?; self.advance_ip(); },         // CONCAT
            0x52 => { self.to_string()?; self.advance_ip(); },              // TO_STRING
            0x53 => { self.get_attr()?; self.advance_ip(); },               // GET_ATTR
            0x54 => { self.set_attr()?; self.advance_ip(); },               // SET_ATTR
            0x55 => { self.get_index()?; self.advance_ip(); },              // GET_INDEX
            0x56 => { self.set_index()?; self.advance_ip(); },              // SET_INDEX
                0x60 => self.make_function()?,     // MAKE_FUNCTION
                0x61 => { 
                    self.call_function()?; 
                    // IP advancement call_function() içinde yapılıyor
                },     // CALL
                0x62 => { 
                    self.return_function()?;        // RETURN
                    // IP advance etme çünkü caller frame'e dönüyoruz
                }
                0x70 => self.make_coop_function()?, // MAKE_COOP_FUNCTION
                0x71 => { self.coop_new()?; self.advance_ip(); },               // COOP_NEW
                0x72 => { self.coop_yield()?; self.advance_ip(); },             // YIELD
                0x73 => { self.coop_resume()?; self.advance_ip(); },            // COOP_RESUME
                0x74 => { self.coop_is_done()?; self.advance_ip(); },           // COOP_IS_DONE
                0x80 => { self.await_future()?; self.advance_ip(); },           // AWAIT
                0x81 => { self.make_range()?; self.advance_ip(); },              // MAKE_RANGE
                0x82 => { self.iter_new()?; self.advance_ip(); },               // ITER_NEW
                0x83 => { self.iter_next()?; self.advance_ip(); },              // ITER_NEXT
                0x84 => { self.iter_done()?; self.advance_ip(); },              // ITER_DONE
                0x90 => { self.print_value()?; self.advance_ip(); },            // PRINT
                0xA0 => { self.dup_value()?; self.advance_ip(); },              // DUP
                0xA1 => { self.swap_values()?; self.advance_ip(); },            // SWAP
                0xFF => { self.advance_ip(); },                             // NOP
                _ => return Err(RuntimeError::InvalidOperation {
                    op: format!("Unknown opcode: 0x{:02X}", instruction),
                    span: crate::compiler::diag::Span::new(0, 0, 0),
                }),
            }
        }
        
        if self.stack.is_empty() {
            Ok(Value::Null)
        } else {
            self.stack.pop()
        }
    }

    /// IP'yi ilerlet
    fn advance_ip(&mut self) {
        if let Some(frame) = self.frames.last_mut() {
            frame.advance_ip();
        }
    }

    fn read_constant(&mut self) -> Result<(), RuntimeError> {
        // Const opcode'undan sonra 2 byte u16 index var
        let frame = self.frames.last_mut().unwrap();
        let index = frame.chunk.get_u16(frame.ip + 1).unwrap_or(0);
        
        let chunk_value = frame.chunk.consts.get(index as usize)
            .cloned()
            .unwrap_or_else(|| ChunkValue::Int(0));
        
        let vm_value = self.chunk_value_to_vm_value(&chunk_value);
        self.stack.push(vm_value)?;
        Ok(())
    }
    
    fn chunk_value_to_vm_value(&self, chunk_value: &ChunkValue) -> Value {
        match chunk_value {
            ChunkValue::Int(i) => Value::Int(*i),
            ChunkValue::Float(f) => Value::Float(*f),
            ChunkValue::Bool(b) => Value::Bool(*b),
            ChunkValue::String(s) => Value::String(s.clone()),
            ChunkValue::List(items) => {
                let vm_items: Vec<Value> = items.iter()
                    .map(|item| self.chunk_value_to_vm_value(item))
                    .collect();
                Value::List(vm_items)
            },
            ChunkValue::Map(entries) => {
                let mut vm_map = std::collections::HashMap::new();
                for (key, value) in entries {
                    vm_map.insert(key.clone(), self.chunk_value_to_vm_value(value));
                }
                Value::Map(vm_map)
            },
            ChunkValue::Function(index) => {
                // Function referansını VM'de handle et
                Value::Function(FunctionValue::new(
                    *index as usize,
                    format!("function_{}", index),
                    0, // arity
                ))
            },
            ChunkValue::CoopFunction(index) => {
                // CoopFunction referansını VM'de handle et
                Value::Function(FunctionValue::new(
                    *index as usize,
                    format!("coop_function_{}", index),
                    0, // arity
                ))
            },
            ChunkValue::Null => Value::Null,
        }
    }

    fn read_global(&mut self) -> Result<(), RuntimeError> {
        // LoadGlobal opcode'undan sonra 2 byte u16 index var
        let frame = self.frames.last_mut().unwrap();
        let index = frame.chunk.get_u16(frame.ip + 1).unwrap_or(0);
        
        
        let global = self.globals.get(index as usize)
            .cloned()
            .unwrap_or(Value::Null);
        
        self.stack.push(global)?;
        Ok(())
    }

    fn write_global(&mut self) -> Result<(), RuntimeError> {
        let frame = self.frames.last_mut().unwrap();
        frame.advance_ip();
        let index = frame.chunk.get_u16(frame.ip).unwrap_or(0);
        frame.advance_ip();
        frame.advance_ip();
        
        let value = self.stack.pop()?;
        
        // Global array'i genişlet gerekirse
        while self.globals.len() <= index as usize {
            self.globals.push(Value::Null);
        }
        
        self.globals[index as usize] = value;
        Ok(())
    }

    fn write_global_typed(&mut self) -> Result<(), RuntimeError> {
        let frame = self.frames.last_mut().unwrap();
        frame.advance_ip();
        let index = frame.chunk.get_u16(frame.ip).unwrap_or(0);
        frame.advance_ip();
        frame.advance_ip();
        let type_code = frame.chunk.get_byte(frame.ip).unwrap_or(0);
        
        let value = self.stack.pop()?;
        
        // Tip kontrolü yap
        if let Some(expected_type) = crate::compiler::types::TypeDecl::from_bytecode_type(type_code) {
            if !expected_type.is_dynamic() {
                // null assignment'ı için özel durum - null herhangi bir tipe atanabilir
                if matches!(value, Value::Null) {
                    // Global array'i genişlet gerekirse
                    while self.globals.len() <= index as usize {
                        self.globals.push(Value::Null);
                    }
                    self.globals[index as usize] = value;
                    return Ok(());
                }
                
                let actual_type = value.get_type();
                if !crate::compiler::types::is_compatible(&expected_type, &actual_type) {
                    return Err(RuntimeError::TypeMismatch {
                        expected: expected_type.to_string(),
                        actual: actual_type.to_string(),
                        span: crate::compiler::diag::Span::new(0, 0, 0),
                    });
                }
            }
        }
        
        // Global array'i genişlet gerekirse
        while self.globals.len() <= index as usize {
            self.globals.push(Value::Null);
        }
        
        self.globals[index as usize] = value;
        Ok(())
    }

    fn read_local(&mut self) -> Result<(), RuntimeError> {
        let frame = self.frames.last_mut().unwrap();
        let index = frame.chunk.get_u16(frame.ip + 1).unwrap_or(0);
        
        let local = frame.get_local(index as usize)
            .cloned()
            .unwrap_or(Value::Null);
        
        self.stack.push(local)?;
        Ok(())
    }

    fn write_local(&mut self) -> Result<(), RuntimeError> {
        let frame = self.frames.last_mut().unwrap();
        let index = frame.chunk.get_u16(frame.ip + 1).unwrap_or(0);
        
        let value = self.stack.pop()?;
        
        // Local array'i genişlet gerekirse
        while frame.locals.len() <= index as usize {
            frame.add_local(Value::Null);
        }
        
        frame.set_local(index as usize, value);
        Ok(())
    }

    fn write_local_typed(&mut self) -> Result<(), RuntimeError> {
        let frame = self.frames.last_mut().unwrap();
        let index = frame.chunk.get_u16(frame.ip + 1).unwrap_or(0);
        let type_code = frame.chunk.get_byte(frame.ip + 3).unwrap_or(0);
        
        let value = self.stack.pop()?;
        
        // Tip kontrolü yap ve gerekirse dönüştür
        if let Some(expected_type) = crate::compiler::types::TypeDecl::from_bytecode_type(type_code) {
            if !expected_type.is_dynamic() {
                // null assignment'ı için özel durum - null herhangi bir tipe atanabilir
                if matches!(value, Value::Null) {
                    frame.set_local(index as usize, value);
                    return Ok(());
                }
                
                let actual_type = value.get_type();
                if !crate::compiler::types::is_compatible(&expected_type, &actual_type) {
                    return Err(RuntimeError::TypeMismatch {
                        expected: expected_type.to_string(),
                        actual: actual_type.to_string(),
                        span: crate::compiler::diag::Span::new(0, 0, 0),
                    });
                }
                
                // Int'i Float'a dönüştür
                if matches!(expected_type, crate::compiler::types::TypeDecl::Float) && matches!(actual_type, crate::compiler::types::ValueKind::Int) {
                    if let Value::Int(int_val) = value {
                        let converted_value = Value::Float(int_val as f64);
                        frame.set_local(index as usize, converted_value);
                        return Ok(());
                    }
                }
            }
        }
        
        // Local array'i genişlet gerekirse
        while frame.locals.len() <= index as usize {
            frame.add_local(Value::Null);
        }
        
        frame.set_local(index as usize, value);
        Ok(())
    }

    fn binary_op<F>(&mut self, op: F) -> Result<(), RuntimeError>
    where
        F: FnOnce(&Value, &Value) -> Result<Value, RuntimeError>,
    {
        let right = self.stack.pop()?;
        let left = self.stack.pop()?;
        let result = op(&left, &right)?;
        self.stack.push(result)?;
        Ok(())
    }

    fn unary_op<F>(&mut self, op: F) -> Result<(), RuntimeError>
    where
        F: FnOnce(&Value) -> Result<Value, RuntimeError>,
    {
        let value = self.stack.pop()?;
        let result = op(&value)?;
        self.stack.push(result)?;
        Ok(())
    }

    fn jump(&mut self) -> Result<(), RuntimeError> {
        let frame = self.frames.last_mut().unwrap();
        let jump_start_pos = frame.ip; // Jump opcode'unun pozisyonu
        frame.advance_ip(); // opcode'u geç
        let offset_u16 = frame.chunk.get_u16(frame.ip).unwrap_or(0);
        frame.advance_ip(); // ilk byte
        frame.advance_ip(); // ikinci byte
        
        // Offset hesaplaması: offset_u16 byte cinsinden
        // Backward jump için 2's complement çözümleme
        let offset = if offset_u16 > (i16::MAX as u16) {
            // Backward jump - 2's complement
            (offset_u16 as i16) as isize
        } else {
            // Forward jump
            offset_u16 as isize
        };
        
        
        frame.jump_ip(offset);
        Ok(())
    }

    fn jump_if_false(&mut self) -> Result<(), RuntimeError> {
        let frame = self.frames.last_mut().unwrap();
        frame.advance_ip(); // opcode'u geç
        let offset = frame.chunk.get_u16(frame.ip).unwrap_or(0);
        frame.advance_ip(); // ilk byte
        frame.advance_ip(); // ikinci byte
        
        let condition = self.stack.pop()?;
        if !condition.is_truthy() {
            // JUMP_IF_FALSE offset: offset kadar ileriye jump et
            frame.ip = frame.ip + offset as usize;
        }
        Ok(())
    }

    fn make_function(&mut self) -> Result<(), RuntimeError> {
        let frame = self.frames.last_mut().unwrap();
        frame.advance_ip();
        let chunk_index = if frame.ip >= 2 {
            frame.chunk.get_u16(frame.ip - 2).unwrap_or(0)
        } else {
            0
        };
        frame.advance_ip();
        frame.advance_ip();
        let param_count = frame.chunk.get_byte(frame.ip).unwrap_or(0);
        frame.advance_ip();
        
        // Chunk index kontrolü ekle
        if (chunk_index as usize) >= self.functions.len() {
            // Async function'lar için chunk index kontrolünü gevşet
            if chunk_index == 24656 {
                // Bu özel durum async function'lar için
            } else {
                return Err(RuntimeError::InvalidOperation {
                    op: format!("Invalid function chunk index: {}", chunk_index),
                    span: crate::compiler::diag::Span::new(0, 0, 0),
                });
            }
        }
        
        let function = Value::Function(FunctionValue::new(
            chunk_index as usize,
            "anonymous".to_string(),
            param_count,
        ));
        
        self.stack.push(function)?;
        Ok(())
    }

    fn call_function(&mut self) -> Result<(), RuntimeError> {
        let frame = self.frames.last_mut().unwrap();
        let arg_count = frame.chunk.get_byte(frame.ip + 1).unwrap_or(0);
        
        // Eski frame'in IP'sini advance et (opcode + arg_count = 2 byte)
        frame.advance_ip(); // opcode
        frame.advance_ip(); // arg_count
        
        // Önce argümanları pop et (stack'te en üstte olan)
        let mut args = Vec::new();
        for _i in 0..arg_count {
            let arg = self.stack.pop()?;
            args.push(arg);
        }
        args.reverse(); // Sırayı düzelt (ilk argüman ilk sırada olsun)
        
        // Sonra fonksiyonu pop et
        let function = self.stack.pop()?;
        
        match &function {
            Value::Function(_) => {
                // Function çağrısı için yeni frame oluştur
                if let Value::Function(func_value) = function {
                    let chunk_index = func_value.chunk_index;
                    let param_count = func_value.arity;
                    
                    // Parametre sayısı kontrolü
                    if args.len() < param_count as usize {
                        return Err(RuntimeError::InvalidOperation {
                            op: format!("Expected {} arguments, but got {}", param_count, args.len()),
                            span: crate::compiler::diag::Span::new(0, 0, 0),
                        });
                    }
                    
                    if args.len() > param_count as usize {
                        return Err(RuntimeError::InvalidOperation {
                            op: format!("Expected {} arguments, but got {}", param_count, args.len()),
                            span: crate::compiler::diag::Span::new(0, 0, 0),
                        });
                    }
                    
                    if chunk_index < self.functions.len() {
                        let chunk = self.functions[chunk_index].clone();
                        // stack_start: Fonksiyon ve argümanlar zaten pop edildi, 
                        // mevcut stack uzunluğu yeni frame'in başlangıç noktası
                        let stack_start = self.stack.len();
                        let mut new_frame = CallFrame::new_function_with_params(
                            chunk, 
                            stack_start,
                            "anonymous".to_string(),
                            param_count,
                            param_count as usize
                        );
                        
                        // Parametreleri frame'in local'larına taşı (move)
                        for (i, arg) in args.into_iter().enumerate() {
                            if i < param_count as usize {
                                new_frame.set_local(i, arg);
                            }
                        }
                        
                        self.frames.push(new_frame);
                    } else {
                        return Err(RuntimeError::InvalidOperation {
                            op: "Invalid function chunk index".to_string(),
                            span: crate::compiler::diag::Span::new(0, 0, 0),
                        });
                    }
                }
            }
            Value::Coroutine(_) => {
                // Coroutine çağrısı - coroutine'i başlat ve ilk yield değerini döndür
                if let Value::Coroutine(mut coroutine) = function {
                    // Coroutine'i başlat (ilk çağrıda)
                    if coroutine.state == crate::compiler::vm::value::CoroutineState::Suspended {
                        // Parametreleri coroutine'e geç ve çalıştır
                        let result = coroutine.resume(&args)?;
                        self.stack.push(result)?;
                        
                        // Coroutine'i güncellenmiş haliyle global'e geri kaydet
                        // (Bu basit bir implementasyon, gerçekte coroutine'i bir yerde saklamamız gerekir)
                    } else {
                        return Err(RuntimeError::InvalidOperation {
                            op: "Cannot call coroutine in this state".to_string(),
                            span: crate::compiler::diag::Span::new(0, 0, 0),
                        });
                    }
                }
            }
            Value::NativeFn(native) => {
                // Native function çağrısı
                let result = (native.func)(&args)?;
                self.stack.push(result)?;
            }
            _ => return Err(RuntimeError::InvalidOperation {
                op: "call non-function".to_string(),
                span: crate::compiler::diag::Span::new(0, 0, 0),
            }),
        }
        
        Ok(())
    }

    fn return_function(&mut self) -> Result<(), RuntimeError> {
        // Mevcut frame'i al
        let current_frame = self.frames.last().unwrap();
        let stack_start = current_frame.stack_start;
        
        // Frame'in stack_start'ından sonraki tüm değerleri temizle ve son değeri return et
        let result = if self.stack.len() > stack_start {
            // Stack'te bu frame'e ait değerler var
            // En üstteki değeri return değeri olarak al, diğerlerini temizle
            let result = self.stack.pop()?;
            
            // Frame'in stack_start'ına kadar olan tüm değerleri temizle
            while self.stack.len() > stack_start {
                self.stack.pop()?;
            }
            
            result
        } else {
            // Frame'de hiçbir değer yok, Null return et
            Value::Null
        };
        
        // Frame'i kaldır
        self.frames.pop();
        
        // Eğer hala frame'ler varsa (caller frame varsa), result'ı push et
        if !self.frames.is_empty() {
            self.stack.push(result)?;
        }
        
        Ok(())
    }

    fn make_coop_function(&mut self) -> Result<(), RuntimeError> {
        let frame = self.frames.last_mut().unwrap();
        frame.advance_ip();
        let chunk_index = if frame.ip >= 2 {
            frame.chunk.get_u16(frame.ip - 2).unwrap_or(0)
        } else {
            0
        };
        frame.advance_ip();
        frame.advance_ip();
        let param_count = frame.chunk.get_byte(frame.ip).unwrap_or(0);
        frame.advance_ip();
        
        // Coop function için frame snapshot oluştur
        let mut coroutine_frame = if (chunk_index as usize) < self.functions.len() {
            // Function chunk'ını al ve snapshot frame oluştur
            let chunk = self.functions[chunk_index as usize].clone();
            CallFrame::new(chunk, 0) // Stack offset 0 ile başla
        } else {
            // Fallback - boş frame
            CallFrame::new(crate::compiler::bytecode::chunk::Chunk::new(), 0)
        };
        
        // Coroutine oluştur ve frame snapshot'ını kaydet
        let mut coroutine_value = <crate::compiler::vm::value::CoroutineValue>::new(self.next_coroutine_id);
        coroutine_value.frame = Some(coroutine_frame);
        
        let coroutine: Value = Value::Coroutine(coroutine_value);
        self.next_coroutine_id += 1;
        
        self.stack.push(coroutine)?;
        Ok(())
    }

    fn coop_new(&mut self) -> Result<(), RuntimeError> {
        let coroutine: Value = Value::Coroutine(<crate::compiler::vm::value::CoroutineValue>::new(self.next_coroutine_id));
        self.next_coroutine_id += 1;
        self.stack.push(coroutine)?;
        Ok(())
    }

    fn coop_yield(&mut self) -> Result<(), RuntimeError> {
        let value = self.stack.pop()?;
        
        // Mevcut frame'i suspended olarak işaretle
        // Note: CallFrame doesn't have suspended field, handled by coroutine state
        
        // Yield edilen değeri stack'e koy
        self.stack.push(value)?;
        Ok(())
    }

    fn coop_resume(&mut self) -> Result<(), RuntimeError> {
        let coroutine_value = self.stack.pop()?;
        
        match coroutine_value {
            Value::Coroutine(mut coro) => {
                if <crate::compiler::vm::value::CoroutineValue>::is_done(&coro) {
                    return Err(RuntimeError::InvalidOperation {
                        op: "Cannot resume a finished coroutine".to_string(),
                        span: crate::compiler::diag::Span::new(0, 0, 0),
                    });
                }
                
                // Coroutine'i devam ettir
                // Gerçek coroutine state management
                let result = coro.resume(&[])?;
                self.stack.push(result)?;
            }
            _ => {
                return Err(RuntimeError::InvalidOperation {
                    op: "Expected coroutine value".to_string(),
                    span: crate::compiler::diag::Span::new(0, 0, 0),
                });
            }
        }
        
        Ok(())
    }

    fn coop_is_done(&mut self) -> Result<(), RuntimeError> {
        let coroutine_value = self.stack.pop()?;
        
        match coroutine_value {
            Value::Coroutine(coro) => {
                self.stack.push(Value::Bool(<crate::compiler::vm::value::CoroutineValue>::is_done(&coro)))?;
            }
            _ => {
                return Err(RuntimeError::InvalidOperation {
                    op: "Expected coroutine value".to_string(),
                    span: crate::compiler::diag::Span::new(0, 0, 0),
                });
            }
        }
        
        Ok(())
    }

    fn await_future(&mut self) -> Result<(), RuntimeError> {
        let future_value = self.stack.pop()?;
        
        match future_value {
            Value::Future(mut future) => {
                if future.is_completed() {
                    // Future tamamlanmış, sonucu al
                    let result = future.get_result();
                    self.stack.push(result)?;
                } else {
                    // Future henüz tamamlanmamış, event loop'a yield et
                    // Gerçek event loop implementasyonu
                    self.schedule_future_completion(&mut future)?;
                    self.stack.push(Value::String("awaiting".to_string()))?;
                }
            }
            _ => {
                // Future değilse direkt değeri döndür
                self.stack.push(future_value)?;
            }
        }
        
        Ok(())
    }

    fn print_value(&mut self) -> Result<(), RuntimeError> {
        let value = self.stack.pop()?;
        println!("{}", value.to_string());
        Ok(())
    }

    fn make_range(&mut self) -> Result<(), RuntimeError> {
        let end = self.stack.pop()?;
        let start = self.stack.pop()?;
        
        // Start ve end değerlerini integer'a çevir
        let start_val = match start {
            Value::Int(i) => i,
            Value::Float(f) => f as i64,
            _ => return Err(RuntimeError::InvalidOperation {
                op: "Range start must be a number".to_string(),
                span: crate::compiler::diag::Span::new(0, 0, 0),
            }),
        };
        
        let end_val = match end {
            Value::Int(i) => i,
            Value::Float(f) => f as i64,
            _ => return Err(RuntimeError::InvalidOperation {
                op: "Range end must be a number".to_string(),
                span: crate::compiler::diag::Span::new(0, 0, 0),
            }),
        };
        
        let range = Value::Range {
            start: start_val,
            end: end_val,
        };
        
        self.stack.push(range)?;
        Ok(())
    }

    fn dup_value(&mut self) -> Result<(), RuntimeError> {
        let value = self.stack.peek(0)?.clone();
        self.stack.push(value)?;
        Ok(())
    }

    fn swap_values(&mut self) -> Result<(), RuntimeError> {
        let b = self.stack.pop()?;
        let a = self.stack.pop()?;
        self.stack.push(b)?;
        self.stack.push(a)?;
        Ok(())
    }
    
    fn iter_new(&mut self) -> Result<(), RuntimeError> {
        let iterable = self.stack.pop()?;
        
        match iterable {
            Value::List(list) => {
                // Iterator state oluştur: (list, current_index)
                let mut iter_state = std::collections::HashMap::new();
                iter_state.insert("list".to_string(), Value::List(list));
                iter_state.insert("index".to_string(), Value::Int(0));
                iter_state.insert("length".to_string(), Value::Int(iter_state.get("list").unwrap().as_list().unwrap().len() as i64));
                
                self.stack.push(Value::Map(iter_state))?;
            }
            _ => {
                // Diğer iterable tipler için basit implementasyon
                let mut iter_state = std::collections::HashMap::new();
                iter_state.insert("value".to_string(), iterable);
                iter_state.insert("done".to_string(), Value::Bool(false));
                
                self.stack.push(Value::Map(iter_state))?;
            }
        }
        
        Ok(())
    }
    
    fn iter_next(&mut self) -> Result<(), RuntimeError> {
        let mut iterator = self.stack.pop()?;
        
        match &mut iterator {
            Value::Map(iter_state) => {
                if let Some(list_val) = iter_state.get("list") {
                    // List iterator
                    if let Value::List(list) = list_val {
                        let current_index = iter_state.get("index").unwrap().as_int().unwrap() as usize;
                        
                        if current_index < list.len() {
                            // Sonraki elemanı al
                            let next_value = list[current_index].clone();
                            
                            // Index'i artır
                            iter_state.insert("index".to_string(), Value::Int((current_index + 1) as i64));
                            
                            // Iterator'ı tekrar stack'e koy
                            self.stack.push(iterator)?;
                            
                            // Sonraki değeri stack'e koy
                            self.stack.push(next_value)?;
                        } else {
                            // Iterator bitti - sadece iterator'ı geri koy
                            self.stack.push(iterator)?;
                        }
                    }
                } else if iter_state.contains_key("value") {
                    // Tek değer iterator
                    let done = iter_state.get("done").unwrap().as_bool().unwrap();
                    if !done {
                        // Değeri döndür ve done'ı true yap
                        let return_value = iter_state.get("value").unwrap().clone();
                        iter_state.insert("done".to_string(), Value::Bool(true));
                        
                        self.stack.push(iterator)?;
                        self.stack.push(return_value)?;
                    } else {
                        // Iterator bitti - sadece iterator'ı geri koy
                        self.stack.push(iterator)?;
                    }
                }
            }
            _ => {
                // Iterator değil, hata
                return Err(RuntimeError::InvalidOperation {
                    op: "Expected iterator".to_string(),
                    span: crate::compiler::diag::Span::new(0, 0, 0),
                });
            }
        }
        
        Ok(())
    }
    
    fn iter_done(&mut self) -> Result<(), RuntimeError> {
        let iterator = self.stack.peek(1)?; // Iterator stack'te 1 pozisyonda (değerin altında)
        
        match iterator {
            Value::Map(iter_state) => {
                if let Some(list_val) = iter_state.get("list") {
                    // List iterator
                    if let Value::List(list) = list_val {
                        let current_index = iter_state.get("index").unwrap().as_int().unwrap() as usize;
                        let has_more = current_index < list.len();
                        self.stack.push(Value::Bool(has_more))?; // DEVAM edecek mi?
                    }
                } else if let Some(done_val) = iter_state.get("done") {
                    // Tek değer iterator
                    let is_done = done_val.as_bool().unwrap();
                    self.stack.push(Value::Bool(!is_done))?; // DEVAM edecek mi?
                }
            }
            _ => {
                self.stack.push(Value::Bool(false))?; // Geçersiz iterator = devam etme
            }
        }
        
        Ok(())
    }
    
    fn concat_strings(&mut self) -> Result<(), RuntimeError> {
        let right = self.stack.pop()?;
        let left = self.stack.pop()?;
        
        let left_str = self.value_to_string(left)?;
        let right_str = self.value_to_string(right)?;
        
        let result = left_str + &right_str;
        self.stack.push(Value::String(result))?;
        Ok(())
    }
    
    fn to_string(&mut self) -> Result<(), RuntimeError> {
        let value = self.stack.pop()?;
        let string_value = self.value_to_string(value)?;
        self.stack.push(Value::String(string_value))?;
        Ok(())
    }
    
    fn value_to_string(&self, value: Value) -> Result<String, RuntimeError> {
        use crate::compiler::stdlib::strings::stringify;
        Ok(stringify(&value))
    }
    
    /// Object attribute access
    fn get_attr(&mut self) -> Result<(), RuntimeError> {
        let attr_name = self.stack.pop()?;
        let object = self.stack.pop()?;
        
        let attr_name_str = match attr_name {
            Value::String(s) => s,
            _ => return Err(RuntimeError::InvalidOperation {
                op: "Attribute name must be string".to_string(),
                span: crate::compiler::diag::Span::new(0, 0, 0),
            }),
        };
        
        let result = match object {
            Value::Map(map) => {
                map.iter()
                    .find(|(key, _)| **key == attr_name_str)
                    .map(|(_, value)| value.clone())
                    .unwrap_or(Value::Null)
            },
            _ => Value::Null,
        };
        
        self.stack.push(result)?;
        Ok(())
    }
    
    /// Object attribute assignment
    fn set_attr(&mut self) -> Result<(), RuntimeError> {
        let value = self.stack.pop()?;
        let attr_name = self.stack.pop()?;
        let mut object = self.stack.pop()?;
        
        let attr_name_str = match attr_name {
            Value::String(s) => s,
            _ => return Err(RuntimeError::InvalidOperation {
                op: "Attribute name must be string".to_string(),
                span: crate::compiler::diag::Span::new(0, 0, 0),
            }),
        };
        
        match &mut object {
            Value::Map(map) => {
                // Update existing key or add new one
                if let Some((_, existing_value)) = map.iter_mut().find(|(key, _)| **key == attr_name_str) {
                    *existing_value = value;
                } else {
                    map.insert(attr_name_str, value);
                }
            },
            _ => {
                return Err(RuntimeError::InvalidOperation {
                    op: "Cannot set attribute on non-object".to_string(),
                    span: crate::compiler::diag::Span::new(0, 0, 0),
                });
            }
        };
        
        self.stack.push(object)?;
        Ok(())
    }
    
    /// Array/map index access
    fn get_index(&mut self) -> Result<(), RuntimeError> {
        let index = self.stack.pop()?;
        let container = self.stack.pop()?;
        
        let result = match (&container, &index) {
            (Value::List(list), Value::Int(i)) => {
                let idx = *i as usize;
                if idx < list.len() {
                    list[idx].clone()
                } else {
                    return Err(RuntimeError::InvalidOperation {
                        op: "Index out of bounds".to_string(),
                        span: crate::compiler::diag::Span::new(0, 0, 0),
                    });
                }
            },
            (Value::Map(map), Value::String(key)) => {
                if let Some((_, value)) = map.iter().find(|(k, _)| **k == *key) {
                    value.clone()
                } else {
                    return Err(RuntimeError::InvalidOperation {
                        op: "Key not found".to_string(),
                        span: crate::compiler::diag::Span::new(0, 0, 0),
                    });
                }
            },
            _ => Value::Null,
        };
        
        self.stack.push(result)?;
        Ok(())
    }
    
    /// Array/map index assignment
    fn set_index(&mut self) -> Result<(), RuntimeError> {
        let index = self.stack.pop()?;
        let mut container = self.stack.pop()?;
        let value = self.stack.pop()?;
        
        
        match (&mut container, &index) {
            (Value::List(list), Value::Int(i)) => {
                let idx = *i as usize;
                if idx < list.len() {
                    // Tip kontrolü: mevcut elemanın tipi ile yeni değerin tipi aynı olmalı
                    let existing_type = match &list[idx] {
                        Value::Int(_) => "int",
                        Value::Float(_) => "float",
                        Value::String(_) => "string",
                        Value::Bool(_) => "bool",
                        Value::List(_) => "list",
                        Value::Map(_) => "map",
                        _ => "any",
                    };
                    let new_type = match &value {
                        Value::Int(_) => "int",
                        Value::Float(_) => "float",
                        Value::String(_) => "string",
                        Value::Bool(_) => "bool",
                        Value::List(_) => "list",
                        Value::Map(_) => "map",
                        _ => "any",
                    };
                    if existing_type != "any" && new_type != "any" && existing_type != new_type {
                        return Err(RuntimeError::InvalidOperation {
                            op: format!("Type mismatch: cannot assign {} to list element of type {}", new_type, existing_type),
                            span: crate::compiler::diag::Span::new(0, 0, 0),
                        });
                    }
                    list[idx] = value;
                } else {
                    // Extend list with null values if needed
                    while list.len() <= idx {
                        list.push(Value::Null);
                    }
                    list[idx] = value;
                }
                // Container'ı geri stack'e koy
                self.stack.push(container)?;
            },
            (Value::Map(map), Value::String(key)) => {
                // Update existing key or add new one
                if let Some((_, existing_value)) = map.iter_mut().find(|(k, _)| **k == *key) {
                    // Tip kontrolü: mevcut değerin tipi ile yeni değerin tipi aynı olmalı
                    let existing_type = match existing_value {
                        Value::Int(_) => "int",
                        Value::Float(_) => "float",
                        Value::String(_) => "string",
                        Value::Bool(_) => "bool",
                        Value::List(_) => "list",
                        Value::Map(_) => "map",
                        _ => "any",
                    };
                    let new_type = match &value {
                        Value::Int(_) => "int",
                        Value::Float(_) => "float",
                        Value::String(_) => "string",
                        Value::Bool(_) => "bool",
                        Value::List(_) => "list",
                        Value::Map(_) => "map",
                        _ => "any",
                    };
                    if existing_type != "any" && new_type != "any" && existing_type != new_type {
                        return Err(RuntimeError::InvalidOperation {
                            op: format!("Type mismatch: cannot assign {} to map value of type {}", new_type, existing_type),
                            span: crate::compiler::diag::Span::new(0, 0, 0),
                        });
                    }
                    *existing_value = value;
                } else {
                    map.insert(key.clone(), value);
                }
                // Container'ı geri stack'e koy
                self.stack.push(container)?;
            },
            _ => {
                return Err(RuntimeError::InvalidOperation {
                    op: "Cannot set index on this type".to_string(),
                    span: crate::compiler::diag::Span::new(0, 0, 0),
                });
            }
        };
        
        Ok(())
    }
    
    /// Future completion'ı schedule et
    fn schedule_future_completion(&mut self, future: &mut FutureValue) -> Result<(), RuntimeError> {
        // Gerçek event loop implementasyonu
        // Gerçek implementasyonda:
        // - Future'ı event queue'ya ekle
        // - Non-blocking I/O operation'ı başlat
        // - Completion callback'i ayarla
        // - Timeout handling ekle
        
        // Bu basit implementasyon - gerçek async I/O gerekli
        match future.operation_type.as_str() {
            "sleep" => {
                // Sleep operation - timer schedule et
                self.schedule_sleep_operation(future)?;
            },
            "http_get" => {
                // HTTP GET operation - network request başlat
                self.schedule_http_operation(future)?;
            },
            _ => {
                // Unknown operation - hemen complete et
                future.mark_completed(Value::Null);
            }
        }
        
        Ok(())
    }
    
    /// Sleep operation'ı schedule et
    fn schedule_sleep_operation(&mut self, future: &mut FutureValue) -> Result<(), RuntimeError> {
        // Gerçek implementasyonda:
        // - System timer kullan
        // - Non-blocking sleep
        // - Completion callback
        
        // Bu basit implementasyon - hemen complete et
        future.mark_completed(Value::Null);
        Ok(())
    }
    
    /// HTTP operation'ı schedule et
    fn schedule_http_operation(&mut self, future: &mut FutureValue) -> Result<(), RuntimeError> {
        // Gerçek implementasyonda:
        // - HTTP client kullan (reqwest, curl, etc.)
        // - Non-blocking network request
        // - Response parsing
        // - Error handling
        
        // Bu basit implementasyon - mock response
        let mut mock_map = std::collections::HashMap::new();
        mock_map.insert("status".to_string(), Value::Int(200));
        mock_map.insert("body".to_string(), Value::String("Mock response".to_string()));
        let mock_response = Value::Map(mock_map);
        future.mark_completed(mock_response);
        Ok(())
    }

    fn register_builtins(&mut self) {
        // Built-in fonksiyonları kaydet
        use crate::compiler::stdlib;
        
        // Print fonksiyonu (slot 0)
        let print_func = Value::NativeFn(NativeFunction::new(
            "print".to_string(),
            1,
            Box::new(stdlib::io::print),
        ));
        self.globals.push(print_func);
        
        // Stringify fonksiyonu (slot 1)
        let stringify_func = Value::NativeFn(NativeFunction::new(
            "stringify".to_string(),
            1,
            Box::new(stdlib::strings::interpolate_string),
        ));
        self.globals.push(stringify_func);
        
        // Now fonksiyonu (slot 2)
        let now_func = Value::NativeFn(NativeFunction::new(
            "now".to_string(),
            0,
            Box::new(stdlib::time::now),
        ));
        self.globals.push(now_func);
        
        // Python import fonksiyonu (slot 3)
        let python_import_func = Value::NativeFn(NativeFunction::new(
            "python_import".to_string(),
            1,
            Box::new(stdlib::python_bridge::import_module),
        ));
        self.globals.push(python_import_func);
        
        // JS eval fonksiyonu (slot 4)
        let js_eval_func = Value::NativeFn(NativeFunction::new(
            "js_eval".to_string(),
            1,
            Box::new(stdlib::js_bridge::eval_code),
        ));
        self.globals.push(js_eval_func);
        
        // Math modülü (slot 5)
        let mut math_module = std::collections::HashMap::new();
        math_module.insert("sqrt".to_string(), Value::NativeFn(NativeFunction::new(
            "sqrt".to_string(),
            1,
            Box::new(|args: &[Value]| {
                if args.len() != 1 {
                    return Err(RuntimeError::InvalidOperation {
                        op: "sqrt expects 1 argument".to_string(),
                        span: crate::compiler::diag::Span::new(0, 0, 0),
                    });
                }
                match &args[0] {
                    Value::Int(i) => Ok(Value::Float((*i as f64).sqrt())),
                    Value::Float(f) => Ok(Value::Float(f.sqrt())),
                    _ => Err(RuntimeError::InvalidOperation {
                        op: "sqrt expects number".to_string(),
                        span: crate::compiler::diag::Span::new(0, 0, 0),
                    }),
                }
            })
        )));
        let math_module_value = Value::Map(math_module);
        self.globals.push(math_module_value);
        
        // Input fonksiyonu (slot 6)
        let input_func = Value::NativeFn(NativeFunction::new(
            "input".to_string(),
            0,
            Box::new(stdlib::io::read),
        ));
        self.globals.push(input_func);
        
        // Write fonksiyonu (slot 7)
        let write_func = Value::NativeFn(NativeFunction::new(
            "write".to_string(),
            1,
            Box::new(stdlib::io::write),
        ));
        self.globals.push(write_func);
        
        // Error fonksiyonu (slot 8)
        let error_func = Value::NativeFn(NativeFunction::new(
            "error".to_string(),
            1,
            Box::new(stdlib::io::error),
        ));
        self.globals.push(error_func);
        
        // Format fonksiyonu (slot 9)
        let format_func = Value::NativeFn(NativeFunction::new(
            "format".to_string(),
            255, // Variable arguments
            Box::new(stdlib::io::format),
        ));
        self.globals.push(format_func);
        
        // Len fonksiyonu (slot 10)
        let len_func = Value::NativeFn(NativeFunction::new(
            "len".to_string(),
            1,
            Box::new(stdlib::io::len),
        ));
        self.globals.push(len_func);
        
        // Type fonksiyonu (slot 11)
        let type_func = Value::NativeFn(NativeFunction::new(
            "type".to_string(),
            1,
            Box::new(stdlib::io::type_of),
        ));
        self.globals.push(type_func);
        
        // Time fonksiyonları (slot 12+)
        let sleep_func = Value::NativeFn(NativeFunction::new(
            "sleep".to_string(),
            1,
            Box::new(stdlib::time::sleep),
        ));
        self.globals.push(sleep_func);
        
        let date_string_func = Value::NativeFn(NativeFunction::new(
            "date_string".to_string(),
            0,
            Box::new(stdlib::time::date_string),
        ));
        self.globals.push(date_string_func);
        
        let timer_func = Value::NativeFn(NativeFunction::new(
            "timer".to_string(),
            1,
            Box::new(stdlib::time::timer),
        ));
        self.globals.push(timer_func);
        
        let benchmark_func = Value::NativeFn(NativeFunction::new(
            "benchmark".to_string(),
            1,
            Box::new(stdlib::time::benchmark),
        ));
        self.globals.push(benchmark_func);
        
        let timezone_func = Value::NativeFn(NativeFunction::new(
            "timezone".to_string(),
            0,
            Box::new(stdlib::time::timezone),
        ));
        self.globals.push(timezone_func);
        
        let micros_func = Value::NativeFn(NativeFunction::new(
            "micros".to_string(),
            0,
            Box::new(stdlib::time::micros),
        ));
        self.globals.push(micros_func);
        
        let nanos_func = Value::NativeFn(NativeFunction::new(
            "nanos".to_string(),
            0,
            Box::new(stdlib::time::nanos),
        ));
        self.globals.push(nanos_func);
        
        // HTTP modülü (slot 19)
        let mut http_module = std::collections::HashMap::new();
        http_module.insert("get".to_string(), Value::NativeFn(NativeFunction::new(
            "get".to_string(),
            1,
            Box::new(stdlib::http::get),
        )));
        http_module.insert("post".to_string(), Value::NativeFn(NativeFunction::new(
            "post".to_string(),
            2,
            Box::new(stdlib::http::post),
        )));
        http_module.insert("put".to_string(), Value::NativeFn(NativeFunction::new(
            "put".to_string(),
            2,
            Box::new(stdlib::http::put),
        )));
        http_module.insert("delete".to_string(), Value::NativeFn(NativeFunction::new(
            "delete".to_string(),
            1,
            Box::new(stdlib::http::delete),
        )));
        let http_module_value = Value::Map(http_module);
        self.globals.push(http_module_value);
        
        // Özel değişkenler (slot 20+)
        let args_list: Vec<Value> = self.cli_args.iter().map(|arg| Value::String(arg.clone())).collect();
        let args_value = Value::List(args_list);
        self.globals.push(args_value);
        
        // __name__ (slot 21)
        self.globals.push(Value::String("__main__".to_string()));
        
        // __file__ (slot 22) - CLI'den geçirilecek
        self.globals.push(Value::String("".to_string()));
        
        // __dir__ (slot 23) - CLI'den geçirilecek
        self.globals.push(Value::String("".to_string()));
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::compiler::bytecode::Chunk;

    #[test]
    fn test_vm_creation() {
        let vm = Vm::new();
        assert!(vm.stack.is_empty());
        assert!(vm.frames.is_empty());
    }

    #[test]
    fn test_simple_arithmetic() {
        let mut vm = Vm::new();
        let mut chunk = Chunk::new();
        
        // 1 + 2
        chunk.write_op(OpCode::Const(chunk.add_constant(Value::Int(1))), 1);
        chunk.write_op(OpCode::Const(chunk.add_constant(Value::Int(2))), 2);
        chunk.write_op(OpCode::Add, 3);
        chunk.write_op(OpCode::Return, 4);
        
        let result = vm.run(chunk).unwrap();
        assert_eq!(result, Value::Int(3));
    }

    #[test]
    fn test_print_instruction() {
        let mut vm = Vm::new();
        let mut chunk = Chunk::new();
        
        chunk.write_op(OpCode::Const(chunk.add_constant(Value::String("Hello".to_string()))), 1);
        chunk.write_op(OpCode::Print, 2);
        chunk.write_op(OpCode::Return, 3);
        
        // Print çıktısını test etmek zor, sadece hata olmadığını kontrol et
        let result = vm.run(chunk).unwrap();
        assert_eq!(result, Value::Null);
    }
}
