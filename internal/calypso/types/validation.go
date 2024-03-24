package types

import "fmt"

func Validate(expected Type, provided Type) (Type, error) {
	fmt.Printf("\t[VALIDATOR] Validating `%s`(provided) || `%s`(expected)\n", provided, expected)

	if provided == LookUp(Unresolved) {
		// should have already been reported
		return expected, nil
	}

	expected = ResolveAliases(expected)
	provided = ResolveAliases(provided)

	expected = UnwrapBounded(expected)
	provided = UnwrapBounded(provided)

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

		if prov := AsTypeParam(provided); prov != nil {
			if expected == prov {
				return expected, nil
			} else if prov.Bound != nil {
				err := Conforms(expected.Constraints, prov.Bound)

				if err != nil {
					return nil, err
				}

				return expected, nil
			} else {
				return nil, fmt.Errorf("params not matching")

			}
		} else {
			err := Conforms(expected.Constraints, provided)

			if err != nil {
				return nil, err
			}

			return expected, nil
		}

	case *Basic:
		panic("bad path")
	}

	var standard error = fmt.Errorf("expected `%s`, received `%s`", expected, provided)

	defExpected := AsDefined(expected)
	defProvided := AsDefined(provided)

	if defExpected == nil {
		fmt.Printf("%T\n", expected)
		panic("bad path")
	}

	// resolve basic
	if typ, ok := defExpected.Parent().(*Basic); ok {

		res, err := validateBasicTypes(typ, provided.Parent())

		if err != nil {
			return nil, err
		}

		if typ == res {
			return defExpected, nil
		} else {
			return defProvided, nil
		}
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

				_, err := Validate(pEx.Unwrapped(), pProv.Unwrapped())

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

func validateBasicTypes(expected *Basic, p Type) (Type, error) {
	provided, ok := p.Parent().(*Basic)

	if !ok {
		return nil, fmt.Errorf("expected `%s`, received `%s`. Type %T, %T", expected, p, expected, p)
	}

	if expected == LookUp(Any) {
		return expected, nil
	}

	// either side
	if IsGroupLiteral(provided) {
		switch {
		case provided.Literal == IntegerLiteral && IsNumeric(expected):
			return expected, nil
		case provided.Literal == FloatLiteral && IsFloatingPoint(expected):
			return expected, nil
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

	return expected, nil
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

	provided := AsDefined(x)
	if provided == nil {
		return fmt.Errorf("%s is not a conforming type, %T", x, x)
	}

	if provided == LookUp(IntegerLiteral) {
		provided = AsDefined(LookUp(Int))
	}

	action := func(s *Standard) error {

		for _, expectedMethod := range s.Signature {
			providedMethod := provided.ResolveMethod(expectedMethod.Name())

			if providedMethod == nil {
				return fmt.Errorf("%s does does not conform to standard: `%s`", x, s)
			}

			_, err := Validate(expectedMethod.Type(), providedMethod)

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
