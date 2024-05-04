package types

type Enum struct {
	Name     string
	Variants []*EnumVariant
}

type EnumVariant struct {
	Name         string
	Discriminant int
	Fields       []*Var
	Owner        *Enum
}

type EnumVariants []*EnumVariant

func (t *Enum) Parent() Type   { return t }
func (t *Enum) String() string { return t.Name }

func (t *EnumVariant) Parent() Type   { return t.Owner }
func (t *EnumVariant) String() string { return t.Owner.String() + "::" + t.Name }

func NewEnum(name string, cases []*EnumVariant) *Enum {
	e := &Enum{
		Name:     name,
		Variants: cases,
	}

	for _, cs := range cases {
		cs.Owner = e
	}

	return e
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
