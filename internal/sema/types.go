package sema

import (
	"fmt"

	"github.com/mburakmmm/sky-lang/internal/ast"
)

// Type tipleri temsil eder
type Type interface {
	String() string
	Equals(Type) bool
	IsAssignableTo(Type) bool
}

// Temel tipler
var (
	IntType    = &BasicType{Name: "int"}
	FloatType  = &BasicType{Name: "float"}
	StringType = &BasicType{Name: "string"}
	BoolType   = &BasicType{Name: "bool"}
	AnyType    = &BasicType{Name: "any"}
	VoidType   = &BasicType{Name: "void"}
	NilType    = &BasicType{Name: "nil"}
)

// BasicType temel tipleri temsil eder
type BasicType struct {
	Name string
}

func (t *BasicType) String() string {
	return t.Name
}

func (t *BasicType) Equals(other Type) bool {
	if o, ok := other.(*BasicType); ok {
		return t.Name == o.Name
	}
	return false
}

func (t *BasicType) IsAssignableTo(target Type) bool {
	// any her şeyi kabul eder
	if target == AnyType || t == AnyType {
		return true
	}

	// Aynı tip ise
	if t.Equals(target) {
		return true
	}

	// int -> float otomatik dönüşüm
	if t == IntType && target == FloatType {
		return true
	}

	return false
}

// ListType list tipini temsil eder [T]
type ListType struct {
	ElementType Type
}

func (t *ListType) String() string {
	return fmt.Sprintf("[%s]", t.ElementType.String())
}

func (t *ListType) Equals(other Type) bool {
	if o, ok := other.(*ListType); ok {
		return t.ElementType.Equals(o.ElementType)
	}
	return false
}

func (t *ListType) IsAssignableTo(target Type) bool {
	if target == AnyType {
		return true
	}

	if o, ok := target.(*ListType); ok {
		return t.ElementType.IsAssignableTo(o.ElementType)
	}

	return false
}

// DictType dictionary tipini temsil eder {K: V}
type DictType struct {
	KeyType   Type
	ValueType Type
}

func (t *DictType) String() string {
	return fmt.Sprintf("{%s: %s}", t.KeyType.String(), t.ValueType.String())
}

func (t *DictType) Equals(other Type) bool {
	if o, ok := other.(*DictType); ok {
		return t.KeyType.Equals(o.KeyType) && t.ValueType.Equals(o.ValueType)
	}
	return false
}

func (t *DictType) IsAssignableTo(target Type) bool {
	if target == AnyType {
		return true
	}

	if o, ok := target.(*DictType); ok {
		return t.KeyType.IsAssignableTo(o.KeyType) && t.ValueType.IsAssignableTo(o.ValueType)
	}

	return false
}

// FunctionType fonksiyon tipini temsil eder (T1, T2) => T3
type FunctionType struct {
	Params     []Type
	ReturnType Type
	Variadic   bool
	MinParams  int // Minimum required parameters (for optional params)
}

func (t *FunctionType) String() string {
	params := ""
	for i, p := range t.Params {
		if i > 0 {
			params += ", "
		}
		params += p.String()
	}
	if t.Variadic {
		params += "..."
	}
	return fmt.Sprintf("(%s) => %s", params, t.ReturnType.String())
}

func (t *FunctionType) Equals(other Type) bool {
	if o, ok := other.(*FunctionType); ok {
		if len(t.Params) != len(o.Params) {
			return false
		}
		for i := range t.Params {
			if !t.Params[i].Equals(o.Params[i]) {
				return false
			}
		}
		return t.ReturnType.Equals(o.ReturnType) && t.Variadic == o.Variadic
	}
	return false
}

func (t *FunctionType) IsAssignableTo(target Type) bool {
	if target == AnyType {
		return true
	}

	if o, ok := target.(*FunctionType); ok {
		// Fonksiyon tiplerinin ataması için parametreler ve dönüş tipi uyumlu olmalı
		if len(t.Params) != len(o.Params) {
			return false
		}
		for i := range t.Params {
			if !t.Params[i].IsAssignableTo(o.Params[i]) {
				return false
			}
		}
		return t.ReturnType.IsAssignableTo(o.ReturnType)
	}

	return false
}

// PointerType pointer tipini temsil eder *T
type PointerType struct {
	PointeeType Type
}

func (t *PointerType) String() string {
	return fmt.Sprintf("*%s", t.PointeeType.String())
}

func (t *PointerType) Equals(other Type) bool {
	if o, ok := other.(*PointerType); ok {
		return t.PointeeType.Equals(o.PointeeType)
	}
	return false
}

func (t *PointerType) IsAssignableTo(target Type) bool {
	if target == AnyType {
		return true
	}

	if o, ok := target.(*PointerType); ok {
		return t.PointeeType.IsAssignableTo(o.PointeeType)
	}

	return false
}

// ClassType sınıf tipini temsil eder
type ClassType struct {
	Name       string
	SuperClass *ClassType
	Methods    map[string]*FunctionType
	Fields     map[string]Type
}

func (t *ClassType) String() string {
	return t.Name
}

func (t *ClassType) Equals(other Type) bool {
	if o, ok := other.(*ClassType); ok {
		return t.Name == o.Name
	}
	return false
}

func (t *ClassType) IsAssignableTo(target Type) bool {
	if target == AnyType {
		return true
	}

	if o, ok := target.(*ClassType); ok {
		// Aynı sınıf veya parent sınıf ise atanabilir
		current := t
		for current != nil {
			if current.Name == o.Name {
				return true
			}
			current = current.SuperClass
		}
	}

	return false
}

// ResolveType AST tip anotasyonunu gerçek Type'a dönüştürür
func ResolveType(typeAnnot ast.TypeAnnotation) Type {
	if typeAnnot == nil {
		return AnyType
	}

	switch t := typeAnnot.(type) {
	case *ast.BasicType:
		switch t.Name {
		case "int":
			return IntType
		case "float":
			return FloatType
		case "string":
			return StringType
		case "bool":
			return BoolType
		case "any":
			return AnyType
		case "void":
			return VoidType
		default:
			return AnyType
		}

	case *ast.ListType:
		elemType := ResolveType(t.ElementType)
		return &ListType{ElementType: elemType}

	case *ast.DictType:
		keyType := ResolveType(t.KeyType)
		valueType := ResolveType(t.ValueType)
		return &DictType{KeyType: keyType, ValueType: valueType}

	case *ast.FunctionType:
		params := make([]Type, len(t.ParamTypes))
		for i, p := range t.ParamTypes {
			params[i] = ResolveType(p)
		}
		returnType := ResolveType(t.ReturnType)
		return &FunctionType{Params: params, ReturnType: returnType}

	case *ast.PointerType:
		pointeeType := ResolveType(t.PointeeType)
		return &PointerType{PointeeType: pointeeType}

	default:
		return AnyType
	}
}

// InferType ifadenin tipini çıkarır
func InferType(expr ast.Expression, scope *Scope, symTable *SymbolTable) Type {
	switch e := expr.(type) {
	case *ast.IntegerLiteral:
		return IntType

	case *ast.FloatLiteral:
		return FloatType

	case *ast.StringLiteral:
		return StringType

	case *ast.BooleanLiteral:
		return BoolType

	case *ast.Identifier:
		if symbol, ok := symTable.Resolve(e.Value); ok {
			return symbol.Type
		}
		return AnyType

	case *ast.ListLiteral:
		if len(e.Elements) == 0 {
			return &ListType{ElementType: AnyType}
		}
		// İlk elemanın tipini al
		elemType := InferType(e.Elements[0], scope, symTable)
		return &ListType{ElementType: elemType}

	case *ast.InfixExpression:
		leftType := InferType(e.Left, scope, symTable)
		rightType := InferType(e.Right, scope, symTable)

		switch e.Operator {
		case "+", "-", "*", "/", "%":
			// Aritmetik operatörler
			if leftType == FloatType || rightType == FloatType {
				return FloatType
			}
			if leftType == IntType && rightType == IntType {
				return IntType
			}
			if leftType == StringType && rightType == StringType && e.Operator == "+" {
				return StringType
			}
			return AnyType

		case "==", "!=", "<", "<=", ">", ">=":
			// Karşılaştırma operatörleri
			return BoolType

		case "&&", "||":
			// Mantıksal operatörler
			return BoolType

		case "=", "+=", "-=", "*=", "/=", "%=":
			// Atama operatörleri
			return leftType

		default:
			return AnyType
		}

	case *ast.PrefixExpression:
		rightType := InferType(e.Right, scope, symTable)
		switch e.Operator {
		case "!":
			return BoolType
		case "-", "+":
			return rightType
		default:
			return AnyType
		}

	case *ast.CallExpression:
		// Fonksiyon çağrısı
		funcType := InferType(e.Function, scope, symTable)
		if ft, ok := funcType.(*FunctionType); ok {
			return ft.ReturnType
		}
		return AnyType

	case *ast.IndexExpression:
		leftType := InferType(e.Left, scope, symTable)
		if lt, ok := leftType.(*ListType); ok {
			return lt.ElementType
		}
		if dt, ok := leftType.(*DictType); ok {
			return dt.ValueType
		}
		return AnyType

	case *ast.MemberExpression:
		// Member access - şimdilik any döndür
		// İleride class type'ları için implement edilecek
		return AnyType

	default:
		return AnyType
	}
}
