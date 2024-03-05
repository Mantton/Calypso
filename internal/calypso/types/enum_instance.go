package types

type EnumInstance struct {
	TypeArgs []Type
	Type     Type
}

func (t *EnumInstance) clyT() {}
func (t *EnumInstance) String() string {

	f := t.Type.String() + "<"

	for i, p := range t.TypeArgs {
		f += p.String()

		if i != len(t.TypeArgs)-1 {
			f += ", "
		}
	}

	f += ">"
	return f
}
func (t *EnumInstance) Parent() Type { return t }

func NewEnumInstance(f Type, args []Type) *EnumInstance {
	return &EnumInstance{
		Type:     f,
		TypeArgs: args,
	}
}
