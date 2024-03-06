package typechecker

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/types"
)

func (c *Checker) validate(expected types.Type, provided types.Type) (types.Type, error) {
	fmt.Printf("Validating `%s`(provided) |> `%s`(expected)\n", provided, expected)

	if provided == unresolved {
		// should have already been reported
		return expected, nil
	}

	gE, isGeneric := expected.(*types.TypeParam)

	if isGeneric {
		// Check Constraints
		err := c.validateConformance(gE.Constraints, provided)

		if err != nil {
			return nil, err
		}

		return provided, nil
	}
	var standard error = fmt.Errorf("expected `%s`, received `%s`", expected, provided)

	// Resolve both sides to their underlying types
	expected = expected.Parent()
	provided = provided.Parent()

	switch expected := expected.(type) {
	case *types.Basic:
		return c.validateBasicTypes(expected, provided)

	case *types.Pointer:
		return c.validatePointerTypes(expected, provided)

	case *types.StructInstance:
		return c.validateStructInstanceTypes(expected, provided)
	case *types.FunctionSignature:
		return c.validateFunctionTypes(expected, provided)
	case *types.EnumInstance:
		return c.validateEnumInstanceType(expected, provided)

	}

	return nil, standard
}

func (c *Checker) validateBasicTypes(expected *types.Basic, p types.Type) (types.Type, error) {
	provided, ok := p.(*types.Basic)

	if !ok {
		return nil, fmt.Errorf("expected `%s`, received `%s`", expected, p)
	}

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

func (c *Checker) validateStructInstanceTypes(expected *types.StructInstance, p types.Type) (types.Type, error) {
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
		return fmt.Errorf("%s is not a conforming type, %T", x, x)
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

func (c *Checker) validateEnumInstanceType(expected *types.EnumInstance, p types.Type) (types.Type, error) {
	var standard error = fmt.Errorf("expected `%s`, received `%s`", expected, p)

	switch provided := p.(type) {
	case *types.EnumInstance:
		// ensure types are the same
		if expected.Type.Parent() != provided.Type.Parent() {
			return nil, standard
		}

		// ensure arguments count match
		if len(expected.TypeArgs) != len(provided.TypeArgs) {
			return nil, fmt.Errorf("expected %d arguments got %d", len(expected.TypeArgs), len(provided.TypeArgs))
		}

		// ensure arguments match

		for idx, eArg := range expected.TypeArgs {
			pArg := provided.TypeArgs[idx]

			_, err := c.validate(eArg, pArg)

			if err != nil {
				return nil, err
			}

		}

	case *types.Enum:
		def, ok := expected.Type.(*types.DefinedType)
		if !ok {
			return nil, standard
		}

		e, ok := def.Parent().(*types.Enum)

		if !ok || e != p {
			return nil, standard
		}

		return expected, nil

	}

	return expected, nil
}
