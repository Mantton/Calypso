package types

type Var struct {
	symbol
	Mutable    bool
	ParamLabel string
}

func NewVar(name string, t Type) *Var {
	return &Var{
		symbol: symbol{
			name: name,
			typ:  t,
		},
	}
}
