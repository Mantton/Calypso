package types

import "github.com/mantton/calypso/internal/calypso/ast"

type SymbolTable struct {
	Symbols         map[Symbol]ast.Node // This links symbols to their corresponding nodes
	Nodes           map[ast.Node]Type   // this links nodes to their corresponding types
	Specializations map[string]Type
}

func NewSymbolTable() *SymbolTable {
	return &SymbolTable{
		Symbols:         make(map[Symbol]ast.Node),
		Nodes:           make(map[ast.Node]Type),
		Specializations: make(map[string]Type),
	}
}

func (t *SymbolTable) SetSymbol(s Symbol, n ast.Node) {
	t.Symbols[s] = n
}

func (t *SymbolTable) GetSymbol(s Symbol) ast.Node {
	return t.Symbols[s]
}

func (t *SymbolTable) SetNodeType(n ast.Node, typ Type) {
	t.Nodes[n] = typ
}

func (t *SymbolTable) GetNodeType(n ast.Node) Type {
	return t.Nodes[n]
}
