package typechecker

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/types"
)

func (c *Checker) validateAssignment(v *types.Var, t types.Type, n ast.Expression) error {
	// fmt.Println("[ASSIGNMENT]", v.Name(), "of Type", v.Type(), "to", t)
	// if LHS has not been assigned a value
	f := v.Type()
	if f == unresolved {

		if t == types.LookUp(types.NilLiteral) {
			return fmt.Errorf("use of unspecialized nil in assignment")
		} else if types.IsGeneric(t) {

			if param := types.AsTypeParam(t); param != nil && param.Bound != nil {
				v.SetType(param.Bound)
			} else {
				err := fmt.Errorf("unable to infer specialization of generic type `%s`", t)
				return err
			}

		} else {
			f = t
			v.SetType(t)
		}

		f = v.Type()
	} else {
		updated, err := c.validate(v.Type(), t)
		if err != nil {
			return err
		}

		f = updated
	}

	c.table.SetNodeType(n, f)
	fmt.Printf("\t[NODE ASSIGNMENT] %p -> %s\n", n, f)
	return nil
}
