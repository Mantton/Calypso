package ast

import (
	"github.com/mantton/calypso/internal/calypso/lexer"
	"github.com/mantton/calypso/internal/calypso/token"
	"github.com/mantton/calypso/internal/calypso/types"
)

type Node interface {
	Range() token.SyntaxRange
	Accept(visitor)
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
	Errors       lexer.ErrorList
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

type StatementDeclaration struct {
	Stmt Statement
}

type StandardDeclaration struct {
	KeyWPos    token.TokenPosition
	Identifier *IdentifierExpression
	Block      *BlockStatement
}

type TypeDeclaration struct {
	KeyWPos       token.TokenPosition
	Identifier    *IdentifierExpression
	EqPos         token.TokenPosition
	GenericParams *GenericParametersClause
	Value         TypeExpression
}

type ExtensionDeclaration struct {
	KeyWPos    token.TokenPosition
	Identifier *IdentifierExpression
	LBracePos  token.TokenPosition
	Content    []*FunctionStatement
	RBracePos  token.TokenPosition
}

type ConformanceDeclaration struct {
	KeyWPos   token.TokenPosition
	Standard  *IdentifierExpression
	Target    *IdentifierExpression
	LBracePos token.TokenPosition
	Content   []*FunctionStatement
	RBracePos token.TokenPosition
}

// * Statements
type BlockStatement struct {
	LBrackPos  token.TokenPosition
	RBrackPos  token.TokenPosition
	Statements []Statement
}

type VariableStatement struct {
	KeyWPos    token.TokenPosition
	Identifier *IdentifierExpression
	Value      Expression
	IsConstant bool
	IsGlobal   bool
}

type FunctionStatement struct {
	Func *FunctionExpression
}

type IfStatement struct {
	KeyWPos     token.TokenPosition
	Condition   Expression
	Action      *BlockStatement
	Alternative *BlockStatement
}

type ReturnStatement struct {
	KeyWPos token.TokenPosition
	Value   Expression
}

type WhileStatement struct {
	KeyWPos   token.TokenPosition
	Condition Expression
	Action    *BlockStatement
}

type ExpressionStatement struct {
	Expr Expression
}

type AliasStatement struct {
	KeyWPos       token.TokenPosition
	EqPos         token.TokenPosition
	Target        TypeExpression
	Identifier    *IdentifierExpression
	GenericParams *GenericParametersClause
}

type StructStatement struct {
	KeyWPos       token.TokenPosition
	Identifier    *IdentifierExpression
	GenericParams *GenericParametersClause
	LBracePos     token.TokenPosition
	RBracePos     token.TokenPosition
	Properties    []*IdentifierExpression
}

// * Expressions

type GroupedExpression struct {
	LParenPos token.TokenPosition
	Expr      Expression
	RParenPos token.TokenPosition
}
type UnaryExpression struct {
	Op         token.Token
	OpPosition token.TokenPosition
	Expr       Expression
}

type BinaryExpression struct {
	Left  Expression
	Op    token.Token
	OpPos token.TokenPosition
	Right Expression
}

type AssignmentExpression struct {
	Target Expression
	OpPos  token.TokenPosition
	Value  Expression
}

type CallExpression struct {
	Target    Expression
	Arguments []Expression
	LParenPos token.TokenPosition
	RParenPos token.TokenPosition
}

type IndexExpression struct {
	Target      Expression
	Index       Expression
	LBracketPos token.TokenPosition
	RBracketPos token.TokenPosition
}

type PropertyExpression struct {
	Target   Expression
	Property Expression
	DotPos   token.TokenPosition
}

type KeyValueExpression struct {
	Key      Expression
	Value    Expression
	ColonPos token.TokenPosition
}

// * Literal Expressions

type IdentifierExpression struct {
	Pos           token.TokenPosition
	Value         string
	AnnotatedType TypeExpression
}

type FunctionExpression struct {
	KeyWPos       token.TokenPosition
	Identifier    *IdentifierExpression
	Body          *BlockStatement
	Params        []*IdentifierExpression
	GenericParams *GenericParametersClause
	RParenPos     token.TokenPosition
	ReturnType    TypeExpression
	Signature     *types.Function
}

// * Literals
type IntegerLiteral struct {
	Pos           token.TokenPosition
	Value         string
	ResolvedValue uint64
}

type FloatLiteral struct {
	Pos           token.TokenPosition
	Value         string
	ResolvedValue float64
}

type StringLiteral struct {
	Pos           token.TokenPosition
	Value         string
	ResolvedValue string
}

type CharLiteral struct {
	Pos           token.TokenPosition
	Value         string
	ResolvedValue rune
}

type BooleanLiteral struct {
	Pos   token.TokenPosition
	Value bool
}

type NullLiteral struct {
	Pos token.TokenPosition
}

type VoidLiteral struct {
	Pos token.TokenPosition
}

type ArrayLiteral struct {
	LBracketPos token.TokenPosition
	Elements    []Expression
	RBracketPos token.TokenPosition
}

type MapLiteral struct {
	LBracePos token.TokenPosition
	Pairs     []*KeyValueExpression
	RBracePos token.TokenPosition
}

type CompositeLiteral struct {
	Identifier *IdentifierExpression
	LBracePos  token.TokenPosition
	RBracePos  token.TokenPosition
	Pairs      []*CompositeLiteralBodyClause
}

type CompositeLiteralBodyClause struct {
	Key      *IdentifierExpression
	Value    Expression
	ColonPos token.TokenPosition
}

// * Types

type TypeExpression interface {
	Node
	typeNode()
}

type IdentifierTypeExpression struct {
	Identifier *IdentifierExpression
	Arguments  *GenericArgumentsClause
}

type GenericArgumentsClause struct {
	Arguments   []TypeExpression
	LChevronPos token.TokenPosition
	RChevronPos token.TokenPosition
}

type ArrayTypeExpression struct {
	LBracketPos token.TokenPosition
	Element     TypeExpression
	RBracketPos token.TokenPosition
}

type MapTypeExpression struct {
	Key         TypeExpression
	Value       TypeExpression
	LBracketPos token.TokenPosition
	RBracketPos token.TokenPosition
}

type FunctionTypeExpression struct {
	Identifier *IdentifierExpression
	Arguments  []TypeExpression
	ReturnType TypeExpression
	// Generic Params
	Params *GenericParametersClause
}

// * Misc
type GenericParametersClause struct {
	Parameters  []*GenericParameterExpression
	LChevronPos token.TokenPosition
	RChevronPos token.TokenPosition
}

type GenericParameterExpression struct {
	Identifier *IdentifierExpression
	Standards  []*IdentifierExpression
}
