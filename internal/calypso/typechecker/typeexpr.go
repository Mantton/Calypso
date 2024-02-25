package t

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/types"
)

func (c *Checker) evaluateTypeExpression(e ast.TypeExpression) types.Type {
	switch expr := e.(type) {
	case *ast.IdentifierTypeExpression:
		return c.evaluateIdentifierTypeExpression(expr)
	default:
		msg := fmt.Sprintf("type expression check not implemented, %T", e)
		panic(msg)
	}
}

func (c *Checker) evaluateIdentifierTypeExpression(expr *ast.IdentifierTypeExpression) types.Type {

	n := expr.Identifier.Value

	def, ok := c.find(n)

	if !ok {
		msg := fmt.Sprintf("Unable to locate `%s`", n)
		c.addError(msg, expr.Range())
		return unresolved
	}

	return def.Type()
}
