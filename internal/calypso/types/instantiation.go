package types

import "fmt"

type Specialization = map[Type]Type

func HashValue(m Specialization, l TypeParams) string {
	str := ""

	if m == nil {
		panic("how!")
	}
	for _, x := range l {
		p, ok := m[x]
		if !ok {
			panic(fmt.Sprintf("%s not found in %s", x.name, m))
		}
		str += p.String()
	}

	return str
}

func Instantiate(t Type, ctx Specialization) Type {

	var out Type

	// Non Generic Type, No Specialization Needed
	if !IsGeneric(t) {
		return t
	}

	if ctx == nil {
		ctx = make(Specialization)
	}

	fmt.Printf("\tSubstituting %s with %s\n", t, ctx)

	switch t := t.(type) {

	case *TypeParam:
		// Check If specialized
		typ, ok := ctx[t]

		// Should always be specialized otherwise there is a problem elsewhere
		if !ok {
			panic("should be specialized")
		}

		// return specialized type
		return typ
	case *DefinedType:
		// declare bound list
		bounds := TypeList{}

		// Collect Bounded Types
		for _, p := range t.TypeParameters {
			arg, ok := ctx[p]

			if !ok {
				panic("should be specialized")
			}

			bounds = append(bounds, arg)
		}
		// Specialize underlying type
		typ := NewSpecializedType(t, bounds)
		return typ
	case *SpecializedType:
		bounds := TypeList{}
		for _, bound := range t.Bounds {
			bounds = append(bounds, Instantiate(bound, ctx))
		}

		return NewSpecializedType(t.InstanceOf, bounds)
	case *FunctionSignature:
		sg := NewFunctionSignature()

		// Parameters
		for _, param := range t.Parameters {
			p := NewVar(param.Name(), LookUp(Unresolved))
			p.SetType(Instantiate(param.Type(), ctx))
			p.ParamLabel = param.ParamLabel
			sg.AddParameter(p)
		}

		// return
		res := t.Result.Type()
		sg.Result.SetType(Instantiate(res, ctx))
		return sg
	case *Alias:
		return Instantiate(t.RHS, ctx)
	case *Pointer:
		cT := t.PointerTo          // Type Pointing To
		uT := Instantiate(cT, ctx) // Instantiate Type with Specialization Map
		out = NewPointer(uT)       // Create new pointer with specialized type
		return out                 // return updated pointer
	default:
		// unimplemented instantiation
		panic(fmt.Sprintf("cannot instantiate type %s", t))
	}
}

// func apply(ctx Specialization, typ Type) Type {

// 	fmt.Printf("\tSubstituting %s with %s\n", typ, ctx)
// 	if !IsGeneric(typ) {
// 		fmt.Println("\tSkipping Non Generic", typ)
// 		return typ
// 	}
// 	switch t := typ.(type) {

// 	case *TypeParam:
// 		sp, ok := ctx[t.Name()]
// 		if !ok {
// 			return LookUp(Unresolved)
// 		}

// 		return sp
// 	case *DefinedType:

// 	case *FunctionSignature:
// 		sg := NewFunctionSignature()

// 		// Parameters
// 		for _, param := range t.Parameters {
// 			p := NewVar(param.Name(), LookUp(Unresolved))
// 			p.SetType(Apply(ctx, param.Type()))
// 			p.ParamLabel = param.ParamLabel
// 			sg.AddParameter(p)
// 		}

// 		// return
// 		res := t.Result.Type()
// 		sg.Result.SetType(Apply(ctx, res))
// 		t.AddInstance(sg, ctx)
// 		return sg
// 	case *Alias:
// 		params := TypeParams{}
// 		for _, p := range t.TypeParameters {

// 			arg, ok := ctx[p.Name()]

// 			if !ok {
// 				return LookUp(Unresolved)
// 			}

// 			switch aT := arg.(type) {
// 			case *TypeParam:
// 				if aT.Bound != nil {
// 					params = append(params, NewTypeParam(aT.Name(), nil, arg, p.Module()))
// 				} else {
// 					params = append(params, NewTypeParam(aT.Name(), aT.Constraints, nil, p.Module()))
// 				}
// 			default:
// 				params = append(params, NewTypeParam(p.Name(), nil, arg, p.Module()))
// 			}
// 		}
// 		alias := NewAlias(t.Name(), nil)

// 		rhs := Apply(ctx, t.RHS)
// 		alias.SetType(rhs)
// 		alias.TypeParameters = params
// 		return alias
// 	case *Pointer:
// 		cT := t.PointerTo
// 		uT := Apply(ctx, cT)
// 		ptr := NewPointer(uT)
// 		return ptr
// 	default:
// 		panic(fmt.Sprintf("cannot instantiate type %s", t))
// 	}
// }

// TODO: Find better way
func cloneWithSpecialization(t Type, ctx Specialization) Type {
	switch parent := t.(type) {
	case *Basic:
		return t // Basic Types cannot be specialized
	case *Struct:
		// Collect Fields
		fields := []*Var{}
		for _, field := range parent.Fields {
			s := Instantiate(field.Type(), ctx)
			spec := NewVar(field.Name(), s)
			fields = append(fields, spec)
		}
		return NewStruct(fields)
	case *Enum:
		variants := EnumVariants{}

		for _, variant := range parent.Variants {
			if len(variant.Fields) == 0 {
				variants = append(variants, NewEnumVariant(variant.Name, variant.Discriminant, nil))
				continue
			}

			fields := []*Var{}

			for _, f := range variant.Fields {
				v := NewVar(f.Name(), nil)
				s := Instantiate(f.Type(), ctx)
				v.SetType(s)
				fields = append(fields, v)
			}

			variants = append(variants, NewEnumVariant(variant.Name, variant.Discriminant, fields))
		}

		return NewEnum(parent.Name, variants)
	default:
		panic(fmt.Sprintf("unhandled case %T", parent))
	}

}
