// Time Module - Zaman Fonksiyonları
// Zaman işlemleri ve bekleme

use crate::compiler::vm::{value::Value, RuntimeError};
use std::time::{SystemTime, UNIX_EPOCH};

/// Şu anki zamanı epoch milliseconds olarak döndür
pub fn now(args: &[Value]) -> Result<Value, RuntimeError> {
    if !args.is_empty() {
        return Err(RuntimeError::InvalidOperation {
            op: "now expects 0 arguments".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        });
    }

    let now = SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .map_err(|_| RuntimeError::InvalidOperation {
            op: "Failed to get current time".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        })?;

    let milliseconds = now.as_millis() as f64;
    Ok(Value::Float(milliseconds))
}

/// Sync sleep (bloklayıcı bekleme)
pub fn sleep(args: &[Value]) -> Result<Value, RuntimeError> {
    if args.len() != 1 {
        return Err(RuntimeError::InvalidOperation {
            op: "sleep expects 1 argument".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        });
    }

    let duration_ms = match &args[0] {
        Value::Int(ms) => *ms as u64,
        Value::Float(ms) => *ms as u64,
        _ => return Err(RuntimeError::InvalidOperation {
            op: "sleep expects int or float argument".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        }),
    };

    if duration_ms > 10000 { // 10 saniye limit
        return Err(RuntimeError::InvalidOperation {
            op: "sleep duration too long (max 10 seconds)".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        });
    }

    std::thread::sleep(std::time::Duration::from_millis(duration_ms));
    Ok(Value::Null)
}

/// Tarih ve saat bilgisini string olarak döndür
pub fn date_string(args: &[Value]) -> Result<Value, RuntimeError> {
    if !args.is_empty() {
        return Err(RuntimeError::InvalidOperation {
            op: "date_string expects 0 arguments".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        });
    }

    let now = SystemTime::now();
    let datetime = now.duration_since(UNIX_EPOCH)
        .map_err(|_| RuntimeError::InvalidOperation {
            op: "Failed to get current time".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        })?;

    // Basit tarih formatı
    let timestamp = datetime.as_secs();
    let date_str = format!("{}", timestamp); // Unix timestamp

    Ok(Value::String(date_str))
}

/// Timer oluştur (basit implementasyon)
pub fn timer(args: &[Value]) -> Result<Value, RuntimeError> {
    if args.len() != 1 {
        return Err(RuntimeError::InvalidOperation {
            op: "timer expects 1 argument".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        });
    }

    let duration_ms = match &args[0] {
        Value::Int(ms) => *ms as u64,
        Value::Float(ms) => *ms as u64,
        _ => return Err(RuntimeError::InvalidOperation {
            op: "timer expects int or float argument".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        }),
    };

    if duration_ms > 60000 { // 1 dakika limit
        return Err(RuntimeError::InvalidOperation {
            op: "timer duration too long (max 1 minute)".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        });
    }

    // Basit timer - sadece duration'ı döndür
    Ok(Value::Int(duration_ms as i64))
}

/// Benchmark fonksiyonu (basit timing)
pub fn benchmark(args: &[Value]) -> Result<Value, RuntimeError> {
    if args.len() != 1 {
        return Err(RuntimeError::InvalidOperation {
            op: "benchmark expects 1 argument".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        });
    }

    let iterations = match &args[0] {
        Value::Int(n) => {
            if *n <= 0 || *n > 1000000 {
                return Err(RuntimeError::InvalidOperation {
                    op: "benchmark iterations must be between 1 and 1000000".to_string(),
                    span: crate::compiler::diag::Span::new(0, 0, 0),
                });
            }
            *n as usize
        }
        _ => return Err(RuntimeError::InvalidOperation {
            op: "benchmark expects int argument".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        }),
    };

    let start = std::time::Instant::now();
    
    // Basit benchmark - sadece döngü
    for _ in 0..iterations {
        let _ = 1 + 1; // Dummy operation
    }
    
    let elapsed = start.elapsed();
    let elapsed_ms = elapsed.as_millis() as f64;
    
    Ok(Value::Float(elapsed_ms))
}

/// Zaman dilimi bilgisi
pub fn timezone(args: &[Value]) -> Result<Value, RuntimeError> {
    if !args.is_empty() {
        return Err(RuntimeError::InvalidOperation {
            op: "timezone expects 0 arguments".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        });
    }

    // Basit implementasyon - UTC döndür
    Ok(Value::String("UTC".to_string()))
}

/// Mikrosaniye cinsinden zaman
pub fn micros(args: &[Value]) -> Result<Value, RuntimeError> {
    if !args.is_empty() {
        return Err(RuntimeError::InvalidOperation {
            op: "micros expects 0 arguments".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        });
    }

    let now = SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .map_err(|_| RuntimeError::InvalidOperation {
            op: "Failed to get current time".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        })?;

    let microseconds = now.as_micros() as f64;
    Ok(Value::Float(microseconds))
}

/// Nanosaniye cinsinden zaman
pub fn nanos(args: &[Value]) -> Result<Value, RuntimeError> {
    if !args.is_empty() {
        return Err(RuntimeError::InvalidOperation {
            op: "nanos expects 0 arguments".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        });
    }

    let now = SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .map_err(|_| RuntimeError::InvalidOperation {
            op: "Failed to get current time".to_string(),
            span: crate::compiler::diag::Span::new(0, 0, 0),
        })?;

    let nanoseconds = now.as_nanos() as f64;
    Ok(Value::Float(nanoseconds))
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_now() {
        let result = now(&[]);
        assert!(result.is_ok());
        
        if let Ok(Value::Float(time)) = result {
            assert!(time > 0.0);
        } else {
            panic!("Expected float result");
        }
    }

    #[test]
    fn test_now_with_args() {
        let args = vec![Value::Int(42)];
        let result = now(&args);
        assert!(result.is_err());
    }

    #[test]
    fn test_sleep() {
        let args = vec![Value::Int(1)]; // 1ms
        let start = std::time::Instant::now();
        let result = sleep(&args);
        let elapsed = start.elapsed();
        
        assert!(result.is_ok());
        assert!(elapsed.as_millis() >= 1);
    }

    #[test]
    fn test_sleep_float() {
        let args = vec![Value::Float(1.5)];
        let result = sleep(&args);
        assert!(result.is_ok());
    }

    #[test]
    fn test_sleep_invalid_args() {
        let args = vec![Value::String("invalid".to_string())];
        let result = sleep(&args);
        assert!(result.is_err());
    }

    #[test]
    fn test_sleep_too_long() {
        let args = vec![Value::Int(15000)]; // 15 seconds
        let result = sleep(&args);
        assert!(result.is_err());
    }

    #[test]
    fn test_date_string() {
        let result = date_string(&[]);
        assert!(result.is_ok());
        
        if let Ok(Value::String(date)) = result {
            assert!(!date.is_empty());
        } else {
            panic!("Expected string result");
        }
    }

    #[test]
    fn test_timer() {
        let args = vec![Value::Int(1000)];
        let result = timer(&args);
        assert!(result.is_ok());
        assert_eq!(result.unwrap(), Value::Int(1000));
    }

    #[test]
    fn test_timer_too_long() {
        let args = vec![Value::Int(120000)]; // 2 minutes
        let result = timer(&args);
        assert!(result.is_err());
    }

    #[test]
    fn test_benchmark() {
        let args = vec![Value::Int(1000)];
        let result = benchmark(&args);
        assert!(result.is_ok());
        
        if let Ok(Value::Float(elapsed)) = result {
            assert!(elapsed >= 0.0);
        } else {
            panic!("Expected float result");
        }
    }

    #[test]
    fn test_benchmark_invalid_iterations() {
        let args = vec![Value::Int(0)];
        let result = benchmark(&args);
        assert!(result.is_err());
        
        let args = vec![Value::Int(2000000)];
        let result = benchmark(&args);
        assert!(result.is_err());
    }

    #[test]
    fn test_timezone() {
        let result = timezone(&[]);
        assert!(result.is_ok());
        assert_eq!(result.unwrap(), Value::String("UTC".to_string()));
    }

    #[test]
    fn test_micros() {
        let result = micros(&[]);
        assert!(result.is_ok());
        
        if let Ok(Value::Float(time)) = result {
            assert!(time > 0.0);
        } else {
            panic!("Expected float result");
        }
    }

    #[test]
    fn test_nanos() {
        let result = nanos(&[]);
        assert!(result.is_ok());
        
        if let Ok(Value::Float(time)) = result {
            assert!(time > 0.0);
        } else {
            panic!("Expected float result");
        }
    }
}
