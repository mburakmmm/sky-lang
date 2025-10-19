package ast

import (
	"fmt"
	"strings"

	"github.com/mburakmmm/sky-lang/internal/lexer"
)

// Node AST'deki tüm düğümlerin temel interface'i
type Node interface {
	TokenLiteral() string
	String() string
	Pos() lexer.Token // Pozisyon bilgisi için
}

// Statement ifadeleri temsil eder
type Statement interface {
	Node
	statementNode()
}

// Expression değer üreten ifadeleri temsil eder
type Expression interface {
	Node
	expressionNode()
}

// ==============================================
// Program - Root Node
// ==============================================

// Program kaynak kodun AST'sini temsil eder
type Program struct {
	Statements []Statement
	File       string
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}
	return ""
}

func (p *Program) String() string {
	var out strings.Builder
	for _, s := range p.Statements {
		out.WriteString(s.String())
	}
	return out.String()
}

func (p *Program) Pos() lexer.Token {
	if len(p.Statements) > 0 {
		return p.Statements[0].Pos()
	}
	return lexer.Token{}
}

// ==============================================
// Statements
// ==============================================

// LetStatement let değişken tanımlama
type LetStatement struct {
	Token lexer.Token // LET token
	Name  *Identifier
	Type  TypeAnnotation // opsiyonel tip anotasyonu
	Value Expression
}

func (ls *LetStatement) statementNode()       {}
func (ls *LetStatement) TokenLiteral() string { return ls.Token.Literal }
func (ls *LetStatement) Pos() lexer.Token     { return ls.Token }
func (ls *LetStatement) String() string {
	var out strings.Builder
	out.WriteString(ls.TokenLiteral() + " ")
	out.WriteString(ls.Name.String())
	if ls.Type != nil {
		out.WriteString(": ")
		out.WriteString(ls.Type.String())
	}
	out.WriteString(" = ")
	if ls.Value != nil {
		out.WriteString(ls.Value.String())
	}
	return out.String()
}

// ConstStatement const sabit tanımlama
type ConstStatement struct {
	Token lexer.Token // CONST token
	Name  *Identifier
	Type  TypeAnnotation
	Value Expression
}

func (cs *ConstStatement) statementNode()       {}
func (cs *ConstStatement) TokenLiteral() string { return cs.Token.Literal }
func (cs *ConstStatement) Pos() lexer.Token     { return cs.Token }
func (cs *ConstStatement) String() string {
	var out strings.Builder
	out.WriteString(cs.TokenLiteral() + " ")
	out.WriteString(cs.Name.String())
	if cs.Type != nil {
		out.WriteString(": ")
		out.WriteString(cs.Type.String())
	}
	out.WriteString(" = ")
	if cs.Value != nil {
		out.WriteString(cs.Value.String())
	}
	return out.String()
}

// ReturnStatement return ifadesi
type ReturnStatement struct {
	Token       lexer.Token // RETURN token
	ReturnValue Expression
}

func (rs *ReturnStatement) statementNode()       {}
func (rs *ReturnStatement) TokenLiteral() string { return rs.Token.Literal }
func (rs *ReturnStatement) Pos() lexer.Token     { return rs.Token }
func (rs *ReturnStatement) String() string {
	var out strings.Builder
	out.WriteString(rs.TokenLiteral())
	if rs.ReturnValue != nil {
		out.WriteString(" ")
		out.WriteString(rs.ReturnValue.String())
	}
	return out.String()
}

// BreakStatement döngüden çıkış
type BreakStatement struct {
	Token lexer.Token // BREAK token
}

func (bs *BreakStatement) statementNode()       {}
func (bs *BreakStatement) TokenLiteral() string { return bs.Token.Literal }
func (bs *BreakStatement) Pos() lexer.Token     { return bs.Token }
func (bs *BreakStatement) String() string       { return "break" }

// ContinueStatement döngüyü devam ettirme
type ContinueStatement struct {
	Token lexer.Token // CONTINUE token
}

func (cs *ContinueStatement) statementNode()       {}
func (cs *ContinueStatement) TokenLiteral() string { return cs.Token.Literal }
func (cs *ContinueStatement) Pos() lexer.Token     { return cs.Token }
func (cs *ContinueStatement) String() string       { return "continue" }

// ExpressionStatement bir expression'ı statement olarak kullanma
type ExpressionStatement struct {
	Token      lexer.Token // expression'ın ilk token'ı
	Expression Expression
}

func (es *ExpressionStatement) statementNode()       {}
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Literal }
func (es *ExpressionStatement) Pos() lexer.Token     { return es.Token }
func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}
	return ""
}

// BlockStatement blok ifadesi (girintili kod blokları)
type BlockStatement struct {
	Token      lexer.Token // INDENT token
	Statements []Statement
}

func (bs *BlockStatement) statementNode()       {}
func (bs *BlockStatement) TokenLiteral() string { return bs.Token.Literal }
func (bs *BlockStatement) Pos() lexer.Token     { return bs.Token }
func (bs *BlockStatement) String() string {
	var out strings.Builder
	for _, s := range bs.Statements {
		out.WriteString(s.String())
		out.WriteString("\n")
	}
	return out.String()
}

// FunctionStatement fonksiyon tanımlama
type FunctionStatement struct {
	Token      lexer.Token // FUNCTION token
	Name       *Identifier
	Async      bool
	Coop       bool // Coroutine/Generator flag
	Parameters []*FunctionParameter
	ReturnType TypeAnnotation
	Body       *BlockStatement
}

func (fs *FunctionStatement) statementNode()       {}
func (fs *FunctionStatement) TokenLiteral() string { return fs.Token.Literal }
func (fs *FunctionStatement) Pos() lexer.Token     { return fs.Token }
func (fs *FunctionStatement) String() string {
	var out strings.Builder
	if fs.Async {
		out.WriteString("async ")
	}
	out.WriteString("function ")
	out.WriteString(fs.Name.String())
	out.WriteString("(")
	params := []string{}
	for _, p := range fs.Parameters {
		params = append(params, p.String())
	}
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(")")
	if fs.ReturnType != nil {
		out.WriteString(": ")
		out.WriteString(fs.ReturnType.String())
	}
	out.WriteString("\n")
	out.WriteString(fs.Body.String())
	out.WriteString("end")
	return out.String()
}

// FunctionParameter fonksiyon parametresi
type FunctionParameter struct {
	Token        lexer.Token
	Name         *Identifier
	Type         TypeAnnotation
	DefaultValue Expression
}

func (fp *FunctionParameter) String() string {
	var out strings.Builder
	out.WriteString(fp.Name.String())
	if fp.Type != nil {
		out.WriteString(": ")
		out.WriteString(fp.Type.String())
	}
	if fp.DefaultValue != nil {
		out.WriteString(" = ")
		out.WriteString(fp.DefaultValue.String())
	}
	return out.String()
}

// IfStatement if-elif-else ifadesi
type IfStatement struct {
	Token       lexer.Token // IF token
	Condition   Expression
	Consequence *BlockStatement
	Elif        []*ElifClause
	Alternative *BlockStatement // else bloğu
}

func (is *IfStatement) statementNode()       {}
func (is *IfStatement) TokenLiteral() string { return is.Token.Literal }
func (is *IfStatement) Pos() lexer.Token     { return is.Token }
func (is *IfStatement) String() string {
	var out strings.Builder
	out.WriteString("if ")
	out.WriteString(is.Condition.String())
	out.WriteString("\n")
	out.WriteString(is.Consequence.String())

	for _, elif := range is.Elif {
		out.WriteString("elif ")
		out.WriteString(elif.Condition.String())
		out.WriteString("\n")
		out.WriteString(elif.Consequence.String())
	}

	if is.Alternative != nil {
		out.WriteString("else\n")
		out.WriteString(is.Alternative.String())
	}
	out.WriteString("end")
	return out.String()
}

// ElifClause elif dalı
type ElifClause struct {
	Token       lexer.Token // ELIF token
	Condition   Expression
	Consequence *BlockStatement
}

// WhileStatement while döngüsü
type WhileStatement struct {
	Token     lexer.Token // WHILE token
	Condition Expression
	Body      *BlockStatement
}

func (ws *WhileStatement) statementNode()       {}
func (ws *WhileStatement) TokenLiteral() string { return ws.Token.Literal }
func (ws *WhileStatement) Pos() lexer.Token     { return ws.Token }
func (ws *WhileStatement) String() string {
	var out strings.Builder
	out.WriteString("while ")
	out.WriteString(ws.Condition.String())
	out.WriteString("\n")
	out.WriteString(ws.Body.String())
	out.WriteString("end")
	return out.String()
}

// ForStatement for döngüsü
type ForStatement struct {
	Token    lexer.Token // FOR token
	Iterator *Identifier
	Iterable Expression
	Body     *BlockStatement
}

func (fs *ForStatement) statementNode()       {}
func (fs *ForStatement) TokenLiteral() string { return fs.Token.Literal }
func (fs *ForStatement) Pos() lexer.Token     { return fs.Token }
func (fs *ForStatement) String() string {
	var out strings.Builder
	out.WriteString("for ")
	out.WriteString(fs.Iterator.String())
	out.WriteString(" in ")
	out.WriteString(fs.Iterable.String())
	out.WriteString("\n")
	out.WriteString(fs.Body.String())
	out.WriteString("end")
	return out.String()
}

// ClassStatement sınıf tanımlama
type ClassStatement struct {
	Token      lexer.Token // CLASS token
	Name       *Identifier
	SuperClass *Identifier // parent class (opsiyonel)
	Body       []Statement // class members
}

func (cs *ClassStatement) statementNode()       {}
func (cs *ClassStatement) TokenLiteral() string { return cs.Token.Literal }
func (cs *ClassStatement) Pos() lexer.Token     { return cs.Token }
func (cs *ClassStatement) String() string {
	var out strings.Builder
	out.WriteString("class ")
	out.WriteString(cs.Name.String())
	if cs.SuperClass != nil {
		out.WriteString("(")
		out.WriteString(cs.SuperClass.String())
		out.WriteString(")")
	}
	out.WriteString("\n")
	for _, stmt := range cs.Body {
		out.WriteString("  ")
		out.WriteString(stmt.String())
		out.WriteString("\n")
	}
	out.WriteString("end")
	return out.String()
}

// ImportStatement import ifadesi
type ImportStatement struct {
	Token lexer.Token // IMPORT token
	Path  []string    // import path segments
	Alias *Identifier // as alias (opsiyonel)
}

func (is *ImportStatement) statementNode()       {}
func (is *ImportStatement) TokenLiteral() string { return is.Token.Literal }
func (is *ImportStatement) Pos() lexer.Token     { return is.Token }
func (is *ImportStatement) String() string {
	var out strings.Builder
	out.WriteString("import ")
	out.WriteString(strings.Join(is.Path, "."))
	if is.Alias != nil {
		out.WriteString(" as ")
		out.WriteString(is.Alias.String())
	}
	return out.String()
}

// UnsafeStatement unsafe bloğu
type UnsafeStatement struct {
	Token lexer.Token // UNSAFE token
	Body  *BlockStatement
}

func (us *UnsafeStatement) statementNode()       {}
func (us *UnsafeStatement) TokenLiteral() string { return us.Token.Literal }
func (us *UnsafeStatement) Pos() lexer.Token     { return us.Token }
func (us *UnsafeStatement) String() string {
	var out strings.Builder
	out.WriteString("unsafe\n")
	out.WriteString(us.Body.String())
	out.WriteString("end")
	return out.String()
}

// ==============================================
// Expressions
// ==============================================

// Identifier tanımlayıcı
type Identifier struct {
	Token lexer.Token
	Value string
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }
func (i *Identifier) Pos() lexer.Token     { return i.Token }
func (i *Identifier) String() string       { return i.Value }

// IntegerLiteral tam sayı literal
type IntegerLiteral struct {
	Token lexer.Token
	Value int64
}

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }
func (il *IntegerLiteral) Pos() lexer.Token     { return il.Token }
func (il *IntegerLiteral) String() string       { return il.Token.Literal }

// FloatLiteral ondalık sayı literal
type FloatLiteral struct {
	Token lexer.Token
	Value float64
}

func (fl *FloatLiteral) expressionNode()      {}
func (fl *FloatLiteral) TokenLiteral() string { return fl.Token.Literal }
func (fl *FloatLiteral) Pos() lexer.Token     { return fl.Token }
func (fl *FloatLiteral) String() string       { return fl.Token.Literal }

// StringLiteral string literal
type StringLiteral struct {
	Token lexer.Token
	Value string
}

func (sl *StringLiteral) expressionNode()      {}
func (sl *StringLiteral) TokenLiteral() string { return sl.Token.Literal }
func (sl *StringLiteral) Pos() lexer.Token     { return sl.Token }
func (sl *StringLiteral) String() string       { return fmt.Sprintf(`"%s"`, sl.Value) }

// BooleanLiteral boolean literal
type BooleanLiteral struct {
	Token lexer.Token
	Value bool
}

func (bl *BooleanLiteral) expressionNode()      {}
func (bl *BooleanLiteral) TokenLiteral() string { return bl.Token.Literal }
func (bl *BooleanLiteral) Pos() lexer.Token     { return bl.Token }
func (bl *BooleanLiteral) String() string       { return bl.Token.Literal }

// PrefixExpression prefix operatörler (!, -, +)
type PrefixExpression struct {
	Token    lexer.Token
	Operator string
	Right    Expression
}

func (pe *PrefixExpression) expressionNode()      {}
func (pe *PrefixExpression) TokenLiteral() string { return pe.Token.Literal }
func (pe *PrefixExpression) Pos() lexer.Token     { return pe.Token }
func (pe *PrefixExpression) String() string {
	return fmt.Sprintf("(%s%s)", pe.Operator, pe.Right.String())
}

// InfixExpression infix operatörler (+, -, *, /, etc.)
type InfixExpression struct {
	Token    lexer.Token
	Left     Expression
	Operator string
	Right    Expression
}

func (ie *InfixExpression) expressionNode()      {}
func (ie *InfixExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *InfixExpression) Pos() lexer.Token     { return ie.Token }
func (ie *InfixExpression) String() string {
	return fmt.Sprintf("(%s %s %s)", ie.Left.String(), ie.Operator, ie.Right.String())
}

// CallExpression fonksiyon çağrısı
type CallExpression struct {
	Token     lexer.Token // LPAREN token
	Function  Expression  // identifier veya function expression
	Arguments []Expression
}

func (ce *CallExpression) expressionNode()      {}
func (ce *CallExpression) TokenLiteral() string { return ce.Token.Literal }
func (ce *CallExpression) Pos() lexer.Token     { return ce.Token }
func (ce *CallExpression) String() string {
	args := []string{}
	for _, a := range ce.Arguments {
		args = append(args, a.String())
	}
	return fmt.Sprintf("%s(%s)", ce.Function.String(), strings.Join(args, ", "))
}

// IndexExpression array/dict indexing
type IndexExpression struct {
	Token lexer.Token // LBRACK token
	Left  Expression
	Index Expression
}

func (ie *IndexExpression) expressionNode()      {}
func (ie *IndexExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *IndexExpression) Pos() lexer.Token     { return ie.Token }
func (ie *IndexExpression) String() string {
	return fmt.Sprintf("(%s[%s])", ie.Left.String(), ie.Index.String())
}

// MemberExpression member access (dot notation)
type MemberExpression struct {
	Token  lexer.Token // DOT token
	Object Expression
	Member *Identifier
}

func (me *MemberExpression) expressionNode()      {}
func (me *MemberExpression) TokenLiteral() string { return me.Token.Literal }
func (me *MemberExpression) Pos() lexer.Token     { return me.Token }
func (me *MemberExpression) String() string {
	return fmt.Sprintf("%s.%s", me.Object.String(), me.Member.String())
}

// ListLiteral array literal
type ListLiteral struct {
	Token    lexer.Token // LBRACK token
	Elements []Expression
}

func (ll *ListLiteral) expressionNode()      {}
func (ll *ListLiteral) TokenLiteral() string { return ll.Token.Literal }
func (ll *ListLiteral) Pos() lexer.Token     { return ll.Token }
func (ll *ListLiteral) String() string {
	elements := []string{}
	for _, el := range ll.Elements {
		elements = append(elements, el.String())
	}
	return fmt.Sprintf("[%s]", strings.Join(elements, ", "))
}

// DictLiteral dictionary literal
type DictLiteral struct {
	Token lexer.Token // LBRACE token
	Pairs map[Expression]Expression
}

func (dl *DictLiteral) expressionNode()      {}
func (dl *DictLiteral) TokenLiteral() string { return dl.Token.Literal }
func (dl *DictLiteral) Pos() lexer.Token     { return dl.Token }
func (dl *DictLiteral) String() string {
	pairs := []string{}
	for key, value := range dl.Pairs {
		pairs = append(pairs, fmt.Sprintf("%s: %s", key.String(), value.String()))
	}
	return fmt.Sprintf("{%s}", strings.Join(pairs, ", "))
}

// AwaitExpression await ifadesi
type AwaitExpression struct {
	Token      lexer.Token // AWAIT token
	Expression Expression
}

func (ae *AwaitExpression) expressionNode()      {}
func (ae *AwaitExpression) TokenLiteral() string { return ae.Token.Literal }
func (ae *AwaitExpression) Pos() lexer.Token     { return ae.Token }
func (ae *AwaitExpression) String() string {
	return fmt.Sprintf("await %s", ae.Expression.String())
}

// YieldExpression yield ifadesi
type YieldExpression struct {
	Token lexer.Token // YIELD token
	Value Expression  // yielded value
}

func (ye *YieldExpression) expressionNode()      {}
func (ye *YieldExpression) TokenLiteral() string { return ye.Token.Literal }
func (ye *YieldExpression) Pos() lexer.Token     { return ye.Token }
func (ye *YieldExpression) String() string {
	if ye.Value != nil {
		return fmt.Sprintf("yield %s", ye.Value.String())
	}
	return "yield"
}

// LambdaExpression anonymous function expression
type LambdaExpression struct {
	Token      lexer.Token          // FUNCTION token
	Parameters []*FunctionParameter // parameters
	ReturnType TypeAnnotation       // return type (optional)
	Body       *BlockStatement      // function body
}

func (le *LambdaExpression) expressionNode()      {}
func (le *LambdaExpression) TokenLiteral() string { return le.Token.Literal }
func (le *LambdaExpression) Pos() lexer.Token     { return le.Token }
func (le *LambdaExpression) String() string {
	params := []string{}
	for _, p := range le.Parameters {
		params = append(params, p.Name.Value)
	}
	return fmt.Sprintf("function(%s)", strings.Join(params, ", "))
}

// ==============================================
// Type Annotations
// ==============================================

// TypeAnnotation tip anotasyonları için interface
type TypeAnnotation interface {
	Node
	typeNode()
}

// BasicType temel tipler (int, float, string, bool, any)
type BasicType struct {
	Token lexer.Token
	Name  string
}

func (bt *BasicType) typeNode()            {}
func (bt *BasicType) TokenLiteral() string { return bt.Token.Literal }
func (bt *BasicType) Pos() lexer.Token     { return bt.Token }
func (bt *BasicType) String() string       { return bt.Name }

// ListType list tip anotasyonu [T]
type ListType struct {
	Token       lexer.Token // LBRACK token
	ElementType TypeAnnotation
}

func (lt *ListType) typeNode()            {}
func (lt *ListType) TokenLiteral() string { return lt.Token.Literal }
func (lt *ListType) Pos() lexer.Token     { return lt.Token }
func (lt *ListType) String() string {
	return fmt.Sprintf("[%s]", lt.ElementType.String())
}

// DictType dictionary tip anotasyonu {K: V}
type DictType struct {
	Token     lexer.Token // LBRACE token
	KeyType   TypeAnnotation
	ValueType TypeAnnotation
}

func (dt *DictType) typeNode()            {}
func (dt *DictType) TokenLiteral() string { return dt.Token.Literal }
func (dt *DictType) Pos() lexer.Token     { return dt.Token }
func (dt *DictType) String() string {
	return fmt.Sprintf("{%s: %s}", dt.KeyType.String(), dt.ValueType.String())
}

// FunctionType fonksiyon tip anotasyonu (T1, T2) => T3
type FunctionType struct {
	Token      lexer.Token // LPAREN token
	ParamTypes []TypeAnnotation
	ReturnType TypeAnnotation
}

func (ft *FunctionType) typeNode()            {}
func (ft *FunctionType) TokenLiteral() string { return ft.Token.Literal }
func (ft *FunctionType) Pos() lexer.Token     { return ft.Token }
func (ft *FunctionType) String() string {
	params := []string{}
	for _, p := range ft.ParamTypes {
		params = append(params, p.String())
	}
	return fmt.Sprintf("(%s) => %s", strings.Join(params, ", "), ft.ReturnType.String())
}

// PointerType pointer tip anotasyonu *T
type PointerType struct {
	Token       lexer.Token // STAR token
	PointeeType TypeAnnotation
}

func (pt *PointerType) typeNode()            {}
func (pt *PointerType) TokenLiteral() string { return pt.Token.Literal }
func (pt *PointerType) Pos() lexer.Token     { return pt.Token }
func (pt *PointerType) String() string {
	return fmt.Sprintf("*%s", pt.PointeeType.String())
}
