package ast

import (
	"github.com/mantton/calypso/internal/calypso/lexer"
	"github.com/mantton/calypso/internal/calypso/token"
)

type Node interface {
	Range() token.SyntaxRange
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

type ExternDeclaration struct {
	KeyWPos   token.TokenPosition
	Target    *StringLiteral
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
	Fields        []*IdentifierExpression
}

type EnumStatement struct {
	KeyWPos       token.TokenPosition
	Identifier    *IdentifierExpression
	GenericParams *GenericParametersClause
	LBracePos     token.TokenPosition
	Variants      []*EnumVariantExpression
	RBracePos     token.TokenPosition
}

type EnumVariantExpression struct {
	Identifier   *IdentifierExpression
	Fields       *FieldListExpression
	Discriminant *EnumDiscriminantExpression
}

type FieldListExpression struct {
	LParenPos token.TokenPosition
	Fields    []TypeExpression
	RParenPos token.TokenPosition
}

type EnumDiscriminantExpression struct {
	EqPos token.TokenPosition
	Value Expression
}

type SwitchStatement struct {
	KeyWPos   token.TokenPosition
	Condition Expression
	LBracePos token.TokenPosition
	Cases     []*SwitchCaseExpression
	RBracePos token.TokenPosition
}

type SwitchCaseExpression struct {
	IsDefault bool
	KeyWPos   token.TokenPosition
	Condition Expression
	ColonPos  token.TokenPosition
	Action    *BlockStatement
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

type FunctionCallExpression struct {
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

type FieldAccessExpression struct {
	Target Expression
	Field  Expression
	DotPos token.TokenPosition
}

type KeyValueExpression struct {
	Key      Expression
	Value    Expression
	ColonPos token.TokenPosition
}

type GenericSpecializationExpression struct {
	Identifier *IdentifierExpression
	Clause     *GenericArgumentsClause
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
}

// * Literals
type IntegerLiteral struct {
	Pos   token.TokenPosition
	Value int64
}

type FloatLiteral struct {
	Pos   token.TokenPosition
	Value float64
}

type StringLiteral struct {
	Pos   token.TokenPosition
	Value string
}

type CharLiteral struct {
	Pos   token.TokenPosition
	Value int64
}

type BooleanLiteral struct {
	Pos   token.TokenPosition
	Value bool
}

type NilLiteral struct {
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
	Identifier    *IdentifierExpression
	TypeArguments *GenericArgumentsClause
	Body          *CompositeLiteralBody
}

type CompositeLiteralBody struct {
	LBracePos token.TokenPosition
	RBracePos token.TokenPosition
	Fields    []*CompositeLiteralField
}

type CompositeLiteralField struct {
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

type PointerTypeExpression struct {
	PointerTo TypeExpression
	StarPos   token.TokenPosition
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
