package types

import (
	"fmt"
)

func Validate(expected Type, provided Type) (Type, error) {
	fmt.Printf("\t[VALIDATOR] Validating `%s`(provided) || `%s`(expected)\n", provided, expected)

	if expected == provided {
		return expected, nil
	}

	if provided == LookUp(Unresolved) {
		// should have already been reported
		return expected, nil
	}

	expected = ResolveAliases(expected)
	provided = ResolveAliases(provided)

	// Instance?
	if expected == LookUp(Placeholder) {
		return provided, nil
	}

	// Non Defined Types
	switch expected := expected.(type) {
	case *Pointer:
		return validatePointerTypes(expected, provided)
	case *FunctionSignature:
		return validateFunctionTypes(expected, provided)
	case *TypeParam:
		return validateTypeParameter(expected, provided)
	case *DefinedType:
		return validateDefinedType(expected, provided)
	case *SpecializedType:
		return validateSpecializedType(expected, provided)
	default:
		panic(fmt.Errorf("unhanled validation case: %T", expected))
	}
}

func validateDefinedType(expected *DefinedType, provided Type) (Type, error) {
	switch provided := provided.(type) {
	case *DefinedType:
		if p, ok := expected.Parent().(*Basic); ok {
			return validateBasicTypes(p, provided.Parent(), expected)
		}

		return nil, fmt.Errorf("expected `%s`, received `%s`", expected, provided)
	default:
		return nil, fmt.Errorf("expected `%s`, received `%s`", expected, provided)
	}

}
func validateBasicTypes(expected *Basic, p Type, e Type) (Type, error) {
	provided, ok := p.Parent().(*Basic)

	if !ok {
		return nil, fmt.Errorf("expected `%s`, received `%s`. Type %T, %T", expected, p, expected, p)
	}

	// either side
	if IsGroupLiteral(provided) {
		switch {
		case provided.Literal == IntegerLiteral && IsNumeric(expected):
			return e, nil
		case provided.Literal == FloatLiteral && IsFloatingPoint(expected):
			return e, nil
		}
	} else if IsGroupLiteral(expected) {
		switch {
		case expected.Literal == IntegerLiteral && IsNumeric(provided):
			return provided, nil
		case expected.Literal == FloatLiteral && IsFloatingPoint(provided):
			return provided, nil
		}
	}

	match := expected == provided
	if !match {
		return nil, fmt.Errorf("expected `%s`, received `%s`", expected, provided)
	}

	return e, nil
}

func validatePointerTypes(expected *Pointer, provided Type) (Type, error) {

	switch provided := provided.(type) {

	case *Pointer:
		_, err := Validate(expected.PointerTo, provided.PointerTo)

		if err != nil {
			return nil, err
		}

		return expected, nil

	default:
		if provided == LookUp(NilLiteral) {
			return expected, nil
		}
	}
	return nil, fmt.Errorf("expected `%s`, received `%s`", expected, provided)

}

func validateFunctionTypes(expected *FunctionSignature, p Type) (Type, error) {
	provided, ok := p.(*FunctionSignature)

	if !ok {
		return nil, fmt.Errorf("expected function signature of %s got %s instead", expected, p)
	}

	if len(expected.Parameters) != len(provided.Parameters) {
		return nil, fmt.Errorf("expected %d parameters, provided %d instead", len(expected.Parameters), len(provided.Parameters))
	}

	for i, eP := range expected.Parameters {
		pP := provided.Parameters[i]

		_, err := Validate(eP.Type(), pP.Type())

		if err != nil {
			return nil, err
		}
	}

	_, err := Validate(expected.Result.Type(), provided.Result.Type())

	if err != nil {
		return nil, err
	}
	return expected, nil
}

func validateTypeParameter(expected *TypeParam, provided Type) (Type, error) {
	err := Conforms(expected.Constraints, provided)
	if err != nil {
		return nil, err
	}
	return expected, nil
}

func validateSpecializedType(expected *SpecializedType, provided Type) (Type, error) {
	switch provided := provided.(type) {
	case *SpecializedType:
		// Is of same instance, check bounds match
		if expected.InstanceOf == provided.InstanceOf {
			for i, eB := range expected.Bounds {
				pB := provided.Bounds[i]

				_, err := Validate(eB, pB)

				if err != nil {
					return nil, err
				}
			}

			// bounds match, return expected
			return expected, nil
		}

	case *DefinedType:
		// is instance of value, should have already validated instance so just pass
		if expected.InstanceOf == provided {
			return expected, nil
		}
	}
	return nil, fmt.Errorf("expected `%s`, received `%s`", expected, provided)
}
func Conforms(constraints []*Standard, x Type) error {
	if provided, ok := x.(*TypeParam); ok {
		seen := make(map[*Standard]bool)
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

	provided := ResolveLiteral(x)

	action := func(s *Standard) error {

		for _, expectedMethod := range s.Signature {
			providedMethod, err := ResolveMethod(provided, expectedMethod.Name())

			if err != nil {
				return err
			}
			if providedMethod == nil {
				return fmt.Errorf("%s does does not conform to standard: `%s`", x, s)
			}

			_, err = Validate(expectedMethod.Type(), providedMethod)

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
