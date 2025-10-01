// Disassembler - Bytecode Ayrıştırıcı
// Bytecode'u okunabilir formata dönüştürür

use super::chunk::{Chunk, Value};

/// Chunk'ı disassemble et
pub fn disassemble_chunk(chunk: &Chunk, name: &str) -> String {
    let mut output = format!("=== {} ===\n", name);
    
    // Hex dump ekle
    output.push_str("=== HEX DUMP ===\n");
    for (i, byte) in chunk.code.iter().enumerate() {
        if i % 16 == 0 {
            output.push_str(&format!("{:04x}: ", i));
        }
        output.push_str(&format!("{:02x} ", byte));
        if i % 16 == 15 {
            output.push_str("\n");
        }
    }
    if chunk.code.len() % 16 != 0 {
        output.push_str("\n");
    }
    output.push_str("=== END HEX DUMP ===\n\n");
    
    let mut offset = 0;
    while offset < chunk.len() {
        offset = disassemble_instruction(chunk, offset, &mut output);
    }
    
    output
}

/// Tek bir instruction'ı disassemble et
pub fn disassemble_instruction(chunk: &Chunk, offset: usize, output: &mut String) -> usize {
    output.push_str(&format!("{:04} ", offset));
    
    if offset > 0 && chunk.get_line(offset) == chunk.get_line(offset - 1) {
        output.push_str("   | ");
    } else {
        output.push_str(&format!("{:4} ", chunk.get_line(offset)));
    }
    
    let instruction = chunk.get_byte(offset).unwrap_or(0xFF);
    
    match instruction {
        0x01 => simple_instruction("CONST", offset, output, 2, chunk),
        0x02 => simple_instruction("LOAD_GLOBAL", offset, output, 2, chunk),
        0x03 => simple_instruction("STORE_GLOBAL", offset, output, 2, chunk),
        0x04 => simple_instruction("LOAD_LOCAL", offset, output, 2, chunk),
        0x05 => simple_instruction("STORE_LOCAL", offset, output, 2, chunk),
        0x06 => typed_store_instruction("STORE_LOCAL_TYPED", offset, chunk, output),
        0x07 => typed_store_instruction("STORE_GLOBAL_TYPED", offset, chunk, output),
        0x10 => simple_instruction("ADD", offset, output, 0, chunk),
        0x11 => simple_instruction("SUB", offset, output, 0, chunk),
        0x12 => simple_instruction("MUL", offset, output, 0, chunk),
        0x13 => simple_instruction("DIV", offset, output, 0, chunk),
        0x14 => simple_instruction("MOD", offset, output, 0, chunk),
        0x20 => simple_instruction("EQUAL", offset, output, 0, chunk),
        0x21 => simple_instruction("NOT_EQUAL", offset, output, 0, chunk),
        0x22 => simple_instruction("LESS", offset, output, 0, chunk),
        0x23 => simple_instruction("LESS_EQUAL", offset, output, 0, chunk),
        0x24 => simple_instruction("GREATER", offset, output, 0, chunk),
        0x25 => simple_instruction("GREATER_EQUAL", offset, output, 0, chunk),
        0x30 => simple_instruction("AND", offset, output, 0, chunk),
        0x31 => simple_instruction("OR", offset, output, 0, chunk),
        0x32 => simple_instruction("NOT", offset, output, 0, chunk),
        0x40 => jump_instruction("JUMP", offset, chunk, output),
        0x41 => jump_instruction("JUMP_IF_FALSE", offset, chunk, output),
        0x50 => simple_instruction("POP", offset, output, 0, chunk),
        0x57 => simple_instruction("DUP", offset, output, 0, chunk),
        0x60 => simple_instruction("MAKE_FUNCTION", offset, output, 2, chunk),
        0x61 => call_instruction(offset, chunk, output),
        0x62 => simple_instruction("RETURN", offset, output, 0, chunk),
        0x70 => simple_instruction("MAKE_COOP_FUNCTION", offset, output, 2, chunk),
        0x71 => simple_instruction("COOP_NEW", offset, output, 0, chunk),
        0x72 => simple_instruction("YIELD", offset, output, 0, chunk),
        0x73 => simple_instruction("COOP_RESUME", offset, output, 0, chunk),
        0x74 => simple_instruction("COOP_IS_DONE", offset, output, 0, chunk),
        0x80 => simple_instruction("AWAIT", offset, output, 0, chunk),
        0x90 => simple_instruction("PRINT", offset, output, 0, chunk),
        0xA0 => simple_instruction("DUP", offset, output, 0, chunk),
        0xA1 => simple_instruction("SWAP", offset, output, 0, chunk),
        0xFF => simple_instruction("NOP", offset, output, 0, chunk),
        _ => {
            output.push_str(&format!("Unknown opcode: 0x{:02X}\n", instruction));
            offset + 1
        }
    }
}

fn simple_instruction(name: &str, offset: usize, output: &mut String, operand_bytes: usize, chunk: &Chunk) -> usize {
    if operand_bytes == 0 {
        output.push_str(&format!("{}\n", name));
        offset + 1
    } else {
        let operand: u16 = if operand_bytes == 1 {
            // 8-bit operand
            if let Some(byte) = chunk.get_byte(offset + 1) {
                byte as u16
            } else {
                0
            }
        } else if operand_bytes == 2 {
            // 16-bit operand
            if let Some(operand) = chunk.get_u16(offset + 1) {
                operand
            } else {
                0
            }
        } else {
            0
        };
        
        output.push_str(&format!("{} {}\n", name, operand));
        offset + 1 + operand_bytes
    }
}

fn jump_instruction(name: &str, offset: usize, chunk: &Chunk, output: &mut String) -> usize {
    if let Some(jump_offset) = chunk.get_u16(offset + 1) {
        let target = offset + 3 + jump_offset as usize;
        output.push_str(&format!("{} {:04} -> {:04}\n", name, jump_offset, target));
    } else {
        output.push_str(&format!("{} [invalid]\n", name));
    }
    offset + 3
}

fn call_instruction(offset: usize, chunk: &Chunk, output: &mut String) -> usize {
    if let Some(argc) = chunk.get_byte(offset + 1) {
        output.push_str(&format!("CALL {}\n", argc));
    } else {
        output.push_str("CALL [invalid]\n");
    }
    offset + 2
}

fn typed_store_instruction(name: &str, offset: usize, chunk: &Chunk, output: &mut String) -> usize {
    eprintln!("DEBUG DIS: typed_store_instruction called, offset={}, chunk.code.len()={}", offset, chunk.code.len());
    if offset + 3 < chunk.code.len() {
        eprintln!("DEBUG DIS: chunk.code[{}]={}, chunk.code[{}]={}, chunk.code[{}]={}, chunk.code[{}]={}", 
            offset, chunk.code[offset],
            offset+1, chunk.code[offset+1],
            offset+2, chunk.code[offset+2],
            offset+3, chunk.code[offset+3]);
    }
    if let Some(slot) = chunk.get_u16(offset + 1) {
        if let Some(type_code) = chunk.get_byte(offset + 3) {
            eprintln!("DEBUG DIS: offset={}, slot={}, type_code={}", offset, slot, type_code);
            let type_name = match type_code {
                0 => "var",
                1 => "int",
                2 => "float",
                3 => "bool",
                4 => "string",
                5 => "list",
                6 => "map",
                7 => "list[param]",
                _ => "unknown",
            };
            output.push_str(&format!("{} {} {}\n", name, slot, type_name));
        } else {
            eprintln!("DEBUG DIS: offset={}, slot={}, NO type_code at offset+3", offset, slot);
            output.push_str(&format!("{} {} [invalid type]\n", name, slot));
        }
    } else {
        output.push_str(&format!("{} [invalid]\n", name));
    }
    offset + 4
}

/// Sabit değeri formatla
pub fn format_constant(constant: &Value) -> String {
    match constant {
        Value::Int(i) => i.to_string(),
        Value::Float(f) => f.to_string(),
        Value::Bool(b) => b.to_string(),
        Value::String(s) => format!("\"{}\"", s),
        Value::List(_) => "[list]".to_string(),
        Value::Map(_) => "{map}".to_string(),
        Value::Function(idx) => format!("<function:{}>", idx),
        Value::CoopFunction(idx) => format!("<coop_function:{}>", idx),
        Value::Null => "null".to_string(),
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use super::chunk::Chunk;

    #[test]
    fn test_disassemble_simple() {
        let mut chunk = Chunk::new();
        chunk.write_op(OpCode::Add, 1);
        chunk.write_op(OpCode::Return, 2);
        
        let disasm = disassemble_chunk(&chunk, "test");
        assert!(disasm.contains("ADD"));
        assert!(disasm.contains("RETURN"));
    }

    #[test]
    fn test_disassemble_with_constants() {
        let mut chunk = Chunk::new();
        let const_idx = chunk.add_constant(Value::Int(42));
        chunk.write_op(OpCode::Const(const_idx), 1);
        chunk.write_op(OpCode::Return, 2);
        
        let disasm = disassemble_chunk(&chunk, "test");
        assert!(disasm.contains("CONST"));
    }

    #[test]
    fn test_format_constant() {
        assert_eq!(format_constant(&Value::Int(42)), "42");
        assert_eq!(format_constant(&Value::Float(3.14)), "3.14");
        assert_eq!(format_constant(&Value::Bool(true)), "true");
        assert_eq!(format_constant(&Value::String("hello".to_string())), "\"hello\"");
        assert_eq!(format_constant(&Value::Null), "null");
    }
}
