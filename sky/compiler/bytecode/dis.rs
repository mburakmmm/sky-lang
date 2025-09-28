// Disassembler - Bytecode Ayrıştırıcı
// Bytecode'u okunabilir formata dönüştürür

use super::chunk::{Chunk, Value};
use super::op::OpCode;

/// Chunk'ı disassemble et
pub fn disassemble_chunk(chunk: &Chunk, name: &str) -> String {
    let mut output = format!("=== {} ===\n", name);
    
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
        0x01 => simple_instruction("CONST", offset, output, 1),
        0x02 => simple_instruction("LOAD_GLOBAL", offset, output, 1),
        0x03 => simple_instruction("STORE_GLOBAL", offset, output, 1),
        0x04 => simple_instruction("LOAD_LOCAL", offset, output, 1),
        0x05 => simple_instruction("STORE_LOCAL", offset, output, 1),
        0x10 => simple_instruction("ADD", offset, output, 0),
        0x11 => simple_instruction("SUB", offset, output, 0),
        0x12 => simple_instruction("MUL", offset, output, 0),
        0x13 => simple_instruction("DIV", offset, output, 0),
        0x14 => simple_instruction("MOD", offset, output, 0),
        0x20 => simple_instruction("EQUAL", offset, output, 0),
        0x21 => simple_instruction("NOT_EQUAL", offset, output, 0),
        0x22 => simple_instruction("LESS", offset, output, 0),
        0x23 => simple_instruction("LESS_EQUAL", offset, output, 0),
        0x24 => simple_instruction("GREATER", offset, output, 0),
        0x25 => simple_instruction("GREATER_EQUAL", offset, output, 0),
        0x30 => simple_instruction("AND", offset, output, 0),
        0x31 => simple_instruction("OR", offset, output, 0),
        0x32 => simple_instruction("NOT", offset, output, 0),
        0x40 => jump_instruction("JUMP", offset, chunk, output),
        0x41 => jump_instruction("JUMP_IF_FALSE", offset, chunk, output),
        0x50 => simple_instruction("POP", offset, output, 0),
        0x60 => simple_instruction("MAKE_FUNCTION", offset, output, 1),
        0x61 => call_instruction(offset, chunk, output),
        0x62 => simple_instruction("RETURN", offset, output, 0),
        0x70 => simple_instruction("MAKE_COOP_FUNCTION", offset, output, 1),
        0x71 => simple_instruction("COOP_NEW", offset, output, 0),
        0x72 => simple_instruction("YIELD", offset, output, 0),
        0x73 => simple_instruction("COOP_RESUME", offset, output, 0),
        0x74 => simple_instruction("COOP_IS_DONE", offset, output, 0),
        0x80 => simple_instruction("AWAIT", offset, output, 0),
        0x90 => simple_instruction("PRINT", offset, output, 0),
        0xA0 => simple_instruction("DUP", offset, output, 0),
        0xA1 => simple_instruction("SWAP", offset, output, 0),
        0xFF => simple_instruction("NOP", offset, output, 0),
        _ => {
            output.push_str(&format!("Unknown opcode: 0x{:02X}\n", instruction));
            offset + 1
        }
    }
}

fn simple_instruction(name: &str, offset: usize, output: &mut String, operand_bytes: usize) -> usize {
    if operand_bytes == 0 {
        output.push_str(&format!("{}\n", name));
        offset + 1
    } else {
        let operand: u16 = if operand_bytes == 1 {
            // 8-bit operand
            if offset + 1 < output.len() {
                output.as_bytes()[offset + 1] as u16
            } else {
                0
            }
        } else if operand_bytes == 2 {
            // 16-bit operand
            if let Some(bytes) = output.as_bytes().get(offset + 1..offset + 3) {
                u16::from_le_bytes([bytes[0], bytes[1]])
            } else {
                0
            }
        } else {
            0
        };
        
        if name == "CONST" {
            output.push_str(&format!("{} {}\n", name, operand));
        } else {
            output.push_str(&format!("{} {}\n", name, operand));
        }
        
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
