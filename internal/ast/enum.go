package ast

import "github.com/mburakmmm/sky-lang/internal/lexer"

// EnumStatement represents an enum declaration
type EnumStatement struct {
	Token    lexer.Token    // ENUM token
	Name     *Identifier    // enum name
	Variants []*EnumVariant // variants
}

func (es *EnumStatement) statementNode()       {}
func (es *EnumStatement) TokenLiteral() string { return es.Token.Literal }
func (es *EnumStatement) Pos() lexer.Token     { return es.Token }
func (es *EnumStatement) String() string {
	return "enum " + es.Name.Value
}

// EnumVariant represents an enum variant
type EnumVariant struct {
	Name    *Identifier      // variant name
	Payload []TypeAnnotation // payload types (optional)
}

// EnumConstructorExpression represents constructing an enum value
type EnumConstructorExpression struct {
	Token    lexer.Token  // enum variant name
	EnumName string       // enum type name
	Variant  string       // variant name
	Args     []Expression // arguments for payload
}

func (ec *EnumConstructorExpression) expressionNode()      {}
func (ec *EnumConstructorExpression) TokenLiteral() string { return ec.Token.Literal }
func (ec *EnumConstructorExpression) Pos() lexer.Token     { return ec.Token }
func (ec *EnumConstructorExpression) String() string {
	return ec.EnumName + "::" + ec.Variant
}
