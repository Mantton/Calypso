package types

// a named module entity such as a struct definition, constant, variable or function
type Symbol interface {
	Name() string
	Type() Type // object type
}

type symbol struct {
	name string
	typ  Type
}

func (e *symbol) Name() string { return e.name }
func (e *symbol) SetType(t Type) {
	e.typ = t
}
func (e *symbol) Type() Type { return e.typ }
