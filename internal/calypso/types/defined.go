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
	return nil

}

func GetTypeParams(t Type) []*TypeParam {
	switch t := t.(type) {
	case *DefinedType:
		return t.TypeParameters
	case *FunctionSignature:
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

// func ResolveField(field string, typ Type) Type {
// 	fmt.Printf("Resolving `%s` for type `%s`\n", field, typ) // Resolve Field
// 	parent := ResolveInstanceParent(typ)

// 	if parent == nil {
// 		return nil
// 	}

// 	params := ResolveTypeParameters(parent)

// 	switch t := parent.Parent().(type) {
// 	case *Enum:
// 		for _, c := range t.Cases {
// 			if c.Name == field {
// 				if len(c.Fields) != 0 {
// 					sg := NewFunctionSignature()
// 					sg.Result.SetType(typ)
// 					sg.TypeParameters = params
// 					for _, f := range c.Fields {
// 						sg.AddParameter(f)
// 					}

// 					if c := AsDefined(typ); c != nil {
// 						sg.TypeParameters = c.TypeParameters
// 					}
// 					return sg
// 				} else {
// 					return typ
// 				}
// 			}
// 		}

// 	case *Struct:
// 		for _, c := range t.Fields {
// 			if c.Name() == field {
// 				return c.Type()
// 			}
// 		}
// 	}

// 	fn, ok := parent.Methods[field]

// 	if !ok {
// 		fmt.Println("not found:", field)
// 		return nil
// 	}

// 	return fn.Sg()
// }
