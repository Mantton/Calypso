package types

type StructInstance struct {
	TypeArgs []Type
	Type     Type
}

func (t *StructInstance) clyT() {}
func (t *StructInstance) String() string {

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
func (t *StructInstance) Parent() Type { return t }

func NewStructInstance(f Type, args []Type) *StructInstance {
	return &StructInstance{
		Type:     f,
		TypeArgs: args,
	}
}
