package sema

import (
	"fmt"

	"github.com/mburakmmm/sky-lang/internal/ast"
)

// Checker semantik analiz yapar
type Checker struct {
	symTable *SymbolTable
	errors   []error

	// Mevcut fonksiyon tipi (return type kontrolü için)
	currentFunction *Symbol

	// Loop içinde miyiz? (break/continue kontrolü için)
	inLoop int
}

// NewChecker yeni bir checker oluşturur
func NewChecker() *Checker {
	return &Checker{
		symTable: NewSymbolTable(),
		errors:   []error{},
		inLoop:   0,
	}
}

// Check programı analiz eder
func (c *Checker) Check(program *ast.Program) []error {
	for _, stmt := range program.Statements {
		c.checkStatement(stmt)
	}
	return c.errors
}

// Errors hataları döndürür
func (c *Checker) Errors() []error {
	return c.errors
}

func (c *Checker) addError(err error) {
	c.errors = append(c.errors, err)
}

func (c *Checker) checkStatement(stmt ast.Statement) {
	if stmt == nil {
		return
	}

	switch s := stmt.(type) {
	case *ast.LetStatement:
		c.checkLetStatement(s)
	case *ast.ConstStatement:
		c.checkConstStatement(s)
	case *ast.ReturnStatement:
		c.checkReturnStatement(s)
	case *ast.ExpressionStatement:
		c.checkExpression(s.Expression)
	case *ast.FunctionStatement:
		c.checkFunctionStatement(s)
	case *ast.IfStatement:
		c.checkIfStatement(s)
	case *ast.WhileStatement:
		c.checkWhileStatement(s)
	case *ast.ForStatement:
		c.checkForStatement(s)
	case *ast.ClassStatement:
		c.checkClassStatement(s)
	case *ast.ImportStatement:
		// Import statements are handled at a higher level
	case *ast.UnsafeStatement:
		c.checkUnsafeStatement(s)
	case *ast.BlockStatement:
		c.checkBlockStatement(s)
	}
}

func (c *Checker) checkLetStatement(stmt *ast.LetStatement) {
	// Değerin tipini çıkar
	var valueType Type = AnyType
	if stmt.Value != nil {
		valueType = c.checkExpression(stmt.Value)
	}

	// Tip anotasyonu varsa kontrol et
	var declaredType Type = AnyType
	if stmt.Type != nil {
		declaredType = ResolveType(stmt.Type)
	} else {
		declaredType = valueType
	}

	// Tip uyumluluğu kontrolü
	if stmt.Type != nil && !valueType.IsAssignableTo(declaredType) {
		c.addError(&SemanticError{
			Message: fmt.Sprintf("type mismatch: cannot assign %s to %s",
				valueType.String(), declaredType.String()),
			Pos: stmt.Token,
		})
	}

	// Sembole ekle
	symbol := &Symbol{
		Name:    stmt.Name.Value,
		Kind:    VariableSymbol,
		Type:    declaredType,
		Pos:     stmt.Token,
		Mutable: true,
		Node:    stmt,
	}

	if err := c.symTable.Define(symbol); err != nil {
		c.addError(err)
	}
}

func (c *Checker) checkConstStatement(stmt *ast.ConstStatement) {
	// Const değeri olmalı
	if stmt.Value == nil {
		c.addError(&SemanticError{
			Message: "const declaration must have a value",
			Pos:     stmt.Token,
		})
		return
	}

	// Değerin tipini çıkar
	valueType := c.checkExpression(stmt.Value)

	// Tip anotasyonu varsa kontrol et
	var declaredType Type = AnyType
	if stmt.Type != nil {
		declaredType = ResolveType(stmt.Type)
	} else {
		declaredType = valueType
	}

	// Tip uyumluluğu kontrolü
	if stmt.Type != nil && !valueType.IsAssignableTo(declaredType) {
		c.addError(&SemanticError{
			Message: fmt.Sprintf("type mismatch: cannot assign %s to %s",
				valueType.String(), declaredType.String()),
			Pos: stmt.Token,
		})
	}

	// Sembole ekle
	symbol := &Symbol{
		Name:    stmt.Name.Value,
		Kind:    ConstantSymbol,
		Type:    declaredType,
		Pos:     stmt.Token,
		Mutable: false, // const değiştirilemez
		Node:    stmt,
	}

	if err := c.symTable.Define(symbol); err != nil {
		c.addError(err)
	}
}

func (c *Checker) checkReturnStatement(stmt *ast.ReturnStatement) {
	// Fonksiyon içinde miyiz kontrol et
	if c.currentFunction == nil {
		c.addError(&SemanticError{
			Message: "return statement outside of function",
			Pos:     stmt.Token,
		})
		return
	}

	// Return değerinin tipi
	var returnType Type = VoidType
	if stmt.ReturnValue != nil {
		returnType = c.checkExpression(stmt.ReturnValue)
	}

	// Fonksiyonun return type'ı ile karşılaştır
	funcType, ok := c.currentFunction.Type.(*FunctionType)
	if ok {
		if !returnType.IsAssignableTo(funcType.ReturnType) {
			c.addError(&SemanticError{
				Message: fmt.Sprintf("return type mismatch: expected %s, got %s",
					funcType.ReturnType.String(), returnType.String()),
				Pos: stmt.Token,
			})
		}
	}
}

func (c *Checker) checkFunctionStatement(stmt *ast.FunctionStatement) {
	// Parametrelerin tiplerini al
	paramTypes := make([]Type, len(stmt.Parameters))
	for i, param := range stmt.Parameters {
		if param.Type != nil {
			paramTypes[i] = ResolveType(param.Type)
		} else {
			paramTypes[i] = AnyType
		}
	}

	// Return type'ı al
	var returnType Type = VoidType
	if stmt.ReturnType != nil {
		returnType = ResolveType(stmt.ReturnType)
	}

	// Fonksiyon tipini oluştur
	funcType := &FunctionType{
		Params:     paramTypes,
		ReturnType: returnType,
	}

	// Fonksiyonu sembole ekle
	funcSymbol := &Symbol{
		Name:    stmt.Name.Value,
		Kind:    FunctionSymbol,
		Type:    funcType,
		Pos:     stmt.Token,
		IsAsync: stmt.Async,
		Node:    stmt,
	}

	if err := c.symTable.Define(funcSymbol); err != nil {
		c.addError(err)
		return
	}

	// Fonksiyon body'sini kontrol et (yeni scope)
	c.symTable.EnterScope()
	oldFunction := c.currentFunction
	c.currentFunction = funcSymbol

	// Parametreleri scope'a ekle
	for i, param := range stmt.Parameters {
		paramSymbol := &Symbol{
			Name:    param.Name.Value,
			Kind:    ParameterSymbol,
			Type:    paramTypes[i],
			Pos:     param.Token,
			Mutable: true,
		}
		if err := c.symTable.Define(paramSymbol); err != nil {
			c.addError(err)
		}
	}

	// Body'yi kontrol et
	if stmt.Body != nil {
		c.checkBlockStatement(stmt.Body)
	}

	c.currentFunction = oldFunction
	c.symTable.ExitScope()
}

func (c *Checker) checkIfStatement(stmt *ast.IfStatement) {
	// Condition'ı kontrol et
	condType := c.checkExpression(stmt.Condition)
	if condType != BoolType && condType != AnyType {
		c.addError(&SemanticError{
			Message: fmt.Sprintf("if condition must be bool, got %s", condType.String()),
			Pos:     stmt.Token,
		})
	}

	// Consequence bloğunu kontrol et
	c.symTable.EnterScope()
	c.checkBlockStatement(stmt.Consequence)
	c.symTable.ExitScope()

	// Elif bloklarını kontrol et
	for _, elif := range stmt.Elif {
		elifCondType := c.checkExpression(elif.Condition)
		if elifCondType != BoolType && elifCondType != AnyType {
			c.addError(&SemanticError{
				Message: fmt.Sprintf("elif condition must be bool, got %s", elifCondType.String()),
				Pos:     elif.Token,
			})
		}

		c.symTable.EnterScope()
		c.checkBlockStatement(elif.Consequence)
		c.symTable.ExitScope()
	}

	// Alternative bloğunu kontrol et
	if stmt.Alternative != nil {
		c.symTable.EnterScope()
		c.checkBlockStatement(stmt.Alternative)
		c.symTable.ExitScope()
	}
}

func (c *Checker) checkWhileStatement(stmt *ast.WhileStatement) {
	// Condition'ı kontrol et
	condType := c.checkExpression(stmt.Condition)
	if condType != BoolType && condType != AnyType {
		c.addError(&SemanticError{
			Message: fmt.Sprintf("while condition must be bool, got %s", condType.String()),
			Pos:     stmt.Token,
		})
	}

	// Body'yi kontrol et (loop içinde)
	c.inLoop++
	c.symTable.EnterScope()
	c.checkBlockStatement(stmt.Body)
	c.symTable.ExitScope()
	c.inLoop--
}

func (c *Checker) checkForStatement(stmt *ast.ForStatement) {
	// Iterable'ın tipini kontrol et
	iterableType := c.checkExpression(stmt.Iterable)

	// Iterator tipini belirle
	var iteratorType Type = AnyType
	if listType, ok := iterableType.(*ListType); ok {
		iteratorType = listType.ElementType
	}

	// Body'yi kontrol et (yeni scope + iterator değişkeni)
	c.inLoop++
	c.symTable.EnterScope()

	// Iterator değişkenini scope'a ekle
	iterSymbol := &Symbol{
		Name:    stmt.Iterator.Value,
		Kind:    VariableSymbol,
		Type:    iteratorType,
		Pos:     stmt.Token,
		Mutable: false, // for iterator'ları read-only
		Node:    stmt,
	}
	if err := c.symTable.Define(iterSymbol); err != nil {
		c.addError(err)
	}

	c.checkBlockStatement(stmt.Body)
	c.symTable.ExitScope()
	c.inLoop--
}

func (c *Checker) checkClassStatement(stmt *ast.ClassStatement) {
	// Class tipini oluştur
	classType := &ClassType{
		Name:    stmt.Name.Value,
		Methods: make(map[string]*FunctionType),
		Fields:  make(map[string]Type),
	}

	// SuperClass varsa
	if stmt.SuperClass != nil {
		if superSymbol, ok := c.symTable.Resolve(stmt.SuperClass.Value); ok {
			if superClass, ok := superSymbol.Type.(*ClassType); ok {
				classType.SuperClass = superClass
			}
		}
	}

	// Class'ı sembole ekle
	classSymbol := &Symbol{
		Name: stmt.Name.Value,
		Kind: ClassSymbol,
		Type: classType,
		Pos:  stmt.Token,
		Node: stmt,
	}

	if err := c.symTable.Define(classSymbol); err != nil {
		c.addError(err)
		return
	}

	// Class body'sini kontrol et (yeni scope)
	c.symTable.EnterScope()

	for _, member := range stmt.Body {
		c.checkStatement(member)
	}

	c.symTable.ExitScope()
}

func (c *Checker) checkUnsafeStatement(stmt *ast.UnsafeStatement) {
	// Unsafe bloğunu kontrol et (yeni scope)
	c.symTable.EnterScope()
	c.checkBlockStatement(stmt.Body)
	c.symTable.ExitScope()
}

func (c *Checker) checkBlockStatement(block *ast.BlockStatement) {
	if block == nil {
		return
	}

	for _, stmt := range block.Statements {
		c.checkStatement(stmt)
	}
}

func (c *Checker) checkExpression(expr ast.Expression) Type {
	if expr == nil {
		return VoidType
	}

	switch e := expr.(type) {
	case *ast.Identifier:
		return c.checkIdentifier(e)
	case *ast.IntegerLiteral:
		return IntType
	case *ast.FloatLiteral:
		return FloatType
	case *ast.StringLiteral:
		return StringType
	case *ast.BooleanLiteral:
		return BoolType
	case *ast.ListLiteral:
		return c.checkListLiteral(e)
	case *ast.DictLiteral:
		return c.checkDictLiteral(e)
	case *ast.PrefixExpression:
		return c.checkPrefixExpression(e)
	case *ast.InfixExpression:
		return c.checkInfixExpression(e)
	case *ast.CallExpression:
		return c.checkCallExpression(e)
	case *ast.IndexExpression:
		return c.checkIndexExpression(e)
	case *ast.MemberExpression:
		return c.checkMemberExpression(e)
	case *ast.AwaitExpression:
		return c.checkAwaitExpression(e)
	case *ast.YieldExpression:
		return c.checkYieldExpression(e)
	default:
		return AnyType
	}
}

func (c *Checker) checkIdentifier(expr *ast.Identifier) Type {
	// Special handling for 'self' and 'super'
	if expr.Value == "self" {
		// 'self' is always valid in method context
		// For now, return AnyType (could be improved with class type tracking)
		return AnyType
	}
	if expr.Value == "super" {
		// 'super' is always valid in methods of classes with superclasses
		return AnyType
	}

	symbol, ok := c.symTable.Resolve(expr.Value)
	if !ok {
		c.addError(&SemanticError{
			Message: fmt.Sprintf("undefined: %s", expr.Value),
			Pos:     expr.Token,
		})
		return AnyType
	}

	return symbol.Type
}

func (c *Checker) checkListLiteral(expr *ast.ListLiteral) Type {
	if len(expr.Elements) == 0 {
		return &ListType{ElementType: AnyType}
	}

	// Tüm elemanların tipini kontrol et
	elemType := c.checkExpression(expr.Elements[0])
	for _, elem := range expr.Elements[1:] {
		t := c.checkExpression(elem)
		if !t.IsAssignableTo(elemType) {
			c.addError(&SemanticError{
				Message: fmt.Sprintf("list element type mismatch: expected %s, got %s",
					elemType.String(), t.String()),
				Pos: expr.Token,
			})
		}
	}

	return &ListType{ElementType: elemType}
}

func (c *Checker) checkDictLiteral(expr *ast.DictLiteral) Type {
	if len(expr.Pairs) == 0 {
		return &DictType{KeyType: AnyType, ValueType: AnyType}
	}

	var keyType, valueType Type
	first := true
	for key, value := range expr.Pairs {
		kt := c.checkExpression(key)
		vt := c.checkExpression(value)

		if first {
			keyType = kt
			valueType = vt
			first = false
		} else {
			if !kt.IsAssignableTo(keyType) {
				c.addError(&SemanticError{
					Message: fmt.Sprintf("dict key type mismatch: expected %s, got %s",
						keyType.String(), kt.String()),
					Pos: expr.Token,
				})
			}
			if !vt.IsAssignableTo(valueType) {
				c.addError(&SemanticError{
					Message: fmt.Sprintf("dict value type mismatch: expected %s, got %s",
						valueType.String(), vt.String()),
					Pos: expr.Token,
				})
			}
		}
	}

	return &DictType{KeyType: keyType, ValueType: valueType}
}

func (c *Checker) checkPrefixExpression(expr *ast.PrefixExpression) Type {
	rightType := c.checkExpression(expr.Right)

	switch expr.Operator {
	case "!":
		if rightType != BoolType && rightType != AnyType {
			c.addError(&SemanticError{
				Message: fmt.Sprintf("operator ! cannot be applied to %s", rightType.String()),
				Pos:     expr.Token,
			})
		}
		return BoolType

	case "-", "+":
		if rightType != IntType && rightType != FloatType && rightType != AnyType {
			c.addError(&SemanticError{
				Message: fmt.Sprintf("operator %s cannot be applied to %s", expr.Operator, rightType.String()),
				Pos:     expr.Token,
			})
		}
		return rightType

	default:
		return AnyType
	}
}

func (c *Checker) checkInfixExpression(expr *ast.InfixExpression) Type {
	leftType := c.checkExpression(expr.Left)
	rightType := c.checkExpression(expr.Right)

	// Assignment operatörleri
	if expr.Operator == "=" || expr.Operator == "+=" || expr.Operator == "-=" ||
		expr.Operator == "*=" || expr.Operator == "/=" || expr.Operator == "%=" {

		// Sol taraf identifier olmalı
		if ident, ok := expr.Left.(*ast.Identifier); ok {
			if symbol, ok := c.symTable.Resolve(ident.Value); ok {
				// Const değişkene atama yapılamaz
				if !symbol.Mutable {
					c.addError(&SemanticError{
						Message: fmt.Sprintf("cannot assign to const variable '%s'", ident.Value),
						Pos:     expr.Token,
					})
				}
			}
		}

		// Tip uyumluluğu kontrolü
		if !rightType.IsAssignableTo(leftType) {
			c.addError(&SemanticError{
				Message: fmt.Sprintf("type mismatch: cannot assign %s to %s",
					rightType.String(), leftType.String()),
				Pos: expr.Token,
			})
		}

		return leftType
	}

	// Comparison operatörleri
	if expr.Operator == "==" || expr.Operator == "!=" ||
		expr.Operator == "<" || expr.Operator == "<=" ||
		expr.Operator == ">" || expr.Operator == ">=" {
		return BoolType
	}

	// Logical operatörler
	if expr.Operator == "&&" || expr.Operator == "||" {
		return BoolType
	}

	// Arithmetic operatörler
	if leftType == FloatType || rightType == FloatType {
		return FloatType
	}
	if leftType == IntType && rightType == IntType {
		return IntType
	}
	if leftType == StringType && rightType == StringType && expr.Operator == "+" {
		return StringType
	}

	return AnyType
}

func (c *Checker) checkCallExpression(expr *ast.CallExpression) Type {
	funcType := c.checkExpression(expr.Function)

	if ft, ok := funcType.(*FunctionType); ok {
		// Parametre sayısı kontrolü
		if !ft.Variadic && len(expr.Arguments) != len(ft.Params) {
			c.addError(&SemanticError{
				Message: fmt.Sprintf("wrong number of arguments: expected %d, got %d",
					len(ft.Params), len(expr.Arguments)),
				Pos: expr.Token,
			})
		}

		// Argüman tiplerini kontrol et
		for i, arg := range expr.Arguments {
			argType := c.checkExpression(arg)
			if i < len(ft.Params) {
				if !argType.IsAssignableTo(ft.Params[i]) {
					c.addError(&SemanticError{
						Message: fmt.Sprintf("argument %d type mismatch: expected %s, got %s",
							i+1, ft.Params[i].String(), argType.String()),
						Pos: expr.Token,
					})
				}
			}
		}

		return ft.ReturnType
	}

	return AnyType
}

func (c *Checker) checkIndexExpression(expr *ast.IndexExpression) Type {
	leftType := c.checkExpression(expr.Left)
	indexType := c.checkExpression(expr.Index)

	if listType, ok := leftType.(*ListType); ok {
		if indexType != IntType && indexType != AnyType {
			c.addError(&SemanticError{
				Message: fmt.Sprintf("list index must be int, got %s", indexType.String()),
				Pos:     expr.Token,
			})
		}
		return listType.ElementType
	}

	if dictType, ok := leftType.(*DictType); ok {
		if !indexType.IsAssignableTo(dictType.KeyType) {
			c.addError(&SemanticError{
				Message: fmt.Sprintf("dict key type mismatch: expected %s, got %s",
					dictType.KeyType.String(), indexType.String()),
				Pos: expr.Token,
			})
		}
		return dictType.ValueType
	}

	return AnyType
}

func (c *Checker) checkMemberExpression(expr *ast.MemberExpression) Type {
	// Şimdilik basit implementasyon
	c.checkExpression(expr.Object)
	return AnyType
}

func (c *Checker) checkAwaitExpression(expr *ast.AwaitExpression) Type {
	// Async fonksiyon içinde miyiz kontrol et
	if c.currentFunction == nil || !c.currentFunction.IsAsync {
		c.addError(&SemanticError{
			Message: "await can only be used in async functions",
			Pos:     expr.Token,
		})
	}

	return c.checkExpression(expr.Expression)
}

func (c *Checker) checkYieldExpression(expr *ast.YieldExpression) Type {
	// Coop fonksiyon içinde miyiz kontrol et (şimdilik basit kontrol)
	if c.currentFunction == nil {
		c.addError(&SemanticError{
			Message: "yield can only be used in coop functions",
			Pos:     expr.Token,
		})
	}

	if expr.Value != nil {
		return c.checkExpression(expr.Value)
	}

	return VoidType
}
