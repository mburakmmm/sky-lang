// String Utilities - String yardımcı fonksiyonları
// stringify, interpolation ve string işlemleri

use crate::compiler::vm::{Value, RuntimeError};

/// Value'yu string'e dönüştür
pub fn stringify(value: &Value) -> String {
    match value {
        Value::Null => "null".to_string(),
        Value::Bool(b) => b.to_string(),
        Value::Int(i) => i.to_string(),
        Value::Float(f) => f.to_string(),
        Value::String(s) => s.clone(),
        Value::List(items) => {
            let mut result = String::new();
            result.push('[');
            for (i, item) in items.iter().enumerate() {
                if i > 0 {
                    result.push_str(", ");
                }
                result.push_str(&stringify(item));
            }
            result.push(']');
            result
        }
        Value::Map(entries) => {
            let mut result = String::new();
            result.push('{');
            for (i, (key, value)) in entries.iter().enumerate() {
                if i > 0 {
                    result.push_str(", ");
                }
                result.push_str(key);
                result.push_str(": ");
                result.push_str(&stringify(value));
            }
            result.push('}');
            result
        }
        Value::Function(func) => format!("<function {}>", func.name),
        Value::NativeFn(func) => format!("<native function {}>", func.name),
        Value::Future(_) => "<future>".to_string(),
        Value::Coroutine(coro) => format!("<coroutine {}>", coro.id),
        Value::Range { start, end } => format!("{}..{}", start, end),
    }
}

/// String interpolation fonksiyonu
pub fn interpolate_string(args: &[Value]) -> Result<Value, RuntimeError> {
    if args.len() != 1 {
        return Err(RuntimeError::InvalidOperation {
            op: "stringify expects 1 argument".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        });
    }
    
    let result = stringify(&args[0]);
    Ok(Value::String(result))
}

/// String birleştirme fonksiyonu
pub fn concat_strings(args: &[Value]) -> Result<Value, RuntimeError> {
    if args.len() < 2 {
        return Err(RuntimeError::InvalidOperation {
            op: "concat expects at least 2 arguments".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        });
    }
    
    let mut result = String::new();
    for arg in args {
        result.push_str(&stringify(arg));
    }
    
    Ok(Value::String(result))
}

/// String uzunluğu fonksiyonu
pub fn string_length(args: &[Value]) -> Result<Value, RuntimeError> {
    if args.len() != 1 {
        return Err(RuntimeError::InvalidOperation {
            op: "string.length expects 1 argument".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        });
    }
    
    match &args[0] {
        Value::String(s) => Ok(Value::Int(s.len() as i64)),
        _ => Err(RuntimeError::InvalidOperation {
            op: "string.length argument must be string".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        }),
    }
}

/// String parçalama fonksiyonu
pub fn string_split(args: &[Value]) -> Result<Value, RuntimeError> {
    if args.len() != 2 {
        return Err(RuntimeError::InvalidOperation {
            op: "string.split expects 2 arguments".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        });
    }
    
    let string = match &args[0] {
        Value::String(s) => s,
        _ => return Err(RuntimeError::InvalidOperation {
            op: "string.split first argument must be string".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        }),
    };
    
    let delimiter = match &args[1] {
        Value::String(s) => s,
        _ => return Err(RuntimeError::InvalidOperation {
            op: "string.split second argument must be string".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        }),
    };
    
    let parts: Vec<Value> = string
        .split(delimiter)
        .map(|s| Value::String(s.to_string()))
        .collect();
    
    Ok(Value::List(parts))
}

/// String birleştirme fonksiyonu
pub fn string_join(args: &[Value]) -> Result<Value, RuntimeError> {
    if args.len() != 2 {
        return Err(RuntimeError::InvalidOperation {
            op: "string.join expects 2 arguments".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        });
    }
    
    let list = match &args[0] {
        Value::List(items) => items,
        _ => return Err(RuntimeError::InvalidOperation {
            op: "string.join first argument must be list".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        }),
    };
    
    let delimiter = match &args[1] {
        Value::String(s) => s,
        _ => return Err(RuntimeError::InvalidOperation {
            op: "string.join second argument must be string".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        }),
    };
    
    let parts: Vec<String> = list
        .iter()
        .map(|item| stringify(item))
        .collect();
    
    let result = parts.join(delimiter);
    Ok(Value::String(result))
}

/// String büyük harfe çevirme
pub fn string_upper(args: &[Value]) -> Result<Value, RuntimeError> {
    if args.len() != 1 {
        return Err(RuntimeError::InvalidOperation {
            op: "string.upper expects 1 argument".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        });
    }
    
    match &args[0] {
        Value::String(s) => Ok(Value::String(s.to_uppercase())),
        _ => Err(RuntimeError::InvalidOperation {
            op: "string.upper argument must be string".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        }),
    }
}

/// String küçük harfe çevirme
pub fn string_lower(args: &[Value]) -> Result<Value, RuntimeError> {
    if args.len() != 1 {
        return Err(RuntimeError::InvalidOperation {
            op: "string.lower expects 1 argument".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        });
    }
    
    match &args[0] {
        Value::String(s) => Ok(Value::String(s.to_lowercase())),
        _ => Err(RuntimeError::InvalidOperation {
            op: "string.lower argument must be string".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        }),
    }
}

/// String trim (başındaki ve sonundaki boşlukları temizle)
pub fn string_trim(args: &[Value]) -> Result<Value, RuntimeError> {
    if args.len() != 1 {
        return Err(RuntimeError::InvalidOperation {
            op: "string.trim expects 1 argument".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        });
    }
    
    match &args[0] {
        Value::String(s) => Ok(Value::String(s.trim().to_string())),
        _ => Err(RuntimeError::InvalidOperation {
            op: "string.trim argument must be string".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        }),
    }
}

/// String içinde arama
pub fn string_contains(args: &[Value]) -> Result<Value, RuntimeError> {
    if args.len() != 2 {
        return Err(RuntimeError::InvalidOperation {
            op: "string.contains expects 2 arguments".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        });
    }
    
    let string = match &args[0] {
        Value::String(s) => s,
        _ => return Err(RuntimeError::InvalidOperation {
            op: "string.contains first argument must be string".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        }),
    };
    
    let substring = match &args[1] {
        Value::String(s) => s,
        _ => return Err(RuntimeError::InvalidOperation {
            op: "string.contains second argument must be string".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        }),
    };
    
    Ok(Value::Bool(string.contains(substring)))
}

/// String ile başlama kontrolü
pub fn string_starts_with(args: &[Value]) -> Result<Value, RuntimeError> {
    if args.len() != 2 {
        return Err(RuntimeError::InvalidOperation {
            op: "string.startsWith expects 2 arguments".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        });
    }
    
    let string = match &args[0] {
        Value::String(s) => s,
        _ => return Err(RuntimeError::InvalidOperation {
            op: "string.startsWith first argument must be string".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        }),
    };
    
    let prefix = match &args[1] {
        Value::String(s) => s,
        _ => return Err(RuntimeError::InvalidOperation {
            op: "string.startsWith second argument must be string".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        }),
    };
    
    Ok(Value::Bool(string.starts_with(prefix)))
}

/// String ile bitme kontrolü
pub fn string_ends_with(args: &[Value]) -> Result<Value, RuntimeError> {
    if args.len() != 2 {
        return Err(RuntimeError::InvalidOperation {
            op: "string.endsWith expects 2 arguments".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        });
    }
    
    let string = match &args[0] {
        Value::String(s) => s,
        _ => return Err(RuntimeError::InvalidOperation {
            op: "string.endsWith first argument must be string".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        }),
    };
    
    let suffix = match &args[1] {
        Value::String(s) => s,
        _ => return Err(RuntimeError::InvalidOperation {
            op: "string.endsWith second argument must be string".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        }),
    };
    
    Ok(Value::Bool(string.ends_with(suffix)))
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_stringify_null() {
        assert_eq!(stringify(&Value::Null), "null");
    }

    #[test]
    fn test_stringify_bool() {
        assert_eq!(stringify(&Value::Bool(true)), "true");
        assert_eq!(stringify(&Value::Bool(false)), "false");
    }

    #[test]
    fn test_stringify_int() {
        assert_eq!(stringify(&Value::Int(42)), "42");
        assert_eq!(stringify(&Value::Int(-10)), "-10");
    }

    #[test]
    fn test_stringify_float() {
        assert_eq!(stringify(&Value::Float(3.14)), "3.14");
        assert_eq!(stringify(&Value::Float(-2.5)), "-2.5");
    }

    #[test]
    fn test_stringify_string() {
        assert_eq!(stringify(&Value::String("hello".to_string())), "hello");
    }

    #[test]
    fn test_stringify_list() {
        let list = Value::List(vec![
            Value::Int(1),
            Value::String("test".to_string()),
            Value::Bool(true),
        ]);
        assert_eq!(stringify(&list), "[1, test, true]");
    }

    #[test]
    fn test_stringify_map() {
        let mut map = std::collections::HashMap::new();
        map.insert("key".to_string(), Value::String("value".to_string()));
        map.insert("num".to_string(), Value::Int(42));
        let map_value = Value::Map(map);
        
        // Map order is not guaranteed, so we check if it contains expected parts
        let result = stringify(&map_value);
        assert!(result.contains("key: value"));
        assert!(result.contains("num: 42"));
        assert!(result.starts_with("{"));
        assert!(result.ends_with("}"));
    }

    #[test]
    fn test_stringify_function() {
        let func = crate::compiler::vm::Function::new(
            "test".to_string(),
            Vec::new(),
            crate::compiler::bytecode::Chunk::new(),
        );
        let func_value = Value::Function(Box::new(func));
        assert!(stringify(&func_value).starts_with("<function test>"));
    }

    #[test]
    fn test_interpolate_string() {
        let result = interpolate_string(&[Value::Int(42)]);
        assert_eq!(result, Ok(Value::String("42".to_string())));
    }

    #[test]
    fn test_concat_strings() {
        let args = vec![
            Value::String("hello".to_string()),
            Value::String(" ".to_string()),
            Value::String("world".to_string()),
        ];
        let result = concat_strings(&args);
        assert_eq!(result, Ok(Value::String("hello world".to_string())));
    }

    #[test]
    fn test_string_length() {
        let result = string_length(&[Value::String("hello".to_string())]);
        assert_eq!(result, Ok(Value::Int(5)));
    }

    #[test]
    fn test_string_split() {
        let args = vec![
            Value::String("a,b,c".to_string()),
            Value::String(",".to_string()),
        ];
        let result = string_split(&args);
        
        if let Ok(Value::List(parts)) = result {
            assert_eq!(parts.len(), 3);
            assert_eq!(parts[0], Value::String("a".to_string()));
            assert_eq!(parts[1], Value::String("b".to_string()));
            assert_eq!(parts[2], Value::String("c".to_string()));
        } else {
            panic!("Expected list result");
        }
    }

    #[test]
    fn test_string_join() {
        let args = vec![
            Value::List(vec![
                Value::String("a".to_string()),
                Value::String("b".to_string()),
                Value::String("c".to_string()),
            ]),
            Value::String(",".to_string()),
        ];
        let result = string_join(&args);
        assert_eq!(result, Ok(Value::String("a,b,c".to_string())));
    }

    #[test]
    fn test_string_upper() {
        let result = string_upper(&[Value::String("hello".to_string())]);
        assert_eq!(result, Ok(Value::String("HELLO".to_string())));
    }

    #[test]
    fn test_string_lower() {
        let result = string_lower(&[Value::String("HELLO".to_string())]);
        assert_eq!(result, Ok(Value::String("hello".to_string())));
    }

    #[test]
    fn test_string_trim() {
        let result = string_trim(&[Value::String("  hello  ".to_string())]);
        assert_eq!(result, Ok(Value::String("hello".to_string())));
    }

    #[test]
    fn test_string_contains() {
        let args = vec![
            Value::String("hello world".to_string()),
            Value::String("world".to_string()),
        ];
        let result = string_contains(&args);
        assert_eq!(result, Ok(Value::Bool(true)));
    }

    #[test]
    fn test_string_starts_with() {
        let args = vec![
            Value::String("hello world".to_string()),
            Value::String("hello".to_string()),
        ];
        let result = string_starts_with(&args);
        assert_eq!(result, Ok(Value::Bool(true)));
    }

    #[test]
    fn test_string_ends_with() {
        let args = vec![
            Value::String("hello world".to_string()),
            Value::String("world".to_string()),
        ];
        let result = string_ends_with(&args);
        assert_eq!(result, Ok(Value::Bool(true)));
    }
}
