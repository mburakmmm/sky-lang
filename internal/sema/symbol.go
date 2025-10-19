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
		{"print", &FunctionType{
			Params:     []Type{AnyType},
			ReturnType: VoidType,
			Variadic:   true,
		}},
		{"len", &FunctionType{
			Params:     []Type{AnyType},
			ReturnType: IntType,
		}},
		{"range", &FunctionType{
			Params:     []Type{IntType},
			ReturnType: &ListType{ElementType: IntType},
		}},
	}

	for _, builtin := range builtins {
		globalScope.Define(&Symbol{
			Name: builtin.name,
			Kind: FunctionSymbol,
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
