// Prelude - Global Fonksiyonlar
// Global scope'a eklenen temel fonksiyonlar

use crate::compiler::vm::{value::Value, RuntimeError};
use crate::compiler::vm::value::NativeFunction;
use std::collections::HashMap;

/// Global print fonksiyonu
pub fn print(args: &[Value]) -> Result<Value, RuntimeError> {
    crate::compiler::stdlib::io::print(args)
}

/// Global len fonksiyonu
pub fn len(args: &[Value]) -> Result<Value, RuntimeError> {
    crate::compiler::stdlib::io::len(args)
}

/// Global type fonksiyonu
pub fn type_of(args: &[Value]) -> Result<Value, RuntimeError> {
    crate::compiler::stdlib::io::type_of(args)
}

/// Range fonksiyonu
pub fn range(args: &[Value]) -> Result<Value, RuntimeError> {
    if args.len() < 1 || args.len() > 3 {
        return Err(RuntimeError::InvalidOperation {
            op: "range expects 1-3 arguments".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        });
    }

    let start = if args.len() == 1 {
        0
    } else {
        match &args[0] {
            Value::Int(n) => *n,
            _ => return Err(RuntimeError::InvalidOperation {
                op: "range arguments must be int".to_string(),
                span: crate::compiler::diag::Span::new(0, 0, 0),
            }),
        }
    };

    let end = if args.len() == 1 {
        match &args[0] {
            Value::Int(n) => *n,
            _ => return Err(RuntimeError::InvalidOperation {
                op: "range arguments must be int".to_string(),
                span: crate::compiler::diag::Span::new(0, 0, 0),
            }),
        }
    } else {
        match &args[1] {
            Value::Int(n) => *n,
            _ => return Err(RuntimeError::InvalidOperation {
                op: "range arguments must be int".to_string(),
                span: crate::compiler::diag::Span::new(0, 0, 0),
            }),
        }
    };

    let step = if args.len() == 3 {
        match &args[2] {
            Value::Int(n) => *n,
            _ => return Err(RuntimeError::InvalidOperation {
                op: "range arguments must be int".to_string(),
                span: crate::compiler::diag::Span::new(0, 0, 0),
            }),
        }
    } else {
        1
    };

    if step == 0 {
        return Err(RuntimeError::InvalidOperation {
            op: "range step cannot be zero".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        });
    }

    let mut result = Vec::new();
    let mut current = start;

    if step > 0 {
        while current < end {
            result.push(Value::Int(current));
            current += step;
        }
    } else {
        while current > end {
            result.push(Value::Int(current));
            current += step;
        }
    }

    Ok(Value::List(result))
}

/// Min fonksiyonu
pub fn min(args: &[Value]) -> Result<Value, RuntimeError> {
    if args.is_empty() {
        return Err(RuntimeError::InvalidOperation {
            op: "min expects at least 1 argument".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        });
    }

    let mut min_val = &args[0];

    for arg in &args[1..] {
        match (min_val, arg) {
            (Value::Int(a), Value::Int(b)) => {
                if b < a {
                    min_val = arg;
                }
            }
            (Value::Float(a), Value::Float(b)) => {
                if b < a {
                    min_val = arg;
                }
            }
            (Value::Int(a), Value::Float(b)) => {
                if *b < *a as f64 {
                    min_val = arg;
                }
            }
            (Value::Float(a), Value::Int(b)) => {
                if (*b as f64) < *a {
                    min_val = arg;
                }
            }
            _ => return Err(RuntimeError::InvalidOperation {
                op: "min arguments must be numbers".to_string(),
                span: crate::compiler::diag::Span::new(0, 0, 0),
            }),
        }
    }

    Ok(min_val.clone())
}

/// Max fonksiyonu
pub fn max(args: &[Value]) -> Result<Value, RuntimeError> {
    if args.is_empty() {
        return Err(RuntimeError::InvalidOperation {
            op: "max expects at least 1 argument".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        });
    }

    let mut max_val = &args[0];

    for arg in &args[1..] {
        match (max_val, arg) {
            (Value::Int(a), Value::Int(b)) => {
                if b > a {
                    max_val = arg;
                }
            }
            (Value::Float(a), Value::Float(b)) => {
                if b > a {
                    max_val = arg;
                }
            }
            (Value::Int(a), Value::Float(b)) => {
                if *b > *a as f64 {
                    max_val = arg;
                }
            }
            (Value::Float(a), Value::Int(b)) => {
                if *b as f64 > *a {
                    max_val = arg;
                }
            }
            _ => return Err(RuntimeError::InvalidOperation {
                op: "max arguments must be numbers".to_string(),
                span: crate::compiler::diag::Span::new(0, 0, 0),
            }),
        }
    }

    Ok(max_val.clone())
}

/// Abs fonksiyonu
pub fn abs(args: &[Value]) -> Result<Value, RuntimeError> {
    if args.len() != 1 {
        return Err(RuntimeError::InvalidOperation {
            op: "abs expects 1 argument".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        });
    }

    match &args[0] {
        Value::Int(n) => Ok(Value::Int(n.abs())),
        Value::Float(n) => Ok(Value::Float(n.abs())),
        _ => Err(RuntimeError::InvalidOperation {
            op: "abs argument must be a number".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        }),
    }
}

/// Round fonksiyonu
pub fn round(args: &[Value]) -> Result<Value, RuntimeError> {
    if args.len() != 1 {
        return Err(RuntimeError::InvalidOperation {
            op: "round expects 1 argument".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        });
    }

    match &args[0] {
        Value::Float(n) => Ok(Value::Float(n.round())),
        Value::Int(n) => Ok(Value::Int(*n)),
        _ => Err(RuntimeError::InvalidOperation {
            op: "round argument must be a number".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        }),
    }
}

/// Floor fonksiyonu
pub fn floor(args: &[Value]) -> Result<Value, RuntimeError> {
    if args.len() != 1 {
        return Err(RuntimeError::InvalidOperation {
            op: "floor expects 1 argument".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        });
    }

    match &args[0] {
        Value::Float(n) => Ok(Value::Float(n.floor())),
        Value::Int(n) => Ok(Value::Int(*n)),
        _ => Err(RuntimeError::InvalidOperation {
            op: "floor argument must be a number".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        }),
    }
}

/// Ceil fonksiyonu
pub fn ceil(args: &[Value]) -> Result<Value, RuntimeError> {
    if args.len() != 1 {
        return Err(RuntimeError::InvalidOperation {
            op: "ceil expects 1 argument".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        });
    }

    match &args[0] {
        Value::Float(n) => Ok(Value::Float(n.ceil())),
        Value::Int(n) => Ok(Value::Int(*n)),
        _ => Err(RuntimeError::InvalidOperation {
            op: "ceil argument must be a number".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        }),
    }
}

/// Pow fonksiyonu
pub fn pow(args: &[Value]) -> Result<Value, RuntimeError> {
    if args.len() != 2 {
        return Err(RuntimeError::InvalidOperation {
            op: "pow expects 2 arguments".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        });
    }

    let base = match &args[0] {
        Value::Int(n) => *n as f64,
        Value::Float(n) => *n,
        _ => return Err(RuntimeError::InvalidOperation {
            op: "pow arguments must be numbers".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        }),
    };

    let exponent = match &args[1] {
        Value::Int(n) => *n as f64,
        Value::Float(n) => *n,
        _ => return Err(RuntimeError::InvalidOperation {
            op: "pow arguments must be numbers".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        }),
    };

    let result = base.powf(exponent);
    Ok(Value::Float(result))
}

/// Sqrt fonksiyonu
pub fn sqrt(args: &[Value]) -> Result<Value, RuntimeError> {
    if args.len() != 1 {
        return Err(RuntimeError::InvalidOperation {
            op: "sqrt expects 1 argument".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        });
    }

    let n = match &args[0] {
        Value::Int(n) => *n as f64,
        Value::Float(n) => *n,
        _ => return Err(RuntimeError::InvalidOperation {
            op: "sqrt argument must be a number".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        }),
    };

    if n < 0.0 {
        return Err(RuntimeError::InvalidOperation {
            op: "sqrt of negative number".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        });
    }

    Ok(Value::Float(n.sqrt()))
}

/// Prelude fonksiyonlarını döndür
pub fn get_prelude_functions() -> HashMap<String, NativeFunction> {
    let mut functions = HashMap::new();

    functions.insert("print".to_string(), NativeFunction::new("print".to_string(), 1, Box::new(print)));
    functions.insert("len".to_string(), NativeFunction::new("len".to_string(), 1, Box::new(len)));
    functions.insert("type".to_string(), NativeFunction::new("type".to_string(), 1, Box::new(type_of)));
    functions.insert("range".to_string(), NativeFunction::new("range".to_string(), 3, Box::new(range)));
    functions.insert("min".to_string(), NativeFunction::new("min".to_string(), 255, Box::new(min))); // Variable args
    functions.insert("max".to_string(), NativeFunction::new("max".to_string(), 255, Box::new(max))); // Variable args
    functions.insert("abs".to_string(), NativeFunction::new("abs".to_string(), 1, Box::new(abs)));
    functions.insert("round".to_string(), NativeFunction::new("round".to_string(), 1, Box::new(round)));
    functions.insert("floor".to_string(), NativeFunction::new("floor".to_string(), 1, Box::new(floor)));
    functions.insert("ceil".to_string(), NativeFunction::new("ceil".to_string(), 1, Box::new(ceil)));
    functions.insert("pow".to_string(), NativeFunction::new("pow".to_string(), 2, Box::new(pow)));
    functions.insert("sqrt".to_string(), NativeFunction::new("sqrt".to_string(), 1, Box::new(sqrt)));

    functions
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_range_single_arg() {
        let args = vec![Value::Int(5)];
        let result = range(&args);
        assert!(result.is_ok());
        
        if let Ok(Value::List(list)) = result {
            assert_eq!(list.len(), 5);
            assert_eq!(list[0], Value::Int(0));
            assert_eq!(list[4], Value::Int(4));
        } else {
            panic!("Expected List result");
        }
    }

    #[test]
    fn test_range_two_args() {
        let args = vec![Value::Int(2), Value::Int(5)];
        let result = range(&args);
        assert!(result.is_ok());
        
        if let Ok(Value::List(list)) = result {
            assert_eq!(list.len(), 3);
            assert_eq!(list[0], Value::Int(2));
            assert_eq!(list[2], Value::Int(4));
        } else {
            panic!("Expected List result");
        }
    }

    #[test]
    fn test_range_three_args() {
        let args = vec![Value::Int(0), Value::Int(10), Value::Int(2)];
        let result = range(&args);
        assert!(result.is_ok());
        
        if let Ok(Value::List(list)) = result {
            assert_eq!(list.len(), 5);
            assert_eq!(list[0], Value::Int(0));
            assert_eq!(list[1], Value::Int(2));
            assert_eq!(list[4], Value::Int(8));
        } else {
            panic!("Expected List result");
        }
    }

    #[test]
    fn test_range_zero_step() {
        let args = vec![Value::Int(0), Value::Int(10), Value::Int(0)];
        let result = range(&args);
        assert!(result.is_err());
    }

    #[test]
    fn test_min() {
        let args = vec![Value::Int(3), Value::Int(1), Value::Int(5)];
        let result = min(&args);
        assert!(result.is_ok());
        assert_eq!(result.unwrap(), Value::Int(1));
    }

    #[test]
    fn test_max() {
        let args = vec![Value::Int(3), Value::Int(1), Value::Int(5)];
        let result = max(&args);
        assert!(result.is_ok());
        assert_eq!(result.unwrap(), Value::Int(5));
    }

    #[test]
    fn test_abs_int() {
        let args = vec![Value::Int(-42)];
        let result = abs(&args);
        assert!(result.is_ok());
        assert_eq!(result.unwrap(), Value::Int(42));
    }

    #[test]
    fn test_abs_float() {
        let args = vec![Value::Float(-3.14)];
        let result = abs(&args);
        assert!(result.is_ok());
        assert_eq!(result.unwrap(), Value::Float(3.14));
    }

    #[test]
    fn test_round() {
        let args = vec![Value::Float(3.7)];
        let result = round(&args);
        assert!(result.is_ok());
        assert_eq!(result.unwrap(), Value::Float(4.0));
    }

    #[test]
    fn test_floor() {
        let args = vec![Value::Float(3.9)];
        let result = floor(&args);
        assert!(result.is_ok());
        assert_eq!(result.unwrap(), Value::Float(3.0));
    }

    #[test]
    fn test_ceil() {
        let args = vec![Value::Float(3.1)];
        let result = ceil(&args);
        assert!(result.is_ok());
        assert_eq!(result.unwrap(), Value::Float(4.0));
    }

    #[test]
    fn test_pow() {
        let args = vec![Value::Int(2), Value::Int(3)];
        let result = pow(&args);
        assert!(result.is_ok());
        assert_eq!(result.unwrap(), Value::Float(8.0));
    }

    #[test]
    fn test_sqrt() {
        let args = vec![Value::Int(16)];
        let result = sqrt(&args);
        assert!(result.is_ok());
        assert_eq!(result.unwrap(), Value::Float(4.0));
    }

    #[test]
    fn test_sqrt_negative() {
        let args = vec![Value::Int(-1)];
        let result = sqrt(&args);
        assert!(result.is_err());
    }

    #[test]
    fn test_prelude_functions() {
        let functions = get_prelude_functions();
        assert!(functions.contains_key("print"));
        assert!(functions.contains_key("len"));
        assert!(functions.contains_key("range"));
        assert!(functions.contains_key("min"));
        assert!(functions.contains_key("max"));
        assert!(functions.contains_key("abs"));
        assert!(functions.contains_key("pow"));
        assert!(functions.contains_key("sqrt"));
    }
}
