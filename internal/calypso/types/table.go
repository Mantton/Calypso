package types

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/ast"
)

type SymbolTable struct {

	// Scopes of nodes
	Main   *Scope
	scopes map[ast.Node]*Scope
	fns    map[*ast.FunctionExpression]*Function
	revFns map[*Function]*ast.FunctionExpression
	tNodes map[ast.Node]Type
}

func NewSymbolTable() *SymbolTable {
	return &SymbolTable{
		scopes: make(map[ast.Node]*Scope),
		fns:    make(map[*ast.FunctionExpression]*Function),
		revFns: make(map[*Function]*ast.FunctionExpression),
		tNodes: make(map[ast.Node]Type),
	}
}

func (t *SymbolTable) AddScope(n ast.Node, s *Scope) {
	t.scopes[n] = s
}

func (t *SymbolTable) GetScope(n ast.Node) (*Scope, bool) {
	v, ok := t.scopes[n]
	return v, ok
}

func (t *SymbolTable) DefineFunction(n *ast.FunctionExpression, typ *Function) {
	t.fns[n] = typ
	t.revFns[typ] = n
}

func (t *SymbolTable) GetFunction(n *ast.FunctionExpression) *Function {
	return t.fns[n]
}

func (t *SymbolTable) GetRevFunction(n *Function) *ast.FunctionExpression {
	return t.revFns[n]
}

func (t *SymbolTable) SetNodeType(n ast.Node, typ Type) {
	t.tNodes[n] = typ
}

func (t *SymbolTable) GetNodeType(n ast.Node) Type {
	return t.tNodes[n]
}

func (t *SymbolTable) DebugPrintScopes() {

	fmt.Println("GLOBAL")
	// fmt.Println(GlobalScope)
	GlobalScope.DebugPrintChildrenScopes()

	fmt.Println("PARENT")
	fmt.Println(t.Main)

	fmt.Println("NESTED")
	for _, scope := range t.scopes {
		if scope.IsEmpty() {
			continue
		}
		fmt.Println(scope)
	}

}
