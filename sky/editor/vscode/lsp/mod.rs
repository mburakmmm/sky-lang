// Sky LSP - Language Server Protocol
// Minimal LSP implementasyonu: diagnostics, hover, completion

use std::collections::HashMap;
use std::sync::Arc;
use tokio::sync::RwLock;
use tower_lsp::jsonrpc::Result;
use tower_lsp::lsp_types::*;
use tower_lsp::{Client, LanguageServer, LspService, Server};

use sky::compiler::diag::{Diagnostic, Span, SourceMap, Emitter};
use sky::compiler::lexer::lex;
use sky::compiler::parser::parse;
use sky::compiler::binder::bind;
use sky::compiler::types::decl::TypeDecl;

/// Sky LSP Server
pub struct SkyLspServer {
    client: Client,
    documents: Arc<RwLock<HashMap<Url, String>>>,
    source_map: Arc<RwLock<SourceMap>>,
}

impl SkyLspServer {
    pub fn new(client: Client) -> Self {
        Self {
            client,
            documents: Arc::new(RwLock::new(HashMap::new())),
            source_map: Arc::new(RwLock::new(SourceMap::new())),
        }
    }

    /// Dosya içeriğini güncelle
    async fn update_document(&self, uri: Url, content: String) {
        let mut documents = self.documents.write().await;
        documents.insert(uri.clone(), content);
        
        let mut source_map = self.source_map.write().await;
        source_map.add_file(uri.path().to_string(), content);
    }

    /// Dosyayı analiz et ve diagnostic'leri döndür
    async fn analyze_document(&self, uri: &Url) -> Vec<Diagnostic> {
        let documents = self.documents.read().await;
        let content = match documents.get(uri) {
            Some(content) => content,
            None => return vec![],
        };

        let mut diagnostics = Vec::new();

        // Lexer
        match lex(content) {
            Ok(_tokens) => {
                // Parser
                match parse(&lex(content).unwrap()) {
                    Ok(_ast) => {
                        // Binder
                        match bind(&parse(&lex(content).unwrap()).unwrap()) {
                            Ok(_bound_ast) => {
                                // Başarılı analiz
                            }
                            Err(diag) => {
                                diagnostics.push(diag);
                            }
                        }
                    }
                    Err(diag) => {
                        diagnostics.push(diag);
                    }
                }
            }
            Err(diag) => {
                diagnostics.push(diag);
            }
        }

        diagnostics
    }

    /// Diagnostic'i LSP Diagnostic'e çevir
    fn convert_diagnostic(&self, diag: &Diagnostic, uri: &Url) -> tower_lsp::lsp_types::Diagnostic {
        let severity = match diag.code.as_str() {
            "E0001" | "E0101" | "E0201" | "E0202" => DiagnosticSeverity::ERROR,
            "E1001" | "E2001" | "E3001" | "E3002" => DiagnosticSeverity::ERROR,
            _ => DiagnosticSeverity::WARNING,
        };

        let range = Range {
            start: Position {
                line: diag.span.start_line as u32,
                character: diag.span.start_column as u32,
            },
            end: Position {
                line: diag.span.end_line as u32,
                character: diag.span.end_column as u32,
            },
        };

        tower_lsp::lsp_types::Diagnostic {
            range,
            severity: Some(severity),
            code: Some(NumberOrString::String(diag.code.clone())),
            source: Some("sky".to_string()),
            message: diag.message.clone(),
            related_information: None,
            tags: None,
            code_description: None,
            data: None,
        }
    }

    /// Hover bilgisi oluştur
    async fn get_hover_info(&self, uri: &Url, position: Position) -> Option<Hover> {
        let documents = self.documents.read().await;
        let content = documents.get(uri)?;
        
        let lines: Vec<&str> = content.lines().collect();
        let line = lines.get(position.line as usize)?;
        
        // Basit identifier tespiti
        let chars: Vec<char> = line.chars().collect();
        let mut start = position.character as usize;
        let mut end = position.character as usize;
        
        // Sola doğru genişlet
        while start > 0 && chars[start - 1].is_alphanumeric() || chars[start - 1] == '_' {
            start -= 1;
        }
        
        // Sağa doğru genişlet
        while end < chars.len() && (chars[end].is_alphanumeric() || chars[end] == '_') {
            end += 1;
        }
        
        if start >= end {
            return None;
        }
        
        let identifier: String = chars[start..end].iter().collect();
        
        // Basit tip bilgisi
        let hover_content = if identifier == "int" {
            "**Sky Type**: `int`\n\nTamsayı tipi. Örnek: `int sayı = 42`"
        } else if identifier == "float" {
            "**Sky Type**: `float`\n\nOndalıklı sayı tipi. Örnek: `float pi = 3.14`"
        } else if identifier == "bool" {
            "**Sky Type**: `bool`\n\nBoolean tipi. Örnek: `bool doğru = true`"
        } else if identifier == "string" {
            "**Sky Type**: `string`\n\nMetin tipi. Örnek: `string mesaj = \"Merhaba\"`"
        } else if identifier == "list" {
            "**Sky Type**: `list`\n\nListe tipi. Örnek: `list sayılar = [1, 2, 3]`"
        } else if identifier == "map" {
            "**Sky Type**: `map`\n\nSözlük tipi. Örnek: `map sözlük = {\"ad\": \"Sky\"}`"
        } else if identifier == "var" {
            "**Sky Type**: `var`\n\nDinamik tip. Örnek: `var değer = 42`"
        } else if identifier == "function" {
            "**Sky Keyword**: `function`\n\nFonksiyon tanımlama. Örnek: `function selam(isim: string)`"
        } else if identifier == "async" {
            "**Sky Keyword**: `async`\n\nAsenkron fonksiyon. Örnek: `async function indir(url: string)`"
        } else if identifier == "coop" {
            "**Sky Keyword**: `coop`\n\nCoroutine fonksiyon. Örnek: `coop function say(n: int)`"
        } else if identifier == "if" {
            "**Sky Keyword**: `if`\n\nKoşul ifadesi. Örnek: `if x > 0`"
        } else if identifier == "for" {
            "**Sky Keyword**: `for`\n\nDöngü. Örnek: `for elem: var in liste`"
        } else if identifier == "while" {
            "**Sky Keyword**: `while`\n\nDöngü. Örnek: `while x > 0`"
        } else if identifier == "return" {
            "**Sky Keyword**: `return`\n\nDönüş ifadesi. Örnek: `return 42`"
        } else if identifier == "yield" {
            "**Sky Keyword**: `yield`\n\nCoroutine yield. Örnek: `yield i`"
        } else if identifier == "await" {
            "**Sky Keyword**: `await`\n\nAsenkron bekleme. Örnek: `await sleep(1000)`"
        } else if identifier == "import" {
            "**Sky Keyword**: `import`\n\nModül import. Örnek: `import math`"
        } else if identifier == "true" || identifier == "false" {
            format!("**Sky Literal**: `{}`\n\nBoolean değer", identifier)
        } else if identifier == "null" {
            "**Sky Literal**: `null`\n\nNull değer"
        } else {
            // Değişken veya fonksiyon adı
            format!("**Sky Identifier**: `{}`\n\nTanımlanmış değişken veya fonksiyon", identifier)
        };
        
        Some(Hover {
            contents: HoverContents::Markup(MarkupContent {
                kind: MarkupKind::Markdown,
                value: hover_content,
            }),
            range: Some(Range {
                start: Position {
                    line: position.line,
                    character: start as u32,
                },
                end: Position {
                    line: position.line,
                    character: end as u32,
                },
            }),
        })
    }
}

#[tower_lsp::async_trait]
impl LanguageServer for SkyLspServer {
    async fn initialize(&self, _: InitializeParams) -> Result<InitializeResult> {
        Ok(InitializeResult {
            server_info: Some(ServerInfo {
                name: "sky-lsp".to_string(),
                version: Some("0.1.0".to_string()),
            }),
            capabilities: ServerCapabilities {
                text_document_sync: Some(TextDocumentSyncCapability::Kind(
                    TextDocumentSyncKind::FULL,
                )),
                hover_provider: Some(HoverProviderCapability::Simple(true)),
                completion_provider: Some(CompletionOptions {
                    resolve_provider: Some(false),
                    trigger_characters: Some(vec![".".to_string(), ":".to_string()]),
                    ..Default::default()
                }),
                diagnostic_provider: Some(DiagnosticServerCapabilities::Options(
                    DiagnosticOptions {
                        identifier: Some("sky".to_string()),
                        inter_file_dependencies: true,
                        workspace_diagnostics: false,
                        ..Default::default()
                    },
                )),
                ..Default::default()
            },
            ..Default::default()
        })
    }

    async fn initialized(&self, _: InitializedParams) {
        self.client
            .log_message(MessageType::INFO, "Sky LSP server initialized!")
            .await;
    }

    async fn shutdown(&self) -> Result<()> {
        Ok(())
    }

    async fn did_open(&self, params: DidOpenTextDocumentParams) {
        let uri = params.text_document.uri;
        let content = params.text_document.text;
        
        self.update_document(uri.clone(), content).await;
        
        // Diagnostic'leri gönder
        let diagnostics = self.analyze_document(&uri).await;
        let lsp_diagnostics: Vec<tower_lsp::lsp_types::Diagnostic> = diagnostics
            .iter()
            .map(|diag| self.convert_diagnostic(diag, &uri))
            .collect();
        
        self.client
            .publish_diagnostics(uri, lsp_diagnostics, None)
            .await;
    }

    async fn did_change(&self, params: DidChangeTextDocumentParams) {
        let uri = params.text_document.uri;
        let content = params.content_changes[0].text.clone();
        
        self.update_document(uri.clone(), content).await;
        
        // Diagnostic'leri gönder
        let diagnostics = self.analyze_document(&uri).await;
        let lsp_diagnostics: Vec<tower_lsp::lsp_types::Diagnostic> = diagnostics
            .iter()
            .map(|diag| self.convert_diagnostic(diag, &uri))
            .collect();
        
        self.client
            .publish_diagnostics(uri, lsp_diagnostics, None)
            .await;
    }

    async fn did_save(&self, params: DidSaveTextDocumentParams) {
        let uri = params.text_document.uri;
        
        // Diagnostic'leri gönder
        let diagnostics = self.analyze_document(&uri).await;
        let lsp_diagnostics: Vec<tower_lsp::lsp_types::Diagnostic> = diagnostics
            .iter()
            .map(|diag| self.convert_diagnostic(diag, &uri))
            .collect();
        
        self.client
            .publish_diagnostics(uri, lsp_diagnostics, None)
            .await;
    }

    async fn did_close(&self, params: DidCloseTextDocumentParams) {
        let uri = params.text_document.uri;
        let mut documents = self.documents.write().await;
        documents.remove(&uri);
    }

    async fn hover(&self, params: HoverParams) -> Result<Option<Hover>> {
        let uri = params.text_document_position_params.text_document.uri;
        let position = params.text_document_position_params.position;
        
        Ok(self.get_hover_info(&uri, position).await)
    }

    async fn completion(&self, params: CompletionParams) -> Result<Option<CompletionResponse>> {
        let uri = params.text_document_position.text_document.uri;
        let position = params.text_document_position.position;
        
        let documents = self.documents.read().await;
        let content = match documents.get(&uri) {
            Some(content) => content,
            None => return Ok(None),
        };
        
        let lines: Vec<&str> = content.lines().collect();
        let line = match lines.get(position.line as usize) {
            Some(line) => line,
            None => return Ok(None),
        };
        
        // Basit completion listesi
        let mut completions = Vec::new();
        
        // Tip anahtar kelimeleri
        let types = vec!["int", "float", "bool", "string", "list", "map", "var"];
        for ty in types {
            completions.push(CompletionItem {
                label: ty.to_string(),
                kind: Some(CompletionItemKind::KEYWORD),
                detail: Some(format!("Sky type: {}", ty)),
                documentation: Some(Documentation::MarkupContent(MarkupContent {
                    kind: MarkupKind::Markdown,
                    value: format!("**Sky Type**: `{}`", ty),
                })),
                ..Default::default()
            });
        }
        
        // Anahtar kelimeler
        let keywords = vec![
            "function", "async", "coop", "if", "elif", "else", "for", "while",
            "break", "continue", "return", "yield", "await", "import", "true", "false", "null"
        ];
        for kw in keywords {
            completions.push(CompletionItem {
                label: kw.to_string(),
                kind: Some(CompletionItemKind::KEYWORD),
                detail: Some(format!("Sky keyword: {}", kw)),
                documentation: Some(Documentation::MarkupContent(MarkupContent {
                    kind: MarkupKind::Markdown,
                    value: format!("**Sky Keyword**: `{}`", kw),
                })),
                ..Default::default()
            });
        }
        
        // Operatörler
        let operators = vec!["+", "-", "*", "/", "%", "==", "!=", "<", "<=", ">", ">=", "and", "or"];
        for op in operators {
            completions.push(CompletionItem {
                label: op.to_string(),
                kind: Some(CompletionItemKind::OPERATOR),
                detail: Some(format!("Sky operator: {}", op)),
                ..Default::default()
            });
        }
        
        Ok(Some(CompletionResponse::Array(completions)))
    }

    async fn diagnostic(&self, params: DocumentDiagnosticParams) -> Result<DocumentDiagnosticReportResult> {
        let uri = params.text_document.uri;
        let diagnostics = self.analyze_document(&uri).await;
        
        let lsp_diagnostics: Vec<tower_lsp::lsp_types::Diagnostic> = diagnostics
            .iter()
            .map(|diag| self.convert_diagnostic(diag, &uri))
            .collect();
        
        Ok(DocumentDiagnosticReportResult::Report(
            DocumentDiagnosticReport {
                kind: ReportKind::Full,
                result_id: None,
                items: lsp_diagnostics,
            },
        ))
    }
}

/// LSP server'ı başlat
pub async fn start_lsp_server() {
    let (service, socket) = LspService::new(SkyLspServer::new);
    Server::new(stdin(), stdout(), socket).serve(service).await;
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_lsp_server_creation() {
        // Mock client ile test
        let client = Client::new(stdin(), stdout());
        let server = SkyLspServer::new(client);
        assert_eq!(server.documents.blocking_read().len(), 0);
    }

    #[test]
    fn test_hover_info() {
        let client = Client::new(stdin(), stdout());
        let server = SkyLspServer::new(client);
        
        // Test için basit bir hover bilgisi
        let uri = Url::parse("file:///test.sky").unwrap();
        let position = Position { line: 0, character: 0 };
        
        // Bu test async olduğu için runtime gerekiyor
        // Gerçek testlerde tokio::test macro kullanılmalı
    }
}
