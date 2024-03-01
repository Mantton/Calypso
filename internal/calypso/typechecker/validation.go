package typechecker

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/types"
)

func (c *Checker) validate(expected types.Type, provided types.Type) (types.Type, error) {
	fmt.Printf("Validating `%s`(provided) |> `%s`(expected)\n", provided, expected)

	// Resolve both sides to their underlying types
	expected = expected.Parent()
	provided = provided.Parent()

	_, isGeneric := expected.(*types.TypeParam)

	if isGeneric {
		// Check Constraints
		panic("generic specialization & checking not implemented")
	}

	var standard error = fmt.Errorf("expected `%s`, received `%s`", expected, provided)
	switch expected := expected.(type) {
	case *types.Basic:
		p := provided.(*types.Basic)

		if p == nil {
			return nil, standard
		}

		return c.validateBasicTypes(expected, p)

	case *types.Pointer:
		return c.validatePointerTypes(expected, provided)

	case *types.StructInstance:
		return c.validateInstanceTypes(expected, provided)
	case *types.FunctionSignature:
		return c.validateFunctionTypes(expected, provided)

	}

	return nil, fmt.Errorf("expected `%s`, received `%s`", expected, provided)
}

func (c *Checker) validateBasicTypes(expected *types.Basic, provided *types.Basic) (types.Type, error) {

	if expected == types.LookUp(types.Any) {
		return expected, nil
	}

	// either side
	if types.IsGroupLiteral(provided) {
		switch {
		case provided.Literal == types.IntegerLiteral && types.IsNumeric(expected):
			return expected, nil
		case provided.Literal == types.FloatLiteral && types.IsFloatingPoint(expected):
			return expected, nil
		}
	}

	match := expected == provided
	if !match {
		return nil, fmt.Errorf("expected `%s`, received `%s`", expected, provided)
	}

	return expected, nil
}

func (c *Checker) validatePointerTypes(expected *types.Pointer, provided types.Type) (types.Type, error) {

	switch provided := provided.(type) {

	case *types.Pointer:
		_, err := c.validate(expected.PointerTo, provided.PointerTo)

		if err != nil {
			return nil, err
		}

		return expected, nil

	default:
		if provided == types.LookUp(types.NilLiteral) {
			return expected, nil
		}
	}
	return nil, fmt.Errorf("expected `%s`, received `%s`", expected, provided)

}

func (c *Checker) validateInstanceTypes(expected *types.StructInstance, p types.Type) (types.Type, error) {
	provided, ok := p.(*types.StructInstance)

	if !ok || expected.Type.Parent() != provided.Type.Parent() {
		return nil, fmt.Errorf("expected instance of %s got %s instead", expected, p)
	}

	// both types are instances of the same base type, ensure matching arguments

	for i, eA := range expected.TypeArgs {
		pA := provided.TypeArgs[i]

		_, err := c.validate(eA, pA)

		if err != nil {
			return nil, err
		}

	}

	return expected, nil
}

func (c *Checker) validateFunctionTypes(expected *types.FunctionSignature, p types.Type) (types.Type, error) {
	provided, ok := p.(*types.FunctionSignature)

	if !ok {
		return nil, fmt.Errorf("expected function signature of %s got %s instead", expected, p)
	}

	if len(expected.Parameters) != len(provided.Parameters) {
		return nil, fmt.Errorf("expected %d parameters, provided %d instead", len(expected.Parameters), len(provided.Parameters))
	}

	for i, eP := range expected.Parameters {
		pP := provided.Parameters[i]

		_, err := c.validate(eP.Type(), pP.Type())

		if err != nil {
			return nil, err
		}
	}

	_, err := c.validate(expected.Result.Type(), provided.Result.Type())

	if err != nil {
		return nil, err
	}
	return expected, nil
}

func (c *Checker) validateConformance(constraints []*types.Standard, x types.Type) error {
	if provided, ok := x.(*types.TypeParam); ok {
		seen := make(map[*types.Standard]bool)
		for _, o := range provided.Constraints {
			seen[o] = true
		}

		for _, o := range constraints {
			_, ok := seen[o]

			if !ok {
				return fmt.Errorf("%s does not conform to %s", provided, o.Name)
			}
		}

		return nil
	}

	provided, ok := x.(*types.DefinedType)

	if !ok {
		return fmt.Errorf("%s is not a conforming type", x)
	}
	action := func(s *types.Standard) error {

		for _, expectedMethod := range s.Dna {
			providedMethod, ok := provided.Methods[expectedMethod.Name()]

			if !ok {
				return fmt.Errorf("%s does does not conform to standard `%s`", x, s)
			}

			_, err := c.validate(expectedMethod.Type(), providedMethod.Type())

			if err != nil {
				return err
			}
		}

		return nil
	}

	for _, s := range constraints {
		err := action(s)

		if err != nil {
			return err
		}
	}

	return nil
}
