package types

type DefinedType struct {
	symbol
	wrapped        Type
	TypeParameters TypeParams
	Methods        map[string]*Function
}

func NewDefinedType(n string, t Type, p TypeParams) *DefinedType {
	return &DefinedType{
		symbol: symbol{
			name: n,
		},
		TypeParameters: p,
		wrapped:        t,
		Methods:        make(map[string]*Function),
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

func (s *DefinedType) AddMethod(n string, f *Function) bool {
	_, ok := s.Methods[n]

	if ok {
		return false
	}

	s.Methods[n] = f
	return true

}
