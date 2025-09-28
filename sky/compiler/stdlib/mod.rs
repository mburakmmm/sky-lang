// Stdlib - Standart Kütüphane
// MVP için küçük ama kullanışlı bir çekirdek

pub mod io;
pub mod time;
pub mod http;
pub mod async_rt;
pub mod coop;
pub mod prelude;
pub mod strings;
pub mod python_bridge;
pub mod js_bridge;

use crate::compiler::vm::value::NativeFunction;
use std::collections::HashMap;

/// Stdlib modülü
pub struct Stdlib {
    pub modules: HashMap<String, Module>,
}

/// Modül
pub struct Module {
    pub name: String,
    pub functions: HashMap<String, NativeFunction>,
}

impl Stdlib {
    pub fn new() -> Self {
        let mut stdlib = Self {
            modules: HashMap::new(),
        };
        
        stdlib.register_modules();
        stdlib
    }

    fn register_modules(&mut self) {
        // IO modülü
        let mut io_module = Module {
            name: "io".to_string(),
            functions: HashMap::new(),
        };
        
        io_module.functions.insert("print".to_string(), 
            NativeFunction::new("print".to_string(), 1, Box::new(io::print)));
        
        self.modules.insert("io".to_string(), io_module);

        // Time modülü
        let mut time_module = Module {
            name: "time".to_string(),
            functions: HashMap::new(),
        };
        
        time_module.functions.insert("now".to_string(),
            NativeFunction::new("now".to_string(), 0, Box::new(time::now)));
        time_module.functions.insert("sleep".to_string(),
            NativeFunction::new("sleep".to_string(), 1, Box::new(time::sleep)));
        
        self.modules.insert("time".to_string(), time_module);

        // HTTP modülü
        let mut http_module = Module {
            name: "http".to_string(),
            functions: HashMap::new(),
        };
        
        http_module.functions.insert("get".to_string(),
            NativeFunction::new("get".to_string(), 1, Box::new(http::get)));
        
        self.modules.insert("http".to_string(), http_module);

        // Async modülü
        let mut async_module = Module {
            name: "async".to_string(),
            functions: HashMap::new(),
        };
        
        async_module.functions.insert("sleep".to_string(),
            NativeFunction::new("sleep".to_string(), 1, Box::new(async_rt::sleep)));
        
        self.modules.insert("async".to_string(), async_module);

        // Coop modülü
        let mut coop_module = Module {
            name: "coop".to_string(),
            functions: HashMap::new(),
        };
        
        coop_module.functions.insert("new".to_string(),
            NativeFunction::new("new".to_string(), 1, Box::new(coop::new)));
        coop_module.functions.insert("resume".to_string(),
            NativeFunction::new("resume".to_string(), 1, Box::new(coop::resume)));
        coop_module.functions.insert("is_done".to_string(),
            NativeFunction::new("is_done".to_string(), 1, Box::new(coop::is_done)));
        
        self.modules.insert("coop".to_string(), coop_module);
        
        // String modülü
        let mut string_module = Module {
            name: "string".to_string(),
            functions: HashMap::new(),
        };
        
        string_module.functions.insert("stringify".to_string(),
            NativeFunction::new("stringify".to_string(), 1, Box::new(strings::interpolate_string)));
        string_module.functions.insert("concat".to_string(),
            NativeFunction::new("concat".to_string(), 255, Box::new(strings::concat_strings)));
        string_module.functions.insert("length".to_string(),
            NativeFunction::new("length".to_string(), 1, Box::new(strings::string_length)));
        string_module.functions.insert("split".to_string(),
            NativeFunction::new("split".to_string(), 2, Box::new(strings::string_split)));
        string_module.functions.insert("join".to_string(),
            NativeFunction::new("join".to_string(), 2, Box::new(strings::string_join)));
        string_module.functions.insert("upper".to_string(),
            NativeFunction::new("upper".to_string(), 1, Box::new(strings::string_upper)));
        string_module.functions.insert("lower".to_string(),
            NativeFunction::new("lower".to_string(), 1, Box::new(strings::string_lower)));
        string_module.functions.insert("trim".to_string(),
            NativeFunction::new("trim".to_string(), 1, Box::new(strings::string_trim)));
        string_module.functions.insert("contains".to_string(),
            NativeFunction::new("contains".to_string(), 2, Box::new(strings::string_contains)));
        string_module.functions.insert("startsWith".to_string(),
            NativeFunction::new("startsWith".to_string(), 2, Box::new(strings::string_starts_with)));
        string_module.functions.insert("endsWith".to_string(),
            NativeFunction::new("endsWith".to_string(), 2, Box::new(strings::string_ends_with)));
        
        self.modules.insert("string".to_string(), string_module);
        
        // Python bridge modülü
        let mut python_module = Module {
            name: "python".to_string(),
            functions: HashMap::new(),
        };
        
        python_module.functions.insert("import".to_string(),
            NativeFunction::new("import".to_string(), 1, Box::new(python_bridge::import_module)));
        
        self.modules.insert("python".to_string(), python_module);
        
        // JS bridge modülü
        let mut js_module = Module {
            name: "js".to_string(),
            functions: HashMap::new(),
        };
        
        js_module.functions.insert("eval".to_string(),
            NativeFunction::new("eval".to_string(), 1, Box::new(js_bridge::eval_code)));
        
        self.modules.insert("js".to_string(), js_module);
    }

    /// Modülü al
    pub fn get_module(&self, name: &str) -> Option<&Module> {
        self.modules.get(name)
    }

    /// Fonksiyonu al
    pub fn get_function(&self, module_name: &str, function_name: &str) -> Option<&NativeFunction> {
        self.modules.get(module_name)?.functions.get(function_name)
    }

    /// Tüm modülleri al
    pub fn all_modules(&self) -> &HashMap<String, Module> {
        &self.modules
    }

    /// Modül sayısını al
    pub fn module_count(&self) -> usize {
        self.modules.len()
    }

    /// Belirli bir modüldeki fonksiyon sayısını al
    pub fn function_count(&self, module_name: &str) -> Option<usize> {
        self.modules.get(module_name).map(|m| m.functions.len())
    }

    /// Toplam fonksiyon sayısını al
    pub fn total_function_count(&self) -> usize {
        self.modules.values().map(|m| m.functions.len()).sum()
    }
}

impl Module {
    pub fn new(name: String) -> Self {
        Self {
            name,
            functions: HashMap::new(),
        }
    }

    pub fn add_function(&mut self, name: String, function: NativeFunction) {
        self.functions.insert(name, function);
    }

    pub fn get_function(&self, name: &str) -> Option<&NativeFunction> {
        self.functions.get(name)
    }

    pub fn function_names(&self) -> Vec<&String> {
        self.functions.keys().collect()
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_stdlib_creation() {
        let stdlib = Stdlib::new();
        assert!(stdlib.module_count() > 0);
        assert!(stdlib.total_function_count() > 0);
    }

    #[test]
    fn test_module_access() {
        let stdlib = Stdlib::new();
        
        assert!(stdlib.get_module("io").is_some());
        assert!(stdlib.get_module("time").is_some());
        assert!(stdlib.get_module("http").is_some());
        assert!(stdlib.get_module("nonexistent").is_none());
    }

    #[test]
    fn test_function_access() {
        let stdlib = Stdlib::new();
        
        assert!(stdlib.get_function("io", "print").is_some());
        assert!(stdlib.get_function("time", "now").is_some());
        assert!(stdlib.get_function("http", "get").is_some());
        assert!(stdlib.get_function("io", "nonexistent").is_none());
        assert!(stdlib.get_function("nonexistent", "print").is_none());
    }

    #[test]
    fn test_module_creation() {
        let mut module = Module::new("test".to_string());
        assert_eq!(module.name, "test");
        assert_eq!(module.functions.len(), 0);
        
        let func = NativeFunction::new("test_func".to_string(), 0, |_| Ok(Value::Null));
        module.add_function("test_func".to_string(), func);
        
        assert_eq!(module.functions.len(), 1);
        assert!(module.get_function("test_func").is_some());
    }
}
