// Python Bridge - Stdlib Integration
// Python bridge'i stdlib'de kullanmak için wrapper

use crate::compiler::vm::{Value, RuntimeError};
use crate::compiler::bridge::python::PythonBridge;

/// Python modülünü import et
pub fn import_module(args: &[Value]) -> Result<Value, RuntimeError> {
    if args.len() != 1 {
        return Err(RuntimeError::InvalidOperation {
            op: "python.import requires exactly 1 argument".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        });
    }

    let module_name = match &args[0] {
        Value::String(s) => s,
        _ => {
            return Err(RuntimeError::InvalidOperation {
                op: "python.import requires a string argument".to_string(),
                span: crate::compiler::diag::Span::new(0, 0, 0),
            });
        }
    };

    // Python bridge'i kullan
    let mut bridge = PythonBridge::new()?;
    bridge.import_module(module_name)
}
