use std::path::PathBuf;

#[derive(Debug, Clone)]
pub struct FormatterConfig {
    pub indent_size: usize,
    pub max_line_length: usize,
    pub use_tabs: bool,
    pub trailing_commas: bool,
    pub newline_at_end: bool,
    pub remove_trailing_spaces: bool,
}

impl Default for FormatterConfig {
    fn default() -> Self {
        Self {
            indent_size: 4,
            max_line_length: 100,
            use_tabs: false,
            trailing_commas: true,
            newline_at_end: true,
            remove_trailing_spaces: true,
        }
    }
}

pub struct SkyFormatter {
    config: FormatterConfig,
}

impl SkyFormatter {
    pub fn new(config: FormatterConfig) -> Self {
        Self { config }
    }

    pub fn format_file(&self, path: &PathBuf) -> Result<String, String> {
        std::fs::read_to_string(path)
            .map_err(|e| format!("Dosya okuma hatası: {}", e))
    }
}

#[derive(Debug, Clone)]
pub struct FormatResult {
    pub changes_made: bool,
    pub lines_changed: usize,
    pub formatted_code: String,
    pub diagnostics: Vec<crate::compiler::diag::Diagnostic>,
}

pub fn format_and_save(path: &PathBuf, config: Option<FormatterConfig>) -> Result<FormatResult, String> {
    let config = config.unwrap_or_default();
    let formatter = SkyFormatter::new(config);
    
    let formatted_content = formatter.format_file(path)?;
    
    std::fs::write(path, &formatted_content)
        .map_err(|e| format!("Dosya yazma hatası: {}", e))?;
    
    Ok(FormatResult {
        changes_made: true,
        lines_changed: 1, // Basit implementasyon
        formatted_code: formatted_content,
        diagnostics: vec![],
    })
}
