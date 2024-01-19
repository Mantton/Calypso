package ast

import "github.com/mantton/calypso/internal/calypso/token"

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
	ModuleName   string
	Declarations []Declaration
	Errors       []string
}

// * Declarations
// - Imports, Modules, Structs, Types
type ConstantDeclaration struct {
	Stmt *VariableStatement
	// Type  Expression
}

type FunctionDeclaration struct {
	Func *FunctionLiteral
}

func (d *ConstantDeclaration) declarationNode() {}
func (d *FunctionDeclaration) declarationNode() {}

// * Statements
type BlockStatement struct {
	Statements []Statement
}

type VariableStatement struct {
	Identifier string
	Value      Expression
	IsConstant bool
}

type FunctionStatement struct {
	Func *FunctionLiteral
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

type AssignmentStatement struct {
	Ident string
	Value Expression
}

func (s *IfStatement) statementNode()         {}
func (s *AssignmentStatement) statementNode() {}
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

func (e *GroupedExpression) expressionNode() {}
func (e *UnaryExpression) expressionNode()   {}
func (e *BinaryExpression) expressionNode()  {}

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

type FunctionLiteral struct {
	Name string
	Body Statement
}

type IdentifierLiteral struct {
	Value string
}

func (e *IntegerLiteral) expressionNode()    {}
func (e *FloatLiteral) expressionNode()      {}
func (e *StringLiteral) expressionNode()     {}
func (e *BooleanLiteral) expressionNode()    {}
func (e *NullLiteral) expressionNode()       {}
func (e *VoidLiteral) expressionNode()       {}
func (e *IdentifierLiteral) expressionNode() {}
