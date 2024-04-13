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

func (t *Enum) Parent() Type   { return t }
func (t *Enum) String() string { return t.Name }

func (t *EnumVariant) Parent() Type   { return t }
func (t *EnumVariant) String() string { return t.Name }

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

func (e *Enum) IsUnion() bool {
	for _, v := range e.Variants {
		if len(v.Fields) != 0 {
			return true
		}
	}

	return false
}

func (e *Enum) FindVariant(n string) *EnumVariant {
	for _, x := range e.Variants {
		if x.Name == n {
			return x
		}
	}

	return nil
}

func IsUnionEnum(t Type) bool {
	x, ok := t.Parent().(*Enum)

	if !ok {
		return false
	}

	return x.IsUnion()
}
