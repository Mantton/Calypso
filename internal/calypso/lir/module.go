package lir

import (
	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/types"
)

type Module struct {
	Functions       map[string]*Function
	GlobalConstants map[string]*Global
	Composites      map[types.Type]*Composite
	TModule         *types.Module
}

func NewModule(mod *types.Module) *Module {
	return &Module{
		Functions:       make(map[string]*Function),
		GlobalConstants: make(map[string]*Global),
		Composites:      make(map[types.Type]*Composite),
		TModule:         mod,
	}
}

func (m *Module) Name() string {
	return m.TModule.Name()
}

// func (m *Module) TypeTable() *types.SymbolTable {
// 	return m.TModule.Table
// }

func (m *Module) FileSet() *ast.FileSet {
	return m.TModule.AST.Set
}
