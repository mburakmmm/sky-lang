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
	case *ast.AbstractClassStatement:
		c.checkAbstractClassStatement(s)
	case *ast.AbstractMethodStatement:
		c.checkAbstractMethodStatement(s)
	case *ast.StaticMethodStatement:
		c.checkStaticMethodStatement(s)
	case *ast.StaticPropertyStatement:
		c.checkStaticPropertyStatement(s)
	case *ast.ImportStatement:
		// Import statements are handled at a higher level
	case *ast.UnsafeStatement:
		c.checkUnsafeStatement(s)
	case *ast.EnumStatement:
		c.checkEnumStatement(s)
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
	minParams := 0
	hasVarargs := false
	for i, param := range stmt.Parameters {
		if param.Type != nil {
			paramTypes[i] = ResolveType(param.Type)
		} else {
			paramTypes[i] = AnyType
		}
		// Check for varargs
		if param.Variadic {
			hasVarargs = true
			// Varargs itself is optional, so minParams stays at current value
		} else if param.DefaultValue == nil && !hasVarargs {
			// Count required parameters (those without default values)
			minParams = i + 1
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
		MinParams:  minParams,
		Variadic:   hasVarargs,
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

	// Multiple inheritance - check all superclasses
	for _, superClassIdent := range stmt.SuperClasses {
		if superSymbol, ok := c.symTable.Resolve(superClassIdent.Value); ok {
			if superClass, ok := superSymbol.Type.(*ClassType); ok {
				classType.SuperClasses = append(classType.SuperClasses, superClass)
			} else {
				c.addError(&SemanticError{
					Message: fmt.Sprintf("%s is not a class", superClassIdent.Value),
					Pos:     superClassIdent.Token,
				})
			}
		} else {
			c.addError(&SemanticError{
				Message: fmt.Sprintf("undefined superclass: %s", superClassIdent.Value),
				Pos:     superClassIdent.Token,
			})
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

func (c *Checker) checkEnumStatement(stmt *ast.EnumStatement) {
	// Enum tipini tanımla
	enumSymbol := &Symbol{
		Name: stmt.Name.Value,
		Kind: ClassSymbol, // Enum'u class gibi ele al
		Type: AnyType,
		Pos:  stmt.Token,
	}

	if err := c.symTable.Define(enumSymbol); err != nil {
		c.addError(err)
		return
	}

	// Her variant için constructor fonksiyonu tanımla
	for _, variant := range stmt.Variants {
		paramTypes := make([]Type, len(variant.Payload))
		for i, payloadType := range variant.Payload {
			paramTypes[i] = ResolveType(payloadType)
		}

		constructorType := &FunctionType{
			Params:     paramTypes,
			ReturnType: AnyType, // Enum instance döndürür
		}

		variantSymbol := &Symbol{
			Name: variant.Name.Value,
			Kind: FunctionSymbol,
			Type: constructorType,
			Pos:  stmt.Token,
		}

		if err := c.symTable.Define(variantSymbol); err != nil {
			c.addError(err)
		}
	}
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
		argCount := len(expr.Arguments)
		minRequired := ft.MinParams
		if minRequired == 0 {
			minRequired = len(ft.Params) // Backward compatibility
		}

		if ft.Variadic {
			// Varargs function: check minimum required arguments
			if argCount < minRequired {
				c.addError(&SemanticError{
					Message: fmt.Sprintf("wrong number of arguments: expected at least %d, got %d",
						minRequired, argCount),
					Pos: expr.Token,
				})
			}
		} else {
			// Non-varargs: check if argument count is within valid range
			if argCount < minRequired || argCount > len(ft.Params) {
				if minRequired == len(ft.Params) {
					c.addError(&SemanticError{
						Message: fmt.Sprintf("wrong number of arguments: expected %d, got %d",
							len(ft.Params), argCount),
						Pos: expr.Token,
					})
				} else {
					c.addError(&SemanticError{
						Message: fmt.Sprintf("wrong number of arguments: expected %d-%d, got %d",
							minRequired, len(ft.Params), argCount),
						Pos: expr.Token,
					})
				}
			}
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

// checkAbstractClassStatement checks abstract class statements
func (c *Checker) checkAbstractClassStatement(stmt *ast.AbstractClassStatement) {
	// For now, treat abstract class like a regular class in type checking
	// Just register the class name as a class type
	c.checkClassStatement(&ast.ClassStatement{
		Token:        stmt.Token,
		Name:         stmt.Name,
		SuperClasses: stmt.SuperClasses,
		Body:         stmt.Body,
	})
}

// checkAbstractMethodStatement checks abstract method statements
func (c *Checker) checkAbstractMethodStatement(stmt *ast.AbstractMethodStatement) {
	// Abstract methods are handled within abstract classes
	// This is just a placeholder
}

// checkStaticMethodStatement checks static method statements
func (c *Checker) checkStaticMethodStatement(stmt *ast.StaticMethodStatement) {
	// Static methods are treated like regular functions
	var returnType Type = AnyType
	if stmt.ReturnType != nil {
		returnType = ResolveType(stmt.ReturnType)
	}

	paramTypes := []Type{}
	for _, param := range stmt.Parameters {
		if param.Type != nil {
			paramTypes = append(paramTypes, ResolveType(param.Type))
		} else {
			paramTypes = append(paramTypes, AnyType)
		}
	}

	funcSymbol := &Symbol{
		Name: stmt.Name.Value,
		Type: &FunctionType{
			Params:     paramTypes,
			ReturnType: returnType,
		},
		Kind: FunctionSymbol,
	}

	if err := c.symTable.Define(funcSymbol); err != nil {
		c.addError(err)
	}
}

// checkStaticPropertyStatement checks static property statements
func (c *Checker) checkStaticPropertyStatement(stmt *ast.StaticPropertyStatement) {
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

	// Symbol table'a ekle
	symbol := &Symbol{
		Name:    stmt.Name.Value,
		Type:    declaredType,
		Kind:    VariableSymbol,
		Mutable: true,
	}

	if err := c.symTable.Define(symbol); err != nil {
		c.addError(err)
	}
}
