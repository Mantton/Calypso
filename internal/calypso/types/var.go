package types

type Var struct {
	symbol
	Mutable     bool
	ParamLabel  string
	StructIndex int
}

func NewVar(name string, t Type, mod *Module) *Var {
	return &Var{
		symbol: symbol{
			name: name,
			typ:  t,
			mod:  mod,
		},
		StructIndex: -1,
	}
}

func AsVar(t Symbol) *Var {
	if t == nil {
		return nil
	}
	v, ok := t.(*Var)

	if !ok {
		return nil
	}

	return v
}
