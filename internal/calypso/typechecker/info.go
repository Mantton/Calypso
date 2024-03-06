package typechecker

import (
	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/types"
)

type SymbolTable struct {

	// Scopes of nodes
	Main        *types.Scope
	scopes      map[ast.Node]*types.Scope
	fns         map[*ast.FunctionExpression]*types.Function
	tNodes      map[ast.Node]types.Type
	fnInstances map[*ast.FunctionCallExpression]*types.FunctionInstance
}

func NewSymbolTable() *SymbolTable {
	return &SymbolTable{
		scopes:      make(map[ast.Node]*types.Scope),
		fns:         make(map[*ast.FunctionExpression]*types.Function),
		tNodes:      make(map[ast.Node]types.Type),
		fnInstances: make(map[*ast.FunctionCallExpression]*types.FunctionInstance),
	}
}

func (t *SymbolTable) AddScope(n ast.Node, s *types.Scope) {
	t.scopes[n] = s
}

func (t *SymbolTable) GetScope(n ast.Node) (*types.Scope, bool) {
	v, ok := t.scopes[n]
	return v, ok
}

func (t *SymbolTable) DefineFunction(n *ast.FunctionExpression, typ *types.Function) {
	t.fns[n] = typ
}

func (t *SymbolTable) GetFunction(n *ast.FunctionExpression) *types.Function {
	return t.fns[n]
}

func (t *SymbolTable) SetNodeType(n ast.Node, typ types.Type) {
	t.tNodes[n] = typ
}

func (t *SymbolTable) GetNodeType(n ast.Node) types.Type {
	return t.tNodes[n]
}

func (t *SymbolTable) SetFunctionInstance(n *ast.FunctionCallExpression, typ *types.FunctionInstance) {
	t.fnInstances[n] = typ
}

func (t *SymbolTable) GetFunctionInstance(n *ast.FunctionCallExpression) *types.FunctionInstance {
	return t.fnInstances[n]
}
