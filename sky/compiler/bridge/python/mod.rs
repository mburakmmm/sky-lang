// Python Bridge - PyO3 ile Python entegrasyonu
// CPython'a gömülü çağrılar

use crate::compiler::vm::{value::Value, value::NativeFunction, RuntimeError};
use std::collections::HashMap;
use pyo3::prelude::*;
use pyo3::types::{PyList, PyDict};

/// Python bridge yöneticisi
pub struct PythonBridge {
    pub imported_modules: HashMap<String, PyModule>,
    pub call_count: u64,
}

/// Python modülü
#[derive(Debug, Clone)]
pub struct PyModule {
    pub name: String,
}

/// Python bridge istatistikleri
#[derive(Debug, Clone)]
pub struct PythonStats {
    pub modules_imported: usize,
    pub total_calls: u64,
}

impl PythonBridge {
    pub fn new() -> Result<Self, RuntimeError> {
        Ok(Self {
            imported_modules: HashMap::new(),
            call_count: 0,
        })
    }

    /// Python modülünü import et
    pub fn import_module(&mut self, module_name: &str) -> Result<Value, RuntimeError> {
        self.call_count += 1;

        // Python runtime'ını initialize et (gerekirse)
        pyo3::prepare_freethreaded_python();

        // Gerçek Python import işlemi
        Python::with_gil(|py| {
            match module_name {
                "math" => {
                    let math = py.import_bound("math").map_err(|e| RuntimeError::InvalidOperation {
                        op: format!("Failed to import math: {}", e),
                        span: crate::compiler::diag::Span::new(0, 0, 0),
                    })?;
                    let mut math_module = HashMap::new();
                    
                    // sqrt fonksiyonu - kalıcı çözüm
                    let sqrt_fn = math.getattr("sqrt")?;
                    math_module.insert("sqrt".to_string(), Value::NativeFn(NativeFunction::new(
                        "sqrt".to_string(), 1, Box::new(move |args: &[Value]| {
                            if args.len() != 1 {
                                return Err(RuntimeError::InvalidOperation {
                                    op: "sqrt expects 1 argument".to_string(),
                                    span: crate::compiler::diag::Span::new(0, 0, 0),
                                });
                            }

                            let num = match &args[0] {
                                Value::Int(i) => *i as f64,
                                Value::Float(f) => *f,
                                _ => return Err(RuntimeError::InvalidOperation {
                                    op: "sqrt argument must be number".to_string(),
                                    span: crate::compiler::diag::Span::new(0, 0, 0),
                                }),
                            };
                            
                            // Her çağrıda fresh Python context oluştur
                            Python::with_gil(|py| {
                                let math = py.import_bound("math")?;
                                let sqrt_fn = math.getattr("sqrt")?;
                                let result = sqrt_fn.call1((num,))?;
                                let result_value: f64 = result.extract()?;
                                Ok(Value::Float(result_value))
                            }).map_err(|e: pyo3::PyErr| RuntimeError::InvalidOperation {
                                op: format!("Python sqrt error: {}", e),
                                span: crate::compiler::diag::Span::new(0, 0, 0),
                            })
                        })
                    )));
                    
                    // pi ve e constants
                    let pi: f64 = math.getattr("pi")?.extract()?;
                    let e: f64 = math.getattr("e")?.extract()?;
                    math_module.insert("pi".to_string(), Value::Float(pi));
                    math_module.insert("e".to_string(), Value::Float(e));
                    
                    let module = PyModule {
                        name: module_name.to_string(),
                    };
                    self.imported_modules.insert(module_name.to_string(), module);
                    
                    Ok(Value::Map(math_module))
                },
                "sys" => {
                    let sys = py.import_bound("sys")?;
                    let mut sys_module = HashMap::new();
                    
                    let version: String = sys.getattr("version")?.extract()?;
                    let platform: String = sys.getattr("platform")?.extract()?;
                    
                    sys_module.insert("version".to_string(), Value::String(version));
                    sys_module.insert("platform".to_string(), Value::String(platform));
                    
                    let module = PyModule {
                        name: module_name.to_string(),
                    };
                    self.imported_modules.insert(module_name.to_string(), module);
                    
                    Ok(Value::Map(sys_module))
                },
                "json" => {
                    let json = py.import_bound("json")?;
                    let mut json_module = HashMap::new();
                    
                    // json.dumps fonksiyonu - kalıcı çözüm
                    let dumps_fn = json.getattr("dumps")?;
                    json_module.insert("dumps".to_string(), Value::NativeFn(NativeFunction::new(
                        "dumps".to_string(), 1, Box::new(move |args: &[Value]| {
                            if args.len() != 1 {
                                return Err(RuntimeError::InvalidOperation {
                                    op: "json.dumps expects 1 argument".to_string(),
                                    span: crate::compiler::diag::Span::new(0, 0, 0),
                                });
                            }
                            
                            // Her çağrıda fresh Python context oluştur
                            Python::with_gil(|py| {
                                let json = py.import_bound("json")?;
                                let dumps_fn = json.getattr("dumps")?;
                                let sky_value = &args[0];
                                let py_value = sky_value_to_python(py, sky_value)?;
                                let result = dumps_fn.call1((py_value,))?;
                                let result_str: String = result.extract()?;
                                Ok(Value::String(result_str))
                            }).map_err(|e: pyo3::PyErr| RuntimeError::InvalidOperation {
                                op: format!("Python json.dumps error: {}", e),
                                span: crate::compiler::diag::Span::new(0, 0, 0),
                            })
                        })
                    )));
                    
                    let module = PyModule {
                        name: module_name.to_string(),
                    };
                    self.imported_modules.insert(module_name.to_string(), module);
                    
                    Ok(Value::Map(json_module))
                },
                _ => {
                    // Genel modül import
                    match py.import_bound(module_name) {
                        Ok(_module) => {
                            let module = PyModule {
                                name: module_name.to_string(),
                            };
                            self.imported_modules.insert(module_name.to_string(), module);
                            
                            Ok(Value::Map(HashMap::new()))
                        },
                        Err(e) => Err(RuntimeError::InvalidOperation {
                            op: format!("Failed to import module '{}': {}", module_name, e),
                            span: crate::compiler::diag::Span::new(0, 0, 0),
                        })
                    }
                }
            }
        })
    }

    /// Python fonksiyonunu çağır
    pub fn call_function(&mut self, module_name: &str, func_name: &str, args: &[Value]) -> Result<Value, RuntimeError> {
        self.call_count += 1;
        
        Ok(Python::with_gil(|py| {
            // Modülü import et
            let module = py.import_bound(module_name)?;
            let func = module.getattr(func_name)?;
            
            // Sky değerlerini Python değerlerine çevir
            let py_args: Vec<PyObject> = args.iter()
                .map(|arg| sky_value_to_python(py, arg))
                .collect::<Result<Vec<_>, _>>()?;
            
            // Fonksiyonu çağır
            let result = func.call1((py_args,))?;
            
            // Sonucu Sky değerine çevir
            python_value_to_sky(result.into())
        })?)
    }

    /// Bridge'i temizle
    pub fn cleanup(&mut self) {
        self.imported_modules.clear();
    }

    /// Bridge sağlığını kontrol et
    pub fn is_healthy(&self) -> bool {
        Python::with_gil(|_py| {
            Ok::<bool, PyErr>(true)
        }).unwrap_or(false)
    }

    /// Bridge istatistiklerini al
    pub fn stats(&self) -> PythonStats {
        PythonStats {
            modules_imported: self.imported_modules.len(),
            total_calls: self.call_count,
        }
    }
}

/// Sky değerini Python değerine çevir
fn sky_value_to_python(py: Python<'_>, value: &Value) -> PyResult<PyObject> {
    match value {
        Value::Int(i) => Ok(i.into_py(py)),
        Value::Float(f) => Ok(f.into_py(py)),
        Value::Bool(b) => Ok(b.into_py(py)),
        Value::String(s) => Ok(s.into_py(py)),
        Value::Null => Ok(py.None()),
        Value::List(list) => {
            let py_list = PyList::new_bound(py, [] as [&str; 0]);
            for item in list {
                let py_item = sky_value_to_python(py, item)?;
                py_list.append(py_item)?;
            }
            Ok(py_list.into())
        },
        Value::Map(map) => {
            let py_dict = PyDict::new_bound(py);
            for (key, val) in map {
                let py_val = sky_value_to_python(py, val)?;
                py_dict.set_item(key, py_val)?;
            }
            Ok(py_dict.into())
        },
        _ => Err(PyErr::new::<pyo3::exceptions::PyTypeError, _>("Unsupported value type"))
    }
}

/// Python değerini Sky değerine çevir
fn python_value_to_sky(value: PyObject) -> PyResult<Value> {
    Python::with_gil(|py| {
        let value = value.bind(py);
        if value.is_instance_of::<pyo3::types::PyInt>() {
            let i: i64 = value.extract()?;
            Ok(Value::Int(i))
        } else if value.is_instance_of::<pyo3::types::PyFloat>() {
            let f: f64 = value.extract()?;
            Ok(Value::Float(f))
        } else if value.is_instance_of::<pyo3::types::PyBool>() {
            let b: bool = value.extract()?;
            Ok(Value::Bool(b))
        } else if value.is_instance_of::<pyo3::types::PyString>() {
            let s: String = value.extract()?;
            Ok(Value::String(s))
        } else if value.is_none() {
            Ok(Value::Null)
        } else if value.is_instance_of::<pyo3::types::PyList>() {
            let list = value.downcast::<pyo3::types::PyList>()?;
            let mut sky_list = Vec::new();
            for item in list {
                let sky_item = python_value_to_sky(item.into())?;
                sky_list.push(sky_item);
            }
            Ok(Value::List(sky_list))
        } else if value.is_instance_of::<pyo3::types::PyDict>() {
            let dict = value.downcast::<pyo3::types::PyDict>()?;
            let mut sky_map = HashMap::new();
            for (key, val) in dict {
                let key_str: String = key.extract()?;
                let sky_val = python_value_to_sky(val.into())?;
                sky_map.insert(key_str, sky_val);
            }
            Ok(Value::Map(sky_map))
        } else {
            Err(PyErr::new::<pyo3::exceptions::PyTypeError, _>("Unsupported Python type"))
        }
    })
}

/// Python modülünü import et (public API)
pub fn python_import(module_name: &str) -> Result<Value, RuntimeError> {
    let mut bridge = PythonBridge::new()?;
    bridge.import_module(module_name)
}

/// Python fonksiyonunu çağır (public API)
pub fn python_call(module_name: &str, func_name: &str, args: &[Value]) -> Result<Value, RuntimeError> {
    let mut bridge = PythonBridge::new()?;
    bridge.call_function(module_name, func_name, args)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_python_bridge_creation() {
        let bridge = PythonBridge::new();
        assert!(bridge.is_ok());
    }

    #[test]
    fn test_python_bridge_stats() {
        let bridge = PythonBridge::new().unwrap();
        let stats = bridge.stats();
        assert_eq!(stats.modules_imported, 0);
        assert_eq!(stats.total_calls, 0);
    }
}