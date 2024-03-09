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

	// Non Defined Types
	switch expected := expected.(type) {
	case *types.Pointer:
		return c.validatePointerTypes(expected, provided)
	case *types.FunctionSignature:
		return c.validateFunctionTypes(expected, provided)
	case *types.TypeParam:
		if expected == provided {
			return expected, nil
		}
	}
	var standard error = fmt.Errorf("expected `%s`, received `%s`", expected, provided)

	defExpected := types.AsDefined(expected)
	defProvided := types.AsDefined(provided)

	if defExpected == nil {
		panic("bad path")
	}

	if defProvided == nil {
		// resolve basic
		if typ, ok := defExpected.Parent().(*types.Basic); ok {
			return c.validateBasicTypes(typ, provided.Parent())
		}

		fmt.Println("[Validation] not a defined type", defExpected, defProvided, "Actual", expected, provided)
		return nil, standard
	}

	// resolve basic
	if typ, ok := defExpected.Parent().(*types.Basic); ok {
		return c.validateBasicTypes(typ, provided.Parent())
	}

	if defExpected.InstanceOf == provided {
		print("parent, child")
	} else if defExpected.InstanceOf == defProvided.InstanceOf {
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

	fmt.Println(defExpected, defProvided)

	return nil, standard
}

func (c *Checker) validateBasicTypes(expected *types.Basic, p types.Type) (types.Type, error) {
	provided, ok := p.(*types.Basic)

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

	provided, ok := x.(*types.DefinedType)
	if !ok {
		return fmt.Errorf("%s is not a conforming type, %T", x, x)
	}

	if provided == types.GlobalScope.MustResolve("literal int").Type() {
		provided = types.AsDefined(types.GlobalScope.MustResolve("int").Type())
	}

	action := func(s *types.Standard) error {

		for _, expectedMethod := range s.Dna {
			providedMethod, ok := provided.Methods[expectedMethod.Name()]

			if !ok {
				return fmt.Errorf("%s does does not conform to standard: `%s`", x, s)
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
