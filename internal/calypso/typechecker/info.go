package typechecker

import (
	"go/constant"

	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/types"
)

type SymbolTable struct {

	// Scopes of nodes
	scopes map[ast.Node]*types.Scope
	nodes  map[ast.Node]*TypedNode
}

type TypedNode struct {
	Type   types.Type
	Value  constant.Value
	Symbol types.Symbol
}

func NewSymbolTable() *SymbolTable {
	return &SymbolTable{
		scopes: make(map[ast.Node]*types.Scope),
		nodes:  make(map[ast.Node]*TypedNode),
	}
}

func (t *SymbolTable) AddScope(n ast.Node, s *types.Scope) {
	t.scopes[n] = s
}

func (t *SymbolTable) GetScope(n ast.Node) (*types.Scope, bool) {
	v, ok := t.scopes[n]
	return v, ok
}

func (t *SymbolTable) AddNode(n ast.Node, typ types.Type, val constant.Value, sym types.Symbol) {
	t.nodes[n] = &TypedNode{
		Type:   typ,
		Value:  val,
		Symbol: sym,
	}
}

func (t *SymbolTable) GetNode(n ast.Node) (*TypedNode, bool) {
	v, ok := t.nodes[n]

	return v, ok
}
