package typechecker

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/types"
)

func (c *Checker) validateAssignment(v *types.Var, t types.Type, n ast.Expression) error {

	// fmt.Println("[ASSIGNMENT]", v.Name(), "of Type", v.Type(), "to", t)
	// if LHS has not been assigned a value
	f := v.Type()
	if f == unresolved {

		if t == types.LookUp(types.NilLiteral) {
			return fmt.Errorf("use of unspecialized nil in assignment")
		} else if types.IsGeneric(t) {

			if param := types.AsTypeParam(t); param != nil && param.Bound != nil {
				v.SetType(param.Bound)
			} else {
				err := fmt.Errorf("unable to infer specialization of generic type `%s`", t)
				return err
			}

		} else {
			f = t
			v.SetType(t)
		}

		f = v.Type()
	} else {
		updated, err := c.validate(v.Type(), t)
		if err != nil {
			return err
		}

		f = updated
	}

	c.table.SetNodeType(n, f)
	fmt.Printf("\t[NODE ASSIGNMENT] %p -> %s\n", n, f)
	return nil
}

type mappings = map[string]types.Type

func instantiate(t types.Type, args []types.Type, ctx mappings) types.Type {

	if !types.IsGeneric(t) {
		return t
	}

	if ctx == nil {
		fmt.Println("\nNew Instantiation", t, "with", args)
		ctx = make(mappings)
	} else {
		fmt.Println("Instantiating Nested Type", t, "with", args)
	}

	switch t := t.(type) {

	case *types.TypeParam:
		if v, ok := ctx[t.Name()]; ok {

			return v
		}
		return t
	case *types.DefinedType:

		// Instantiate Gens?
		if len(t.TypeParameters) != len(args) {
			panic(fmt.Errorf("expecting %d arguments, got %d", len(t.TypeParameters), len(args)))
		}

		if len(t.TypeParameters) == 0 {
			return t
		}

		// Params
		params := types.TypeParams{}
		for i, p := range t.TypeParameters {

			// check constraints
			arg := args[i]
			fmt.Println("\tMapping", arg, "to", p)

			ctx[p.Name()] = arg

			switch aT := arg.(type) {
			case *types.TypeParam:
				if aT.Bound != nil {
					params = append(params, types.NewTypeParam(aT.Name(), nil, arg))
				} else {
					params = append(params, types.NewTypeParam(aT.Name(), aT.Constraints, nil))
				}
			default:
				params = append(params, types.NewTypeParam(p.Name(), nil, arg))
			}
		}

		// Methods
		methods := make(map[string]*types.Function)
		for _, m := range t.Methods {
			fn := apply(ctx, m.Sg()).(*types.FunctionSignature)
			fmt.Println("\tInstantiated Method:", fn)
			methods[m.Name()] = types.NewFunction(m.Name(), fn)
		}

		switch parent := t.Parent().(type) {
		case *types.Basic:
			// if parent.Literal == types.Unresolved {
			// 	return unresolved
			// }
			panic(fmt.Sprintf("cannot instantiate basic type, \"%s\"", parent))
		case *types.Struct:
			fields := []*types.Var{}
			for _, field := range parent.Fields {
				s := apply(ctx, field.Type())
				fmt.Println("\tInstantiated Field", field.Name(), field.Type(), "to", s)
				spec := types.NewVar(field.Name(), s)
				fields = append(fields, spec)
			}
			copy := types.NewDefinedType(t.Name(), types.NewStruct(fields), params, t.Scope.Parent)
			copy.Methods = methods
			if t.InstanceOf == nil {
				copy.InstanceOf = t
			} else {
				copy.InstanceOf = t.InstanceOf
			}
			return copy
		case *types.Enum:

			variants := types.EnumVariants{}

			for _, variant := range parent.Variants {
				if len(variant.Fields) == 0 {
					variants = append(variants, types.NewEnumVariant(variant.Name, variant.Discriminant, nil))
					continue
				}

				fields := []*types.Var{}

				for _, f := range variant.Fields {
					v := types.NewVar(f.Name(), nil)
					s := apply(ctx, f.Type())
					v.SetType(s)
					fields = append(fields, v)
				}

				variants = append(variants, types.NewEnumVariant(variant.Name, variant.Discriminant, fields))
			}
			e := types.NewEnum(parent.Name, variants)
			copy := types.NewDefinedType(t.Name(), e, params, t.Scope.Parent)
			copy.Methods = methods

			if t.InstanceOf == nil {
				copy.InstanceOf = t
			} else {
				copy.InstanceOf = t.InstanceOf
			}
			return copy

		default:
			panic(fmt.Sprintf("unhandled case %T", parent))

		}
	case *types.FunctionSignature:

		for i, p := range t.Parameters {
			arg := args[i]
			if !types.IsGeneric(p.Type()) {
				fmt.Println("non generic type", p.Type())
				continue
			}

			switch t1 := p.Type().(type) {
			case *types.TypeParam:
				// TODO: Validation
				ctx[t1.Name()] = arg
			default:
				instantiate(t1, args, ctx)
			}
		}
		return apply(ctx, t)
	case *types.Alias:
		// Instantiate Gens?
		if len(t.TypeParameters) != len(args) {
			panic(fmt.Errorf("expecting %d arguments, got %d", len(t.TypeParameters), len(args)))
		}

		if len(t.TypeParameters) == 0 {
			return t
		}

		// Params
		params := types.TypeParams{}
		for i, p := range t.TypeParameters {

			// check constraints
			arg := args[i]
			fmt.Println("\tMapping", arg, "to", p)

			ctx[p.Name()] = arg

			switch aT := arg.(type) {
			case *types.TypeParam:
				if aT.Bound != nil {
					params = append(params, types.NewTypeParam(aT.Name(), nil, arg))
				} else {
					params = append(params, types.NewTypeParam(aT.Name(), aT.Constraints, nil))
				}
			default:
				params = append(params, types.NewTypeParam(p.Name(), nil, arg))
			}
		}

		alias := types.NewAlias(t.Name(), nil)

		rhs := apply(ctx, t.RHS)
		alias.SetType(rhs)
		alias.TypeParameters = params
		return alias
	case *types.Pointer:
		cT := t.PointerTo
		uT := apply(ctx, cT)

		ptr := types.NewPointer(uT)
		return ptr
	default:
		panic(fmt.Sprintf("cannot instantiate type %s", t))
	}
}

func apply(ctx mappings, typ types.Type) types.Type {

	fmt.Printf("\tSubstituting %s with %s\n", typ, ctx)
	if !types.IsGeneric(typ) {
		fmt.Println("\tSkipping Non Generic", typ)
		return typ
	}
	switch t := typ.(type) {

	case *types.TypeParam:
		sp, ok := ctx[t.Name()]
		if !ok {
			return unresolved
		}

		return sp
	case *types.DefinedType:
		params := types.TypeParams{}
		for _, p := range t.TypeParameters {

			arg, ok := ctx[p.Name()]

			if !ok {
				return unresolved
			}

			switch aT := arg.(type) {
			case *types.TypeParam:
				if aT.Bound != nil {
					params = append(params, types.NewTypeParam(aT.Name(), nil, arg))
				} else {
					params = append(params, types.NewTypeParam(aT.Name(), aT.Constraints, nil))
				}
			default:
				params = append(params, types.NewTypeParam(p.Name(), nil, arg))
			}
		}

		// Methods
		methods := make(map[string]*types.Function)
		for _, m := range t.Methods {
			methods[m.Name()] = types.NewFunction(m.Name(), apply(ctx, m.Sg()).(*types.FunctionSignature))
		}
		switch parent := t.Parent().(type) {
		case *types.Basic:
			return typ // basic cannot have instantiation
		case *types.Struct:
			fields := []*types.Var{}
			for _, field := range parent.Fields {
				s := apply(ctx, field.Type())
				spec := types.NewVar(field.Name(), s)
				fields = append(fields, spec)
			}

			copy := types.NewDefinedType(t.Name(), types.NewStruct(fields), params, t.Scope.Parent)
			copy.Methods = methods

			if t.InstanceOf == nil {
				copy.InstanceOf = t
			} else {
				copy.InstanceOf = t.InstanceOf
			}
			return copy
		case *types.Enum:

			variants := types.EnumVariants{}

			for _, variant := range parent.Variants {
				if len(variant.Fields) == 0 {
					variants = append(variants, types.NewEnumVariant(variant.Name, variant.Discriminant, nil))
					continue
				}

				fields := []*types.Var{}

				for _, f := range variant.Fields {
					v := types.NewVar(f.Name(), nil)
					s := apply(ctx, f.Type())
					v.SetType(s)
					fields = append(fields, v)
				}

				variants = append(variants, types.NewEnumVariant(variant.Name, variant.Discriminant, fields))
			}

			e := types.NewEnum(parent.Name, variants)
			copy := types.NewDefinedType(t.Name(), e, params, t.Scope.Parent)
			copy.Methods = methods

			if t.InstanceOf == nil {
				copy.InstanceOf = t
			} else {
				copy.InstanceOf = t.InstanceOf
			}
			return copy

		default:
			panic(fmt.Sprintf("unhandled case %T", parent))

		}
	case *types.FunctionSignature:
		sg := types.NewFunctionSignature()

		// Parameters
		for _, param := range t.Parameters {
			p := types.NewVar(param.Name(), unresolved)
			p.SetType(apply(ctx, param.Type()))
			sg.AddParameter(p)
		}

		// return
		res := t.Result.Type()
		sg.Result.SetType(apply(ctx, res))

		return sg
	case *types.Alias:
		params := types.TypeParams{}
		for _, p := range t.TypeParameters {

			arg, ok := ctx[p.Name()]

			if !ok {
				return unresolved
			}

			switch aT := arg.(type) {
			case *types.TypeParam:
				if aT.Bound != nil {
					params = append(params, types.NewTypeParam(aT.Name(), nil, arg))
				} else {
					params = append(params, types.NewTypeParam(aT.Name(), aT.Constraints, nil))
				}
			default:
				params = append(params, types.NewTypeParam(p.Name(), nil, arg))
			}
		}
		alias := types.NewAlias(t.Name(), nil)

		rhs := apply(ctx, t.RHS)
		alias.SetType(rhs)
		alias.TypeParameters = params
		return alias
	case *types.Pointer:
		cT := t.PointerTo
		uT := apply(ctx, cT)

		ptr := types.NewPointer(uT)
		return ptr
	default:
		panic(fmt.Sprintf("cannot instantiate type %s", t))
	}
}
