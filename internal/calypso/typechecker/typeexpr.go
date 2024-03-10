package typechecker

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/types"
)

func (c *Checker) evaluateTypeExpression(e ast.TypeExpression, tPs []*types.TypeParam) types.Type {
	switch expr := e.(type) {
	case *ast.IdentifierTypeExpression:
		return c.evaluateIdentifierTypeExpression(expr, tPs)
	case *ast.PointerTypeExpression:
		return c.evaluatePointerTypeExpression(expr, tPs)
	default:
		msg := fmt.Sprintf("type expression check not implemented, %T", e)
		panic(msg)
	}
}

func (c *Checker) evaluateIdentifierTypeExpression(expr *ast.IdentifierTypeExpression, tPs []*types.TypeParam) types.Type {

	n := expr.Identifier.Value

	def, ok := c.find(n)

	var typ types.Type

	if !ok {
		// Find in T Params
		for _, p := range tPs {
			if p.Name() == n {
				typ = p
				break
			}
		}
	} else {
		typ = def.Type()
	}

	if typ == nil {
		msg := fmt.Sprintf("Unable to locate `%s`", n)
		c.addError(msg, expr.Range())
		return unresolved
	}

	var eArgs []types.Type
	var tParams types.TypeParams

	if expr.Arguments != nil {
		for _, n := range expr.Arguments.Arguments {
			eArgs = append(eArgs, c.evaluateTypeExpression(n, tPs))
		}
	}

	if x, ok := typ.(*types.DefinedType); ok {
		tParams = append(tParams, x.TypeParameters...)
	}

	if len(eArgs) != len(tParams) {
		msg := fmt.Sprintf("expected %d type parameter(s), provided %d", len(tParams), len(eArgs))
		c.addError(msg, expr.Range())
		return unresolved
	}

	// not a generic instance
	if len(tParams) == 0 {
		return typ
	}

	hasErrors := false
	// ensure conformance of LHS Arguments into RHS Parameters
	for i, arg := range eArgs {
		p := tParams[i]

		err := c.validateConformance(p.Constraints, arg)

		if err != nil {
			hasErrors = true
			c.addError(err.Error(), expr.Range())
			continue
		}

	}

	if hasErrors {
		return unresolved
	}

	o := instantiate(typ, eArgs, nil)
	fmt.Println("Instantiated:", o, "from", typ)
	fmt.Println()
	return o

}

func (c *Checker) evaluatePointerTypeExpression(expr *ast.PointerTypeExpression, tPs []*types.TypeParam) types.Type {

	n := expr.PointerTo

	p := c.evaluateTypeExpression(n, tPs)

	v := types.NewPointer(p)

	return v
}

func (c *Checker) evaluateFunctionSignature(e *ast.FunctionExpression) *types.FunctionSignature {

	sg := types.NewFunctionSignature()

	// Parameters
	for _, p := range e.Params {
		t := c.evaluateTypeExpression(p.AnnotatedType, nil)
		v := types.NewVar(p.Value, t)
		sg.AddParameter(v)
	}

	// Annotated Return Type
	if e.ReturnType != nil {
		t := c.evaluateTypeExpression(e.ReturnType, nil)
		sg.Result = types.NewVar("", t)
	} else {
		c.addError("missing return value in function signature", e.Range())
		sg.Result = types.NewVar("", unresolved)
	}

	return sg
}

func (c *Checker) evaluateGenericParameterExpression(e *ast.GenericParameterExpression) types.Type {
	d := types.NewTypeParam(e.Identifier.Value, nil, nil)

	for _, eI := range e.Standards {
		sym, ok := c.find(eI.Value)

		if !ok {
			c.addError(
				fmt.Sprintf("`%s` cannot be found in context.", eI.Value),
				e.Identifier.Range(),
			)
			return unresolved
		}

		s, ok := sym.Type().Parent().(*types.Standard)

		if !ok {
			if !ok {
				c.addError(
					fmt.Sprintf("`%s` is not a standard", eI.Value),
					e.Identifier.Range(),
				)
				return unresolved
			}
		}

		d.AddConstraint(s)

	}
	return d
}
