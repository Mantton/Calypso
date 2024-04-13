package types

import "fmt"

type mappings = map[string]Type

func Instantiate(t Type, args []Type, ctx mappings) Type {

	if !IsGeneric(t) {
		return t
	}

	if ctx == nil {
		fmt.Println("\nNew Instantiation", t, "with", args)
		ctx = make(mappings)
	} else {
		fmt.Println("Instantiating Nested Type", t, "with", args)
	}

	switch t := t.(type) {

	case *TypeParam:
		if v, ok := ctx[t.Name()]; ok {

			return v
		}
		return t
	case *DefinedType:

		// Instantiate Gens?
		if len(t.TypeParameters) != len(args) {
			panic(fmt.Errorf("expecting %d arguments, got %d", len(t.TypeParameters), len(args)))
		}

		if len(t.TypeParameters) == 0 {
			return t
		}

		// Params
		for i, p := range t.TypeParameters {
			// check constraints
			arg := args[i]
			fmt.Println("\tMapping", arg, "to", p)
			ctx[p.Name()] = arg
		}

		return Apply(ctx, t)
	case *FunctionSignature:

		for i, p := range t.Parameters {
			arg := args[i]
			if !IsGeneric(p.Type()) {
				fmt.Println("non generic type", p.Type())
				continue
			}

			switch t1 := p.Type().(type) {
			case *TypeParam:
				// TODO: Validation
				ctx[t1.Name()] = arg
			default:
				Instantiate(t1, args, ctx)
			}
		}
		return Apply(ctx, t)
	case *Alias:
		// Instantiate Gens?
		if len(t.TypeParameters) != len(args) {
			panic(fmt.Errorf("expecting %d arguments, got %d", len(t.TypeParameters), len(args)))
		}

		if len(t.TypeParameters) == 0 {
			return t
		}

		// Params
		for i, p := range t.TypeParameters {

			// check constraints
			arg := args[i]
			fmt.Println("\tMapping", arg, "to", p)

			ctx[p.Name()] = arg
		}

		return Apply(ctx, t)
	case *Pointer:
		cT := t.PointerTo
		uT := Apply(ctx, cT)
		ptr := NewPointer(uT)
		return ptr
	default:
		panic(fmt.Sprintf("cannot instantiate type %s", t))
	}
}

func Apply(ctx mappings, typ Type) Type {

	fmt.Printf("\tSubstituting %s with %s\n", typ, ctx)
	if !IsGeneric(typ) {
		fmt.Println("\tSkipping Non Generic", typ)
		return typ
	}
	switch t := typ.(type) {

	case *TypeParam:
		sp, ok := ctx[t.Name()]
		if !ok {
			return LookUp(Unresolved)
		}

		return sp
	case *DefinedType:
		params := TypeParams{}
		for _, p := range t.TypeParameters {

			arg, ok := ctx[p.Name()]

			if !ok {
				return LookUp(Unresolved)
			}

			switch aT := arg.(type) {
			case *TypeParam:
				if aT.Bound != nil {
					params = append(params, NewTypeParam(aT.Name(), nil, arg))
				} else {
					params = append(params, NewTypeParam(aT.Name(), aT.Constraints, nil))
				}
			default:
				params = append(params, NewTypeParam(p.Name(), nil, arg))
			}
		}

		var internal Type

		switch parent := t.Parent().(type) {
		case *Basic:
			return typ
		case *Struct:
			fields := []*Var{}
			for _, field := range parent.Fields {
				s := Apply(ctx, field.Type())
				spec := NewVar(field.Name(), s)
				fields = append(fields, spec)
			}

			internal = NewStruct(fields)

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
					s := Apply(ctx, f.Type())
					v.SetType(s)
					fields = append(fields, v)
				}

				variants = append(variants, NewEnumVariant(variant.Name, variant.Discriminant, fields))
			}
			internal = NewEnum(parent.Name, variants)

		default:
			panic(fmt.Sprintf("unhandled case %T", parent))

		}

		copy := NewDefinedType(t.Name(), internal, params, nil) // nil scope, potential problem?

		// Is Parent Instance
		var parentInstance *DefinedType
		if t.InstanceOf == nil {
			parentInstance = t
		} else {
			// T is not parent, link to parent of T
			parentInstance = t.InstanceOf
		}

		copy.InstanceOf = parentInstance
		return copy
	case *FunctionSignature:
		sg := NewFunctionSignature()

		// Parameters
		for _, param := range t.Parameters {
			p := NewVar(param.Name(), LookUp(Unresolved))
			p.SetType(Apply(ctx, param.Type()))
			sg.AddParameter(p)
		}

		// return
		res := t.Result.Type()
		sg.Result.SetType(Apply(ctx, res))

		return sg
	case *Alias:
		params := TypeParams{}
		for _, p := range t.TypeParameters {

			arg, ok := ctx[p.Name()]

			if !ok {
				return LookUp(Unresolved)
			}

			switch aT := arg.(type) {
			case *TypeParam:
				if aT.Bound != nil {
					params = append(params, NewTypeParam(aT.Name(), nil, arg))
				} else {
					params = append(params, NewTypeParam(aT.Name(), aT.Constraints, nil))
				}
			default:
				params = append(params, NewTypeParam(p.Name(), nil, arg))
			}
		}
		alias := NewAlias(t.Name(), nil)

		rhs := Apply(ctx, t.RHS)
		alias.SetType(rhs)
		alias.TypeParameters = params
		return alias
	case *Pointer:
		cT := t.PointerTo
		uT := Apply(ctx, cT)
		ptr := NewPointer(uT)
		return ptr
	default:
		panic(fmt.Sprintf("cannot instantiate type %s", t))
	}
}
