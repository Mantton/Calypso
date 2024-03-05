package types

type FunctionInstance struct {
	Signature *FunctionSignature
	Arguments []Type
}

func NewFunctionInstance(sg *FunctionSignature, args []Type) *FunctionInstance {
	return &FunctionInstance{
		Signature: sg,
		Arguments: args,
	}
}

func (t *FunctionInstance) clyT() {}
func (t *FunctionInstance) String() string {

	f := t.Signature.String() + "<"

	for i, p := range t.Arguments {
		f += p.String()

		if i != len(t.Arguments)-1 {
			f += ", "
		}
	}

	f += ">"
	return f
}

func (t *FunctionInstance) Parent() Type { return t.Signature }
