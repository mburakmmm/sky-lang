// Bridge Modules - Python ve JS Köprüleri
// Native dil entegrasyonu

pub mod python;
pub mod js;

use crate::compiler::vm::{value::Value, RuntimeError, value::NativeFunction};
use std::collections::HashMap;

/// Bridge yöneticisi
pub struct BridgeManager {
    pub python_bridge: python::PythonBridge,
    pub js_bridge: js::JSBridge,
}

impl BridgeManager {
    pub fn new() -> Result<Self, RuntimeError> {
        Ok(Self {
            python_bridge: python::PythonBridge::new()?,
            js_bridge: js::JSBridge::new()?,
        })
    }

    /// Python modülünü import et
    pub fn python_import(&mut self, module_name: &str) -> Result<Value, RuntimeError> {
        self.python_bridge.import_module(module_name)
    }

    /// JS kodu eval et
    pub fn js_eval(&mut self, code: &str) -> Result<Value, RuntimeError> {
        self.js_bridge.eval_code(code)
    }

    /// Bridge'leri temizle
    pub fn cleanup(&mut self) {
        self.python_bridge.cleanup();
        self.js_bridge.cleanup();
    }

    /// Bridge durumunu kontrol et
    pub fn is_healthy(&self) -> bool {
        self.python_bridge.is_healthy() && self.js_bridge.is_healthy()
    }

    /// Bridge istatistiklerini al
    pub fn stats(&self) -> BridgeStats {
        BridgeStats {
            python_stats: self.python_bridge.stats(),
            js_stats: self.js_bridge.stats(),
        }
    }
}

/// Bridge istatistikleri
#[derive(Debug, Clone)]
pub struct BridgeStats {
    pub python_stats: python::PythonStats,
    pub js_stats: js::JSStats,
}

/// Bridge fonksiyonlarını döndür
pub fn get_bridge_functions() -> HashMap<String, NativeFunction> {
    let mut functions = HashMap::new();

    // Python bridge fonksiyonları
    functions.insert("python_import".to_string(), 
        NativeFunction::new("python_import".to_string(), 1, Box::new(|args| {
            if args.len() != 1 {
                return Err(RuntimeError::InvalidOperation {
                    op: "python_import expects 1 argument".to_string(),
                    span: crate::compiler::diag::Span::new(0, 0, 0),
                });
            }
            if let Value::String(module_name) = &args[0] {
                python::python_import(module_name)
            } else {
                Err(RuntimeError::TypeMismatch {
                    expected: "string".to_string(),
                    actual: format!("{:?}", args[0]),
                    span: crate::compiler::diag::Span::new(0, 0, 0),
                })
            }
        })));
    
    functions.insert("python_call".to_string(), 
        NativeFunction::new("python_call".to_string(), 3, Box::new(|args| {
            if args.len() != 3 {
                return Err(RuntimeError::InvalidOperation {
                    op: "python_call expects 3 arguments".to_string(),
                    span: crate::compiler::diag::Span::new(0, 0, 0),
                });
            }
            let module_name = if let Value::String(s) = &args[0] { s } else { 
                return Err(RuntimeError::TypeMismatch {
                    expected: "string".to_string(),
                    actual: format!("{:?}", args[0]),
                    span: crate::compiler::diag::Span::new(0, 0, 0),
                });
            };
            let func_name = if let Value::String(s) = &args[1] { s } else { 
                return Err(RuntimeError::TypeMismatch {
                    expected: "string".to_string(),
                    actual: format!("{:?}", args[1]),
                    span: crate::compiler::diag::Span::new(0, 0, 0),
                });
            };
            let func_args = if let Value::List(l) = &args[2] { l } else { 
                return Err(RuntimeError::TypeMismatch {
                    expected: "list".to_string(),
                    actual: format!("{:?}", args[2]),
                    span: crate::compiler::diag::Span::new(0, 0, 0),
                });
            };
            python::python_call(module_name, func_name, func_args)
        })));

    // JS bridge fonksiyonları
    functions.insert("js_eval".to_string(), 
        NativeFunction::new("js_eval".to_string(), 1, Box::new(|args| {
            if args.len() != 1 {
                return Err(RuntimeError::InvalidOperation {
                    op: "js_eval expects 1 argument".to_string(),
                    span: crate::compiler::diag::Span::new(0, 0, 0),
                });
            }
            if let Value::String(code) = &args[0] {
                js::eval_js_code(code)
            } else {
                Err(RuntimeError::TypeMismatch {
                    expected: "string".to_string(),
                    actual: format!("{:?}", args[0]),
                    span: crate::compiler::diag::Span::new(0, 0, 0),
                })
            }
        })));
    
    functions.insert("js_call".to_string(), 
        NativeFunction::new("js_call".to_string(), 3, Box::new(|args| {
            if args.len() != 2 {
                return Err(RuntimeError::InvalidOperation {
                    op: "js_call expects 2 arguments".to_string(),
                    span: crate::compiler::diag::Span::new(0, 0, 0),
                });
            }
            let func_name = if let Value::String(s) = &args[0] { s } else { 
                return Err(RuntimeError::TypeMismatch {
                    expected: "string".to_string(),
                    actual: format!("{:?}", args[0]),
                    span: crate::compiler::diag::Span::new(0, 0, 0),
                });
            };
            let func_args = if let Value::List(l) = &args[1] { l } else { 
                return Err(RuntimeError::TypeMismatch {
                    expected: "list".to_string(),
                    actual: format!("{:?}", args[1]),
                    span: crate::compiler::diag::Span::new(0, 0, 0),
                });
            };
            js::call_js_function(func_name, func_args)
        })));

    functions
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_bridge_manager_creation() {
        // Bridge manager oluşturma testi
        // Gerçek implementasyonda Python/JS bağımlılıkları olacağı için
        // şimdilik sadece struct oluşturma test edilebilir
    }

    #[test]
    fn test_bridge_functions() {
        let functions = get_bridge_functions();
        assert!(functions.contains_key("python_import"));
        assert!(functions.contains_key("python_call"));
        assert!(functions.contains_key("js_eval"));
        assert!(functions.contains_key("js_call"));
    }
}
