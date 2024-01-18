package ast

type Node interface {
	Start() int
	End() int
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

// * Declarations
// - Imports, Modules, Structs, Types
type ConstantDeclaration struct {
	Ident Expression
	Value Expression
	// Type  Expression
}

type File struct {
	ModuleName   string
	Declarations []Declaration
	Errors       []string
}
