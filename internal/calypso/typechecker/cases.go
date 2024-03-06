package typechecker

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/types"
)

func (c *Checker) validateAssignment(v *types.Var, t types.Type, n ast.Expression) error {

	// if LHS has not been assigned a value
	f := v.Type()
	if f == unresolved {
		if t == types.LookUp(types.NilLiteral) {
			return fmt.Errorf("use of unspecialized nil in assignment")
		} else if types.IsGeneric(t) {
			err := fmt.Errorf("unable to infer specialization of generic type `%s`", t)
			return err
		}
		v.SetType(t)
	} else {
		updated, err := c.validate(v.Type(), t)
		if err != nil {
			return err
		}

		f = updated
	}
	c.table.SetNodeType(n, f)
	fmt.Printf("[Validator] %p -> %s\n", n, f)
	return nil
}
