package typechecker

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/types"
)

func (c *Checker) validateAssignment(v *types.Var, t types.Type, n ast.Expression) error {

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
	fmt.Printf("[Validator] %p -> %s\n", n, f)
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
		if v, ok := ctx[t.Name]; ok {

			return v
		}
		panic("????")
		return t
	case *types.DefinedType:

		// Instantiate Gens?
		if len(t.TypeParameters) != len(args) {
			panic(fmt.Errorf("expecting %d arguments, got %d", len(t.TypeParameters), len(args)))
		}

		if len(t.TypeParameters) == 0 {
			return t
		}

		params := types.TypeParams{}
		for i, p := range t.TypeParameters {

			// check constraints
			arg := args[i]
			fmt.Println("Mapping", arg, "to", p)

			ctx[p.Name] = arg
			switch aT := arg.(type) {
			case *types.TypeParam:
				if aT.Bound != nil {
					params = append(params, types.NewTypeParam(arg.String(), nil, arg))
				} else {
					params = append(params, types.NewTypeParam(aT.Name, aT.Constraints, nil))
				}
			default:
				params = append(params, types.NewTypeParam(arg.String(), nil, arg))
			}
		}

		switch parent := t.Parent().(type) {
		case *types.Basic:
			panic("cannot instantiate basic type")
		case *types.Struct:
			fields := []*types.Var{}
			for _, field := range parent.Fields {
				s := apply(ctx, field.Type())
				fmt.Println("Instantiated Field", field.Name(), field.Type(), "to", s)
				spec := types.NewVar(field.Name(), s)
				fields = append(fields, spec)
			}
			copy := types.NewDefinedType(t.Name(), types.NewStruct(fields), params, t.Scope.Parent)
			return copy

		}
		panic("not done!")
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
				ctx[t1.Name] = arg
			default:
				instantiate(t1, args, ctx)
			}
		}
		return apply(ctx, t)
	default:
		fmt.Println(t)
		panic("cannot instantiate type")
	}
}

func apply(ctx mappings, typ types.Type) types.Type {

	fmt.Printf("Substituting %s with %s\n", typ, ctx)
	if !types.IsGeneric(typ) {
		return typ
	}
	switch t := typ.(type) {

	case *types.TypeParam:
		sp, ok := ctx[t.Name]
		if !ok {
			fmt.Printf("NOT FOUND: %s\n", t)
			panic("generic not specialized")
		}

		return sp
	case *types.DefinedType:
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

			params := types.TypeParams{}
			for _, p := range t.TypeParameters {

				arg, ok := ctx[p.Name]

				if !ok {
					panic("--")
				}

				switch aT := arg.(type) {
				case *types.TypeParam:
					if aT.Bound != nil {
						params = append(params, types.NewTypeParam(arg.String(), nil, arg))
					} else {
						params = append(params, types.NewTypeParam(aT.Name, aT.Constraints, nil))
					}
				default:
					params = append(params, types.NewTypeParam(arg.String(), nil, arg))
				}
			}
			copy := types.NewDefinedType(t.Name(), types.NewStruct(fields), params, t.Scope.Parent)
			return copy
		default:
			panic("not implemented")

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
	default:
		fmt.Println(t)
		panic("cannot instantiate type")
	}
}
