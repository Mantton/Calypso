package typechecker

import (
	"errors"
	"fmt"
)

type SpecializationTable map[*SymbolInfo]*SymbolInfo

// validates that two types
func (c *Checker) validate(expected, provided *SymbolInfo) error {
	fmt.Printf("\nValidating `%s`(provided) |> `%s`(expected)\n", provided.Name, expected.Name)
	// Provided is unresolved
	if expected == unresolved {
		return fmt.Errorf("`%s` is unresolved", expected.Name)
	}

	if provided == unresolved {
		return fmt.Errorf("`%s` is unresolved", provided.Name)
	}

	// expected is any, always validate
	if expected == c.resolveLiteral(ANY) {
		return nil
	}

	// Ensure Expected is a Type
	if !c.isType(expected.Type) {
		return fmt.Errorf("`%s` is not a type. this is a typechecker error. report", expected.Name)
	}

	// Ensure Provided is a Type
	if !c.isType(provided.Type) {
		return fmt.Errorf("`%s` is not a type. this is a typechecker error. report", expected.Name)
	}

	// Provided is Expected
	if provided == expected {
		return nil
	}

	// TODO: This should account for packages/modules
	hasError := false

	specializations := make(SpecializationTable)
	// Resolve Specializations, and get map containing the specialized types
	rExpected, err := c.resolveSpecialization(expected, specializations)

	if err != nil {
		return err
	}

	rProvided, err := c.resolveSpecialization(provided, specializations)

	if err != nil {
		return err
	}

	// Resolve Any Aliases On Both Sides
	rExpected, rExpectedStandards := c.resolveAlias(rExpected)
	rProvided, rProvidedStandards := c.resolveAlias(rProvided)

	// Iterate till neither side has any generics or
	depth := 0
	for rExpected.AliasOf != nil || rProvided.AliasOf != nil || rExpected.ConcreteOf != nil || rProvided.ConcreteOf != nil {
		// Resolve Specializations, and get map containing the specialized types
		rExpected, err = c.resolveSpecialization(rExpected, specializations)

		if err != nil {
			return err
		}

		rProvided, err = c.resolveSpecialization(rProvided, specializations)

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
				fmt.Sprintf("`%s` does not conform/implement the `%s` standard.", provided.Name, value.Name),
				c.currentNode.Range(),
			)
			hasError = true
			continue
		}

		// Standards have same key identifier but do not match for some reason.
		if p != value {
			c.addError(
				fmt.Sprintf("`%s` does not match standard of the same identifier. please report this issue", p.Name),
				c.currentNode.Range(),
			)
			hasError = true
			continue
		}
	}

	if hasError {
		return errors.New("REPORTED")
	}

	// If validating generics & Constraints of the Generic Param T have been met
	if rExpected.Type == GenericTypeSymbol {
		return nil
	}

	// At this point, the expected is not a generic so both resolved types should be the exact same, with only checks of the arguments left
	if rExpected != rProvided {
		return fmt.Errorf("cannot assign `%s` to `%s`", rProvided.Name, rExpected.Name)
	}

	// Both Types are the same, check arguments
	// TODO: do we possibly need to compare the length of both arg arrays?
	// for i, arg := range rExpected.GenericArguments {
	// 	expectedArg, ok := specializations.get(arg)

	// 	if !ok {
	// 		panic("UNABLE TO RESOLVE GENERIC SPECIALIZATION")
	// 	}

	// 	providedArg, ok := specializations.get(rProvided.GenericArguments[i])

	// 	if !ok {
	// 		panic("UNABLE TO RESOLVE GENERIC SPECIALIZATION")
	// 	}

	// 	ok = c.validate(expectedArg, providedArg)

	// 	if !ok {
	// 		c.addError(
	// 			fmt.Sprintf("Cannot assign `%s` to `%s`", expectedArg.Name, providedArg.Name),
	// 			c.currentNode.Range(),
	// 		)

	// 		hasError = true

	// 	}
	// }

	return nil
}

func (c *Checker) resolveAlias(s *SymbolInfo) (*SymbolInfo, map[string]*SymbolInfo) {
	o := s
	// list of standards to conform to using this alias
	constraints := make(map[string]*SymbolInfo)

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
func (c *Checker) resolveSpecialization(s *SymbolInfo, t SpecializationTable) (*SymbolInfo, error) {

	generic := s.ConcreteOf

	// does not have generic
	if generic == nil {
		return s, nil
	}

	if generic.ConcreteOf != nil {
		panic("Cannot have a concrete of a concrete")
	}

	for i, arg := range s.GenericArguments {
		a := generic.GenericArguments[i]
		if a.Type == GenericTypeSymbol {
			err := c.add(t, a, arg)

			if err != nil {
				return nil, err
			}
		}

	}

	return generic, nil
}

func (c *Checker) add(t SpecializationTable, k, v *SymbolInfo) error {

	// If generic, find all where key
	if v.Type == GenericTypeSymbol {

		oldVal, ok := t[k]
		if !ok {
			t[k] = t[v]
			return nil
		}

		newVal, ok := t[v]
		if !ok {
			panic("Trying to get generic type that has not been specialized")
		}

		err := c.validate(newVal, oldVal)

		if err != nil {
			t[k] = unresolved
			return err
		}

		t[k] = t[v]
	} else {

		oldVal, ok := t[k]

		// Not Stored, Store
		if !ok {
			t[k] = v
			return nil
		}

		// Stored, Validate Change
		err := c.validate(v, oldVal)

		if err != nil {
			t[k] = unresolved
			return err
		}

		t[k] = v

	}

	// fmt.Println("\n DICT")
	// for key, value := range t {
	// 	fmt.Println(" >>>>>>", key.Name, key.ID, "Maps To", value.Name, value.ID, "Args")
	// }
	return nil
}

func (t SpecializationTable) get(s *SymbolInfo) (*SymbolInfo, bool) {

	if s.Type != GenericTypeSymbol {
		return s, true
	}
	v, ok := t[s]
	return v, ok
}

func (c *Checker) debugPrintArguments(s *SymbolInfo) {
	for _, arg := range s.GenericArguments {
		fmt.Println("[DEBUG]", arg.Name, "For", s.Name)
	}
}
