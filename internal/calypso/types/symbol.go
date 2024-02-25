package types

type Symbol interface {
	Owner() *Scope
	Name() string
	Type() Type // object type
	SetScope(*Scope)
	SetType(Type)
}

type symbol struct {
	owner *Scope
	name  string
	typ   Type
}

func (e *symbol) Name() string { return e.name }

func (e *symbol) Owner() *Scope     { return e.owner }
func (e *symbol) SetScope(s *Scope) { e.owner = s }

func (e *symbol) SetType(t Type) { e.typ = t }
func (e *symbol) Type() Type     { return e.typ }
