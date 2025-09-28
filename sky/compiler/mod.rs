// Sky Compiler - Ana Modül
// Tüm derleyici bileşenlerini organize eder

pub mod diag;
pub mod lexer;
pub mod parser;
// pub mod ast; // AST structures are in parser/ast.rs
pub mod types;
pub mod binder;
pub mod bytecode;
pub mod vm;
pub mod runtime;
pub mod stdlib;
pub mod bridge;
// pub mod formatter; // Formatter is in sky/formatter/mod.rs

// Re-exports
pub use diag::{Diagnostic, Span, SourceMap, Emitter};
pub use lexer::{lex, token::Token};
pub use parser::{parse, ast::Ast};
pub use types::{decl::TypeDecl, decl::Param};
pub use binder::{bind, BoundAst};
pub use bytecode::{chunk::Chunk, op::OpCode};
pub use vm::{vm::Vm, value::Value, RuntimeError};
