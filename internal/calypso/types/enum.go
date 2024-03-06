package types

type Enum struct {
	Name  string
	Cases []*EnumVariant
}

type EnumVariant struct {
	Name         string
	Discriminant int
	Fields       []*Var
}

func (t *Enum) clyT()          {}
func (t *Enum) Parent() Type   { return t }
func (t *Enum) String() string { return t.Name }

func NewEnum(name string, cases []*EnumVariant) *Enum {
	return &Enum{
		Name:  name,
		Cases: cases,
	}
}

func NewEnumVariant(name string, discriminant int, fields []*Var) *EnumVariant {
	return &EnumVariant{
		Name:         name,
		Discriminant: discriminant,
		Fields:       fields,
	}
}
