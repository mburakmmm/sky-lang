package interpreter

import (
	"github.com/mburakmmm/sky-lang/internal/ast"
)

// Note: Enum and Match implementation requires deep integration with interpreter
// This is a placeholder showing the structure
// Full implementation requires:
// 1. Extending Value interface/types in value.go
// 2. Adding enum evaluation in main interpreter loop
// 3. Pattern matching logic

// EnumTypeInfo stores enum type metadata
type EnumTypeInfo struct {
	Name     string
	Variants map[string]*VariantInfo
}

// VariantInfo stores variant metadata
type VariantInfo struct {
	Name         string
	PayloadCount int
}

// EnumInstance represents a runtime enum value
type EnumInstance struct {
	TypeName string
	Variant  string
	Payload  []interface{}
}

// TODO: Full implementation requires modifying value.go and interpreter.go
// to add proper enum support throughout the entire system

// Placeholder functions (not integrated yet)
func evalEnumStatementPlaceholder(stmt *ast.EnumStatement) {
	// Register enum type
	_ = &EnumTypeInfo{
		Name:     stmt.Name.Value,
		Variants: make(map[string]*VariantInfo),
	}
}

func evalMatchExpressionPlaceholder(expr *ast.MatchExpression) {
	// Pattern matching logic
	_ = &EnumInstance{
		TypeName: "Example",
		Variant:  "Some",
		Payload:  []interface{}{},
	}
}
