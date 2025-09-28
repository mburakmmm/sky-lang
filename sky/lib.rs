// Sky Programming Language - Main Library
// Girintili sözdizimi, tip beyanı zorunlu, dinamik & güçlü tip (runtime checks)

pub mod compiler;
pub mod formatter;
// pub mod cli; // CLI is in sky/cli/main.rs

// Re-exports for convenience
pub use compiler::diag::Diagnostic;
pub use compiler::lexer::lex;
pub use compiler::parser::parse;
pub use compiler::vm::Vm;
