package t

import (
	"fmt"
	"reflect"

	"github.com/mantton/calypso/internal/calypso/types"
)

func (c *Checker) validate(expected types.Type, provided types.Type) (types.Type, error) {
	fmt.Printf("Validating `%s`(provided) |> `%s`(expected)\n", provided, expected)

	_, isGeneric := expected.(*types.TypeParam)

	if isGeneric {
		// Check Constraints
		panic("generic specialization & checking not implemented")
	}

	match := reflect.TypeOf(expected) == reflect.TypeOf(provided)

	if !match {
		return nil, fmt.Errorf("expected `%s`, received `%s`", expected, provided)
	}

	switch expected := expected.(type) {
	case *types.Basic:
		return c.validateBasicTypes(expected, provided.(*types.Basic))
	}
	return nil, fmt.Errorf("expected `%s`, received `%s`", expected, provided)
}

func (c *Checker) validateBasicTypes(expected *types.Basic, provided *types.Basic) (types.Type, error) {

	if expected == types.LookUp(types.Any) {
		return expected, nil
	}

	match := expected == provided
	if !match {
		return nil, fmt.Errorf("expected `%s`, received `%s`", expected, provided)
	}

	return expected, nil
}
