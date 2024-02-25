package types

type TypeDef struct {
	symbol
}

func NewTypeDef(name string, t Type) *TypeDef {
	return &TypeDef{
		symbol: symbol{
			name: name,
			typ:  t,
		},
	}
}
