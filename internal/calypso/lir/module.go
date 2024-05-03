package lir

import (
	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/types"
)

type Module struct {
	Functions       map[string]*Function
	GlobalConstants map[string]*Global
	Composites      map[string]*Composite
	TModule         *types.Module
	Imports         map[string]*Module
	GFunctions      map[string]*GenericFunction // Maps symbols to their generic functions
}

func NewModule(mod *types.Module) *Module {
	return &Module{
		Functions:       make(map[string]*Function),
		GFunctions:      make(map[string]*GenericFunction),
		GlobalConstants: make(map[string]*Global),
		Composites:      make(map[string]*Composite),
		Imports:         make(map[string]*Module),
		TModule:         mod,
	}
}

func (m *Module) Name() string {
	return m.TModule.Name()
}

func (m *Module) FileSet() *ast.FileSet {
	return m.TModule.AST.Set
}

func (m *Module) Yields() types.Type {
	return m.TModule
}

func (m *Module) Find(s string) Value {

	symbol := m.TModule.Scope.MustResolve(s)

	s = symbol.SymbolName()
	if v, ok := m.Functions[s]; ok {
		return v
	}

	if v, ok := m.GlobalConstants[s]; ok {
		return v
	}

	if v, ok := m.Composites[s]; ok {
		return v
	}

	return nil
}
