// Bytecode - Opcode & Chunk
// VM'in yürüteceği bayt kodları, sabit havuzu, AST'ten kod üretimi

pub mod op;
pub mod chunk;
pub mod dis;
pub mod compiler;

use crate::compiler::binder::BoundAst;
use crate::compiler::diag::Diagnostic;

// Re-exports
pub use op::{OpCode, OpCode::*};
pub use chunk::{Chunk, LineInfo};
pub use dis::disassemble_chunk;
pub use compiler::Compiler;

/// Ana derleme fonksiyonu
pub fn compile(ast: BoundAst) -> Result<Chunk, Diagnostic> {
    let mut compiler = Compiler::new();
    compiler.compile_ast(ast)
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::compiler::lexer::lex;
    use crate::compiler::parser::parse;
    use crate::compiler::binder::bind;

    #[test]
    fn test_simple_arithmetic_compilation() {
        let tokens = lex("int x = 1 + 2").unwrap();
        let ast = parse(tokens).unwrap();
        let bound_ast = bind(ast).unwrap();
        let chunk = compile(bound_ast).unwrap();
        
        // Chunk'ta en azından CONST, ADD, STORE opcode'ları olmalı
        assert!(!chunk.code.is_empty());
    }

    #[test]
    fn test_function_compilation() {
        let tokens = lex("function test()\n  return 42").unwrap();
        let ast = parse(tokens).unwrap();
        let bound_ast = bind(ast).unwrap();
        let chunk = compile(bound_ast).unwrap();
        
        // Chunk'ta MAKE_FUNCTION ve RETURN opcode'ları olmalı
        assert!(!chunk.code.is_empty());
    }
}
