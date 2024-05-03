package types

import (
	"fmt"
)

// a named module entity such as a struct definition, constant, variable or function
type Symbol interface {
	Name() string
	Type() Type // object type
	Module() *Module
	SymbolName() string
	IsPublic() bool
	IsVisible(from *Module) bool
}

type symbol struct {
	name string
	typ  Type
	mod  *Module
	pub  bool
}

func (e *symbol) Name() string { return e.name }
func (e *symbol) SetType(t Type) {
	e.typ = t
}
func (e *symbol) Type() Type      { return e.typ }
func (e *symbol) Module() *Module { return e.mod }
func (e *symbol) SymbolName() string {
	v := fmt.Sprintf("%s::%s", e.mod.SymbolName(), e.name)
	return v
}

func (e *symbol) IsVisible(from *Module) bool {
	if e.mod == nil {
		return true
	}
	if from == e.mod {
		return true
	}

	return e.IsPublic()
}

func (e *symbol) IsPublic() bool {
	return e.pub
}

func (e *symbol) SetVisibility(b bool) {
	e.pub = b
}
