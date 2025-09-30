// Module Cache - Modül cache sistemi

use std::collections::HashMap;
use super::Module;

pub struct ModuleCache {
    modules: HashMap<String, Module>,
}

impl ModuleCache {
    pub fn new() -> Self {
        Self {
            modules: HashMap::new(),
        }
    }
    
    pub fn get(&self, name: &str) -> Option<&Module> {
        self.modules.get(name)
    }
    
    pub fn insert(&mut self, name: String, module: Module) {
        self.modules.insert(name, module);
    }
    
    pub fn contains(&self, name: &str) -> bool {
        self.modules.contains_key(name)
    }
}
