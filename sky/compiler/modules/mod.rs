// Modules - Import ve Modül Sistemi
// Çoklu dosya desteği, modül çözümleme, cache

pub mod resolver;
pub mod cache;

use std::collections::HashMap;
use std::path::PathBuf;
use crate::compiler::diag::Span;
use crate::compiler::vm::Value;

/// Modül sistemi
pub struct ModuleSystem {
    /// Yüklenmiş modüllerin cache'i
    pub modules: HashMap<String, Module>,
    /// Modül çözümleme yolları
    pub search_paths: Vec<PathBuf>,
}

/// Bir modül
#[derive(Debug, Clone)]
pub struct Module {
    /// Modül adı (örn. "utils", "io.file")
    pub name: String,
    /// Modül dosya yolu
    pub file_path: PathBuf,
    /// Public semboller (export edilenler)
    pub exports: HashMap<String, ExportSymbol>,
    /// Modül global'leri
    pub globals: HashMap<String, Value>,
    /// Modül derlenmiş mi?
    pub compiled: bool,
}

/// Export edilen sembol
#[derive(Debug, Clone)]
pub struct ExportSymbol {
    /// Sembol adı
    pub name: String,
    /// Sembol tipi (function, variable)
    pub kind: SymbolKind,
    /// Sembol değeri
    pub value: Value,
    /// Sembol span'i
    pub span: Span,
}

#[derive(Debug, Clone)]
pub enum SymbolKind {
    Function,
    Variable,
}

impl ModuleSystem {
    pub fn new() -> Self {
        let mut system = Self {
            modules: HashMap::new(),
            search_paths: Vec::new(),
        };
        
        // Varsayılan arama yolları
        system.search_paths.push(PathBuf::from(".")); // Çalışan dosya klasörü
        system.search_paths.push(PathBuf::from("./stdlib")); // Stdlib
        
        system
    }
    
    /// Modül import et
    pub fn import_module(&mut self, name: &str, current_file: &PathBuf) -> Result<(), String> {
        // Cache'de var mı kontrol et
        if self.modules.contains_key(name) {
            return Ok(());
        }
        
        // Modül dosyasını bul
        let module_path = self.resolve_module_path(name, current_file)?;
        
        // Modülü yükle ve derle
        let module = self.load_module(name, &module_path)?;
        
        // Cache'e ekle
        self.modules.insert(name.to_string(), module);
        
        Ok(())
    }
    
    /// Modül dosya yolunu çözümle
    fn resolve_module_path(&self, name: &str, current_file: &PathBuf) -> Result<PathBuf, String> {
        let module_name = name.replace(".", "/") + ".sky";
        
        // Arama yollarında ara
        for search_path in &self.search_paths {
            let full_path = search_path.join(&module_name);
            if full_path.exists() {
                return Ok(full_path);
            }
        }
        
        // Mevcut dosya klasöründe ara
        if let Some(parent) = current_file.parent() {
            let full_path = parent.join(&module_name);
            if full_path.exists() {
                return Ok(full_path);
            }
        }
        
        Err(format!("Module '{}' not found", name))
    }
    
    /// Modülü yükle ve derle
    fn load_module(&self, name: &str, path: &PathBuf) -> Result<Module, String> {
        // Dosyayı oku
        let content = std::fs::read_to_string(path)
            .map_err(|e| format!("Failed to read module file: {}", e))?;
        
        // Lex
        let tokens = crate::compiler::lexer::lex(&content)
            .map_err(|e| format!("Lex error: {:?}", e))?;
        
        // Parse
        let ast = crate::compiler::parser::parse_with_source(tokens, content)
            .map_err(|e| format!("Parse error: {:?}", e))?;
        
        // Bind
        let bound_ast = crate::compiler::binder::bind(ast)
            .map_err(|e| format!("Bind error: {:?}", e))?;
        
        // Compile
        let mut compiler = crate::compiler::bytecode::compiler::Compiler::new();
        let _chunk = compiler.compile_ast(bound_ast.clone())
            .map_err(|e| format!("Compile error: {:?}", e))?;
        
        // Public sembolleri topla
        let mut exports = HashMap::new();
        let mut globals = HashMap::new();
        
        for symbol in &bound_ast.symbols {
            if symbol.is_public() && !symbol.name.starts_with("__") {
                // Export edilebilir sembol
                let export_symbol = ExportSymbol {
                    name: symbol.name.clone(),
                    kind: match symbol.kind {
                        crate::compiler::binder::symbols::SymbolKind::Function => SymbolKind::Function,
                        crate::compiler::binder::symbols::SymbolKind::Variable => SymbolKind::Variable,
                        crate::compiler::binder::symbols::SymbolKind::Parameter => continue, // Parameters are not exported
                    },
                    value: Value::Null, // Placeholder
                    span: symbol.span,
                };
                exports.insert(symbol.name.clone(), export_symbol);
            }
        }
        
        Ok(Module {
            name: name.to_string(),
            file_path: path.clone(),
            exports,
            globals,
            compiled: true,
        })
    }
}
