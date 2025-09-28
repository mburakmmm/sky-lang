# Changelog

All notable changes to Sky will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial implementation of Sky programming language
- **String Interpolation**: Dart-style (`$var`, `${expr}`) and Python f-string (`f"{expr}"`) support
- Lexer with Unicode identifier support and indentation handling
- Parser with AST generation
- Type system with mandatory type declarations
- Binder for symbol resolution and scoping
- Bytecode compiler and VM
- Garbage collection (Mark-Sweep)
- Async/await support with event loop
- Coroutines (coop/yield) with cooperative multitasking
- Python bridge using pyo3
- JavaScript bridge using QuickJS
- Standard library (io, time, http, async, coop, prelude)
- CLI with run, repl, fmt, parse, lex, compile, check commands
- Code formatter with indentation standardization
- VSCode syntax highlighting
- Basic LSP server
- Comprehensive documentation and examples

### Changed
- N/A (Initial release)

### Deprecated
- N/A (Initial release)

### Removed
- N/A (Initial release)

### Fixed
- N/A (Initial release)

### Security
- N/A (Initial release)

## [0.1.0] - 2024-01-01

### Added
- **Core Language Features**
  - Python-like indented syntax
  - Mandatory type declarations (`var`, `int`, `float`, `bool`, `string`, `list`, `map`)
  - Dynamic & strongly typed runtime with type checking
  - Unicode identifier support (Turkish characters: ç, ğ, ı, ö, ş, ü)
  
- **Concurrency**
  - Async/await with single-threaded event loop
  - Coroutines (coop/yield) with cooperative multitasking
  - Future and Coroutine value types
  
- **VM and Runtime**
  - Stack-based bytecode interpreter
  - Mark-Sweep garbage collection
  - String interning
  - Call frames and stack management
  
- **Bridges**
  - Python bridge using pyo3 (CPython embedded)
  - JavaScript bridge using QuickJS
  - Type conversion between Sky and Python/JS
  
- **Standard Library**
  - `io.print()` - Output functions
  - `time.now()` - Time utilities
  - `http.get()` - HTTP client (mock implementation)
  - `async.sleep()` - Async sleep function
  - `coop.*` - Coroutine helper functions
  - `string.*` - String utilities (stringify, concat, split, join, etc.)
  - Prelude with global functions
  
- **Tooling**
  - CLI with comprehensive commands
  - Code formatter with indentation rules
  - VSCode syntax highlighting
  - Basic LSP server with diagnostics and hover
  - Comprehensive test suite
  
- **Documentation**
  - Complete grammar specification
  - Coroutine documentation with examples
  - Bridge documentation with usage examples
  - String interpolation documentation with examples
  - README with quick start guide
  - Example programs demonstrating all features

### Technical Details
- **Lexer**: chumsky-based with INDENT/DEDENT tokens
- **Parser**: Pratt parser with operator precedence
- **VM**: Stack-based with frame management
- **GC**: Single-threaded Mark-Sweep
- **Error System**: Positioned diagnostics with error codes
- **Build System**: Rust with Cargo
- **CI/CD**: GitHub Actions with multi-platform builds

### Error Codes
- **E0001**: Missing type annotation
- **E0101**: Invalid indentation (tabs or inconsistent)
- **E0201**: await outside async function
- **E0202**: yield outside coop function
- **E1001**: Type mismatch at runtime
- **E2001**: Coroutine already finished
- **E3001**: Python bridge error
- **E3002**: JavaScript bridge error
- **E4101**: Unterminated interpolation: missing '}'
- **E4102**: Invalid identifier after '$'
- **E4103**: Nested interpolation not allowed
- **E4104**: F-string: stray '}' or '{' (use '{{' or '}}')

### Example Programs
- `hello.sky` - Basic language features
- `async.sky` - Async/await examples
- `coop_basic.sky` - Coroutine examples
- `py_bridge.sky` - Python bridge examples
- `js_bridge.sky` - JavaScript bridge examples
- `strings.sky` - String interpolation examples

### CLI Commands
- `sky run <file>` - Execute Sky program
- `sky repl` - Interactive REPL
- `sky fmt <file>` - Format code
- `sky parse <file>` - Parse and show AST
- `sky lex <file>` - Tokenize and show tokens
- `sky compile <file>` - Compile to bytecode
- `sky check <file>` - Syntax and type checking
- `sky version` - Show version information
- `sky about` - Show about information

### Supported Platforms
- Linux x64
- Windows x64
- macOS x64

### Dependencies
- Rust 1.79+
- Python 3.8+ (for Python bridge)
- Node.js (for JavaScript bridge)

---

## Version History

- **v0.1.0** - Initial release with core language features, async/await, coroutines, bridges, and comprehensive tooling
