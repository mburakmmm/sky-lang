// Async Runtime - Async İşlemler
// Async/await semantiği için runtime

use crate::compiler::vm::{value::Value, RuntimeError};
use std::collections::HashMap;

/// Async sleep (non-blocking)
pub fn sleep(args: &[Value]) -> Result<Value, RuntimeError> {
    if args.len() != 1 {
        return Err(RuntimeError::InvalidOperation {
            op: "async.sleep expects 1 argument".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        });
    }

    let duration_ms = match &args[0] {
        Value::Int(ms) => *ms as u64,
        Value::Float(ms) => *ms as u64,
        _ => return Err(RuntimeError::InvalidOperation {
            op: "async.sleep expects int or float argument".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        }),
    };

    if duration_ms > 10000 { // 10 saniye limit
        return Err(RuntimeError::InvalidOperation {
            op: "async.sleep duration too long (max 10 seconds)".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        });
    }

    // Future oluştur (async işlem simülasyonu)
    let future: Value = Value::Future(crate::compiler::vm::value::FutureValue::new(1, "sleep".to_string())); // Mock ID sistemi
    
    Ok(future)
}

/// Async delay (sleep ile aynı)
pub fn delay(args: &[Value]) -> Result<Value, RuntimeError> {
    sleep(args)
}

/// Future oluştur
pub fn future(args: &[Value]) -> Result<Value, RuntimeError> {
    if args.len() != 1 {
        return Err(RuntimeError::InvalidOperation {
            op: "async.future expects 1 argument".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        });
    }

    // Future oluştur
    let future: Value = Value::Future(crate::compiler::vm::value::FutureValue::new(2, "delay".to_string())); // Mock ID sistemi
    
    Ok(future)
}

/// Promise oluştur (future ile aynı)
pub fn promise(args: &[Value]) -> Result<Value, RuntimeError> {
    future(args)
}

/// Async task oluştur
pub fn task(args: &[Value]) -> Result<Value, RuntimeError> {
    if args.len() != 1 {
        return Err(RuntimeError::InvalidOperation {
            op: "async.task expects 1 argument".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        });
    }

    // Task oluştur (basit implementasyon)
    let task_id = match &args[0] {
        Value::String(name) => name.clone(),
        _ => "anonymous".to_string(),
    };

    let mut task_info = HashMap::new();
    task_info.insert("id".to_string(), Value::String(task_id));
    task_info.insert("status".to_string(), Value::String("pending".to_string()));
    task_info.insert("result".to_string(), Value::Null);

    Ok(Value::Map(task_info))
}

/// Event loop bilgisi
pub fn event_loop_info(args: &[Value]) -> Result<Value, RuntimeError> {
    if !args.is_empty() {
        return Err(RuntimeError::InvalidOperation {
            op: "async.event_loop_info expects 0 arguments".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        });
    }

    let mut info = HashMap::new();
    info.insert("is_running".to_string(), Value::Bool(true));
    info.insert("pending_tasks".to_string(), Value::Int(0));
    info.insert("completed_tasks".to_string(), Value::Int(0));
    info.insert("failed_tasks".to_string(), Value::Int(0));

    Ok(Value::Map(info))
}

/// Async timeout oluştur
pub fn timeout(args: &[Value]) -> Result<Value, RuntimeError> {
    if args.len() != 2 {
        return Err(RuntimeError::InvalidOperation {
            op: "async.timeout expects 2 arguments (duration, callback)".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        });
    }

    let duration_ms = match &args[0] {
        Value::Int(ms) => *ms as u64,
        Value::Float(ms) => *ms as u64,
        _ => return Err(RuntimeError::InvalidOperation {
            op: "async.timeout first argument must be int or float".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        }),
    };

    if duration_ms > 60000 { // 1 dakika limit
        return Err(RuntimeError::InvalidOperation {
            op: "async.timeout duration too long (max 1 minute)".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        });
    }

    // Timeout oluştur
    let mut timeout_info = HashMap::new();
    timeout_info.insert("duration_ms".to_string(), Value::Int(duration_ms as i64));
    timeout_info.insert("callback".to_string(), args[1].clone());
    timeout_info.insert("status".to_string(), Value::String("pending".to_string()));

    Ok(Value::Map(timeout_info))
}

/// Async interval oluştur
pub fn interval(args: &[Value]) -> Result<Value, RuntimeError> {
    if args.len() != 2 {
        return Err(RuntimeError::InvalidOperation {
            op: "async.interval expects 2 arguments (duration, callback)".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        });
    }

    let interval_ms = match &args[0] {
        Value::Int(ms) => *ms as u64,
        Value::Float(ms) => *ms as u64,
        _ => return Err(RuntimeError::InvalidOperation {
            op: "async.interval first argument must be int or float".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        }),
    };

    if interval_ms < 100 { // Minimum 100ms
        return Err(RuntimeError::InvalidOperation {
            op: "async.interval duration too short (min 100ms)".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        });
    }

    if interval_ms > 3600000 { // Maximum 1 saat
        return Err(RuntimeError::InvalidOperation {
            op: "async.interval duration too long (max 1 hour)".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        });
    }

    // Interval oluştur
    let mut interval_info = HashMap::new();
    interval_info.insert("interval_ms".to_string(), Value::Int(interval_ms as i64));
    interval_info.insert("callback".to_string(), args[1].clone());
    interval_info.insert("status".to_string(), Value::String("active".to_string()));
    interval_info.insert("count".to_string(), Value::Int(0));

    Ok(Value::Map(interval_info))
}

/// Async debounce oluştur
pub fn debounce(args: &[Value]) -> Result<Value, RuntimeError> {
    if args.len() != 2 {
        return Err(RuntimeError::InvalidOperation {
            op: "async.debounce expects 2 arguments (delay, callback)".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        });
    }

    let delay_ms = match &args[0] {
        Value::Int(ms) => *ms as u64,
        Value::Float(ms) => *ms as u64,
        _ => return Err(RuntimeError::InvalidOperation {
            op: "async.debounce first argument must be int or float".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        }),
    };

    if delay_ms > 5000 { // Maximum 5 saniye
        return Err(RuntimeError::InvalidOperation {
            op: "async.debounce delay too long (max 5 seconds)".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        });
    }

    // Debounce oluştur
    let mut debounce_info = HashMap::new();
    debounce_info.insert("delay_ms".to_string(), Value::Int(delay_ms as i64));
    debounce_info.insert("callback".to_string(), args[1].clone());
    debounce_info.insert("status".to_string(), Value::String("ready".to_string()));
    debounce_info.insert("last_call".to_string(), Value::Int(0));

    Ok(Value::Map(debounce_info))
}

/// Async throttle oluştur
pub fn throttle(args: &[Value]) -> Result<Value, RuntimeError> {
    if args.len() != 2 {
        return Err(RuntimeError::InvalidOperation {
            op: "async.throttle expects 2 arguments (limit, callback)".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        });
    }

    let limit_ms = match &args[0] {
        Value::Int(ms) => *ms as u64,
        Value::Float(ms) => *ms as u64,
        _ => return Err(RuntimeError::InvalidOperation {
            op: "async.throttle first argument must be int or float".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        }),
    };

    if limit_ms < 10 { // Minimum 10ms
        return Err(RuntimeError::InvalidOperation {
            op: "async.throttle limit too short (min 10ms)".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        });
    }

    // Throttle oluştur
    let mut throttle_info = HashMap::new();
    throttle_info.insert("limit_ms".to_string(), Value::Int(limit_ms as i64));
    throttle_info.insert("callback".to_string(), args[1].clone());
    throttle_info.insert("status".to_string(), Value::String("ready".to_string()));
    throttle_info.insert("last_execution".to_string(), Value::Int(0));

    Ok(Value::Map(throttle_info))
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_async_sleep() {
        let args = vec![Value::Int(1000)];
        let result = sleep(&args);
        assert!(result.is_ok());
        
        if let Ok(Value::Future(_)) = result {
            // Expected
        } else {
            panic!("Expected Future result");
        }
    }

    #[test]
    fn test_async_sleep_invalid_args() {
        let args = vec![Value::String("invalid".to_string())];
        let result = sleep(&args);
        assert!(result.is_err());
    }

    #[test]
    fn test_async_sleep_too_long() {
        let args = vec![Value::Int(15000)]; // 15 seconds
        let result = sleep(&args);
        assert!(result.is_err());
    }

    #[test]
    fn test_async_delay() {
        let args = vec![Value::Int(500)];
        let result = delay(&args);
        assert!(result.is_ok());
    }

    #[test]
    fn test_async_future() {
        let args = vec![Value::String("test".to_string())];
        let result = future(&args);
        assert!(result.is_ok());
    }

    #[test]
    fn test_async_task() {
        let args = vec![Value::String("test_task".to_string())];
        let result = task(&args);
        assert!(result.is_ok());
        
        if let Ok(Value::Map(task_info)) = result {
            assert!(task_info.contains_key("id"));
            assert!(task_info.contains_key("status"));
        } else {
            panic!("Expected Map result");
        }
    }

    #[test]
    fn test_event_loop_info() {
        let result = event_loop_info(&[]);
        assert!(result.is_ok());
        
        if let Ok(Value::Map(info)) = result {
            assert!(info.contains_key("is_running"));
            assert!(info.contains_key("pending_tasks"));
        } else {
            panic!("Expected Map result");
        }
    }

    #[test]
    fn test_timeout() {
        let args = vec![
            Value::Int(1000),
            Value::String("callback".to_string()),
        ];
        let result = timeout(&args);
        assert!(result.is_ok());
    }

    #[test]
    fn test_timeout_too_long() {
        let args = vec![
            Value::Int(120000), // 2 minutes
            Value::String("callback".to_string()),
        ];
        let result = timeout(&args);
        assert!(result.is_err());
    }

    #[test]
    fn test_interval() {
        let args = vec![
            Value::Int(1000),
            Value::String("callback".to_string()),
        ];
        let result = interval(&args);
        assert!(result.is_ok());
    }

    #[test]
    fn test_interval_too_short() {
        let args = vec![
            Value::Int(50), // Too short
            Value::String("callback".to_string()),
        ];
        let result = interval(&args);
        assert!(result.is_err());
    }

    #[test]
    fn test_debounce() {
        let args = vec![
            Value::Int(500),
            Value::String("callback".to_string()),
        ];
        let result = debounce(&args);
        assert!(result.is_ok());
    }

    #[test]
    fn test_throttle() {
        let args = vec![
            Value::Int(100),
            Value::String("callback".to_string()),
        ];
        let result = throttle(&args);
        assert!(result.is_ok());
    }

    #[test]
    fn test_throttle_too_short() {
        let args = vec![
            Value::Int(5), // Too short
            Value::String("callback".to_string()),
        ];
        let result = throttle(&args);
        assert!(result.is_err());
    }
}
