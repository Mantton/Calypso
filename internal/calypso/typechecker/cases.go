package typechecker

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/types"
)

func (c *Checker) validateAssignment(variable *types.Var, provided types.Type, node ast.Node, allowGeneric bool) error {
	// fmt.Println("[ASSIGNMENT]", v.Name(), "of Type", v.Type(), "to", t)
	// if LHS has not been assigned a value
	expected := variable.Type()

	if types.IsUnresolved(expected) {
		switch {
		case provided == types.LookUp(types.NilLiteral):
			return fmt.Errorf("use of unspecialized nil in assignment")
		case types.IsGeneric(provided):
			if allowGeneric {
				expected = provided
			} else {
				err := fmt.Errorf("unable to infer specialization of generic type `%s`", provided)
				return err
			}
		default:
			expected = provided
		}
	} else {

		updated, err := c.validate(expected, provided)
		if err != nil {
			return err
		}
		expected = updated
	}

	expected = types.ResolveLiteral(expected)

	if !types.IsAssignable(expected) {
		return fmt.Errorf("%s, cannot assign", expected)
	}

	variable.SetType(expected)
	c.module.Table.SetNodeType(node, expected)
	return nil
}
