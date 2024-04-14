package types

import "fmt"

// a named module entity such as a struct definition, constant, variable or function
type Symbol interface {
	Name() string
	Type() Type // object type
	Module() *Module
	SymbolName() string
}

type symbol struct {
	name string
	typ  Type
	mod  *Module
}

func (e *symbol) Name() string { return e.name }
func (e *symbol) SetType(t Type) {
	e.typ = t
}
func (e *symbol) Type() Type      { return e.typ }
func (e *symbol) Module() *Module { return e.mod }
func (e *symbol) SymbolName() string {
	v := fmt.Sprintf("%s::%s::%s", e.mod.pkg.name, e.mod.name, e.name)
	return v
}
