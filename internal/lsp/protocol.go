package lsp

import "encoding/json"

// LSP Protocol Types - LSP 3.17 specification

// Position pozisyon (line, character)
type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

// Range metin aralığı
type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

// Location dosya konumu
type Location struct {
	URI   string `json:"uri"`
	Range Range  `json:"range"`
}

// TextDocumentIdentifier doküman identifier
type TextDocumentIdentifier struct {
	URI string `json:"uri"`
}

// VersionedTextDocumentIdentifier versiyonlu doküman
type VersionedTextDocumentIdentifier struct {
	TextDocumentIdentifier
	Version int `json:"version"`
}

// TextDocumentItem doküman
type TextDocumentItem struct {
	URI        string `json:"uri"`
	LanguageID string `json:"languageId"`
	Version    int    `json:"version"`
	Text       string `json:"text"`
}

// TextDocumentContentChangeEvent doküman değişikliği
type TextDocumentContentChangeEvent struct {
	Range       *Range `json:"range,omitempty"`
	RangeLength int    `json:"rangeLength,omitempty"`
	Text        string `json:"text"`
}

// Diagnostic hata/uyarı
type Diagnostic struct {
	Range    Range              `json:"range"`
	Severity DiagnosticSeverity `json:"severity"`
	Code     string             `json:"code,omitempty"`
	Source   string             `json:"source,omitempty"`
	Message  string             `json:"message"`
}

// DiagnosticSeverity severity seviyeleri
type DiagnosticSeverity int

const (
	SeverityError       DiagnosticSeverity = 1
	SeverityWarning     DiagnosticSeverity = 2
	SeverityInformation DiagnosticSeverity = 3
	SeverityHint        DiagnosticSeverity = 4
)

// CompletionItem tamamlama önerisi
type CompletionItem struct {
	Label         string             `json:"label"`
	Kind          CompletionItemKind `json:"kind"`
	Detail        string             `json:"detail,omitempty"`
	Documentation string             `json:"documentation,omitempty"`
	InsertText    string             `json:"insertText,omitempty"`
	SortText      string             `json:"sortText,omitempty"`
}

// CompletionItemKind completion türleri
type CompletionItemKind int

const (
	CompletionText          CompletionItemKind = 1
	CompletionMethod        CompletionItemKind = 2
	CompletionFunction      CompletionItemKind = 3
	CompletionConstructor   CompletionItemKind = 4
	CompletionField         CompletionItemKind = 5
	CompletionVariable      CompletionItemKind = 6
	CompletionClass         CompletionItemKind = 7
	CompletionInterface     CompletionItemKind = 8
	CompletionModule        CompletionItemKind = 9
	CompletionProperty      CompletionItemKind = 10
	CompletionKeyword       CompletionItemKind = 14
	CompletionSnippet       CompletionItemKind = 15
)

// Hover hover bilgisi
type Hover struct {
	Contents string `json:"contents"`
	Range    *Range `json:"range,omitempty"`
}

// SignatureHelp fonksiyon imza yardımı
type SignatureHelp struct {
	Signatures      []SignatureInformation `json:"signatures"`
	ActiveSignature int                    `json:"activeSignature"`
	ActiveParameter int                    `json:"activeParameter"`
}

// SignatureInformation fonksiyon imzası
type SignatureInformation struct {
	Label         string                 `json:"label"`
	Documentation string                 `json:"documentation,omitempty"`
	Parameters    []ParameterInformation `json:"parameters,omitempty"`
}

// ParameterInformation parametre bilgisi
type ParameterInformation struct {
	Label         string `json:"label"`
	Documentation string `json:"documentation,omitempty"`
}

// DocumentSymbol doküman sembolleri
type DocumentSymbol struct {
	Name           string           `json:"name"`
	Detail         string           `json:"detail,omitempty"`
	Kind           SymbolKind       `json:"kind"`
	Range          Range            `json:"range"`
	SelectionRange Range            `json:"selectionRange"`
	Children       []DocumentSymbol `json:"children,omitempty"`
}

// SymbolKind sembol türleri
type SymbolKind int

const (
	SymbolFile        SymbolKind = 1
	SymbolModule      SymbolKind = 2
	SymbolNamespace   SymbolKind = 3
	SymbolPackage     SymbolKind = 4
	SymbolClass       SymbolKind = 5
	SymbolMethod      SymbolKind = 6
	SymbolProperty    SymbolKind = 7
	SymbolField       SymbolKind = 8
	SymbolConstructor SymbolKind = 9
	SymbolFunction    SymbolKind = 12
	SymbolVariable    SymbolKind = 13
	SymbolConstant    SymbolKind = 14
)

// ReferenceContext referans context
type ReferenceContext struct {
	IncludeDeclaration bool `json:"includeDeclaration"`
}

// FormattingOptions format ayarları
type FormattingOptions struct {
	TabSize      int  `json:"tabSize"`
	InsertSpaces bool `json:"insertSpaces"`
}

// TextEdit metin düzenleme
type TextEdit struct {
	Range   Range  `json:"range"`
	NewText string `json:"newText"`
}

// InitializeParams initialize parametreleri
type InitializeParams struct {
	ProcessID int                `json:"processId"`
	RootURI   string             `json:"rootUri"`
	Capabilities ClientCapabilities `json:"capabilities"`
}

// ClientCapabilities client yetenekleri
type ClientCapabilities struct {
	TextDocument TextDocumentClientCapabilities `json:"textDocument"`
}

// TextDocumentClientCapabilities doküman yetenekleri
type TextDocumentClientCapabilities struct {
	Completion    *CompletionClientCapabilities    `json:"completion,omitempty"`
	Hover         *HoverClientCapabilities         `json:"hover,omitempty"`
	SignatureHelp *SignatureHelpClientCapabilities `json:"signatureHelp,omitempty"`
}

// CompletionClientCapabilities completion yetenekleri
type CompletionClientCapabilities struct {
	CompletionItem *struct {
		SnippetSupport bool `json:"snippetSupport"`
	} `json:"completionItem,omitempty"`
}

// HoverClientCapabilities hover yetenekleri
type HoverClientCapabilities struct {
	ContentFormat []string `json:"contentFormat,omitempty"`
}

// SignatureHelpClientCapabilities signature help yetenekleri
type SignatureHelpClientCapabilities struct {
	SignatureInformation *struct {
		DocumentationFormat []string `json:"documentationFormat,omitempty"`
	} `json:"signatureInformation,omitempty"`
}

// InitializeResult initialize sonucu
type InitializeResult struct {
	Capabilities ServerCapabilities `json:"capabilities"`
}

// ServerCapabilities server yetenekleri
type ServerCapabilities struct {
	TextDocumentSync   int                           `json:"textDocumentSync"`
	CompletionProvider *CompletionOptions            `json:"completionProvider,omitempty"`
	HoverProvider      bool                          `json:"hoverProvider"`
	SignatureHelpProvider *SignatureHelpOptions      `json:"signatureHelpProvider,omitempty"`
	DefinitionProvider bool                          `json:"definitionProvider"`
	ReferencesProvider bool                          `json:"referencesProvider"`
	DocumentSymbolProvider bool                      `json:"documentSymbolProvider"`
	DocumentFormattingProvider bool                  `json:"documentFormattingProvider"`
}

// CompletionOptions completion ayarları
type CompletionOptions struct {
	TriggerCharacters []string `json:"triggerCharacters,omitempty"`
}

// SignatureHelpOptions signature help ayarları
type SignatureHelpOptions struct {
	TriggerCharacters []string `json:"triggerCharacters,omitempty"`
}

// Request LSP request
type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// Response LSP response
type Response struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
}

// Notification LSP notification
type Notification struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// RPCError RPC hata
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Error codes
const (
	ParseError     = -32700
	InvalidRequest = -32600
	MethodNotFound = -32601
	InvalidParams  = -32602
	InternalError  = -32603
)

