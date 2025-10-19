# LSP Implementation

## Overview

SKY Language Server implements **LSP 3.17** (Language Server Protocol) for editor integration.

## Implementation (`internal/lsp/`)

### Files

- `server.go` (568 lines) - Main LSP server
- `protocol.go` (233 lines) - Protocol types

## Protocol

### Transport

- **JSON-RPC 2.0** over stdio
- **Content-Length** headers for message framing

```
Content-Length: 123\r\n
\r\n
{"jsonrpc":"2.0","id":1,"method":"initialize",...}
```

## Supported Features

### Initialization

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "initialize",
  "params": {
    "processId": 1234,
    "rootUri": "file:///path/to/project",
    "capabilities": { ... }
  }
}
```

Response:
```json
{
  "capabilities": {
    "textDocumentSync": 1,
    "completionProvider": {
      "triggerCharacters": [".", "("]
    },
    "hoverProvider": true,
    "definitionProvider": true,
    ...
  }
}
```

### Text Synchronization

#### didOpen
```json
{
  "method": "textDocument/didOpen",
  "params": {
    "textDocument": {
      "uri": "file:///test.sky",
      "languageId": "sky",
      "version": 1,
      "text": "function main\n  print(\"hello\")\nend"
    }
  }
}
```

#### didChange
```json
{
  "method": "textDocument/didChange",
  "params": {
    "textDocument": {"uri": "...", "version": 2},
    "contentChanges": [
      {"text": "new content"}
    ]
  }
}
```

### Diagnostics

Automatic error publishing:

```json
{
  "method": "textDocument/publishDiagnostics",
  "params": {
    "uri": "file:///test.sky",
    "diagnostics": [
      {
        "range": {
          "start": {"line": 5, "character": 2},
          "end": {"line": 5, "character": 12}
        },
        "severity": 1,
        "source": "semantic",
        "message": "cannot assign to const variable 'PI'"
      }
    ]
  }
}
```

### Auto-Completion

```json
{
  "method": "textDocument/completion",
  "params": {
    "textDocument": {"uri": "..."},
    "position": {"line": 2, "character": 5}
  }
}
```

Response:
```json
{
  "result": [
    {
      "label": "print",
      "kind": 3,
      "detail": "print(value...)",
      "insertText": "print($1)"
    },
    {
      "label": "function",
      "kind": 14,
      "detail": "keyword"
    }
  ]
}
```

### Hover Information

```json
{
  "method": "textDocument/hover",
  "params": {
    "textDocument": {"uri": "..."},
    "position": {"line": 3, "character": 10}
  }
}
```

Response:
```json
{
  "result": {
    "contents": "function add(a: int, b: int): int",
    "range": {...}
  }
}
```

### Go to Definition

```json
{
  "method": "textDocument/definition",
  "params": {
    "textDocument": {"uri": "..."},
    "position": {"line": 10, "character": 5}
  }
}
```

### Find References

```json
{
  "method": "textDocument/references",
  "params": {
    "textDocument": {"uri": "..."},
    "position": {"line": 5, "character": 8},
    "context": {"includeDeclaration": true}
  }
}
```

### Document Symbols

```json
{
  "method": "textDocument/documentSymbol",
  "params": {
    "textDocument": {"uri": "..."}
  }
}
```

Response:
```json
{
  "result": [
    {
      "name": "main",
      "kind": 12,
      "range": {...},
      "selectionRange": {...}
    }
  ]
}
```

## Editor Setup

### VS Code

`.vscode/settings.json`:
```json
{
  "sky.languageServer": {
    "command": "skyls",
    "args": []
  }
}
```

### Vim (coc.nvim)

`:CocConfig`:
```json
{
  "languageserver": {
    "sky": {
      "command": "skyls",
      "filetypes": ["sky"],
      "rootPatterns": ["sky.project.json"]
    }
  }
}
```

### Emacs (lsp-mode)

```elisp
(add-to-list 'lsp-language-id-configuration '(sky-mode . "sky"))
(lsp-register-client
 (make-lsp-client :new-connection (lsp-stdio-connection "skyls")
                  :major-modes '(sky-mode)
                  :server-id 'skyls))
```

## Performance

- **Initialization**: <100ms
- **Parse + Analysis**: <50ms per file
- **Completion**: <10ms
- **Hover**: <5ms

## Implementation Details

### Document Management

Thread-safe document storage:

```go
type Document struct {
    URI     string
    Version int
    Text    string
    AST     *ast.Program
    Symbols *sema.SymbolTable
    Errors  []Diagnostic
    mu      sync.RWMutex
}
```

### Analysis Pipeline

```
Text → Lexer → Parser → AST → Sema → Diagnostics
```

Real-time on every change!

## Supported Clients

- ✅ VS Code
- ✅ Vim/Neovim (coc.nvim, nvim-lspconfig)
- ✅ Emacs (lsp-mode)
- ✅ Sublime Text (LSP package)
- ✅ Any LSP-compatible editor

## Future Features

- [ ] Code actions
- [ ] Refactoring
- [ ] Semantic tokens
- [ ] Call hierarchy
- [ ] Type hierarchy

