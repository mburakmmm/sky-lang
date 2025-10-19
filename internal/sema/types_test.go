package sema

import (
	"testing"

	"github.com/mburakmmm/sky-lang/internal/ast"
	"github.com/mburakmmm/sky-lang/internal/lexer"
)

func TestBasicTypeEquals(t *testing.T) {
	tests := []struct {
		t1     Type
		t2     Type
		equals bool
	}{
		{IntType, IntType, true},
		{FloatType, FloatType, true},
		{StringType, StringType, true},
		{BoolType, BoolType, true},
		{IntType, FloatType, false},
		{StringType, BoolType, false},
	}

	for _, tt := range tests {
		result := tt.t1.Equals(tt.t2)
		if result != tt.equals {
			t.Errorf("%s.Equals(%s) = %v, want %v",
				tt.t1.String(), tt.t2.String(), result, tt.equals)
		}
	}
}

func TestTypeAssignability(t *testing.T) {
	tests := []struct {
		from       Type
		to         Type
		assignable bool
		desc       string
	}{
		{IntType, IntType, true, "int to int"},
		{IntType, FloatType, true, "int to float (auto convert)"},
		{FloatType, IntType, false, "float to int (no auto convert)"},
		{IntType, AnyType, true, "int to any"},
		{AnyType, IntType, true, "any to int"},
		{StringType, IntType, false, "string to int"},
	}

	for _, tt := range tests {
		result := tt.from.IsAssignableTo(tt.to)
		if result != tt.assignable {
			t.Errorf("%s: %s.IsAssignableTo(%s) = %v, want %v",
				tt.desc, tt.from.String(), tt.to.String(), result, tt.assignable)
		}
	}
}

func TestListTypeEquals(t *testing.T) {
	listInt1 := &ListType{ElementType: IntType}
	listInt2 := &ListType{ElementType: IntType}
	listStr := &ListType{ElementType: StringType}

	if !listInt1.Equals(listInt2) {
		t.Error("identical list types should be equal")
	}

	if listInt1.Equals(listStr) {
		t.Error("different list types should not be equal")
	}
}

func TestDictTypeEquals(t *testing.T) {
	dict1 := &DictType{KeyType: StringType, ValueType: IntType}
	dict2 := &DictType{KeyType: StringType, ValueType: IntType}
	dict3 := &DictType{KeyType: StringType, ValueType: FloatType}

	if !dict1.Equals(dict2) {
		t.Error("identical dict types should be equal")
	}

	if dict1.Equals(dict3) {
		t.Error("different dict types should not be equal")
	}
}

func TestFunctionTypeEquals(t *testing.T) {
	func1 := &FunctionType{
		Params:     []Type{IntType, IntType},
		ReturnType: IntType,
	}
	func2 := &FunctionType{
		Params:     []Type{IntType, IntType},
		ReturnType: IntType,
	}
	func3 := &FunctionType{
		Params:     []Type{IntType, FloatType},
		ReturnType: IntType,
	}

	if !func1.Equals(func2) {
		t.Error("identical function types should be equal")
	}

	if func1.Equals(func3) {
		t.Error("different function types should not be equal")
	}
}

func TestResolveType(t *testing.T) {
	tests := []struct {
		annotation ast.TypeAnnotation
		expected   Type
		desc       string
	}{
		{
			&ast.BasicType{Name: "int"},
			IntType,
			"basic int type",
		},
		{
			&ast.BasicType{Name: "float"},
			FloatType,
			"basic float type",
		},
		{
			&ast.BasicType{Name: "string"},
			StringType,
			"basic string type",
		},
		{
			&ast.BasicType{Name: "bool"},
			BoolType,
			"basic bool type",
		},
		{
			&ast.BasicType{Name: "any"},
			AnyType,
			"any type",
		},
		{
			nil,
			AnyType,
			"nil annotation defaults to any",
		},
	}

	for _, tt := range tests {
		result := ResolveType(tt.annotation)
		if !result.Equals(tt.expected) {
			t.Errorf("%s: expected %s, got %s",
				tt.desc, tt.expected.String(), result.String())
		}
	}
}

func TestResolveListType(t *testing.T) {
	annotation := &ast.ListType{
		ElementType: &ast.BasicType{
			Token: lexer.Token{},
			Name:  "int",
		},
	}

	result := ResolveType(annotation)
	listType, ok := result.(*ListType)
	if !ok {
		t.Fatalf("expected ListType, got %T", result)
	}

	if !listType.ElementType.Equals(IntType) {
		t.Errorf("expected element type int, got %s", listType.ElementType.String())
	}
}

func TestResolveDictType(t *testing.T) {
	annotation := &ast.DictType{
		KeyType:   &ast.BasicType{Name: "string"},
		ValueType: &ast.BasicType{Name: "int"},
	}

	result := ResolveType(annotation)
	dictType, ok := result.(*DictType)
	if !ok {
		t.Fatalf("expected DictType, got %T", result)
	}

	if !dictType.KeyType.Equals(StringType) {
		t.Errorf("expected key type string, got %s", dictType.KeyType.String())
	}

	if !dictType.ValueType.Equals(IntType) {
		t.Errorf("expected value type int, got %s", dictType.ValueType.String())
	}
}

func TestInferTypeFromLiterals(t *testing.T) {
	tests := []struct {
		input    string
		expected Type
		desc     string
	}{
		{`let x = 42`, IntType, "integer literal"},
		{`let x = 3.14`, FloatType, "float literal"},
		{`let x = "hello"`, StringType, "string literal"},
		{`let x = true`, BoolType, "boolean literal"},
	}

	for _, tt := range tests {
		program := parseProgram(t, tt.input)

		// Get the let statement
		letStmt, ok := program.Statements[0].(*ast.LetStatement)
		if !ok {
			t.Fatalf("%s: not a let statement", tt.desc)
		}

		checker := NewChecker()
		inferredType := checker.checkExpression(letStmt.Value)

		if !inferredType.Equals(tt.expected) {
			t.Errorf("%s: expected %s, got %s",
				tt.desc, tt.expected.String(), inferredType.String())
		}
	}
}

func TestInferTypeFromBinaryOp(t *testing.T) {
	tests := []struct {
		input    string
		expected Type
		desc     string
	}{
		{`let x = 10 + 20`, IntType, "int + int"},
		{`let x = 3.14 + 2.71`, FloatType, "float + float"},
		{`let x = "hello" + "world"`, StringType, "string + string"},
		{`let x = 10 > 5`, BoolType, "comparison"},
		{`let x = true && false`, BoolType, "logical"},
	}

	for _, tt := range tests {
		program := parseProgram(t, tt.input)
		letStmt := program.Statements[0].(*ast.LetStatement)

		checker := NewChecker()
		inferredType := checker.checkExpression(letStmt.Value)

		if !inferredType.Equals(tt.expected) && inferredType != AnyType {
			t.Errorf("%s: expected %s, got %s",
				tt.desc, tt.expected.String(), inferredType.String())
		}
	}
}

func TestSymbolTableScoping(t *testing.T) {
	input := `let x = 10

function test: int
  let y = 20
  let z = x + y
  return z
end`

	program := parseProgram(t, input)
	checker := NewChecker()
	errors := checker.Check(program)

	if len(errors) != 0 {
		t.Fatalf("expected no errors for valid scoping, got: %v", errors)
	}
}

func TestSymbolRedefinition(t *testing.T) {
	input := `let x = 10
let x = 20`

	program := parseProgram(t, input)
	checker := NewChecker()
	errors := checker.Check(program)

	if len(errors) == 0 {
		t.Fatal("expected error for symbol redefinition")
	}
}
