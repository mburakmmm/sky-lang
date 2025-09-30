// Diagnostic System - Hata Mesajları ve Konum Bilgisi
// Konumlu hata mesajları ve renkli CLI çıktısı

use colored::*;
use std::collections::HashMap;
use std::fmt;

/// Dosya içindeki konum bilgisi
#[derive(Debug, Clone, Copy, PartialEq)]
pub struct Span {
    pub file_id: usize,
    pub start: usize,
    pub end: usize,
}

impl Span {
    pub fn new(file_id: usize, start: usize, end: usize) -> Self {
        Self { file_id, start, end }
    }

    pub fn len(&self) -> usize {
        self.end - self.start
    }

    pub fn start_line(&self) -> usize {
        // Basit implementation - gerçekte source map'ten hesaplanmalı
        (self.start / 80) + 1 // Satır başına ~80 karakter varsayımı
    }
}

/// Hata kodu ve mesajı
#[derive(Debug, Clone)]
pub struct Diagnostic {
    pub code: String,
    pub message: String,
    pub span: Span,
    pub notes: Vec<String>,
    pub severity: Severity,
}

#[derive(Debug, Clone, Copy, PartialEq)]
pub enum Severity {
    Error,
    Warning,
    Info,
}

impl Diagnostic {
    pub fn error(code: &str, message: &str, span: Span) -> Self {
        Self {
            code: code.to_string(),
            message: message.to_string(),
            span,
            notes: Vec::new(),
            severity: Severity::Error,
        }
    }

    pub fn with_note(mut self, note: &str) -> Self {
        self.notes.push(note.to_string());
        self
    }

    pub fn warning(code: &str, message: &str, span: Span) -> Self {
        Self {
            code: code.to_string(),
            message: message.to_string(),
            span,
            notes: Vec::new(),
            severity: Severity::Warning,
        }
    }
}

/// Kaynak dosya haritası ve satır indeksleri
pub struct SourceMap {
    files: HashMap<usize, String>,
    line_starts: HashMap<usize, Vec<usize>>,
}

impl SourceMap {
    pub fn new() -> Self {
        Self {
            files: HashMap::new(),
            line_starts: HashMap::new(),
        }
    }

    pub fn add_file(&mut self, file_id: usize, content: String) {
        let mut starts = vec![0];
        for (i, ch) in content.char_indices() {
            if ch == '\n' {
                starts.push(i + 1);
            }
        }
        self.files.insert(file_id, content);
        self.line_starts.insert(file_id, starts);
    }

    pub fn get_line_col(&self, file_id: usize, pos: usize) -> Option<(usize, usize)> {
        let starts = self.line_starts.get(&file_id)?;
        let line = starts.binary_search(&pos).unwrap_or_else(|i| i - 1);
        let col = pos - starts[line];
        Some((line + 1, col + 1))
    }

    pub fn get_line(&self, file_id: usize, line: usize) -> Option<&str> {
        let content = self.files.get(&file_id)?;
        let starts = self.line_starts.get(&file_id)?;
        if line == 0 || line > starts.len() {
            return None;
        }
        let start = starts[line - 1];
        let end = if line < starts.len() {
            starts[line] - 1
        } else {
            content.len()
        };
        Some(&content[start..end])
    }
}

/// Hata mesajı yazıcısı
pub struct Emitter {
    source_map: SourceMap,
}

impl Emitter {
    pub fn new(source_map: SourceMap) -> Self {
        Self { source_map }
    }

    pub fn emit(&self, diag: &Diagnostic) -> String {
        let mut output = String::new();
        
        // Hata kodu ve mesajı
        let severity_color = match diag.severity {
            Severity::Error => "error".red().bold(),
            Severity::Warning => "warning".yellow().bold(),
            Severity::Info => "info".blue().bold(),
        };
        
        output.push_str(&format!(
            "{}[{}]: {}\n",
            severity_color,
            diag.code,
            diag.message
        ));

        // Konum bilgisi
        if let Some((line, col)) = self.source_map.get_line_col(diag.span.file_id, diag.span.start) {
            output.push_str(&format!("  --> main.sky:{}:{}\n", line, col));
        }

        // Kaynak satırı ve caret
        if let Some((line_num, col)) = self.source_map.get_line_col(diag.span.file_id, diag.span.start) {
            if let Some(source_line) = self.source_map.get_line(diag.span.file_id, line_num) {
                output.push_str(&format!(" {} |{}\n", line_num, source_line));
            
                // Caret pozisyonu
                let caret_pos = if diag.span.len() > 0 {
                    format!("{:>width$}", "^".repeat(diag.span.len()), width = col + 3)
                } else {
                    format!("{:>width$}", "^", width = col + 3)
                };
                output.push_str(&format!("   |{}\n", caret_pos));
            }
        }

        // Notlar
        for note in &diag.notes {
            output.push_str(&format!("   = help: {}\n", note.green()));
        }

        output
    }
}

// Standart hata kodları
pub mod codes {
    pub const MISSING_TYPE_ANNOTATION: &str = "E0001";
    pub const INVALID_INDENTATION: &str = "E0101";
    pub const AWAIT_OUTSIDE_ASYNC: &str = "E0201";
    pub const YIELD_OUTSIDE_COOP: &str = "E0202";
    pub const TYPE_MISMATCH: &str = "E1001";
    pub const COROUTINE_FINISHED: &str = "E2001";
    pub const PYTHON_BRIDGE_ERROR: &str = "E3001";
    pub const JS_BRIDGE_ERROR: &str = "E3002";
    
    // Entry point and visibility errors
    pub const INVALID_MAIN_SIGNATURE: &str = "E6001";
    pub const BRACE_BODY_NON_MAIN: &str = "E6002";
    pub const MAIN_MUST_USE_BRACES: &str = "E6003";
    pub const PRIVATE_SYMBOL_ACCESS: &str = "E0404";
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_diagnostic_creation() {
        let span = Span::new(0, 0, 5);
        let diag = Diagnostic::error("E0001", "missing type annotation", span);
        assert_eq!(diag.code, "E0001");
        assert_eq!(diag.message, "missing type annotation");
        assert_eq!(diag.span, span);
    }

    #[test]
    fn test_source_map() {
        let mut source_map = SourceMap::new();
        let content = "int x = 3\nprint(x)".to_string();
        source_map.add_file(0, content);
        
        assert_eq!(source_map.get_line_col(0, 0), Some((1, 1)));
        assert_eq!(source_map.get_line_col(0, 10), Some((2, 1)));
        assert_eq!(source_map.get_line(0, 1), Some("int x = 3"));
        assert_eq!(source_map.get_line(0, 2), Some("print(x)"));
    }
}
