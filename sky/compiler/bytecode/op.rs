// OpCode Definitions - Bytecode Operasyon Kodları
// VM'in yürüteceği tüm operasyon kodlarını tanımlar

/// Bytecode operasyon kodları
#[derive(Debug, Clone, Copy, PartialEq)]
pub enum OpCode {
    // Constants
    Const(u16),      // Sabit değer yükle (consts tablosundan)
    
    // Variables
    LoadGlobal(u16), // Global değişken yükle
    StoreGlobal(u16), // Global değişkene yaz
    LoadLocal(u16),  // Yerel değişken yükle
    StoreLocal(u16), // Yerel değişkene yaz
    
    // Type-checked variable operations
    StoreLocalTyped(u16, u8), // Yerel değişkene tip kontrolü ile yaz (slot, type)
    StoreGlobalTyped(u16, u8), // Global değişkene tip kontrolü ile yaz (slot, type)
    
    // Arithmetic
    Add,             // Toplama
    Sub,             // Çıkarma
    Mul,             // Çarpma
    Div,             // Bölme
    Mod,             // Modulo
    
    // Comparison
    Equal,           // Eşitlik
    NotEqual,        // Eşitsizlik
    Less,            // Küçük
    LessEqual,       // Küçük eşit
    Greater,         // Büyük
    GreaterEqual,    // Büyük eşit
    
    // Logical
    And,             // Mantıksal VE
    Or,              // Mantıksal VEYA
    Not,             // Mantıksal DEĞİL
    
    // Control Flow
    Jump(u16),       // Koşulsuz atlama
    JumpIfFalse(u16), // Yanlışsa atla
    Pop,             // Stack'ten çıkar
    
    // String operations
    Concat,          // Stack top-2 string'i birleştir
    ToString,        // Non-string Value → String dönüştürme
    
    // Object operations
    GetAttr,         // Object attribute access
    SetAttr,         // Object attribute assignment
    GetIndex,        // Array/map index access
    SetIndex,        // Array/map index assignment
    
    // Functions
    MakeFunction(u16, u8), // Fonksiyon oluştur (chunk index, param count)
    Call(u8),        // Fonksiyon çağır (arg count)
    Return,          // Fonksiyondan dön
    
    // Coroutines
    MakeCoopFunction(u16, u8), // Coop fonksiyon oluştur (chunk index, param count)
    CoopNew,         // Yeni coroutine oluştur
    Yield,           // Coroutine'den çık
    CoopResume,      // Coroutine'i devam ettir
    CoopIsDone,      // Coroutine bitti mi?
    
    // Async
    Await,           // Async bekleme
    
    // Range
    MakeRange,       // Range value oluştur (start, end)
    
    // Built-ins
    Print,           // Yazdır
    
    // Stack operations
    Dup,             // Stack'in üstünü kopyala
    Swap,            // İki üst elemanı değiştir
    
    // Iterator operations (for loops)
    IterNew,         // Yeni iterator oluştur
    IterNext,        // Iterator'dan sonraki elemanı al
    IterDone,        // Iterator bitti mi?
    
    // Special
    Nop,             // Hiçbir şey yapma
}

impl OpCode {
    /// Opcode'un byte uzunluğunu döndür
    pub fn size(&self) -> usize {
        match self {
            Self::Const(_) | Self::LoadGlobal(_) | Self::StoreGlobal(_) |
            Self::LoadLocal(_) | Self::StoreLocal(_) |
            Self::StoreLocalTyped(_, _) | Self::StoreGlobalTyped(_, _) |
            Self::MakeFunction(_, _) | Self::MakeCoopFunction(_, _) |
            Self::Jump(_) | Self::JumpIfFalse(_) => 4, // 1 byte opcode + 2 byte chunk index + 1 byte param count
            
            Self::Call(_) => 2, // 1 byte opcode + 1 byte arg count
            
            _ => 1, // Diğerleri 1 byte
        }
    }

    /// Opcode'un operand gerektirip gerektirmediğini kontrol et
    pub fn has_operand(&self) -> bool {
        matches!(self,
            Self::Const(_) | Self::LoadGlobal(_) | Self::StoreGlobal(_) |
            Self::LoadLocal(_) | Self::StoreLocal(_) |
            Self::StoreLocalTyped(_, _) | Self::StoreGlobalTyped(_, _) |
            Self::MakeFunction(_, _) | Self::MakeCoopFunction(_, _) |
            Self::Jump(_) | Self::JumpIfFalse(_) | Self::Call(_)
        )
    }

    /// Opcode'un jump operasyonu olup olmadığını kontrol et
    pub fn is_jump(&self) -> bool {
        matches!(self, Self::Jump(_) | Self::JumpIfFalse(_))
    }

    /// Opcode'un function operasyonu olup olmadığını kontrol et
    pub fn is_function(&self) -> bool {
        matches!(self, Self::MakeFunction(_, _) | Self::MakeCoopFunction(_, _))
    }

    /// Opcode'un variable operasyonu olup olmadığını kontrol et
    pub fn is_variable(&self) -> bool {
        matches!(self,
            Self::LoadGlobal(_) | Self::StoreGlobal(_) |
            Self::LoadLocal(_) | Self::StoreLocal(_) |
            Self::StoreLocalTyped(_, _) | Self::StoreGlobalTyped(_, _)
        )
    }
}

impl std::fmt::Display for OpCode {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            Self::Const(idx) => write!(f, "CONST {}", idx),
            Self::LoadGlobal(idx) => write!(f, "LOAD_GLOBAL {}", idx),
            Self::StoreGlobal(idx) => write!(f, "STORE_GLOBAL {}", idx),
            Self::LoadLocal(idx) => write!(f, "LOAD_LOCAL {}", idx),
            Self::StoreLocal(idx) => write!(f, "STORE_LOCAL {}", idx),
            Self::StoreLocalTyped(slot, ty) => write!(f, "STORE_LOCAL_TYPED {} {}", slot, ty),
            Self::StoreGlobalTyped(slot, ty) => write!(f, "STORE_GLOBAL_TYPED {} {}", slot, ty),
            Self::Add => write!(f, "ADD"),
            Self::Sub => write!(f, "SUB"),
            Self::Mul => write!(f, "MUL"),
            Self::Div => write!(f, "DIV"),
            Self::Mod => write!(f, "MOD"),
            Self::Equal => write!(f, "EQUAL"),
            Self::NotEqual => write!(f, "NOT_EQUAL"),
            Self::Less => write!(f, "LESS"),
            Self::LessEqual => write!(f, "LESS_EQUAL"),
            Self::Greater => write!(f, "GREATER"),
            Self::GreaterEqual => write!(f, "GREATER_EQUAL"),
            Self::And => write!(f, "AND"),
            Self::Or => write!(f, "OR"),
            Self::Not => write!(f, "NOT"),
            Self::Jump(offset) => write!(f, "JUMP {}", offset),
            Self::JumpIfFalse(offset) => write!(f, "JUMP_IF_FALSE {}", offset),
            Self::Pop => write!(f, "POP"),
            Self::MakeFunction(idx, params) => write!(f, "MAKE_FUNCTION {} {}", idx, params),
            Self::Call(argc) => write!(f, "CALL {}", argc),
            Self::Return => write!(f, "RETURN"),
            Self::MakeCoopFunction(idx, params) => write!(f, "MAKE_COOP_FUNCTION {} {}", idx, params),
            Self::CoopNew => write!(f, "COOP_NEW"),
            Self::Yield => write!(f, "YIELD"),
            Self::CoopResume => write!(f, "COOP_RESUME"),
            Self::CoopIsDone => write!(f, "COOP_IS_DONE"),
            Self::Await => write!(f, "AWAIT"),
            Self::MakeRange => write!(f, "MAKE_RANGE"),
            Self::Print => write!(f, "PRINT"),
            Self::Dup => write!(f, "DUP"),
            Self::Swap => write!(f, "SWAP"),
            Self::IterNew => write!(f, "ITER_NEW"),
            Self::IterNext => write!(f, "ITER_NEXT"),
            Self::IterDone => write!(f, "ITER_DONE"),
            Self::Nop => write!(f, "NOP"),
            Self::Concat => write!(f, "CONCAT"),
            Self::ToString => write!(f, "TO_STRING"),
            Self::GetAttr => write!(f, "GET_ATTR"),
            Self::SetAttr => write!(f, "SET_ATTR"),
            Self::GetIndex => write!(f, "GET_INDEX"),
            Self::SetIndex => write!(f, "SET_INDEX"),
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_opcode_sizes() {
        assert_eq!(OpCode::Const(0).size(), 3);
        assert_eq!(OpCode::LoadLocal(5).size(), 3);
        assert_eq!(OpCode::Jump(100).size(), 3);
        assert_eq!(OpCode::MakeFunction(0, 2).size(), 4);
        assert_eq!(OpCode::Call(2).size(), 2);
        assert_eq!(OpCode::Add.size(), 1);
        assert_eq!(OpCode::Return.size(), 1);
    }

    #[test]
    fn test_opcode_properties() {
        assert!(OpCode::Const(0).has_operand());
        assert!(OpCode::Jump(10).has_operand());
        assert!(!OpCode::Add.has_operand());
        assert!(!OpCode::Return.has_operand());

        assert!(OpCode::Jump(0).is_jump());
        assert!(OpCode::JumpIfFalse(0).is_jump());
        assert!(!OpCode::Add.is_jump());

        assert!(OpCode::MakeFunction(0, 2).is_function());
        assert!(OpCode::MakeCoopFunction(0, 1).is_function());
        assert!(!OpCode::Call(0).is_function());

        assert!(OpCode::LoadGlobal(0).is_variable());
        assert!(OpCode::StoreLocal(0).is_variable());
        assert!(!OpCode::Add.is_variable());
    }

    #[test]
    fn test_opcode_display() {
        assert_eq!(format!("{}", OpCode::Const(5)), "CONST 5");
        assert_eq!(format!("{}", OpCode::LoadLocal(3)), "LOAD_LOCAL 3");
        assert_eq!(format!("{}", OpCode::Add), "ADD");
        assert_eq!(format!("{}", OpCode::Jump(100)), "JUMP 100");
    }
}
