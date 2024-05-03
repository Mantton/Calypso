package typechecker

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/types"
)

func (c *Checker) instantiateWithArguments(t types.Type, args types.TypeList, expr ast.Node) (types.Type, error) {

	if !types.IsGeneric(t) {
		return nil, fmt.Errorf("%s is not a generic type", t)
	}

	spec := make(types.Specialization)
	tparams := types.GetTypeParams(t)

	x := len(tparams)
	y := len(args)
	if x != y {
		return nil, fmt.Errorf("expected %d arguments, got %d", x, y)
	}

	for i, p := range tparams {
		arg := args[i]
		_, err := c.validate(p, arg)

		if err != nil {
			c.addError(err.Error(), expr.Range())
		}

		spec[p] = arg
	}

	instantiation := types.Instantiate(t, spec, c.module)
	if types.IsUnresolved(instantiation) {
		return nil, fmt.Errorf("failed to instantiate type")
	}
	// fmt.Printf("Instantiated %s, from %s\n", instantiation, t)
	return instantiation, nil
}

func (c *Checker) instantiateWithSpecialization(t types.Type, s types.Specialization) (types.Type, error) {

	if !types.IsGeneric(t) {
		return nil, fmt.Errorf("%s is not a generic type", t)

	}
	instantiation := types.Instantiate(t, s, c.module)
	if types.IsUnresolved(instantiation) {
		return nil, fmt.Errorf("failed to instantiate type")
	}
	// fmt.Printf("Instantiated %s, from %s\n", instantiation, t)
	return instantiation, nil
}
