package types

type Instance struct {
	TypeArgs []Type
	Type     Type
}

func (t *Instance) clyT() {}
func (t *Instance) String() string {

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
func (t *Instance) Parent() Type { return t }

func NewInstance(f Type, args []Type) *Instance {
	return &Instance{
		Type:     f,
		TypeArgs: args,
	}
}
