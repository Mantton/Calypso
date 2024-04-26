package typechecker

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/types"
)

func (c *Checker) evaluateTypeExpression(e ast.TypeExpression, tPs []*types.TypeParam, ctx *NodeContext) types.Type {
	switch expr := e.(type) {
	case *ast.IdentifierExpression:
		return c.evaluateIdentifierTypeExpression(expr, tPs, ctx)
	case *ast.PointerTypeExpression:
		return c.evaluatePointerTypeExpression(expr, tPs, ctx)
	case *ast.ArrayTypeExpression:
		return c.evaluateArrayTypeExpression(expr, tPs, ctx)
	case *ast.MapTypeExpression:
		return c.evaluateMapTypeExpression(expr, tPs, ctx)
	case *ast.SpecializationExpression:
		return c.evaluateTypeSpecializationExpression(expr, tPs, ctx)
	case *ast.FieldAccessExpression:
		return c.evaluateTypeFieldAccessExpression(expr, tPs, ctx)
	default:
		msg := fmt.Sprintf("type expression check not implemented, %T", e)
		panic(msg)
	}
}

func (c *Checker) evaluateIdentifierTypeExpression(expr *ast.IdentifierExpression, tPs []*types.TypeParam, ctx *NodeContext) types.Type {

	n := expr.Value

	def, ok := ctx.scope.Resolve(n, c.ParentScope())

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

	return typ
}

func (c *Checker) evaluatePointerTypeExpression(expr *ast.PointerTypeExpression, tPs []*types.TypeParam, ctx *NodeContext) types.Type {

	n := expr.PointerTo

	p := c.evaluateTypeExpression(n, tPs, ctx)

	v := types.NewPointer(p)

	return v
}

func (c *Checker) evaluateTypeParamterStandards(e *ast.GenericParameterExpression, tP *types.TypeParam, ctx *NodeContext) {
	for _, eI := range e.Standards {
		sym, ok := ctx.scope.Resolve(eI.Value, c.ParentScope())

		if !ok {
			c.addError(
				fmt.Sprintf("`%s` cannot be found in context.", eI.Value),
				e.Identifier.Range(),
			)
			return
		}

		s, ok := sym.Type().Parent().(*types.Standard)

		if !ok {
			if !ok {
				c.addError(
					fmt.Sprintf("`%s` is not a standard", eI.Value),
					e.Identifier.Range(),
				)
				return
			}
		}

		tP.AddConstraint(s)

	}
}

func (c *Checker) evaluateArrayTypeExpression(expr *ast.ArrayTypeExpression, tPs []*types.TypeParam, ctx *NodeContext) types.Type {
	element := c.evaluateTypeExpression(expr.Element, tPs, ctx)
	sym, ok := ctx.scope.Resolve("Array", c.ParentScope()) // TODO: This should be different

	if !ok {
		c.addError("unable to find array type", expr.Range())
		return unresolved
	}

	inst, err := c.instantiateWithArguments(sym.Type(), types.TypeList{element}, expr)
	if err != nil {
		c.addError(err.Error(), expr.Range())
		return unresolved
	}

	return inst
}

func (c *Checker) evaluateMapTypeExpression(expr *ast.MapTypeExpression, tPs []*types.TypeParam, ctx *NodeContext) types.Type {

	key := c.evaluateTypeExpression(expr.Key, tPs, ctx)
	value := c.evaluateTypeExpression(expr.Value, tPs, ctx)

	// TODO: This should use the std/map module
	sym, ok := ctx.scope.Resolve("Map", c.ParentScope())

	if !ok {
		c.addError("unable to find map type", expr.Range())
		return unresolved
	}
	inst, err := c.instantiateWithArguments(sym.Type(), types.TypeList{key, value}, expr)

	if err != nil {
		c.addError(err.Error(), expr.Range())
		return unresolved
	}

	return inst
}

func (c *Checker) evaluateTypeSpecializationExpression(expr *ast.SpecializationExpression, tPs []*types.TypeParam, ctx *NodeContext) types.Type {

	typ := c.evaluateTypeExpression(expr.Expression, tPs, ctx)

	if types.IsUnresolved(typ) {
		return unresolved
	}

	var eArgs []types.Type
	for _, n := range expr.Clause.Arguments {
		eArgs = append(eArgs, c.evaluateTypeExpression(n, tPs, ctx))
	}

	inst, err := c.instantiateWithArguments(typ, eArgs, expr.Clause)

	if err != nil {
		c.addError(err.Error(), expr.Range())
		return unresolved
	}

	return inst
}

func (c *Checker) evaluateTypeFieldAccessExpression(expr *ast.FieldAccessExpression, tPs []*types.TypeParam, ctx *NodeContext) types.Type {

	t := c.evaluateTypeExpression(expr.Target, tPs, ctx)

	switch t := t.(type) {

	case *types.Module:
		nCtx := NewContext(t.Scope, ctx.sg, ctx.lhs)
		f := c.evaluateTypeExpression(expr.Field, tPs, nCtx)
		return f

	default:
		panic("unimplemented type field access expression")
	}

}
