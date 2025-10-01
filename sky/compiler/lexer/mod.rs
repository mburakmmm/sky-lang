// Lexer - Tokenization ve Girinti Mantığı
// Kaynağı token'lara böl, INDENT/DEDENT üreterek blokları belirle

pub mod token;
pub mod indent;

use token::Token;
use indent::IndentTracker;
use crate::compiler::diag::{Diagnostic, Span, codes};

/// Ana lexer fonksiyonu
pub fn lex(src: &str) -> Result<Vec<Token>, Diagnostic> {
    let mut tokens: Vec<Token> = Vec::new();
    let mut indent_tracker = IndentTracker::new();
    let mut chars = src.char_indices().peekable();
    
    let mut line_start = 0;
    let mut in_line = true;
    let mut current_pos = 0;
    
    while let Some((pos, ch)) = chars.next() {
        current_pos = pos;
        
        // Satır başında girinti kontrolü (newline hariç)
        if in_line && ch.is_whitespace() && ch != '\n' {
            if ch == '\t' {
                return Err(Diagnostic::error(
                    codes::INVALID_INDENTATION,
                    "Use spaces for indentation",
                    Span::new(0, pos, pos + 1)
                ));
            }
            continue;
        }
        
        // Satır başında girinti tokenları üret
        if in_line {
            // Satır başındaki boşluk sayısını hesapla
            let mut indent_level = 0;
            let mut temp_pos = line_start;
            while temp_pos < pos && temp_pos < src.len() {
                let byte_at_temp = src.as_bytes().get(temp_pos);
                match byte_at_temp {
                    Some(b' ') => indent_level += 1,
                    Some(b'\t') => {
                        return Err(Diagnostic::error(
                            codes::INVALID_INDENTATION,
                            "Use spaces for indentation",
                            Span::new(0, temp_pos, temp_pos + 1)
                        ));
                    }
                    _ => break,
                }
                temp_pos += 1;
            }
            
        match indent_tracker.process_indent(indent_level) {
            Ok(Some(indent_tokens)) => {
                tokens.extend(indent_tokens);
            }
            Ok(None) => {
            }
            Err(e) => {
                return Err(e);
            }
        }
            in_line = false;
        }
        
        // Token türüne göre işle
        match ch {
            '\n' => {
                tokens.push(Token::new(token::TokenKind::Newline, Span::new(0, pos, pos + 1)));
                line_start = pos + 1;
                in_line = true;
            }
            
            // Yorumlar
            '#' => {
                let start = pos;
                while let Some((_, next_ch)) = chars.peek() {
                    if *next_ch == '\n' {
                        break;
                    }
                    chars.next();
                }
                let end_pos = chars.peek().map(|(p, _)| *p).unwrap_or(src.len());
                tokens.push(Token::new(
                    token::TokenKind::Comment,
                    Span::new(0, start, end_pos)
                ));
            }
            
            // String literals
            '"' => {
                let start = pos;
                let mut escaped = false;
                let mut string_content = String::new();
                
                while let Some((_, next_ch)) = chars.next() {
                    if escaped {
                        // Escape sequence işle
                        match next_ch {
                            'n' => string_content.push('\n'),
                            't' => string_content.push('\t'),
                            'r' => string_content.push('\r'),
                            '"' => string_content.push('"'),
                            '\\' => string_content.push('\\'),
                            _ => {
                                // Geçersiz escape sequence - karakteri olduğu gibi ekle
                                string_content.push('\\');
                                string_content.push(next_ch);
                            }
                        }
                        escaped = false;
                        continue;
                    }
                    
                    if next_ch == '\\' {
                        escaped = true;
                        continue;
                    }
                    
                    if next_ch == '"' {
                        break;
                    }
                    
                    string_content.push(next_ch);
                }
                
                let end_pos = chars.peek().map(|(p, _)| *p).unwrap_or(src.len());
                let mut token = Token::new(
                    token::TokenKind::String,
                    Span::new(0, start, end_pos)
                );
                token.value = Some(string_content);
                tokens.push(token);
            }
            
            // Sayılar
            '0'..='9' => {
                let start = pos;
                let mut has_dot = false;
                while let Some((_, next_ch)) = chars.peek() {
                    match next_ch {
                        '0'..='9' => { chars.next(); }
                        '.' if !has_dot => {
                            // İkinci noktayı kontrol et (Range operatörü için)
                            let mut lookahead = chars.clone();
                            lookahead.next(); // İlk noktayı geç
                            if let Some((_, '.')) = lookahead.peek() {
                                // İkinci nokta var, bu Range operatörü, sayı parsing'i bitir
                                break;
                            } else {
                                // Tek nokta, Float parsing devam et
                                has_dot = true;
                                chars.next();
                            }
                        }
                        _ => break,
                    }
                }
                let kind = if has_dot {
                    token::TokenKind::Float
                } else {
                    token::TokenKind::Int
                };
                tokens.push(Token::new(kind, Span::new(0, start, chars.peek().map(|(p, _)| *p).unwrap_or(src.len()))));
            }
            
            // Operatörler ve semboller
            '+' => tokens.push(Token::new(token::TokenKind::Plus, Span::new(0, pos, pos + 1))),
            '-' => tokens.push(Token::new(token::TokenKind::Minus, Span::new(0, pos, pos + 1))),
            '*' => tokens.push(Token::new(token::TokenKind::Star, Span::new(0, pos, pos + 1))),
            '/' => tokens.push(Token::new(token::TokenKind::Slash, Span::new(0, pos, pos + 1))),
            '%' => tokens.push(Token::new(token::TokenKind::Percent, Span::new(0, pos, pos + 1))),
            '=' => {
                if chars.peek().map(|(_, ch)| *ch) == Some('=') {
                    chars.next();
                    tokens.push(Token::new(token::TokenKind::EqualEqual, Span::new(0, pos, pos + 2)));
                } else {
                    tokens.push(Token::new(token::TokenKind::Equal, Span::new(0, pos, pos + 1)));
                }
            }
            '!' => {
                if chars.peek().map(|(_, ch)| *ch) == Some('=') {
                    chars.next();
                    tokens.push(Token::new(token::TokenKind::BangEqual, Span::new(0, pos, pos + 2)));
                } else {
                    tokens.push(Token::new(token::TokenKind::Bang, Span::new(0, pos, pos + 1)));
                }
            }
            '<' => {
                if chars.peek().map(|(_, ch)| *ch) == Some('=') {
                    chars.next();
                    tokens.push(Token::new(token::TokenKind::LessEqual, Span::new(0, pos, pos + 2)));
                } else {
                    tokens.push(Token::new(token::TokenKind::Less, Span::new(0, pos, pos + 1)));
                }
            }
            '>' => {
                if chars.peek().map(|(_, ch)| *ch) == Some('=') {
                    chars.next();
                    tokens.push(Token::new(token::TokenKind::GreaterEqual, Span::new(0, pos, pos + 2)));
                } else {
                    tokens.push(Token::new(token::TokenKind::Greater, Span::new(0, pos, pos + 1)));
                }
            }
            ':' => tokens.push(Token::new(token::TokenKind::Colon, Span::new(0, pos, pos + 1))),
            ',' => tokens.push(Token::new(token::TokenKind::Comma, Span::new(0, pos, pos + 1))),
            '?' => tokens.push(Token::new(token::TokenKind::Question, Span::new(0, pos, pos + 1))),
            '(' => tokens.push(Token::new(token::TokenKind::LeftParen, Span::new(0, pos, pos + 1))),
            ')' => tokens.push(Token::new(token::TokenKind::RightParen, Span::new(0, pos, pos + 1))),
            '[' => tokens.push(Token::new(token::TokenKind::LeftBracket, Span::new(0, pos, pos + 1))),
            ']' => tokens.push(Token::new(token::TokenKind::RightBracket, Span::new(0, pos, pos + 1))),
            '{' => tokens.push(Token::new(token::TokenKind::LeftBrace, Span::new(0, pos, pos + 1))),
            '}' => tokens.push(Token::new(token::TokenKind::RightBrace, Span::new(0, pos, pos + 1))),
            '.' => {
                if chars.peek().map(|(_, ch)| *ch) == Some('.') {
                    chars.next();
                    tokens.push(Token::new(token::TokenKind::Range, Span::new(0, pos, pos + 2)));
                } else {
                    // Tek nokta - Dot token'ı
                    tokens.push(Token::new(token::TokenKind::Dot, Span::new(0, pos, pos + 1)));
                }
            }
            
            // Whitespace (satır başı hariç)
            ' ' | '\t' | '\r' => {
                // Ignore whitespace
            }
            
            // Identifiers ve keywords
            ch if ch.is_alphabetic() || ch == '_' => {
                let start = pos;
                while let Some((_, next_ch)) = chars.peek() {
                    if next_ch.is_alphanumeric() || *next_ch == '_' {
                        chars.next();
                    } else if *next_ch == '.' {
                        // Dot karakteri için özel istisna - identifier parsing'i durdur
                        // Dot token'ı main loop'ta oluşturulacak
                        break;
                    } else {
                        break;
                    }
                }
                let span = Span::new(0, start, chars.peek().map(|(p, _)| *p).unwrap_or(src.len()));
                let text = &src[start..span.end];
                let kind = token::TokenKind::from_keyword(text).unwrap_or(token::TokenKind::Identifier);
                tokens.push(Token::new(kind, span));
            }
            
            _ => {
                return Err(Diagnostic::error(
                    "E0000",
                    &format!("Unexpected character: '{}'", ch),
                    Span::new(0, pos, pos + 1)
                ));
            }
        }
    }
    
    // Dosya sonunda girinti kapanışları
    if let Some(dedent_tokens) = indent_tracker.finalize()? {
        tokens.extend(dedent_tokens);
    }
    
    // EOF token
    tokens.push(Token::new(token::TokenKind::Eof, Span::new(0, current_pos, current_pos)));
    
    Ok(tokens)
}


#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_basic_tokens() {
        let tokens = lex("int x = 42").unwrap();
        assert_eq!(tokens.len(), 6); // int, x, =, 42, Eof
        assert!(matches!(tokens[0].kind, token::TokenKind::Int));
        assert!(matches!(tokens[1].kind, token::TokenKind::Identifier));
        assert!(matches!(tokens[2].kind, token::TokenKind::Equal));
        assert!(matches!(tokens[3].kind, token::TokenKind::Int));
    }

    #[test]
    fn test_indentation() {
        let tokens = lex("if true\n  print(1)\nprint(2)").unwrap();
        let has_indent = tokens.iter().any(|t| matches!(t.kind, token::TokenKind::Indent));
        let has_dedent = tokens.iter().any(|t| matches!(t.kind, token::TokenKind::Dedent));
        assert!(has_indent);
        assert!(has_dedent);
    }

    #[test]
    fn test_tab_error() {
        let result = lex("if true\n\tprint(1)");
        assert!(result.is_err());
        if let Err(diag) = result {
            assert_eq!(diag.code, codes::INVALID_INDENTATION);
        }
    }

    #[test]
    fn test_unicode_identifiers() {
        let tokens = lex("int çorba = 1").unwrap();
        assert!(matches!(tokens[0].kind, token::TokenKind::Int));
        assert!(matches!(tokens[1].kind, token::TokenKind::Identifier));
    }
}
