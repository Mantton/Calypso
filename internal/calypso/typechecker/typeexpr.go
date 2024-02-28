package typechecker

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/types"
)

func (c *Checker) evaluateTypeExpression(e ast.TypeExpression) types.Type {
	switch expr := e.(type) {
	case *ast.IdentifierTypeExpression:
		return c.evaluateIdentifierTypeExpression(expr)
	case *ast.PointerTypeExpression:
		return c.evaluatePointerTypeExpression(expr)
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

func (c *Checker) evaluatePointerTypeExpression(expr *ast.PointerTypeExpression) types.Type {

	n := expr.PointerTo

	p := c.evaluateTypeExpression(n)

	v := types.NewPointer(p)

	return v
}

func (c *Checker) evaluateFunctionSignature(e *ast.FunctionExpression) *types.FunctionSignature {

	sg := types.NewFunctionSignature()

	// Parameters
	for _, p := range e.Params {
		t := c.evaluateTypeExpression(p.AnnotatedType)
		v := types.NewVar(p.Value, t)
		sg.AddParameter(v)
	}

	// Annotated Return Type
	if e.ReturnType != nil {
		t := c.evaluateTypeExpression(e.ReturnType)
		sg.Result = types.NewVar("", t)
	} else {
		c.addError("missing return value in function signature", e.Range())
		sg.Result = types.NewVar("", unresolved)
	}

	return sg
}
