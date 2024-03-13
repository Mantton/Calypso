package typechecker

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/types"
)

func (c *Checker) validate(expected types.Type, provided types.Type) (types.Type, error) {
	fmt.Printf("\t[VALIDATOR] Validating `%s`(provided) || `%s`(expected)\n", provided, expected)

	if provided == unresolved {
		// should have already been reported
		return expected, nil
	}

	expected = types.ResolveAliases(expected)
	provided = types.ResolveAliases(provided)

	// Instance?
	if expected == types.LookUp(types.Placeholder) {
		return provided, nil
	}

	// Non Defined Types
	switch expected := expected.(type) {
	case *types.Pointer:
		return c.validatePointerTypes(expected, provided)
	case *types.FunctionSignature:
		return c.validateFunctionTypes(expected, provided)
	case *types.TypeParam:

		if prov := types.AsTypeParam(provided); prov != nil {
			if expected == prov {
				return expected, nil
			} else if prov.Bound != nil {
				err := c.validateConformance(expected.Constraints, prov.Bound)

				if err != nil {
					return nil, err
				}

				return expected, nil
			} else {
				return nil, fmt.Errorf("params not matching")

			}
		} else {
			err := c.validateConformance(expected.Constraints, provided)

			if err != nil {
				return nil, err
			}

			return expected, nil
		}

	case *types.Basic:
		return c.validateBasicTypes(expected, provided)
	}

	var standard error = fmt.Errorf("expected `%s`, received `%s`", expected, provided)

	defExpected := types.AsDefined(expected)
	defProvided := types.AsDefined(provided)

	if defExpected == nil {
		fmt.Printf("%T\n", expected)
		panic("bad path")
	}

	// resolve basic
	if typ, ok := defExpected.Parent().(*types.Basic); ok {
		return c.validateBasicTypes(typ, provided.Parent())
	}

	if defProvided == nil {
		fmt.Printf("[VALIDATOR] %s is not a DefinedType: %T\n", provided, provided)
		return nil, standard
	}

	// TODO: this needs some serious work...
	if defExpected.InstanceOf == nil && defProvided.InstanceOf == nil {
		if defExpected == defProvided {
			return expected, nil
		}
	} else if defExpected.InstanceOf != nil && defProvided.InstanceOf == nil {
		if defExpected.InstanceOf == defProvided {
			return defExpected, nil
		}
	} else if defExpected.InstanceOf == nil && defProvided.InstanceOf != nil {
		if defProvided.InstanceOf == defExpected {
			return defExpected, nil
		}
	} else {
		// Both Instances are Non nil
		if defExpected.InstanceOf == defProvided.InstanceOf {
			// same instance, rather than compare each field, compare type arguments instead
			// safety check, theoretically not possible
			if len(defExpected.TypeParameters) != len(defProvided.TypeParameters) {
				err := fmt.Errorf("expected %d type arguments got %d instead", len(defExpected.TypeParameters), len(defProvided.TypeParameters))
				return nil, err
			}

			for idx, pEx := range defExpected.TypeParameters {
				pProv := defProvided.TypeParameters[idx]

				_, err := c.validate(pEx.Unwrapped(), pProv.Unwrapped())

				if err != nil {
					return nil, err
				}
			}

			return expected, nil
		}
	}

	fmt.Printf("\t[VALIDATOR] Failed: `%s`(provided) || `%s`(expected)\n", provided, expected)
	return nil, standard
}

func (c *Checker) validateBasicTypes(expected *types.Basic, p types.Type) (types.Type, error) {
	provided, ok := p.Parent().(*types.Basic)

	if !ok {
		return nil, fmt.Errorf("expected `%s`, received `%s`. Type %T, %T", expected, p, expected, p)
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
	} else if types.IsGroupLiteral(expected) {
		switch {
		case expected.Literal == types.IntegerLiteral && types.IsNumeric(provided):
			return provided, nil
		case expected.Literal == types.FloatLiteral && types.IsFloatingPoint(provided):
			return provided, nil
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
				return fmt.Errorf("%s does not conform to standard: %s", provided, o.Name)
			}
		}

		return nil
	}

	provided := types.AsDefined(x)
	if provided == nil {
		panic(
			fmt.Errorf("%s is not a conforming type, %T", x, x),
		)
	}

	if provided == types.LookUp(types.IntegerLiteral) {
		provided = types.AsDefined(types.LookUp(types.Int))
	}

	action := func(s *types.Standard) error {

		for _, expectedMethod := range s.Signature {
			providedMethod := provided.ResolveMethod(expectedMethod.Name())

			if providedMethod == nil {
				return fmt.Errorf("%s does does not conform to standard: `%s`", x, s)
			}

			_, err := c.validate(expectedMethod.Type(), providedMethod)

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
