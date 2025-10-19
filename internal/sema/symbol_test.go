package sema

import (
	"testing"

	"github.com/mburakmmm/sky-lang/internal/lexer"
)

func TestNewScope(t *testing.T) {
	scope := NewScope(nil)

	if scope == nil {
		t.Fatal("NewScope returned nil")
	}

	if scope.parent != nil {
		t.Error("global scope should have no parent")
	}

	if scope.depth != 0 {
		t.Errorf("global scope depth should be 0, got %d", scope.depth)
	}
}

func TestScopeDefine(t *testing.T) {
	scope := NewScope(nil)
	symbol := &Symbol{
		Name: "x",
		Kind: VariableSymbol,
		Type: IntType,
		Pos:  lexer.Token{Line: 1, Column: 1},
	}

	err := scope.Define(symbol)
	if err != nil {
		t.Fatalf("Define failed: %v", err)
	}

	// Try to define again - should error
	err = scope.Define(symbol)
	if err == nil {
		t.Error("expected error for redefinition")
	}
}

func TestScopeResolve(t *testing.T) {
	parent := NewScope(nil)
	child := parent.NewChild()

	// Define in parent
	parentSym := &Symbol{
		Name: "x",
		Kind: VariableSymbol,
		Type: IntType,
		Pos:  lexer.Token{},
	}
	parent.Define(parentSym)

	// Resolve from child
	resolved, ok := child.Resolve("x")
	if !ok {
		t.Fatal("failed to resolve symbol from parent scope")
	}

	if resolved.Name != "x" {
		t.Errorf("resolved wrong symbol: %s", resolved.Name)
	}
}

func TestScopeResolveLocal(t *testing.T) {
	parent := NewScope(nil)
	child := parent.NewChild()

	// Define in parent
	parentSym := &Symbol{
		Name: "x",
		Kind: VariableSymbol,
		Type: IntType,
		Pos:  lexer.Token{},
	}
	parent.Define(parentSym)

	// ResolveLocal from child should not find it
	_, ok := child.ResolveLocal("x")
	if ok {
		t.Error("ResolveLocal should not find parent symbols")
	}

	// But Resolve should find it
	_, ok = child.Resolve("x")
	if !ok {
		t.Error("Resolve should find parent symbols")
	}
}

func TestNestedScopes(t *testing.T) {
	global := NewScope(nil)
	func1 := global.NewChild()
	func2 := func1.NewChild()

	if global.depth != 0 {
		t.Errorf("global depth should be 0, got %d", global.depth)
	}

	if func1.depth != 1 {
		t.Errorf("func1 depth should be 1, got %d", func1.depth)
	}

	if func2.depth != 2 {
		t.Errorf("func2 depth should be 2, got %d", func2.depth)
	}

	// Define at each level
	global.Define(&Symbol{Name: "a", Kind: VariableSymbol, Type: IntType})
	func1.Define(&Symbol{Name: "b", Kind: VariableSymbol, Type: IntType})
	func2.Define(&Symbol{Name: "c", Kind: VariableSymbol, Type: IntType})

	// Resolve from func2
	if _, ok := func2.Resolve("a"); !ok {
		t.Error("should resolve global variable")
	}
	if _, ok := func2.Resolve("b"); !ok {
		t.Error("should resolve parent variable")
	}
	if _, ok := func2.Resolve("c"); !ok {
		t.Error("should resolve local variable")
	}
}

func TestSymbolTable(t *testing.T) {
	st := NewSymbolTable()

	if st.globalScope == nil {
		t.Fatal("symbol table should have global scope")
	}

	if st.currentScope != st.globalScope {
		t.Error("current scope should start as global")
	}
}

func TestSymbolTableEnterExitScope(t *testing.T) {
	st := NewSymbolTable()

	initialScope := st.currentScope

	st.EnterScope()
	if st.currentScope == initialScope {
		t.Error("EnterScope should create new scope")
	}

	st.ExitScope()
	if st.currentScope != initialScope {
		t.Error("ExitScope should return to previous scope")
	}
}

func TestSymbolTableBuiltins(t *testing.T) {
	st := NewSymbolTable()

	// Check built-in functions
	builtins := []string{"print", "len", "range"}

	for _, name := range builtins {
		symbol, ok := st.Resolve(name)
		if !ok {
			t.Errorf("built-in function %s not found", name)
		}

		if symbol.Kind != FunctionSymbol {
			t.Errorf("%s should be a function, got %s", name, symbol.Kind.String())
		}
	}
}

func TestSymbolKindString(t *testing.T) {
	tests := []struct {
		kind     SymbolKind
		expected string
	}{
		{VariableSymbol, "variable"},
		{ConstantSymbol, "constant"},
		{FunctionSymbol, "function"},
		{ParameterSymbol, "parameter"},
		{ClassSymbol, "class"},
	}

	for _, tt := range tests {
		result := tt.kind.String()
		if result != tt.expected {
			t.Errorf("SymbolKind(%d).String() = %s, want %s",
				tt.kind, result, tt.expected)
		}
	}
}

func TestSemanticError(t *testing.T) {
	err := &SemanticError{
		Message: "test error",
		Pos:     lexer.Token{File: "test.sky", Line: 10, Column: 5},
	}

	errStr := err.Error()
	if !contains(errStr, "test error") {
		t.Errorf("error string should contain message: %s", errStr)
	}

	if !contains(errStr, "test.sky") {
		t.Errorf("error string should contain filename: %s", errStr)
	}
}

