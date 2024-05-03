package types

type DefinedType struct {
	symbol
	wrapped        Type
	TypeParameters TypeParams
	scope          *Scope
}

func NewBaseDefinedType(name string, wrapped Type, params TypeParams, scope *Scope, mod *Module) *DefinedType {
	return &DefinedType{
		symbol: symbol{
			name: name,
			mod:  mod,
		},
		TypeParameters: params,
		wrapped:        wrapped,
		scope:          scope,
	}
}

func (s *DefinedType) AddTypeParameter(t *TypeParam) error {
	s.TypeParameters = append(s.TypeParameters, t)
	return s.scope.Define(t)
}

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
	e.wrapped = t
}

func (s *DefinedType) AddMethod(n string, f *Function) error {
	return s.scope.Define(f)
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

func (n *DefinedType) ResolveMethod(s string) Type {

	symbol := n.scope.ResolveInCurrent(s)

	// Not found in current, find & specialize from instance
	if symbol == nil {
		return nil
	}

	// match function types
	switch fn := symbol.(type) {
	case *Function:
		return fn.Type()
	case *FunctionSet:
		return fn
	}

	return nil
}

func (n *DefinedType) ResolveType(s string) Type {
	symbol := n.scope.ResolveInCurrent(s)

	// Not found in current, find & specialize from instance
	if symbol == nil {
		return nil
	}

	typ, ok := symbol.(Type)

	if !ok {
		return nil
	}

	return typ
}

func (n *DefinedType) ResolveField(s string) Type {

	// Is Method
	if method := n.ResolveMethod(s); method != nil {
		return method
	}

	// Is Field

	// Access Field
	switch parent := n.Parent().(type) {
	case *Struct:
		target, ok := parent.Map[s]

		if ok {
			return target.Type()
		}

	case *Enum:
		for _, v := range parent.Variants {
			if v.Name == s {
				// Not tuple type, return parent type
				if len(v.Fields) == 0 {
					return n
				}

				// Tuple Type, Return Function Returning Parent Type
				sg := NewFunctionSignature()
				for _, p := range v.Fields {
					sg.AddParameter(p)
				}

				sg.Result.SetType(n)
				sg.TypeParameters = n.TypeParameters
				return sg
			}
		}
	default:
		return nil
	}

	return nil

}

func (n *DefinedType) GetScope() *Scope {
	return n.scope
}

type HType interface {
	ResolveField(string) Type
	ResolveType(string) Type
	ResolveMethod(string) Type
}
