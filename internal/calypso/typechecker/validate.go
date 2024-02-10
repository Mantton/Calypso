package typechecker

import (
	"errors"
	"fmt"

	"github.com/mantton/calypso/internal/calypso/symbols"
)

// validates that two types
func (c *Checker) validate(expected, provided *symbols.SymbolInfo, t symbols.SpecializationTable) error {
	fmt.Printf("Validating `%s`(provided) |> `%s`(expected)\n", provided, expected)

	// Provided is unresolved
	if expected == unresolved {
		return fmt.Errorf("`%s` is unresolved", expected)
	}

	if provided == unresolved {
		return fmt.Errorf("`%s` is unresolved", provided)
	}

	// expected is any, always validate
	if expected == c.resolveLiteral(symbols.ANY) {
		return nil
	}

	// Ensure Expected is a Type
	if !c.isType(expected.Type) {
		return fmt.Errorf("`%s` is not a type. this is a typechecker error. report", expected)
	}

	// Ensure Provided is a Type
	if !c.isType(provided.Type) {
		return fmt.Errorf("`%s` is not a type. this is a typechecker error. report", expected)
	}

	// Provided is Expected
	if provided == expected {
		return nil
	}

	hasError := false

	if t == nil {
		t = make(symbols.SpecializationTable)
	}

	// Resolve Specializations, and get map containing the specialized types
	rExpected, err := c.resolveSpecialization(expected, t)

	if err != nil {
		return err
	}

	rProvided, err := c.resolveSpecialization(provided, t)

	if err != nil {
		return err
	}

	// Resolve Any Aliases On Both Sides
	rExpected, rExpectedStandards := c.resolveAlias(rExpected)
	rProvided, rProvidedStandards := c.resolveAlias(rProvided)

	// Iterate till neither side has any generics or
	depth := 0
	for rExpected.AliasOf != nil || rProvided.AliasOf != nil || rExpected.SpecializedOf != nil || rProvided.SpecializedOf != nil {
		// Resolve Specializations, and get map containing the specialized types
		rExpected, err = c.resolveSpecialization(rExpected, t)

		if err != nil {
			return err
		}

		rProvided, err = c.resolveSpecialization(rProvided, t)

		if err != nil {
			return err
		}

		// Resolve Any Aliases On Both Sides
		rExpected, rExpectedStandards = c.resolveAlias(rExpected)
		rProvided, rProvidedStandards = c.resolveAlias(rProvided)

		depth += 1
		if depth > 100 {
			panic("TOO MANY NESTED RESULTS")
		}
	}
	// Ensure Provided conforms to all standards of the expected type
	for key, value := range rExpectedStandards {
		p, ok := rProvidedStandards[key]

		// Does not conform to standard
		if !ok {
			c.addError(
				fmt.Sprintf("`%s` does not conform/implement the `%s` standard.", provided, value),
				c.currentNode.Range(),
			)
			hasError = true
			continue
		}

		// Standards have same key identifier but do not match for some reason.
		if p != value {
			c.addError(
				fmt.Sprintf("`%s` does not match standard of the same identifier. please report this issue", p),
				c.currentNode.Range(),
			)
			hasError = true
			continue
		}
	}

	if hasError {
		return errors.New("REPORTED")
	}

	// // If validating generics & Constraints of the Generic Param T have been met
	// if rExpected.Type == symbols.GenericTypeSymbol && rProvided.Type == symbols.GenericTypeSymbol {
	// 	return nil
	// }

	if rExpected.Type == symbols.GenericTypeSymbol {
		return nil
	}

	// At this point, the expected is not a generic so both resolved types should be the exact same, with only checks of the arguments left
	if rExpected != rProvided {
		return fmt.Errorf("expected `%s`, received `%s`", rExpected, rProvided)
	}
	return nil
}

func (c *Checker) resolveAlias(s *symbols.SymbolInfo) (*symbols.SymbolInfo, map[string]*symbols.SymbolInfo) {
	o := s
	// list of standards to conform to using this alias
	constraints := make(map[string]*symbols.SymbolInfo)

	// Add from Current Object
	for key, value := range o.Constraints {
		constraints[key] = value
	}

	// Loop till the base type, collect all standards this alias conforms to
	// TODO: is this necessary, the highest alias would have all the constraints already
	for o.AliasOf != nil {
		// Add Constraints
		for key, value := range o.Constraints {
			constraints[key] = value
		}
		o = o.AliasOf
	}

	return o, constraints
}

// Resolves Specializations of Generic Types.
func (c *Checker) resolveSpecialization(s *symbols.SymbolInfo, t symbols.SpecializationTable) (*symbols.SymbolInfo, error) {

	// does not have generic
	if s.SpecializedOf == nil {
		return s, nil
	}

	generic := s

	for generic != nil {

		for key, value := range generic.Specializations {
			// fmt.Println("[Resolve]", key, value)
			err := c.specialize(t, key, value)
			if err != nil {
				// fmt.Println("err", err)
				return nil, err
			}
		}

		if generic.SpecializedOf == nil {
			return generic, nil
		} else {
			generic = generic.SpecializedOf
		}

	}
	return generic, nil
}

func (c *Checker) specialize(t symbols.SpecializationTable, k, v *symbols.SymbolInfo) error {
	// fmt.Println("Specializing", k, "as", v, "In Table")

	// ensure is generic
	if k.Type != symbols.GenericTypeSymbol {
		return fmt.Errorf("`%s` is not a generic type", k)
	}

	currentType, ok := t.Get(k)

	// `K` has not been defined, define & exit
	if !ok {
		t[k] = v
		return nil
	}

	// `K` has been defined check

	// Specialization of `K` is a generic
	if currentType.Type == symbols.GenericTypeSymbol {
		err := c.validate(currentType, v, t)

		if err != nil {
			return err
		}

		t[k] = v

		// fmt.Println("Specializing [UPDATING]", currentType, "as", v)
		t[currentType] = v

	} else {
		// Current Value is not a generic, compare

		// New Value is generic, resolve
		if v.Type == symbols.GenericTypeSymbol {
			spec, ok := t.Get(k)

			if !ok {
				panic("unable to resolve generic")
			}

			err := c.validate(currentType, spec, t)
			if err != nil {
				return err
			}

		} else {
			err := c.validate(currentType, v, t)
			if err != nil {
				return err
			}
		}

	}

	return nil
}
