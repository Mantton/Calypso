package typechecker

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/types"
)

func (c *Checker) evaluateTypeExpression(e ast.TypeExpression, tPs []*types.TypeParam, ctx *NodeContext) types.Type {
	switch expr := e.(type) {
	case *ast.IdentifierTypeExpression:
		return c.evaluateIdentifierTypeExpression(expr, tPs, ctx)
	case *ast.PointerTypeExpression:
		return c.evaluatePointerTypeExpression(expr, tPs, ctx)
	case *ast.ArrayTypeExpression:
		return c.evaluateArrayTypeExpression(expr, tPs, ctx)
	case *ast.MapTypeExpression:
		return c.evaluateMapTypeExpression(expr, tPs, ctx)
	default:
		msg := fmt.Sprintf("type expression check not implemented, %T", e)
		panic(msg)
	}
}

func (c *Checker) evaluateIdentifierTypeExpression(expr *ast.IdentifierTypeExpression, tPs []*types.TypeParam, ctx *NodeContext) types.Type {

	n := expr.Identifier.Value

	def, ok := ctx.scope.Resolve(n)

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
			eArgs = append(eArgs, c.evaluateTypeExpression(n, tPs, ctx))
		}
	}

	tParams = types.GetTypeParams(typ)

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

		err := types.Conforms(p.Constraints, arg)

		if err != nil {
			hasErrors = true
			c.addError(err.Error(), expr.Range())
			continue
		}

	}

	if hasErrors {
		return unresolved
	}

	o := types.Instantiate(typ, eArgs, nil)
	fmt.Println("Instantiated:", o, "from", typ)
	fmt.Println()
	return o

}

func (c *Checker) evaluatePointerTypeExpression(expr *ast.PointerTypeExpression, tPs []*types.TypeParam, ctx *NodeContext) types.Type {

	n := expr.PointerTo

	p := c.evaluateTypeExpression(n, tPs, ctx)

	v := types.NewPointer(p)

	return v
}

func (c *Checker) evaluateFunctionSignature(e *ast.FunctionExpression, ctx *NodeContext) *types.FunctionSignature {

	sg := types.NewFunctionSignature()

	if (e.GenericParams) != nil {
		panic("TODO")
	}

	// Parameters
	for _, p := range e.Parameters {
		t := c.evaluateTypeExpression(p.Type, nil, ctx)
		v := types.NewVar(p.Name.Value, t)
		sg.AddParameter(v)
	}

	// Annotated Return Type
	if e.ReturnType != nil {
		t := c.evaluateTypeExpression(e.ReturnType, nil, ctx)
		sg.Result = types.NewVar("", t)
	} else {
		c.addError("missing return value in function signature", e.Range())
		sg.Result = types.NewVar("", unresolved)
	}

	return sg
}

func (c *Checker) evaluateGenericParameterExpression(e *ast.GenericParameterExpression, ctx *NodeContext) types.Type {
	d := types.NewTypeParam(e.Identifier.Value, nil, nil)

	for _, eI := range e.Standards {
		sym, ok := ctx.scope.Resolve(eI.Value)

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

func (c *Checker) evaluateArrayTypeExpression(expr *ast.ArrayTypeExpression, tPs []*types.TypeParam, ctx *NodeContext) types.Type {
	element := c.evaluateTypeExpression(expr.Element, tPs, ctx)
	sym, ok := ctx.scope.Resolve("Array") // TODO: This should be different

	if !ok {
		c.addError("unable to find array type", expr.Range())
		return unresolved
	}
	typ := types.AsDefined(sym.Type())
	inst := types.Instantiate(typ, []types.Type{element}, nil)
	return inst
}

func (c *Checker) evaluateMapTypeExpression(expr *ast.MapTypeExpression, tPs []*types.TypeParam, ctx *NodeContext) types.Type {

	key := c.evaluateTypeExpression(expr.Key, tPs, ctx)
	value := c.evaluateTypeExpression(expr.Value, tPs, ctx)

	sym, ok := ctx.scope.Resolve("Map")

	if !ok {
		c.addError("unable to find map type", expr.Range())
		return unresolved
	}
	typ := types.AsDefined(sym.Type())
	inst := types.Instantiate(typ, []types.Type{key, value}, nil)

	return inst
}
