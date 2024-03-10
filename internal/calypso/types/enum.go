package types

type Enum struct {
	Name     string
	Variants []*EnumVariant
}

type EnumVariant struct {
	Name         string
	Discriminant int
	Fields       []*Var
}

type EnumVariants []*EnumVariant

func (t *Enum) clyT()          {}
func (t *Enum) Parent() Type   { return t }
func (t *Enum) String() string { return t.Name }

func NewEnum(name string, cases []*EnumVariant) *Enum {
	return &Enum{
		Name:     name,
		Variants: cases,
	}
}

func NewEnumVariant(name string, discriminant int, fields []*Var) *EnumVariant {
	return &EnumVariant{
		Name:         name,
		Discriminant: discriminant,
		Fields:       fields,
	}
}
