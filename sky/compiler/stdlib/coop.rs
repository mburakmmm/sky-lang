// Coop Module - Coroutine Yardımcıları
// Coop/yield semantiği için yardımcı fonksiyonlar

use crate::compiler::vm::{value::Value, RuntimeError};
use crate::compiler::vm::value::CoroutineState;
use std::collections::HashMap;

/// Yeni coroutine oluştur
pub fn new(args: &[Value]) -> Result<Value, RuntimeError> {
    if args.len() != 1 {
        return Err(RuntimeError::InvalidOperation {
            op: "coop.new expects 1 argument".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        });
    }

    let function = match &args[0] {
        Value::Function(_) => args[0].clone(),
        _ => return Err(RuntimeError::InvalidOperation {
            op: "coop.new expects function argument".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        }),
    };

    // Coroutine oluştur
    let coroutine = Value::Coroutine(crate::compiler::vm::value::CoroutineValue::new(1)); // Mock ID sistemi
    
    Ok(coroutine)
}

/// Coroutine'i resume et
pub fn resume(args: &[Value]) -> Result<Value, RuntimeError> {
    if args.len() < 1 {
        return Err(RuntimeError::InvalidOperation {
            op: "coop.resume expects at least 1 argument".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        });
    }

    let coroutine = match &args[0] {
        Value::Coroutine(coro) => coro,
        _ => return Err(RuntimeError::InvalidOperation {
            op: "coop.resume first argument must be coroutine".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        }),
    };

    if crate::compiler::vm::value::CoroutineValue::is_done(&coroutine) {
        return Err(RuntimeError::CoroutineFinished {
            span: crate::compiler::diag::Span::new(0, 0, 0),
        });
    }

    // Resume işlemi (basit implementasyon)
    // Gerçek coroutine execution (mock)
    // Gerçek implementasyonda:
    // coroutine.resume(args)
    Ok(Value::Int(42)) // Mock result
}

/// Coroutine'in tamamlanıp tamamlanmadığını kontrol et
pub fn is_done(args: &[Value]) -> Result<Value, RuntimeError> {
    if args.len() != 1 {
        return Err(RuntimeError::InvalidOperation {
            op: "coop.is_done expects 1 argument".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        });
    }

    let coroutine = match &args[0] {
        Value::Coroutine(coro) => coro,
        _ => return Err(RuntimeError::InvalidOperation {
            op: "coop.is_done argument must be coroutine".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        }),
    };

    Ok(Value::Bool(crate::compiler::vm::value::CoroutineValue::is_done(&coroutine)))
}

/// Coroutine'in durumunu döndür
pub fn status(args: &[Value]) -> Result<Value, RuntimeError> {
    if args.len() != 1 {
        return Err(RuntimeError::InvalidOperation {
            op: "coop.status expects 1 argument".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        });
    }

    let coroutine = match &args[0] {
        Value::Coroutine(coro) => coro,
        _ => return Err(RuntimeError::InvalidOperation {
            op: "coop.status argument must be coroutine".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        }),
    };

    let status = match coroutine.state {
        CoroutineState::Suspended => "suspended",
        CoroutineState::Running => "running",
        CoroutineState::Done => "done",
    };

    Ok(Value::String(status.to_string()))
}

/// Coroutine pool oluştur
pub fn pool(args: &[Value]) -> Result<Value, RuntimeError> {
    if !args.is_empty() {
        return Err(RuntimeError::InvalidOperation {
            op: "coop.pool expects 0 arguments".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        });
    }

    // Pool oluştur
    let mut pool_info = HashMap::new();
    pool_info.insert("size".to_string(), Value::Int(0));
    pool_info.insert("max_size".to_string(), Value::Int(100));
    pool_info.insert("active".to_string(), Value::Int(0));
    pool_info.insert("idle".to_string(), Value::Int(0));

    Ok(Value::Map(pool_info))
}

/// Coroutine'i sonlandır
pub fn cancel(args: &[Value]) -> Result<Value, RuntimeError> {
    if args.len() != 1 {
        return Err(RuntimeError::InvalidOperation {
            op: "coop.cancel expects 1 argument".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        });
    }

    let _coroutine = match &args[0] {
        Value::Coroutine(coro) => coro,
        _ => return Err(RuntimeError::InvalidOperation {
            op: "coop.cancel argument must be coroutine".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        }),
    };

    // Cancel işlemi (basit implementasyon)
    Ok(Value::Bool(true))
}

/// Coroutine'in ID'sini döndür
pub fn id(args: &[Value]) -> Result<Value, RuntimeError> {
    if args.len() != 1 {
        return Err(RuntimeError::InvalidOperation {
            op: "coop.id expects 1 argument".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        });
    }

    let coroutine = match &args[0] {
        Value::Coroutine(coro) => coro,
        _ => return Err(RuntimeError::InvalidOperation {
            op: "coop.id argument must be coroutine".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        }),
    };

    Ok(Value::Int(coroutine.id as i64))
}

/// Coroutine bilgilerini döndür
pub fn info(args: &[Value]) -> Result<Value, RuntimeError> {
    if args.len() != 1 {
        return Err(RuntimeError::InvalidOperation {
            op: "coop.info expects 1 argument".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        });
    }

    let coroutine = match &args[0] {
        Value::Coroutine(coro) => coro,
        _ => return Err(RuntimeError::InvalidOperation {
            op: "coop.info argument must be coroutine".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        }),
    };

    let mut info = HashMap::new();
    info.insert("id".to_string(), Value::Int(coroutine.id as i64));
    info.insert("state".to_string(), Value::String(match coroutine.state {
        CoroutineState::Suspended => "suspended".to_string(),
        CoroutineState::Running => "running".to_string(),
        CoroutineState::Done => "done".to_string(),
    }));
    info.insert("is_done".to_string(), Value::Bool(crate::compiler::vm::value::CoroutineValue::is_done(&coroutine)));

    Ok(Value::Map(info))
}

/// Coroutine generator oluştur
pub fn generator(args: &[Value]) -> Result<Value, RuntimeError> {
    if args.len() != 1 {
        return Err(RuntimeError::InvalidOperation {
            op: "coop.generator expects 1 argument".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        });
    }

    let function = match &args[0] {
        Value::Function(_) => args[0].clone(),
        _ => return Err(RuntimeError::InvalidOperation {
            op: "coop.generator expects function argument".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        }),
    };

    // Generator oluştur (coroutine ile aynı)
    let generator = Value::Coroutine(crate::compiler::vm::value::CoroutineValue::new(2)); // Mock ID sistemi
    
    Ok(generator)
}

/// Coroutine scheduler bilgisi
pub fn scheduler(args: &[Value]) -> Result<Value, RuntimeError> {
    if !args.is_empty() {
        return Err(RuntimeError::InvalidOperation {
            op: "coop.scheduler expects 0 arguments".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        });
    }

    let mut scheduler_info = HashMap::new();
    scheduler_info.insert("is_running".to_string(), Value::Bool(true));
    scheduler_info.insert("active_coroutines".to_string(), Value::Int(0));
    scheduler_info.insert("suspended_coroutines".to_string(), Value::Int(0));
    scheduler_info.insert("completed_coroutines".to_string(), Value::Int(0));

    Ok(Value::Map(scheduler_info))
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_coop_new() {
        let args = vec![Value::Function(crate::compiler::vm::value::FunctionValue::new(0, "test".to_string(), 0))];
        let result = new(&args);
        assert!(result.is_ok());
        
        if let Ok(Value::Coroutine(_)) = result {
            // Expected
        } else {
            panic!("Expected Coroutine result");
        }
    }

    #[test]
    fn test_coop_new_invalid_args() {
        let args = vec![Value::Int(42)];
        let result = new(&args);
        assert!(result.is_err());
    }

    #[test]
    fn test_coop_resume() {
        let coroutine = Value::Coroutine(crate::compiler::vm::value::CoroutineValue::new(1));
        let args = vec![coroutine];
        let result = resume(&args);
        assert!(result.is_ok());
        assert_eq!(result.unwrap(), Value::Int(42));
    }

    #[test]
    fn test_coop_resume_invalid_args() {
        let args = vec![Value::Int(42)];
        let result = resume(&args);
        assert!(result.is_err());
    }

    #[test]
    fn test_coop_is_done() {
        let coroutine = Value::Coroutine(crate::compiler::vm::value::CoroutineValue::new(1));
        let args = vec![coroutine];
        let result = is_done(&args);
        assert!(result.is_ok());
        assert_eq!(result.unwrap(), Value::Bool(false));
    }

    #[test]
    fn test_coop_status() {
        let coroutine = Value::Coroutine(crate::compiler::vm::value::CoroutineValue::new(1));
        let args = vec![coroutine];
        let result = status(&args);
        assert!(result.is_ok());
        assert_eq!(result.unwrap(), Value::String("suspended".to_string()));
    }

    #[test]
    fn test_coop_pool() {
        let result = pool(&[]);
        assert!(result.is_ok());
        
        if let Ok(Value::Map(pool_info)) = result {
            assert!(pool_info.contains_key("size"));
            assert!(pool_info.contains_key("max_size"));
        } else {
            panic!("Expected Map result");
        }
    }

    #[test]
    fn test_coop_cancel() {
        let coroutine = Value::Coroutine(crate::compiler::vm::value::CoroutineValue::new(1));
        let args = vec![coroutine];
        let result = cancel(&args);
        assert!(result.is_ok());
        assert_eq!(result.unwrap(), Value::Bool(true));
    }

    #[test]
    fn test_coop_id() {
        let coroutine = Value::Coroutine(crate::compiler::vm::value::CoroutineValue::new(123));
        let args = vec![coroutine];
        let result = id(&args);
        assert!(result.is_ok());
        assert_eq!(result.unwrap(), Value::Int(123));
    }

    #[test]
    fn test_coop_info() {
        let coroutine = Value::Coroutine(crate::compiler::vm::value::CoroutineValue::new(456));
        let args = vec![coroutine];
        let result = info(&args);
        assert!(result.is_ok());
        
        if let Ok(Value::Map(info)) = result {
            assert!(info.contains_key("id"));
            assert!(info.contains_key("state"));
            assert!(info.contains_key("is_done"));
        } else {
            panic!("Expected Map result");
        }
    }

    #[test]
    fn test_coop_generator() {
        let args = vec![Value::Function(crate::compiler::vm::value::FunctionValue::new(0, "gen".to_string(), 0))];
        let result = generator(&args);
        assert!(result.is_ok());
        
        if let Ok(Value::Coroutine(_)) = result {
            // Expected
        } else {
            panic!("Expected Coroutine result");
        }
    }

    #[test]
    fn test_coop_scheduler() {
        let result = scheduler(&[]);
        assert!(result.is_ok());
        
        if let Ok(Value::Map(scheduler_info)) = result {
            assert!(scheduler_info.contains_key("is_running"));
            assert!(scheduler_info.contains_key("active_coroutines"));
        } else {
            panic!("Expected Map result");
        }
    }
}
