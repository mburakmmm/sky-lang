// VM - Yorumlayıcı, Değerler, Çağrı Çerçeveleri
// Stack tabanlı VM, çağrı çerçevesi, globaller, fonksiyon çağrısı

pub mod value;
pub mod stack;
pub mod call;
pub mod vm;
pub mod coroutine;


// Re-exports
pub use value::Value;
pub use crate::compiler::types::runtime::ValueKind;
pub use stack::Stack;
pub use call::CallFrame;
pub use vm::Vm;
pub use coroutine::Coroutine;

/// Runtime hata türleri
#[derive(Debug, Clone)]
pub enum RuntimeError {
    TypeMismatch {
        expected: String,
        actual: String,
        span: crate::compiler::diag::Span,
    },
    TypeError {
        expected: String,
        found: String,
        span: crate::compiler::diag::Span,
    },
    UndefinedVariable {
        name: String,
        span: crate::compiler::diag::Span,
    },
    CoroutineFinished {
        span: crate::compiler::diag::Span,
    },
    StackOverflow,
    StackUnderflow,
    DivisionByZero {
        span: crate::compiler::diag::Span,
    },
    InvalidOperation {
        op: String,
        span: crate::compiler::diag::Span,
    },
    PythonBridgeError {
        message: String,
        span: crate::compiler::diag::Span,
    },
    JSBridgeError {
        message: String,
        span: crate::compiler::diag::Span,
    },
    InvalidOpcode {
        opcode: u8,
        span: crate::compiler::diag::Span,
    },
}

impl std::fmt::Display for RuntimeError {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            Self::TypeMismatch { expected, actual, .. } => {
                write!(f, "Type mismatch: expected {}, found {}", expected, actual)
            }
            Self::UndefinedVariable { name, .. } => {
                write!(f, "Undefined variable: {}", name)
            }
            Self::CoroutineFinished { .. } => {
                write!(f, "Coroutine already finished")
            }
            Self::StackOverflow => write!(f, "Stack overflow"),
            Self::StackUnderflow => write!(f, "Stack underflow"),
            Self::DivisionByZero { .. } => write!(f, "Division by zero"),
            Self::InvalidOperation { op, .. } => write!(f, "Invalid operation: {}", op),
            Self::PythonBridgeError { message, .. } => write!(f, "Python bridge error: {}", message),
            Self::JSBridgeError { message, .. } => write!(f, "JS bridge error: {}", message),
            Self::TypeError { expected, found, .. } => write!(f, "Type error: expected {}, found {}", expected, found),
            Self::InvalidOpcode { opcode, .. } => write!(f, "Invalid opcode: 0x{:02X}", opcode),
        }
    }
}

impl std::error::Error for RuntimeError {}

impl From<pyo3::PyErr> for RuntimeError {
    fn from(err: pyo3::PyErr) -> Self {
        RuntimeError::PythonBridgeError {
            message: err.to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_runtime_error_display() {
        let error = RuntimeError::TypeMismatch {
            expected: "int".to_string(),
            actual: "string".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        };
        assert!(format!("{}", error).contains("Type mismatch"));
    }
}
