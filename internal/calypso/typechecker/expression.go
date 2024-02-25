package t

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/types"
)

func (c *Checker) checkExpression(expr ast.Expression) {

	fmt.Printf(
		"Checking Expression: %T @ Line %d\n",
		expr,
		expr.Range().Start.Line,
	)
	switch expr := expr.(type) {
	case *ast.FunctionExpression:
		c.checkFunctionExpression(expr)
	// case *ast.AssignmentExpression:
	default:
		msg := fmt.Sprintf("expression check not implemented, %T", expr)
		panic(msg)
	}
}

func (c *Checker) evaluateExpression(expr ast.Expression) types.Type {
	switch expr := expr.(type) {
	// Literals
	case *ast.IntegerLiteral:
		return types.LookUp(types.Int)
	case *ast.BooleanLiteral:
		return types.LookUp(types.Bool)
	case *ast.FloatLiteral:
		return types.LookUp(types.Float)
	case *ast.StringLiteral:
		return types.LookUp(types.String)
	case *ast.NullLiteral:
		return types.LookUp(types.Null)
	case *ast.VoidLiteral:
		return types.LookUp(types.Void)
	case *ast.IdentifierExpression:
		return c.evaluateIdentifierExpression(expr)
	case *ast.GroupedExpression:
		return c.evaluateGroupedExpression(expr)
	case *ast.CallExpression:
		return c.evaluateCallExpression(expr)
	// case *ast.ArrayLiteral:
	// case *ast.UnaryExpression:
	// case *ast.BinaryExpression:
	// case *ast.AssignmentExpression:
	// case *ast.CompositeLiteral:
	default:
		msg := fmt.Sprintf("expression evaluation not implemented, %T", expr)
		panic(msg)
	}
}

func (c *Checker) checkFunctionExpression(e *ast.FunctionExpression) {

	sg := types.NewFunctionSignature()
	def := types.NewFunction(e.Identifier.Value, sg)
	e.Signature = def
	ok := c.define(def)

	if !ok {
		c.addError(
			fmt.Sprintf("`%s` is already defined", def.Name()),
			e.Identifier.Range(),
		)
		return
	}

	c.enterScope()
	sg.Scope = c.scope
	defer c.leaveScope()

	// Type/Generic Parameters
	if e.GenericParams != nil {
		for _, p := range e.GenericParams.Parameters {
			d := types.NewTypeDef(p.Identifier.Value, unresolved)
			c.scope.Define(d)
			t := types.NewTypeParam(d, []types.Type{})
			t.Definition.SetType(t)
			sg.AddTypeParameter(t)
		}
	}

	// Parameters
	for _, p := range e.Params {
		t := c.evaluateTypeExpression(p.AnnotatedType)
		v := types.NewVar(p.Value, t)
		c.scope.Define(v)
		sg.AddParameter(v)
	}

	// Annotated Return Type
	if e.ReturnType != nil {
		t := c.evaluateTypeExpression(e.ReturnType)
		sg.ReturnType = t
	}

	// Body
	c.checkBlockStatement(e.Body, def)

	// Ensure All Generic Params are used
	// Ensure All Params are used
}

func (c *Checker) evaluateIdentifierExpression(expr *ast.IdentifierExpression) types.Type {

	s, ok := c.find(expr.Value)

	if !ok {
		c.addError(
			fmt.Sprintf("`%s` is not defined", expr.Value),
			expr.Range(),
		)

		return unresolved
	}

	return s.Type()
}

func (c *Checker) evaluateGroupedExpression(expr *ast.GroupedExpression) types.Type {
	return c.evaluateExpression(expr.Expr)
}

func (c *Checker) evaluateCallExpression(expr *ast.CallExpression) types.Type {
	t := c.evaluateExpression(expr.Target)

	fn, ok := t.(*types.FunctionSignature)

	// Ensure Target is callable
	if !ok {
		c.addError(
			fmt.Sprintf("`%s` is not invocable", expr.Target),
			expr.Target.Range(),
		)
		return unresolved
	}

	// Guard Argument Count == Parameter Count
	if len(expr.Arguments) != len(fn.Parameters) {
		c.addError(
			fmt.Sprintf("expected %d arguments, provided %d",
				len(fn.Parameters),
				len(expr.Arguments)),
			expr.Range(),
		)
		return fn.ReturnType
	}

	// Check Arguments
	// TODO: Generics
	for i, arg := range expr.Arguments {

		fmt.Printf("%T\n", arg)
		provided := c.evaluateExpression(arg)
		expected := fn.Parameters[i].Type()

		// validate will return resolved generic
		_, err := c.validate(expected, provided)

		if err != nil {
			c.addError(
				err.Error(),
				arg.Range(),
			)
			continue
		}
	}

	return fn.ReturnType

}
