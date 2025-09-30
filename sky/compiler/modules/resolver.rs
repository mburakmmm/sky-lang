// Module Resolver - Modül dosya yolu çözümleme

use std::path::PathBuf;

pub fn resolve_module_path(name: &str, search_paths: &[PathBuf], current_file: &PathBuf) -> Result<PathBuf, String> {
    let module_name = name.replace(".", "/") + ".sky";
    
    // Arama yollarında ara
    for search_path in search_paths {
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
