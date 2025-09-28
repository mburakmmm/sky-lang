// Chunk - Bytecode Chunk Yapısı
// VM'in yürüteceği bytecode, sabitler ve debug bilgileri

use super::op::OpCode;
use crate::compiler::diag::Span;

/// Bytecode chunk'ı
#[derive(Debug, Clone, PartialEq)]
pub struct Chunk {
    pub code: Vec<u8>,
    pub consts: Vec<Value>,
    pub lines: Vec<LineInfo>,
}

/// Chunk'taki değerler
#[derive(Debug, Clone, PartialEq)]
pub enum Value {
    Int(i64),
    Float(f64),
    Bool(bool),
    String(String),
    List(Vec<Value>),
    Map(Vec<(String, Value)>),
    Function(usize), // Chunk index
    CoopFunction(usize), // Chunk index
    Null,
}

/// Satır bilgisi (debug için)
#[derive(Debug, Clone, PartialEq)]
pub struct LineInfo {
    pub line: usize,
    pub start_offset: usize,
    pub end_offset: usize,
}

impl Chunk {
    pub fn new() -> Self {
        Self {
            code: Vec::new(),
            consts: Vec::new(),
            lines: Vec::new(),
        }
    }

    pub fn disassemble(&self) -> String {
        let mut output = String::new();
        output.push_str("=== BYTECODE DISASSEMBLY ===\n");
        output.push_str("Constants:\n");
        for (i, constant) in self.consts.iter().enumerate() {
            output.push_str(&format!("  [{}] {:?}\n", i, constant));
        }
        output.push_str("\nCode:\n");
        
        let mut ip = 0;
        while ip < self.code.len() {
            let op = self.code[ip];
            output.push_str(&format!("{:04} ", ip));
            
            match op {
                0x00 => { // Const
                    let index = if ip + 2 < self.code.len() {
                        u16::from_le_bytes([self.code[ip + 1], self.code[ip + 2]])
                    } else { 0 };
                    output.push_str(&format!("CONST {}\n", index));
                    ip += 3;
                }
                0x02 => { // LoadGlobal
                    let index = if ip + 2 < self.code.len() {
                        u16::from_le_bytes([self.code[ip + 1], self.code[ip + 2]])
                    } else { 0 };
                    output.push_str(&format!("LOAD_GLOBAL {}\n", index));
                    ip += 3;
                }
                0x03 => { // StoreGlobal
                    let index = if ip + 2 < self.code.len() {
                        u16::from_le_bytes([self.code[ip + 1], self.code[ip + 2]])
                    } else { 0 };
                    output.push_str(&format!("STORE_GLOBAL {}\n", index));
                    ip += 3;
                }
                0x04 => { // LoadLocal
                    let index = if ip + 2 < self.code.len() {
                        u16::from_le_bytes([self.code[ip + 1], self.code[ip + 2]])
                    } else { 0 };
                    output.push_str(&format!("LOAD_LOCAL {}\n", index));
                    ip += 3;
                }
                0x05 => { // StoreLocal
                    let index = if ip + 2 < self.code.len() {
                        u16::from_le_bytes([self.code[ip + 1], self.code[ip + 2]])
                    } else { 0 };
                    output.push_str(&format!("STORE_LOCAL {}\n", index));
                    ip += 3;
                }
                0x10 => { // Add
                    output.push_str("ADD\n");
                    ip += 1;
                }
                0x11 => { // Sub
                    output.push_str("SUB\n");
                    ip += 1;
                }
                0x12 => { // Mul
                    output.push_str("MUL\n");
                    ip += 1;
                }
                0x13 => { // Div
                    output.push_str("DIV\n");
                    ip += 1;
                }
                0x14 => { // Mod
                    output.push_str("MOD\n");
                    ip += 1;
                }
                0x20 => { // Equal
                    output.push_str("EQUAL\n");
                    ip += 1;
                }
                0x21 => { // NotEqual
                    output.push_str("NOT_EQUAL\n");
                    ip += 1;
                }
                0x22 => { // Less
                    output.push_str("LESS\n");
                    ip += 1;
                }
                0x23 => { // LessEqual
                    output.push_str("LESS_EQUAL\n");
                    ip += 1;
                }
                0x24 => { // Greater
                    output.push_str("GREATER\n");
                    ip += 1;
                }
                0x25 => { // GreaterEqual
                    output.push_str("GREATER_EQUAL\n");
                    ip += 1;
                }
                0x30 => { // And
                    output.push_str("AND\n");
                    ip += 1;
                }
                0x31 => { // Or
                    output.push_str("OR\n");
                    ip += 1;
                }
                0x32 => { // Not
                    output.push_str("NOT\n");
                    ip += 1;
                }
                0x40 => { // Jump
                    let offset = if ip + 2 < self.code.len() {
                        u16::from_le_bytes([self.code[ip + 1], self.code[ip + 2]])
                    } else { 0 };
                    output.push_str(&format!("JUMP {}\n", offset));
                    ip += 3;
                }
                0x41 => { // JumpIfFalse
                    let offset = if ip + 2 < self.code.len() {
                        u16::from_le_bytes([self.code[ip + 1], self.code[ip + 2]])
                    } else { 0 };
                    output.push_str(&format!("JUMP_IF_FALSE {}\n", offset));
                    ip += 3;
                }
                0x50 => { // Pop
                    output.push_str("POP\n");
                    ip += 1;
                }
                0x51 => { // Concat
                    output.push_str("CONCAT\n");
                    ip += 1;
                }
                0x52 => { // ToString
                    output.push_str("TO_STRING\n");
                    ip += 1;
                }
                0x53 => { // GetAttr
                    output.push_str("GET_ATTR\n");
                    ip += 1;
                }
                0x54 => { // SetAttr
                    output.push_str("SET_ATTR\n");
                    ip += 1;
                }
                0x55 => { // GetIndex
                    output.push_str("GET_INDEX\n");
                    ip += 1;
                }
                0x56 => { // SetIndex
                    output.push_str("SET_INDEX\n");
                    ip += 1;
                }
                0x60 => { // MakeFunction
                    let index = if ip + 2 < self.code.len() {
                        u16::from_le_bytes([self.code[ip + 1], self.code[ip + 2]])
                    } else { 0 };
                    let param_count = if ip + 3 < self.code.len() { self.code[ip + 3] } else { 0 };
                    output.push_str(&format!("MAKE_FUNCTION {} {}\n", index, param_count));
                    ip += 4;
                }
                0x61 => { // Call
                    let argc = if ip + 1 < self.code.len() { self.code[ip + 1] } else { 0 };
                    output.push_str(&format!("CALL {}\n", argc));
                    ip += 2;
                }
                0x62 => { // Return
                    output.push_str("RETURN\n");
                    ip += 1;
                }
                0x70 => { // MakeCoopFunction
                    let index = if ip + 2 < self.code.len() {
                        u16::from_le_bytes([self.code[ip + 1], self.code[ip + 2]])
                    } else { 0 };
                    let param_count = if ip + 3 < self.code.len() { self.code[ip + 3] } else { 0 };
                    output.push_str(&format!("MAKE_COOP_FUNCTION {} {}\n", index, param_count));
                    ip += 4;
                }
                0x71 => { // CoopNew
                    output.push_str("COOP_NEW\n");
                    ip += 1;
                }
                0x72 => { // Yield
                    output.push_str("YIELD\n");
                    ip += 1;
                }
                0x73 => { // CoopResume
                    output.push_str("COOP_RESUME\n");
                    ip += 1;
                }
                0x74 => { // CoopIsDone
                    output.push_str("COOP_IS_DONE\n");
                    ip += 1;
                }
                0x80 => { // Await
                    output.push_str("AWAIT\n");
                    ip += 1;
                }
                0x90 => { // Print
                    output.push_str("PRINT\n");
                    ip += 1;
                }
                0xA0 => { // Dup
                    output.push_str("DUP\n");
                    ip += 1;
                }
                0xA1 => { // Swap
                    output.push_str("SWAP\n");
                    ip += 1;
                }
                0xFF => { // Nop
                    output.push_str("NOP\n");
                    ip += 1;
                }
                _ => {
                    output.push_str(&format!("UNKNOWN({:02X})\n", op));
                    ip += 1;
                }
            }
        }
        output.push_str("=== END DISASSEMBLY ===\n");
        output
    }

    /// Opcode ekle
    pub fn write_op(&mut self, op: OpCode, line: usize) {
        let start_offset = self.code.len();
        
        match op {
            OpCode::Const(idx) => {
                self.code.push(0x00); // CONST opcode
                self.code.extend_from_slice(&idx.to_le_bytes());
            }
            OpCode::LoadGlobal(idx) => {
                self.code.push(0x02); // LOAD_GLOBAL opcode
                self.code.extend_from_slice(&idx.to_le_bytes());
            }
            OpCode::StoreGlobal(idx) => {
                self.code.push(0x03); // STORE_GLOBAL opcode
                self.code.extend_from_slice(&idx.to_le_bytes());
            }
            OpCode::LoadLocal(idx) => {
                self.code.push(0x04); // LOAD_LOCAL opcode
                self.code.extend_from_slice(&idx.to_le_bytes());
            }
            OpCode::StoreLocal(idx) => {
                self.code.push(0x05); // STORE_LOCAL opcode
                self.code.extend_from_slice(&idx.to_le_bytes());
            }
            OpCode::Add => self.code.push(0x10),
            OpCode::Sub => self.code.push(0x11),
            OpCode::Mul => self.code.push(0x12),
            OpCode::Div => self.code.push(0x13),
            OpCode::Mod => self.code.push(0x14),
            OpCode::Equal => self.code.push(0x20),
            OpCode::NotEqual => self.code.push(0x21),
            OpCode::Less => self.code.push(0x22),
            OpCode::LessEqual => self.code.push(0x23),
            OpCode::Greater => self.code.push(0x24),
            OpCode::GreaterEqual => self.code.push(0x25),
            OpCode::And => self.code.push(0x30),
            OpCode::Or => self.code.push(0x31),
            OpCode::Not => self.code.push(0x32),
            OpCode::Jump(offset) => {
                self.code.push(0x40);
                self.code.extend_from_slice(&offset.to_le_bytes());
            }
            OpCode::JumpIfFalse(offset) => {
                self.code.push(0x41);
                self.code.extend_from_slice(&offset.to_le_bytes());
            }
            OpCode::Pop => self.code.push(0x50),
            OpCode::Concat => self.code.push(0x51),
            OpCode::ToString => self.code.push(0x52),
            OpCode::GetAttr => self.code.push(0x53),
            OpCode::SetAttr => self.code.push(0x54),
            OpCode::GetIndex => self.code.push(0x55),
            OpCode::SetIndex => self.code.push(0x56),
            OpCode::Dup => self.code.push(0x57),
            OpCode::MakeFunction(idx, param_count) => {
                self.code.push(0x60);
                self.code.extend_from_slice(&idx.to_le_bytes());
                self.code.push(param_count);
            }
            OpCode::Call(argc) => {
                self.code.push(0x61);
                self.code.push(argc);
            }
            OpCode::Return => self.code.push(0x62),
            OpCode::MakeCoopFunction(idx, param_count) => {
                self.code.push(0x70);
                self.code.extend_from_slice(&idx.to_le_bytes());
                self.code.push(param_count);
            }
            OpCode::CoopNew => self.code.push(0x71),
            OpCode::Yield => self.code.push(0x72),
            OpCode::CoopResume => self.code.push(0x73),
            OpCode::CoopIsDone => self.code.push(0x74),
            OpCode::Await => self.code.push(0x80),
            OpCode::Print => self.code.push(0x90),
            OpCode::Swap => self.code.push(0xA1),
            OpCode::Nop => self.code.push(0xFF),
            OpCode::IterNew => self.code.push(0xB0),
            OpCode::IterNext => self.code.push(0xB1),
            OpCode::IterDone => self.code.push(0xB2),
        }
        
        let end_offset = self.code.len();
        self.lines.push(LineInfo {
            line,
            start_offset,
            end_offset,
        });
    }

    /// Sabit ekle
    pub fn add_constant(&mut self, value: Value) -> u16 {
        self.consts.push(value);
        (self.consts.len() - 1) as u16
    }

    /// Belirli bir offset'teki satır numarasını bul
    pub fn get_line(&self, offset: usize) -> usize {
        for line_info in &self.lines {
            if offset >= line_info.start_offset && offset < line_info.end_offset {
                return line_info.line;
            }
        }
        0
    }

    /// Chunk'ın byte uzunluğunu döndür
    pub fn len(&self) -> usize {
        self.code.len()
    }
    
    /// Jump offset'ini patch et
    pub fn patch_jump(&mut self, jump_offset: usize, target_offset: usize) {
        if jump_offset + 3 <= self.code.len() {
            let offset = (target_offset - jump_offset - 3) as u16;
            self.code[jump_offset + 1..jump_offset + 3].copy_from_slice(&offset.to_le_bytes());
        }
    }

    /// Chunk'ın boş olup olmadığını kontrol et
    pub fn is_empty(&self) -> bool {
        self.code.is_empty()
    }

    /// Belirli bir offset'teki byte'ı al
    pub fn get_byte(&self, offset: usize) -> Option<u8> {
        self.code.get(offset).copied()
    }

    /// Belirli bir offset'teki 16-bit değeri al (little-endian)
    pub fn get_u16(&self, offset: usize) -> Option<u16> {
        // Bounds check to prevent panic
        if offset < self.code.len() && offset + 1 < self.code.len() {
            Some(u16::from_le_bytes([self.code[offset], self.code[offset + 1]]))
        } else {
            None
        }
    }
}

impl Value {
    /// Değerin türünü döndür
    pub fn kind(&self) -> &'static str {
        match self {
            Self::Int(_) => "int",
            Self::Float(_) => "float",
            Self::Bool(_) => "bool",
            Self::String(_) => "string",
            Self::List(_) => "list",
            Self::Map(_) => "map",
            Self::Function(_) => "function",
            Self::CoopFunction(_) => "coop_function",
            Self::Null => "null",
        }
    }

    /// Değerin truthy olup olmadığını kontrol et
    pub fn is_truthy(&self) -> bool {
        match self {
            Self::Bool(false) | Self::Null => false,
            Self::Int(0) | Self::Float(0.0) => false,
            Self::String(s) => !s.is_empty(),
            Self::List(l) => !l.is_empty(),
            Self::Map(m) => !m.is_empty(),
            _ => true,
        }
    }

    /// Değeri string'e dönüştür
    pub fn to_string(&self) -> String {
        match self {
            Self::Int(i) => i.to_string(),
            Self::Float(f) => f.to_string(),
            Self::Bool(b) => b.to_string(),
            Self::String(s) => s.clone(),
            Self::List(_) => "[list]".to_string(),
            Self::Map(_) => "{map}".to_string(),
            Self::Function(_) => "<function>".to_string(),
            Self::CoopFunction(_) => "<coop_function>".to_string(),
            Self::Null => "null".to_string(),
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_chunk_creation() {
        let chunk = Chunk::new();
        assert!(chunk.is_empty());
        assert_eq!(chunk.len(), 0);
    }

    #[test]
    fn test_write_opcodes() {
        let mut chunk = Chunk::new();
        
        chunk.write_op(OpCode::Add, 1);
        chunk.write_op(OpCode::Const(42), 2);
        chunk.write_op(OpCode::Return, 3);
        
        assert_eq!(chunk.len(), 4); // 1 + 3 + 1 bytes
        assert_eq!(chunk.get_byte(0), Some(0x10)); // ADD
        assert_eq!(chunk.get_byte(1), Some(0x01)); // CONST
        assert_eq!(chunk.get_byte(3), Some(0x62)); // RETURN
    }

    #[test]
    fn test_constants() {
        let mut chunk = Chunk::new();
        
        let idx1 = chunk.add_constant(Value::Int(42));
        let idx2 = chunk.add_constant(Value::String("hello".to_string()));
        
        assert_eq!(idx1, 0);
        assert_eq!(idx2, 1);
        assert_eq!(chunk.consts.len(), 2);
    }

    #[test]
    fn test_line_info() {
        let mut chunk = Chunk::new();
        
        chunk.write_op(OpCode::Add, 1);
        chunk.write_op(OpCode::Const(0), 2);
        chunk.write_op(OpCode::Return, 3);
        
        assert_eq!(chunk.get_line(0), 1);
        assert_eq!(chunk.get_line(1), 2);
        assert_eq!(chunk.get_line(4), 3);
    }

    #[test]
    fn test_value_properties() {
        assert!(Value::Bool(true).is_truthy());
        assert!(!Value::Bool(false).is_truthy());
        assert!(!Value::Null.is_truthy());
        assert!(Value::Int(1).is_truthy());
        assert!(!Value::Int(0).is_truthy());
        assert!(Value::String("hello".to_string()).is_truthy());
        assert!(!Value::String("".to_string()).is_truthy());
    }

    #[test]
    fn test_value_to_string() {
        assert_eq!(Value::Int(42).to_string(), "42");
        assert_eq!(Value::Float(3.14).to_string(), "3.14");
        assert_eq!(Value::Bool(true).to_string(), "true");
        assert_eq!(Value::String("hello".to_string()).to_string(), "hello");
        assert_eq!(Value::Null.to_string(), "null");
    }
}
