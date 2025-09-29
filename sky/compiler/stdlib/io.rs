// IO Module - Input/Output Fonksiyonları
// Temel giriş/çıkış işlemleri

use crate::compiler::vm::{value::Value, RuntimeError};

/// Print fonksiyonu
pub fn print(args: &[Value]) -> Result<Value, RuntimeError> {
    if args.len() != 1 {
        return Err(RuntimeError::InvalidOperation {
            op: "print expects 1 argument".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        });
    }

    eprintln!("DEBUG: print fonksiyonu çağrıldı, argüman: {}", args[0].to_string());
    print!("{}", args[0].to_string());
    std::io::Write::flush(&mut std::io::stdout())
        .map_err(|_| RuntimeError::InvalidOperation {
            op: "Failed to flush stdout".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        })?;
    Ok(Value::Null)
}

/// Read fonksiyonu (stdin'den okur)
pub fn read(args: &[Value]) -> Result<Value, RuntimeError> {
    if !args.is_empty() {
        return Err(RuntimeError::InvalidOperation {
            op: "read expects 0 arguments".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        });
    }

    let mut input = String::new();
    std::io::stdin().read_line(&mut input)
        .map_err(|_| RuntimeError::InvalidOperation {
            op: "Failed to read from stdin".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        })?;

    // Newline karakterini kaldır
    if input.ends_with('\n') {
        input.pop();
    }
    if input.ends_with('\r') {
        input.pop();
    }

    Ok(Value::String(input))
}

/// Write fonksiyonu (stdout'a yazar, newline eklemez)
pub fn write(args: &[Value]) -> Result<Value, RuntimeError> {
    if args.len() != 1 {
        return Err(RuntimeError::InvalidOperation {
            op: "write expects 1 argument".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        });
    }

    print!("{}", args[0].to_string());
    std::io::Write::flush(&mut std::io::stdout())
        .map_err(|_| RuntimeError::InvalidOperation {
            op: "Failed to flush stdout".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        })?;

    Ok(Value::Null)
}

/// Error fonksiyonu (stderr'a yazar)
pub fn error(args: &[Value]) -> Result<Value, RuntimeError> {
    if args.len() != 1 {
        return Err(RuntimeError::InvalidOperation {
            op: "error expects 1 argument".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        });
    }

    eprintln!("{}", args[0].to_string());
    Ok(Value::Null)
}

/// Format fonksiyonu (string formatlama)
pub fn format(args: &[Value]) -> Result<Value, RuntimeError> {
    if args.is_empty() {
        return Err(RuntimeError::InvalidOperation {
            op: "format expects at least 1 argument".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        });
    }

    let format_string = match &args[0] {
        Value::String(s) => s.clone(),
        _ => return Err(RuntimeError::InvalidOperation {
            op: "format first argument must be a string".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        }),
    };

    let mut result = String::new();
    let mut arg_index = 1;
    let mut chars = format_string.chars().peekable();

    while let Some(ch) = chars.next() {
        if ch == '{' && chars.peek() == Some(&'}') {
            chars.next(); // Skip the '}'
            
            if arg_index < args.len() {
                result.push_str(&args[arg_index].to_string());
                arg_index += 1;
            } else {
                return Err(RuntimeError::InvalidOperation {
                    op: "Not enough arguments for format string".to_string(),
                    span: crate::compiler::diag::Span::new(0, 0, 0),
                });
            }
        } else {
            result.push(ch);
        }
    }

    Ok(Value::String(result))
}

/// Len fonksiyonu (string/list/map uzunluğu)
pub fn len(args: &[Value]) -> Result<Value, RuntimeError> {
    if args.len() != 1 {
        return Err(RuntimeError::InvalidOperation {
            op: "len expects 1 argument".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        });
    }

    let length = match &args[0] {
        Value::String(s) => s.len(),
        Value::List(l) => l.len(),
        Value::Map(m) => m.len(),
        _ => return Err(RuntimeError::InvalidOperation {
            op: "len expects string, list, or map".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        }),
    };

    Ok(Value::Int(length as i64))
}

/// Type fonksiyonu (değerin tipini döndürür)
pub fn type_of(args: &[Value]) -> Result<Value, RuntimeError> {
    if args.len() != 1 {
        return Err(RuntimeError::InvalidOperation {
            op: "type expects 1 argument".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        });
    }

    let type_name = match &args[0] {
        Value::Int(_) => "int",
        Value::Float(_) => "float",
        Value::Bool(_) => "bool",
        Value::String(_) => "string",
        Value::List(_) => "list",
        Value::Map(_) => "map",
        Value::Function(_) => "function",
        Value::NativeFn(_) => "native_function",
        Value::Future(_) => "future",
        Value::Coroutine(_) => "coroutine",
        Value::Range { .. } => "range",
        Value::Null => "null",
    };

    Ok(Value::String(type_name.to_string()))
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_print() {
        let args = vec![Value::String("Hello, World!".to_string())];
        let result = print(&args);
        assert!(result.is_ok());
        assert_eq!(result.unwrap(), Value::Null);
    }

    #[test]
    fn test_print_wrong_args() {
        let args = vec![];
        let result = print(&args);
        assert!(result.is_err());
        
        let args = vec![Value::Int(1), Value::Int(2)];
        let result = print(&args);
        assert!(result.is_err());
    }

    #[test]
    fn test_format() {
        let args = vec![
            Value::String("Hello, {}!".to_string()),
            Value::String("World".to_string()),
        ];
        let result = format(&args);
        assert!(result.is_ok());
        assert_eq!(result.unwrap(), Value::String("Hello, World!".to_string()));
    }

    #[test]
    fn test_format_multiple() {
        let args = vec![
            Value::String("{} + {} = {}".to_string()),
            Value::Int(2),
            Value::Int(3),
            Value::Int(5),
        ];
        let result = format(&args);
        assert!(result.is_ok());
        assert_eq!(result.unwrap(), Value::String("2 + 3 = 5".to_string()));
    }

    #[test]
    fn test_format_not_enough_args() {
        let args = vec![
            Value::String("Hello, {}!".to_string()),
        ];
        let result = format(&args);
        assert!(result.is_err());
    }

    #[test]
    fn test_len_string() {
        let args = vec![Value::String("hello".to_string())];
        let result = len(&args);
        assert!(result.is_ok());
        assert_eq!(result.unwrap(), Value::Int(5));
    }

    #[test]
    fn test_len_list() {
        let args = vec![Value::List(vec![
            Value::Int(1),
            Value::Int(2),
            Value::Int(3),
        ])];
        let result = len(&args);
        assert!(result.is_ok());
        assert_eq!(result.unwrap(), Value::Int(3));
    }

    #[test]
    fn test_len_invalid_type() {
        let args = vec![Value::Int(42)];
        let result = len(&args);
        assert!(result.is_err());
    }

    #[test]
    fn test_type_of() {
        assert_eq!(type_of(&[Value::Int(42)]).unwrap(), Value::String("int".to_string()));
        assert_eq!(type_of(&[Value::String("hello".to_string())]).unwrap(), Value::String("string".to_string()));
        assert_eq!(type_of(&[Value::Bool(true)]).unwrap(), Value::String("bool".to_string()));
        assert_eq!(type_of(&[Value::Null]).unwrap(), Value::String("null".to_string()));
    }
}
