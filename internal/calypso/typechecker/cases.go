package typechecker

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/types"
)

func (c *Checker) validateAssignment(v *types.Var, t types.Type, n ast.Expression) error {

	// if LHS has not been assigned a value
	if v.Type() == unresolved {
		if t == types.LookUp(types.NilLiteral) {
			return fmt.Errorf("use of unspecialized nil in assignment")
		}
		v.SetType(t)
	} else {
		_, err := c.validate(v.Type(), t)
		if err != nil {
			return err
		}
	}
	c.table.SetNodeType(n, v.Type())
	fmt.Printf("[Validator] %p -> %s\n", n, v.Type())
	return nil
}
