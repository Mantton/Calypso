package lir

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/types"
)

type Module struct {

	// core
	TModule *types.Module // the typechecked module

	Functions       map[string]*Function  // functions to be built in this module
	GlobalConstants map[string]*Global    // global constants in this module
	Composites      map[string]*Composite // composites defined in this module
	Imports         map[string]*Module    // imports in this module

	// Generics
	GFunctions map[string]*GenericFunction // Maps symbols to their generic functions
	GTypes     map[string]*GenericType     // maps symbols to their generic type

	Enums  map[string]*EnumReference
	GEnums map[string]*GenericEnumReference
}

func NewModule(t *types.Module) *Module {
	return &Module{
		Functions:       make(map[string]*Function),
		GFunctions:      make(map[string]*GenericFunction),
		GlobalConstants: make(map[string]*Global),
		Composites:      make(map[string]*Composite),
		GTypes:          make(map[string]*GenericType),
		Imports:         make(map[string]*Module),
		Enums:           make(map[string]*EnumReference),
		GEnums:          make(map[string]*GenericEnumReference),
		TModule:         t,
	}
}

func (m *Module) ID() int64 {
	return m.TModule.ID()
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

	if symbol == nil {
		panic(fmt.Sprintf("nil symbol, %s", s))
	}
	s = symbol.SymbolName()
	return m.FindSymbol(s)
}

func (m *Module) FindSymbol(s string) Value {
	if v, ok := m.Functions[s]; ok {
		return v
	}

	if v, ok := m.GlobalConstants[s]; ok {
		return v
	}

	if v, ok := m.Enums[s]; ok {
		return v
	}

	if v, ok := m.GEnums[s]; ok {
		return v
	}

	if v, ok := m.Composites[s]; ok {
		return v
	}

	if v, ok := m.GFunctions[s]; ok {
		return v
	}

	if v, ok := m.GTypes[s]; ok {
		return v
	}

	return nil
}
