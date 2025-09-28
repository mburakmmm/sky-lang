// JS Bridge - Stdlib Integration
// JavaScript bridge'i stdlib'de kullanmak için wrapper

use crate::compiler::vm::{Value, RuntimeError};
use crate::compiler::bridge::js::JSBridge;

/// JavaScript kodu çalıştır
pub fn eval_code(args: &[Value]) -> Result<Value, RuntimeError> {
    if args.len() != 1 {
        return Err(RuntimeError::InvalidOperation {
            op: "js.eval requires exactly 1 argument".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        });
    }

    let code = match &args[0] {
        Value::String(s) => s,
        _ => {
            return Err(RuntimeError::InvalidOperation {
                op: "js.eval requires a string argument".to_string(),
                span: crate::compiler::diag::Span::new(0, 0, 0),
            });
        }
    };

    // JS bridge'i kullan
    let mut bridge = JSBridge::new()?;
    bridge.eval_code(code)
}
