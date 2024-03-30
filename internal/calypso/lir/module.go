package lir

import (
	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/typechecker"
	"github.com/mantton/calypso/internal/calypso/types"
)

type Module struct {
	Functions         map[string]*Function
	GlobalConstants   map[string]*Global
	ExternalFunctions map[string]*types.FunctionSignature
	Composites        map[string]types.Symbol
	TModule           *typechecker.Module
}

func NewModule(mod *typechecker.Module) *Module {
	return &Module{
		Functions:       make(map[string]*Function),
		GlobalConstants: make(map[string]*Global),
		Composites:      make(map[string]types.Symbol),
		TModule:         mod,
	}
}

func (m *Module) Name() string {
	return m.TModule.FileSet.ModuleName
}

func (m *Module) TypeTable() *typechecker.SymbolTable {
	return m.TModule.Table
}

func (m *Module) FileSet() *ast.FileSet {
	return m.TModule.FileSet
}
