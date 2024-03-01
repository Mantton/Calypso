package types

type DefinedType struct {
	symbol
	wrapped        Type
	TypeParameters TypeParams
}

func NewDefinedType(n string, t Type, p TypeParams) *DefinedType {
	return &DefinedType{
		symbol: symbol{
			name: n,
		},
		TypeParameters: p,
		wrapped:        t,
	}
}

func (s *DefinedType) AddTypeParameter(t *TypeParam) {
	s.TypeParameters = append(s.TypeParameters, t)
}

func (t *DefinedType) clyT()        {}
func (t *DefinedType) Parent() Type { return t.wrapped.Parent() }

func (t *DefinedType) Type() Type {
	return t
}
func (t *DefinedType) String() string {
	return t.Name()

}
func (e *DefinedType) SetType(t Type) {
	e.wrapped = ResolveLiteral(t)
}
