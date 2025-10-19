package lsp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"
	"sync"

	"github.com/mburakmmm/sky-lang/internal/ast"
	"github.com/mburakmmm/sky-lang/internal/lexer"
	"github.com/mburakmmm/sky-lang/internal/parser"
	"github.com/mburakmmm/sky-lang/internal/sema"
)

// Server LSP sunucusunu temsil eder
type Server struct {
	reader io.Reader
	writer io.Writer

	// Documents
	documents map[string]*Document
	docsMu    sync.RWMutex

	// Capabilities
	initialized bool

	// Logger
	logger *log.Logger
}

// Document açık bir dökümanı temsil eder
type Document struct {
	URI     string
	Version int
	Text    string
	AST     *ast.Program
	Symbols *sema.SymbolTable
	Errors  []Diagnostic
	mu      sync.RWMutex
}

// NewServer yeni bir LSP server oluşturur
func NewServer(reader io.Reader, writer io.Writer, logger *log.Logger) *Server {
	return &Server{
		reader:    reader,
		writer:    writer,
		documents: make(map[string]*Document),
		logger:    logger,
	}
}

// Start sunucuyu başlatır
func (s *Server) Start() error {
	scanner := bufio.NewScanner(s.reader)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024) // 1MB max

	for scanner.Scan() {
		line := scanner.Text()

		// LSP mesajları Content-Length header ile başlar
		if !strings.HasPrefix(line, "Content-Length: ") {
			continue
		}

		var contentLength int
		fmt.Sscanf(line, "Content-Length: %d", &contentLength)

		// Boş satırı atla
		scanner.Scan()

		// İçeriği oku
		content := make([]byte, contentLength)
		totalRead := 0
		for totalRead < contentLength {
			scanner.Scan()
			line := scanner.Text()
			copy(content[totalRead:], line)
			totalRead += len(line)
			if totalRead < contentLength {
				content[totalRead] = '\n'
				totalRead++
			}
		}

		// Mesajı parse et
		var req Request
		if err := json.Unmarshal(content[:contentLength], &req); err != nil {
			s.logError("Failed to parse request: %v", err)
			continue
		}

		// Handle request
		s.handleRequest(&req)
	}

	return scanner.Err()
}

// handleRequest request'i işler
func (s *Server) handleRequest(req *Request) {
	switch req.Method {
	case "initialize":
		s.handleInitialize(req)
	case "initialized":
		// Notification, no response needed
	case "shutdown":
		s.handleShutdown(req)
	case "textDocument/didOpen":
		s.handleDidOpen(req)
	case "textDocument/didChange":
		s.handleDidChange(req)
	case "textDocument/didClose":
		s.handleDidClose(req)
	case "textDocument/completion":
		s.handleCompletion(req)
	case "textDocument/hover":
		s.handleHover(req)
	case "textDocument/definition":
		s.handleDefinition(req)
	case "textDocument/references":
		s.handleReferences(req)
	case "textDocument/documentSymbol":
		s.handleDocumentSymbol(req)
	case "textDocument/formatting":
		s.handleFormatting(req)
	default:
		s.sendError(req.ID, MethodNotFound, fmt.Sprintf("Method not found: %s", req.Method))
	}
}

// handleInitialize initialize request
func (s *Server) handleInitialize(req *Request) {
	result := InitializeResult{
		Capabilities: ServerCapabilities{
			TextDocumentSync: 1, // Full sync
			CompletionProvider: &CompletionOptions{
				TriggerCharacters: []string{".", "("},
			},
			HoverProvider: true,
			SignatureHelpProvider: &SignatureHelpOptions{
				TriggerCharacters: []string{"(", ","},
			},
			DefinitionProvider:         true,
			ReferencesProvider:         true,
			DocumentSymbolProvider:     true,
			DocumentFormattingProvider: true,
		},
	}

	s.initialized = true
	s.sendResponse(req.ID, result)
}

// handleShutdown shutdown request
func (s *Server) handleShutdown(req *Request) {
	s.sendResponse(req.ID, nil)
}

// handleDidOpen didOpen notification
func (s *Server) handleDidOpen(req *Request) {
	var params struct {
		TextDocument TextDocumentItem `json:"textDocument"`
	}

	if err := json.Unmarshal(req.Params, &params); err != nil {
		s.logError("didOpen parse error: %v", err)
		return
	}

	doc := &Document{
		URI:     params.TextDocument.URI,
		Version: params.TextDocument.Version,
		Text:    params.TextDocument.Text,
	}

	s.docsMu.Lock()
	s.documents[params.TextDocument.URI] = doc
	s.docsMu.Unlock()

	// Parse and analyze
	s.analyzeDocument(doc)
}

// handleDidChange didChange notification
func (s *Server) handleDidChange(req *Request) {
	var params struct {
		TextDocument   VersionedTextDocumentIdentifier  `json:"textDocument"`
		ContentChanges []TextDocumentContentChangeEvent `json:"contentChanges"`
	}

	if err := json.Unmarshal(req.Params, &params); err != nil {
		s.logError("didChange parse error: %v", err)
		return
	}

	s.docsMu.RLock()
	doc, ok := s.documents[params.TextDocument.URI]
	s.docsMu.RUnlock()

	if !ok {
		return
	}

	doc.mu.Lock()
	// Full sync - replace entire document
	if len(params.ContentChanges) > 0 {
		doc.Text = params.ContentChanges[0].Text
		doc.Version = params.TextDocument.Version
	}
	doc.mu.Unlock()

	// Re-analyze
	s.analyzeDocument(doc)
}

// handleDidClose didClose notification
func (s *Server) handleDidClose(req *Request) {
	var params struct {
		TextDocument TextDocumentIdentifier `json:"textDocument"`
	}

	if err := json.Unmarshal(req.Params, &params); err != nil {
		return
	}

	s.docsMu.Lock()
	delete(s.documents, params.TextDocument.URI)
	s.docsMu.Unlock()
}

// analyzeDocument dokümanı analiz eder
func (s *Server) analyzeDocument(doc *Document) {
	doc.mu.Lock()
	defer doc.mu.Unlock()

	// Parse
	l := lexer.New(doc.Text, doc.URI)
	p := parser.New(l)
	doc.AST = p.ParseProgram()

	doc.Errors = make([]Diagnostic, 0)

	// Parser errors
	for _, err := range p.Errors() {
		doc.Errors = append(doc.Errors, Diagnostic{
			Range: Range{
				Start: Position{Line: 0, Character: 0},
				End:   Position{Line: 0, Character: 0},
			},
			Severity: SeverityError,
			Source:   "parser",
			Message:  err,
		})
	}

	// Semantic analysis
	checker := sema.NewChecker()
	semErrors := checker.Check(doc.AST)

	for _, err := range semErrors {
		if semErr, ok := err.(*sema.SemanticError); ok {
			doc.Errors = append(doc.Errors, Diagnostic{
				Range: Range{
					Start: Position{Line: semErr.Pos.Line - 1, Character: semErr.Pos.Column - 1},
					End:   Position{Line: semErr.Pos.Line - 1, Character: semErr.Pos.Column + 10},
				},
				Severity: SeverityError,
				Source:   "semantic",
				Message:  semErr.Message,
			})
		}
	}

	// Publish diagnostics
	s.publishDiagnostics(doc.URI, doc.Errors)
}

// handleCompletion completion request
func (s *Server) handleCompletion(req *Request) {
	var params struct {
		TextDocument TextDocumentIdentifier `json:"textDocument"`
		Position     Position               `json:"position"`
	}

	if err := json.Unmarshal(req.Params, &params); err != nil {
		s.sendError(req.ID, InvalidParams, "Invalid params")
		return
	}

	s.docsMu.RLock()
	doc, ok := s.documents[params.TextDocument.URI]
	s.docsMu.RUnlock()

	if !ok {
		s.sendResponse(req.ID, []CompletionItem{})
		return
	}

	items := s.getCompletionItems(doc, params.Position)
	s.sendResponse(req.ID, items)
}

// getCompletionItems completion items oluşturur
func (s *Server) getCompletionItems(doc *Document, pos Position) []CompletionItem {
	items := []CompletionItem{}

	// Keywords
	keywords := []string{
		"function", "end", "class", "let", "const", "if", "elif", "else",
		"for", "while", "return", "async", "await", "coop", "yield",
		"unsafe", "self", "super", "import", "as", "in", "true", "false",
	}

	for _, kw := range keywords {
		items = append(items, CompletionItem{
			Label:  kw,
			Kind:   CompletionKeyword,
			Detail: "keyword",
		})
	}

	// Built-in functions
	builtins := []struct {
		name   string
		detail string
	}{
		{"print", "print(value...)"},
		{"len", "len(collection) -> int"},
		{"range", "range(n) -> [int]"},
	}

	for _, builtin := range builtins {
		items = append(items, CompletionItem{
			Label:  builtin.name,
			Kind:   CompletionFunction,
			Detail: builtin.detail,
		})
	}

	// TODO: Add symbols from current scope

	return items
}

// handleHover hover request
func (s *Server) handleHover(req *Request) {
	var params struct {
		TextDocument TextDocumentIdentifier `json:"textDocument"`
		Position     Position               `json:"position"`
	}

	if err := json.Unmarshal(req.Params, &params); err != nil {
		s.sendError(req.ID, InvalidParams, "Invalid params")
		return
	}

	s.docsMu.RLock()
	doc, ok := s.documents[params.TextDocument.URI]
	s.docsMu.RUnlock()

	if !ok {
		s.sendResponse(req.ID, nil)
		return
	}

	hover := s.getHover(doc, params.Position)
	s.sendResponse(req.ID, hover)
}

// getHover hover bilgisi oluşturur
func (s *Server) getHover(doc *Document, pos Position) *Hover {
	// TODO: Implement actual hover info from AST
	return &Hover{
		Contents: "SKY Language Hover Info",
	}
}

// handleDefinition definition request
func (s *Server) handleDefinition(req *Request) {
	var params struct {
		TextDocument TextDocumentIdentifier `json:"textDocument"`
		Position     Position               `json:"position"`
	}

	if err := json.Unmarshal(req.Params, &params); err != nil {
		s.sendError(req.ID, InvalidParams, "Invalid params")
		return
	}

	s.sendResponse(req.ID, nil)
}

// handleReferences references request
func (s *Server) handleReferences(req *Request) {
	var params struct {
		TextDocument TextDocumentIdentifier `json:"textDocument"`
		Position     Position               `json:"position"`
		Context      ReferenceContext       `json:"context"`
	}

	if err := json.Unmarshal(req.Params, &params); err != nil {
		s.sendError(req.ID, InvalidParams, "Invalid params")
		return
	}

	s.sendResponse(req.ID, []Location{})
}

// handleDocumentSymbol document symbol request
func (s *Server) handleDocumentSymbol(req *Request) {
	var params struct {
		TextDocument TextDocumentIdentifier `json:"textDocument"`
	}

	if err := json.Unmarshal(req.Params, &params); err != nil {
		s.sendError(req.ID, InvalidParams, "Invalid params")
		return
	}

	s.docsMu.RLock()
	doc, ok := s.documents[params.TextDocument.URI]
	s.docsMu.RUnlock()

	if !ok {
		s.sendResponse(req.ID, []DocumentSymbol{})
		return
	}

	symbols := s.getDocumentSymbols(doc)
	s.sendResponse(req.ID, symbols)
}

// getDocumentSymbols doküman sembollerini toplar
func (s *Server) getDocumentSymbols(doc *Document) []DocumentSymbol {
	symbols := []DocumentSymbol{}

	doc.mu.RLock()
	defer doc.mu.RUnlock()

	if doc.AST == nil {
		return symbols
	}

	for _, stmt := range doc.AST.Statements {
		switch s := stmt.(type) {
		case *ast.FunctionStatement:
			symbols = append(symbols, DocumentSymbol{
				Name:   s.Name.Value,
				Kind:   SymbolFunction,
				Detail: "function",
				Range: Range{
					Start: Position{Line: s.Token.Line - 1, Character: s.Token.Column - 1},
					End:   Position{Line: s.Token.Line, Character: 0},
				},
				SelectionRange: Range{
					Start: Position{Line: s.Token.Line - 1, Character: s.Token.Column - 1},
					End:   Position{Line: s.Token.Line - 1, Character: s.Token.Column + len(s.Name.Value)},
				},
			})

		case *ast.ClassStatement:
			symbols = append(symbols, DocumentSymbol{
				Name:   s.Name.Value,
				Kind:   SymbolClass,
				Detail: "class",
				Range: Range{
					Start: Position{Line: s.Token.Line - 1, Character: s.Token.Column - 1},
					End:   Position{Line: s.Token.Line, Character: 0},
				},
				SelectionRange: Range{
					Start: Position{Line: s.Token.Line - 1, Character: s.Token.Column - 1},
					End:   Position{Line: s.Token.Line - 1, Character: s.Token.Column + len(s.Name.Value)},
				},
			})
		}
	}

	return symbols
}

// handleFormatting formatting request
func (s *Server) handleFormatting(req *Request) {
	var params struct {
		TextDocument TextDocumentIdentifier `json:"textDocument"`
		Options      FormattingOptions      `json:"options"`
	}

	if err := json.Unmarshal(req.Params, &params); err != nil {
		s.sendError(req.ID, InvalidParams, "Invalid params")
		return
	}

	// TODO: Implement formatting
	s.sendResponse(req.ID, []TextEdit{})
}

// publishDiagnostics diagnostics publish eder
func (s *Server) publishDiagnostics(uri string, diagnostics []Diagnostic) {
	notification := Notification{
		JSONRPC: "2.0",
		Method:  "textDocument/publishDiagnostics",
		Params: json.RawMessage(mustMarshal(map[string]interface{}{
			"uri":         uri,
			"diagnostics": diagnostics,
		})),
	}

	s.sendNotification(&notification)
}

// sendResponse response gönderir
func (s *Server) sendResponse(id interface{}, result interface{}) {
	resp := Response{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}

	data := mustMarshal(resp)
	s.writeMessage(data)
}

// sendError error response gönderir
func (s *Server) sendError(id interface{}, code int, message string) {
	resp := Response{
		JSONRPC: "2.0",
		ID:      id,
		Error: &RPCError{
			Code:    code,
			Message: message,
		},
	}

	data := mustMarshal(resp)
	s.writeMessage(data)
}

// sendNotification notification gönderir
func (s *Server) sendNotification(notification *Notification) {
	data := mustMarshal(notification)
	s.writeMessage(data)
}

// writeMessage LSP mesajı yazar
func (s *Server) writeMessage(data []byte) {
	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(data))
	s.writer.Write([]byte(header))
	s.writer.Write(data)
}

// logError hata loglar
func (s *Server) logError(format string, args ...interface{}) {
	if s.logger != nil {
		s.logger.Printf(format, args...)
	}
}

func mustMarshal(v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return data
}
