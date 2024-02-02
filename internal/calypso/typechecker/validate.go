package typechecker

import (
	"fmt"
)

// validates that two types
func (c *Checker) validate(expected, provided *SymbolInfo) bool {
	fmt.Printf("\nValidating `%s`(provided) |> `%s`(expected)\n", provided.Name, expected.Name)
	// Provided is unresolved
	if expected == unresolved {
		c.addError(
			fmt.Sprintf("`%s` is unresolved", expected.Name),
			c.currentNode.Range(),
		)
		return false
	}

	if provided == unresolved {
		c.addError(
			fmt.Sprintf("`%s` is unresolved", provided.Name),
			c.currentNode.Range(),
		)
		return false
	}

	// expected is any, always validate
	if expected == c.resolveLiteral(ANY) {
		return true
	}

	// Ensure Expected is a Type
	if !c.isType(expected.Type) {
		c.addError(
			fmt.Sprintf("`%s` is not a type. this is a typechecker error. report", expected.Name),
			c.currentNode.Range(),
		)
		return false
	}

	// Ensure Provided is a Type
	if !c.isType(provided.Type) {
		c.addError(
			fmt.Sprintf("`%s` is not a type. this is a typechecker error. report", expected.Name),
			c.currentNode.Range(),
		)
		return false
	}

	// Provided is Expected
	if provided == expected {
		return true
	}

	// TODO: Check Generics

	// Direct name match,
	// TODO: This should account for packages/modules
	if expected.Name == provided.Name {
		return true
	}

	// Resolve Aliases of Both Sides
	rExpected, rExpectedStandards := c.resolveAlias(expected)
	rProvided, rProvidedStandards := c.resolveAlias(provided)

	// Provided must conform to all standards of the expected type
	hasError := false
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

	// If validating generics
	if rExpected.Type == GenericTypeSymbol && rProvided.Type == GenericTypeSymbol {
		return true
	}
	// Validated if both types have the same parent
	if rExpected != rProvided {
		c.addError(
			fmt.Sprintf("resolved `%s` does not match resolved `%s`. please report this issue", rExpected.Name, rProvided.Name),
			c.currentNode.Range(),
		)
		hasError = true
	}

	return !hasError

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

func (c *Checker) satisfiesConstraint(target *SymbolInfo, constraint *SymbolInfo) bool {
	return false
}
