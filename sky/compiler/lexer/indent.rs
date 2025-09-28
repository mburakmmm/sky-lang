// Indentation Tracker - Girinti Mantığı
// INDENT/DEDENT tokenları üretir ve girinti tutarlılığını kontrol eder

use crate::compiler::diag::{Diagnostic, codes};
use super::token::{Token, TokenKind};
use crate::compiler::diag::Span;

pub struct IndentTracker {
    stack: Vec<usize>,
}

impl IndentTracker {
    pub fn new() -> Self {
        Self {
            stack: vec![0], // Başlangıç seviyesi
        }
    }
    
    /// Satır başındaki girinti seviyesini işle
    pub fn process_indent(&mut self, indent_level: usize) -> Result<Option<Vec<Token>>, Diagnostic> {
        let current = *self.stack.last().unwrap();
        
        if indent_level > current {
            // Yeni girinti seviyesi - INDENT token üret
            self.stack.push(indent_level);
            Ok(Some(vec![Token::new(TokenKind::Indent, Span::new(0, 0, 0))]))
        } else if indent_level < current {
            // Girinti azaldı - DEDENT tokenları üret
            let mut tokens = Vec::new();
            
            while let Some(&level) = self.stack.last() {
                if level <= indent_level {
                    break;
                }
                self.stack.pop();
                tokens.push(Token::new(TokenKind::Dedent, Span::new(0, 0, 0)));
            }
            
            // Girinti seviyesi stack'te yoksa hata
            if self.stack.last().unwrap() != &indent_level {
                return Err(Diagnostic::error(
                    codes::INVALID_INDENTATION,
                    "Inconsistent indentation",
                    Span::new(0, 0, 0)
                ));
            }
            
            Ok(if tokens.is_empty() { None } else { Some(tokens) })
        } else {
            // Aynı seviye - token üretme
            Ok(None)
        }
    }
    
    /// Dosya sonunda kalan girinti seviyelerini kapat
    pub fn finalize(mut self) -> Result<Option<Vec<Token>>, Diagnostic> {
        let mut tokens = Vec::new();
        
        // Başlangıç seviyesi (0) hariç tüm seviyeleri kapat
        while self.stack.len() > 1 {
            self.stack.pop();
            tokens.push(Token::new(TokenKind::Dedent, Span::new(0, 0, 0)));
        }
        
        Ok(if tokens.is_empty() { None } else { Some(tokens) })
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_indent_tracking() {
        let mut tracker = IndentTracker::new();
        
        // İlk girinti
        let tokens = tracker.process_indent(2).unwrap();
        assert!(tokens.is_some());
        assert_eq!(tokens.unwrap().len(), 1);
        
        // Aynı seviye
        let tokens = tracker.process_indent(2).unwrap();
        assert!(tokens.is_none());
        
        // Daha derin girinti
        let tokens = tracker.process_indent(4).unwrap();
        assert!(tokens.is_some());
        assert_eq!(tokens.unwrap().len(), 1);
        
        // Geri dönüş
        let tokens = tracker.process_indent(2).unwrap();
        assert!(tokens.is_some());
        assert_eq!(tokens.unwrap().len(), 1); // Bir DEDENT
    }

    #[test]
    fn test_inconsistent_indentation() {
        let mut tracker = IndentTracker::new();
        
        tracker.process_indent(2).unwrap();
        tracker.process_indent(4).unwrap();
        
        // Tutarsız girinti
        let result = tracker.process_indent(3);
        assert!(result.is_err());
    }

    #[test]
    fn test_finalize() {
        let tracker = IndentTracker::new();
        let mut tracker = tracker;
        
        tracker.process_indent(2).unwrap();
        tracker.process_indent(4).unwrap();
        
        let tokens = tracker.finalize().unwrap();
        assert!(tokens.is_some());
        assert_eq!(tokens.unwrap().len(), 2); // İki DEDENT
    }
}
