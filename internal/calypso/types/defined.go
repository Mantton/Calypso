package types

type DefinedType struct {
	symbol
	wrapped        Type
	TypeParameters TypeParams
	scope          *Scope
}

func NewBaseDefinedType(name string, wrapped Type, params TypeParams, scope *Scope) *DefinedType {
	return &DefinedType{
		symbol: symbol{
			name: name,
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

	if n.scope == nil {
		return n.instantiateMethod(s)
	}

	symbol := n.scope.ResolveInCurrent(s)

	// Not found in current, find & specialize from instance
	if symbol == nil {
		return n.instantiateMethod(s)
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

	if n.scope == nil {
		return n.instantiateType(s)
	}

	symbol := n.scope.ResolveInCurrent(s)

	// Not found in current, find & specialize from instance
	if symbol == nil {
		return n.instantiateType(s)
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
				return sg
			}
		}
	default:
		return nil
	}

	return nil

}

func (n *DefinedType) initializeMappings() {
	// if n.mappings == nil {
	// 	n.mappings = make(map[string]Type)
	// }

	// for _, t := range n.TypeParameters {
	// 	n.mappings[t.name] = t.Unwrapped()
	// }
	panic("???")
}

func (n *DefinedType) GetScope() *Scope {
	return n.scope
}

func (n *DefinedType) instantiateType(sym string) Type {
	// no instance of, is parent with none found
	// if n.InstanceOf == nil {
	// 	return nil
	// }

	// symbol := n.InstanceOf.ResolveType(sym)

	// // checked parent, not found
	// if symbol == nil {
	// 	return nil
	// }

	// if IsGeneric(symbol) {
	// 	n.initializeMappings()

	// 	return Apply(n.mappings, symbol)
	// }

	// return symbol

	panic("not ready!")
}

func (n *DefinedType) instantiateMethod(sym string) Type {
	// no instance of, is parent with none found
	// if n.InstanceOf == nil {
	// 	return nil
	// }

	// symbol := n.InstanceOf.ResolveMethod(sym)

	// // checked parent, not found
	// if symbol == nil {
	// 	return nil
	// }

	// if IsGeneric(symbol) {
	// 	n.initializeMappings()

	// 	return Apply(n.mappings, symbol).(*FunctionSignature)
	// }

	// return symbol
	panic("wut!")
}
