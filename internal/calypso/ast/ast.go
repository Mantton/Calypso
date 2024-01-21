package ast

import (
	"github.com/mantton/calypso/internal/calypso/lexer"
	"github.com/mantton/calypso/internal/calypso/token"
)

type Node interface {
}

type Expression interface {
	Node
	expressionNode()
}

type Statement interface {
	Node
	statementNode()
}

type Declaration interface {
	Node
	declarationNode()
}

type File struct {
	ModuleName string
	Constants  []*ConstantDeclaration
	Functions  []*FunctionDeclaration
	Errors     lexer.ErrorList
}

// * Declarations
// - Imports, Modules, Structs, Types
type ConstantDeclaration struct {
	Stmt *VariableStatement
	// Type  Expression
}

type FunctionDeclaration struct {
	Func *FunctionExpression
}

func (d *ConstantDeclaration) declarationNode() {}
func (d *FunctionDeclaration) declarationNode() {}

// * Statements
type BlockStatement struct {
	Statements []Statement
}

type VariableStatement struct {
	Identifier     string
	Value          Expression
	IsConstant     bool
	TypeAnnotation TypeExpression
}

type FunctionStatement struct {
	Func *FunctionExpression
}

type IfStatement struct {
	Condition   Expression
	Action      *BlockStatement
	Alternative *BlockStatement
}

type ReturnStatement struct {
	Value Expression
}

type WhileStatement struct {
	Condition Expression
	Action    *BlockStatement
}

type ExpressionStatement struct {
	Expr Expression
}

func (s *IfStatement) statementNode()         {}
func (s *ExpressionStatement) statementNode() {}
func (s *WhileStatement) statementNode()      {}
func (s *ReturnStatement) statementNode()     {}
func (s *BlockStatement) statementNode()      {}
func (s *VariableStatement) statementNode()   {}
func (s *FunctionStatement) statementNode()   {}

// * Expressions

type GroupedExpression struct {
	Expr Expression
}
type UnaryExpression struct {
	Op   token.Token
	Expr Expression
}

type BinaryExpression struct {
	Left  Expression
	Op    token.Token
	Right Expression
}

type AssignmentExpression struct {
	Target Expression
	Value  Expression
}

type CallExpression struct {
	Target    Expression
	Arguments []Expression
}

type IndexExpression struct {
	Target Expression
	Index  Expression
}

type PropertyExpression struct {
	Target   Expression
	Property *IdentifierExpression
}

func (e *GroupedExpression) expressionNode()    {}
func (e *CallExpression) expressionNode()       {}
func (e *UnaryExpression) expressionNode()      {}
func (e *BinaryExpression) expressionNode()     {}
func (e *AssignmentExpression) expressionNode() {}
func (e *IndexExpression) expressionNode()      {}
func (e *PropertyExpression) expressionNode()   {}

// * Literal Expressions

type IdentifierExpression struct {
	Value string
}

type FunctionExpression struct {
	Name   string
	Body   *BlockStatement
	Params []*IdentifierExpression
}

func (e *IdentifierExpression) expressionNode() {}
func (e *FunctionExpression) expressionNode()   {}

// * Literals
type IntegerLiteral struct {
	Value int
}

type FloatLiteral struct {
	Value float64
}

type StringLiteral struct {
	Value string
}

type BooleanLiteral struct {
	Value bool
}

type NullLiteral struct{}

type VoidLiteral struct{}

type ArrayLiteral struct {
	Elements []Expression
}

type MapLiteral struct {
	Pairs map[Expression]Expression
}

func (e *IntegerLiteral) expressionNode() {}
func (e *FloatLiteral) expressionNode()   {}
func (e *StringLiteral) expressionNode()  {}
func (e *BooleanLiteral) expressionNode() {}
func (e *NullLiteral) expressionNode()    {}
func (e *VoidLiteral) expressionNode()    {}
func (e *ArrayLiteral) expressionNode()   {}
func (e *MapLiteral) expressionNode()     {}

// * Types

type TypeExpression interface {
	Node
	typeNode()
}

type IdentifierTypeExpression struct {
	Identifier *IdentifierExpression
	Arguments  *GenericArgumentList
}

type GenericArgumentList struct {
	Arguments []TypeExpression
}

type ArrayTypeExpression struct {
	Element TypeExpression
}

type MapTypeExpression struct {
	Key   TypeExpression
	Value TypeExpression
}

func (e *IdentifierTypeExpression) typeNode() {}
func (e *GenericArgumentList) typeNode()      {}
func (e *ArrayTypeExpression) typeNode()      {}
func (e *MapTypeExpression) typeNode()        {}
