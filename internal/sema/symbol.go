package sema

import (
	"github.com/mburakmmm/sky-lang/internal/ast"
	"github.com/mburakmmm/sky-lang/internal/lexer"
)

// SymbolKind sembol tipini belirtir
type SymbolKind int

const (
	VariableSymbol SymbolKind = iota
	ConstantSymbol
	FunctionSymbol
	ParameterSymbol
	ClassSymbol
)

func (sk SymbolKind) String() string {
	switch sk {
	case VariableSymbol:
		return "variable"
	case ConstantSymbol:
		return "constant"
	case FunctionSymbol:
		return "function"
	case ParameterSymbol:
		return "parameter"
	case ClassSymbol:
		return "class"
	default:
		return "unknown"
	}
}

// Symbol bir değişken, fonksiyon veya sınıfı temsil eder
type Symbol struct {
	Name    string
	Kind    SymbolKind
	Type    Type
	Pos     lexer.Token
	IsAsync bool // fonksiyonlar için
	Mutable bool // true = let, false = const
	Scope   *Scope
	Node    ast.Node // tanımın AST düğümü
}

// Scope değişken ve fonksiyon kapsam alanını temsil eder
type Scope struct {
	parent   *Scope
	symbols  map[string]*Symbol
	children []*Scope
	depth    int
}

// NewScope yeni bir scope oluşturur
func NewScope(parent *Scope) *Scope {
	depth := 0
	if parent != nil {
		depth = parent.depth + 1
	}

	return &Scope{
		parent:   parent,
		symbols:  make(map[string]*Symbol),
		children: []*Scope{},
		depth:    depth,
	}
}

// Define scope'a yeni bir sembol ekler
func (s *Scope) Define(symbol *Symbol) error {
	if _, exists := s.symbols[symbol.Name]; exists {
		return &SemanticError{
			Message: "symbol '" + symbol.Name + "' already defined in this scope",
			Pos:     symbol.Pos,
		}
	}

	symbol.Scope = s
	s.symbols[symbol.Name] = symbol
	return nil
}

// Resolve sembolu bu scope veya parent scope'larda arar
func (s *Scope) Resolve(name string) (*Symbol, bool) {
	// Önce bu scope'ta ara
	if symbol, ok := s.symbols[name]; ok {
		return symbol, true
	}

	// Parent scope'ta ara
	if s.parent != nil {
		return s.parent.Resolve(name)
	}

	return nil, false
}

// ResolveLocal sadece bu scope'ta arar (parent'larda aramaz)
func (s *Scope) ResolveLocal(name string) (*Symbol, bool) {
	symbol, ok := s.symbols[name]
	return symbol, ok
}

// NewChild yeni bir child scope oluşturur
func (s *Scope) NewChild() *Scope {
	child := NewScope(s)
	s.children = append(s.children, child)
	return child
}

// Parent parent scope'u döndürür
func (s *Scope) Parent() *Scope {
	return s.parent
}

// Symbols bu scope'taki tüm sembolleri döndürür
func (s *Scope) Symbols() map[string]*Symbol {
	return s.symbols
}

// Depth scope derinliğini döndürür (0 = global)
func (s *Scope) Depth() int {
	return s.depth
}

// SemanticError semantik analiz hatalarını temsil eder
type SemanticError struct {
	Message string
	Pos     lexer.Token
}

func (e *SemanticError) Error() string {
	if e.Pos.File != "" {
		return e.Pos.Position() + ": " + e.Message
	}
	return e.Message
}

// SymbolTable global sembol tablosu yöneticisi
type SymbolTable struct {
	globalScope  *Scope
	currentScope *Scope
}

// NewSymbolTable yeni bir sembol tablosu oluşturur
func NewSymbolTable() *SymbolTable {
	globalScope := NewScope(nil)

	// Built-in fonksiyonları ekle
	builtins := []struct {
		name string
		typ  Type
	}{
		// Core builtins
		{"print", &FunctionType{Params: []Type{AnyType}, ReturnType: VoidType, Variadic: true}},
		{"len", &FunctionType{Params: []Type{AnyType}, ReturnType: IntType}},
		{"range", &FunctionType{Params: []Type{IntType}, ReturnType: &ListType{ElementType: IntType}}},

		// Type conversions
		{"int", &FunctionType{Params: []Type{AnyType}, ReturnType: IntType}},
		{"float", &FunctionType{Params: []Type{AnyType}, ReturnType: FloatType}},
		{"str", &FunctionType{Params: []Type{AnyType}, ReturnType: StringType}},
		{"bool", &FunctionType{Params: []Type{AnyType}, ReturnType: BoolType}},
		{"list", &FunctionType{Params: []Type{AnyType}, ReturnType: AnyType}},
		{"dict", &FunctionType{Params: []Type{AnyType}, ReturnType: AnyType}},

		// Utilities
		{"type", &FunctionType{Params: []Type{AnyType}, ReturnType: StringType}},
		{"isinstance", &FunctionType{Params: []Type{AnyType, AnyType}, ReturnType: BoolType}},
		{"input", &FunctionType{Params: []Type{AnyType}, ReturnType: StringType}},

		// Functional
		{"sum", &FunctionType{Params: []Type{&ListType{ElementType: AnyType}}, ReturnType: AnyType}},
		{"map", &FunctionType{Params: []Type{AnyType, AnyType}, ReturnType: AnyType}},
		{"filter", &FunctionType{Params: []Type{AnyType, AnyType}, ReturnType: AnyType}},
		{"any", &FunctionType{Params: []Type{AnyType}, ReturnType: BoolType}},
		{"all", &FunctionType{Params: []Type{AnyType}, ReturnType: BoolType}},

		// File System
		{"fs_exists", &FunctionType{Params: []Type{StringType}, ReturnType: BoolType}},
		{"fs_read_text", &FunctionType{Params: []Type{StringType}, ReturnType: StringType}},
		{"fs_write_text", &FunctionType{Params: []Type{StringType, StringType}, ReturnType: BoolType}},

		// OS
		{"os_platform", &FunctionType{Params: []Type{}, ReturnType: StringType}},
		{"os_getcwd", &FunctionType{Params: []Type{}, ReturnType: StringType}},
		{"os_getenv", &FunctionType{Params: []Type{StringType}, ReturnType: StringType}},

		// Crypto
		{"crypto_md5", &FunctionType{Params: []Type{StringType}, ReturnType: StringType}},
		{"crypto_sha256", &FunctionType{Params: []Type{StringType}, ReturnType: StringType}},

		// JSON
		{"json_encode", &FunctionType{Params: []Type{AnyType}, ReturnType: StringType}},
		{"json_decode", &FunctionType{Params: []Type{StringType}, ReturnType: AnyType}},

		// Time
		{"time_now", &FunctionType{Params: []Type{}, ReturnType: IntType}},
		{"time_sleep", &FunctionType{Params: []Type{IntType}, ReturnType: VoidType}},

		// Random
		{"rand_int", &FunctionType{Params: []Type{IntType}, ReturnType: IntType}},
		{"rand_uuid", &FunctionType{Params: []Type{}, ReturnType: StringType}},

		// Global utilities
		{"join", &FunctionType{Params: []Type{StringType, &ListType{ElementType: AnyType}}, ReturnType: StringType}},
		{"null", NilType},
		{"nil", NilType},

		// FS Module Functions
		{"fs_read_text", &FunctionType{Params: []Type{StringType}, ReturnType: StringType}},
		{"fs_write_text", &FunctionType{Params: []Type{StringType, StringType}, ReturnType: BoolType}},
		{"fs_exists", &FunctionType{Params: []Type{StringType}, ReturnType: BoolType}},
		{"fs_mkdir", &FunctionType{Params: []Type{StringType}, ReturnType: BoolType}},
		{"fs_list_dir", &FunctionType{Params: []Type{StringType}, ReturnType: &ListType{ElementType: StringType}}},

		// OS Module Functions
		{"os_platform", &FunctionType{Params: []Type{}, ReturnType: StringType}},
		{"os_getcwd", &FunctionType{Params: []Type{}, ReturnType: StringType}},
		{"os_getenv", &FunctionType{Params: []Type{StringType}, ReturnType: StringType}},
		{"os_setenv", &FunctionType{Params: []Type{StringType, StringType}, ReturnType: BoolType}},

		// Time Module Functions
		{"time_format", &FunctionType{Params: []Type{IntType, StringType}, ReturnType: StringType}},
		{"time_parse", &FunctionType{Params: []Type{StringType, StringType}, ReturnType: IntType}},
		{"time_add", &FunctionType{Params: []Type{IntType, IntType, StringType}, ReturnType: IntType}},
		{"time_diff", &FunctionType{Params: []Type{IntType, IntType}, ReturnType: IntType}},

		// Promise Functions
		{"Promise_all", &FunctionType{Params: []Type{&ListType{ElementType: AnyType}}, ReturnType: &ListType{ElementType: AnyType}}},
		{"Promise_allSettled", &FunctionType{Params: []Type{&ListType{ElementType: AnyType}}, ReturnType: &ListType{ElementType: AnyType}}},

		// HTTP Module Functions
		{"http_get", &FunctionType{Params: []Type{StringType}, ReturnType: StringType}},
		{"http_post", &FunctionType{Params: []Type{StringType, StringType}, ReturnType: StringType}},
		{"http_put", &FunctionType{Params: []Type{StringType, StringType}, ReturnType: StringType}},
		{"http_delete", &FunctionType{Params: []Type{StringType}, ReturnType: StringType}},

		// Crypto Module Functions
		{"crypto_md5", &FunctionType{Params: []Type{StringType}, ReturnType: StringType}},
		{"crypto_sha256", &FunctionType{Params: []Type{StringType}, ReturnType: StringType}},
		{"crypto_aes_encrypt", &FunctionType{Params: []Type{StringType, StringType}, ReturnType: StringType}},
	}

	for _, builtin := range builtins {
		kind := FunctionSymbol
		if builtin.name == "null" || builtin.name == "nil" {
			kind = VariableSymbol
		}
		globalScope.Define(&Symbol{
			Name: builtin.name,
			Kind: kind,
			Type: builtin.typ,
		})
	}

	return &SymbolTable{
		globalScope:  globalScope,
		currentScope: globalScope,
	}
}

// EnterScope yeni bir scope'a girer
func (st *SymbolTable) EnterScope() {
	st.currentScope = st.currentScope.NewChild()
}

// ExitScope mevcut scope'tan çıkar
func (st *SymbolTable) ExitScope() {
	if st.currentScope.parent != nil {
		st.currentScope = st.currentScope.parent
	}
}

// Define mevcut scope'a yeni bir sembol ekler
func (st *SymbolTable) Define(symbol *Symbol) error {
	return st.currentScope.Define(symbol)
}

// Resolve sembolu arar
func (st *SymbolTable) Resolve(name string) (*Symbol, bool) {
	return st.currentScope.Resolve(name)
}

// CurrentScope mevcut scope'u döndürür
func (st *SymbolTable) CurrentScope() *Scope {
	return st.currentScope
}

// GlobalScope global scope'u döndürür
func (st *SymbolTable) GlobalScope() *Scope {
	return st.globalScope
}
