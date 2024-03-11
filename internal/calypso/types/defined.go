package types

type DefinedType struct {
	symbol
	wrapped        Type
	InstanceOf     Type
	TypeParameters TypeParams
	Methods        map[string]*Function
	Scope          *Scope
}

func NewDefinedType(n string, t Type, p TypeParams, parentScope *Scope) *DefinedType {
	return &DefinedType{
		symbol: symbol{
			name: n,
		},
		TypeParameters: p,
		wrapped:        t,
		Methods:        make(map[string]*Function),
		Scope:          NewScope(parentScope),
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
	if len(t.TypeParameters) == 0 {
		return t.Name()
	} else {
		f := t.Name() + "<"

		for i, p := range t.TypeParameters {
			f += p.String()

			if i != len(t.TypeParameters)-1 {
				f += ", "
			}
		}
		f += ">"
		return f
	}
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

func AsDefined(t Type) *DefinedType {
	if a, ok := t.(*DefinedType); ok {
		return a
	}

	if b, ok := t.(*Alias); ok {
		return AsDefined(b.RHS)
	}
	return nil

}

func GetTypeParams(t Type) []*TypeParam {
	switch t := t.(type) {
	case *DefinedType:
		return t.TypeParameters
	case *FunctionSignature:
		return t.TypeParameters
	case *Alias:
		return t.TypeParameters
	}

	return nil
}

func ResolveTypeParameters(t Type) TypeParams {
	switch t := t.(type) {
	case *DefinedType:
		return t.TypeParameters
	case *FunctionSignature:
		return t.TypeParameters
	}
	return nil
}
