package typechecker

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/types"
)

func (c *Checker) registerFunctionExpression(e *ast.FunctionExpression, scope *types.Scope) *types.Function {
	// Create new function

	sg := types.NewFunctionSignature()
	def := types.NewFunction(e.Identifier.Value, sg, c.module)
	def.SetVisibility(e.Visibility == ast.PUBLIC)
	// Enter Function Scope
	def.Scope = types.NewScope(scope, e.Identifier.Value)

	// Type/Generic Parameters
	if e.GenericParams != nil {
		for _, p := range e.GenericParams.Parameters {
			d := types.NewTypeParam(p.Identifier.Value, nil)
			err := sg.AddTypeParameter(d)

			if err != nil {
				c.addError(err.Error(), p.Identifier.Range())
			}
		}
	}

	// Parameters
	for _, p := range e.Parameters {

		// Placeholder / Discard

		v := types.NewVar(p.Name.Value, unresolved, c.module)

		// Parameter Has Required Label
		if p.Label.Value != "_" {
			v.ParamLabel = p.Label.Value
		}

		sg.AddParameter(v)

		if p.Name.Value == "_" {
			continue
		}
		err := def.Scope.Define(v)

		if err != nil {
			c.addError(err.Error(), p.Range())
		}
	}

	// Annotated Return Type
	if e.ReturnType != nil {
		sg.Result = types.NewVar("result", unresolved, c.module)
	} else {
		sg.Result = types.NewVar("result", types.LookUp(types.Void), c.module)
	}

	// At this point the signature has been constructed fully, add to scope
	err := scope.Define(def)

	if err != nil {
		c.addError(err.Error(), e.Identifier.Range())
	}
	c.module.Table.SetNodeType(e, def.Sg())
	c.module.Table.SetSymbol(def, e)

	return def
}

func (c *Checker) registerConformance(d *ast.ConformanceDeclaration) {
	// Check for Standard
	sSymbol := c.ParentScope().MustResolve(d.Standard.Value)
	if sSymbol == nil {
		c.addError(fmt.Sprintf("%s cannot be found in the current context", d.Target.Value), d.Standard.Range())
		return
	}

	standard := types.AsStandard(sSymbol.Type().Parent())

	if standard == nil {
		c.addError(fmt.Sprintf("%s is not a standard", d.Standard.Value), d.Standard.Range())
		return
	}

	// Check for Type
	tSymbol := c.ParentScope().MustResolve(d.Target.Value)
	if tSymbol == nil {
		c.addError(fmt.Sprintf("%s cannot be found in the current context", d.Target.Value), d.Target.Range())
		return
	}

	tDefinition := types.AsDefined(tSymbol.Type())

	if tDefinition == nil {
		c.addError(fmt.Sprintf("%s is not a defined type", d.Target.Value), d.Target.Range())
		return
	}

	// Types
	ctx := NewContext(tDefinition.GetScope(), nil, nil)
	for _, t := range d.Types {
		c.defineAlias(t, ctx)
	}

	// Functions
	for _, t := range d.Signatures {
		c.registerFunctionExpression(t.Func, ctx.scope)
	}
}

func (c *Checker) define(n *ast.IdentifierExpression, core ast.Node, parent *types.Scope) *types.DefinedType {
	scope := types.NewScope(parent, n.Value)
	def := types.NewBaseDefinedType(n.Value, unresolved, nil, scope, c.module)
	err := parent.Define(def)

	if err != nil {
		c.addError(
			fmt.Sprintf(err.Error(), def.Name()),
			n.Range(),
		)
		return nil
	}

	c.module.Table.SetNodeType(core, def)
	return def
}

func (c *Checker) defineAlias(n *ast.TypeStatement, ctx *NodeContext) *types.Alias {
	// 1 - Define
	name := n.Identifier.Value
	alias := types.NewAlias(name, types.LookUp(types.Placeholder))
	err := ctx.scope.Define(alias)
	if err != nil {
		c.addError(
			fmt.Sprintf("`%s` is already defined in context", name),
			n.Identifier.Range(),
		)
		return nil
	}

	return alias
}

func (c *Checker) resolve(n *ast.IdentifierExpression, core ast.Node, scope *types.Scope) *types.DefinedType {

	def := scope.MustResolve(n.Value)

	if def != nil {
		return types.AsDefined(def.Type())
	}

	return c.define(n, core, scope)
}

func (c *Checker) registerTypeParameters(g *ast.GenericParametersClause, t *types.DefinedType) {
	if g == nil || t == nil {
		return
	}

	for _, p := range g.Parameters {
		d := types.NewTypeParam(p.Identifier.Value, nil)
		err := t.AddTypeParameter(d)

		if err != nil {
			c.addError(err.Error(), p.Identifier.Range())
		}
	}
}

func (c *Checker) registerFunctionSignatures(e *ast.FunctionExpression) *types.FunctionSignature {

	sg := c.module.Table.GetNodeType(e).(*types.FunctionSignature)

	if sg == nil {
		panic("unregistered node")
	}

	fn := sg.Function

	ctx := NewContext(fn.Scope, sg, nil)
	//  Generic Params
	if e.GenericParams != nil {
		for i, p := range e.GenericParams.Parameters {
			tP := sg.TypeParameters[i]
			c.evaluateTypeParamterStandards(p, tP, ctx)
		}
	}

	// Parameters
	for i, p := range e.Parameters {
		param := sg.Parameters[i]

		// Type Check Parameter Value
		t := c.evaluateTypeExpression(p.Type, sg.TypeParameters, ctx)
		err := c.validateAssignment(param, t, p, true)
		if err != nil {
			c.addError(err.Error(), p.Range())
		}
	}

	// Annotated Return Type
	if e.ReturnType != nil {
		t := c.evaluateTypeExpression(e.ReturnType, sg.TypeParameters, ctx)
		err := c.validateAssignment(sg.Result, t, e.ReturnType, true)

		if err != nil {
			c.addError(err.Error(), e.ReturnType.Range())
		}
	}

	fn.IsAsync = e.IsAsync
	fn.IsMutating = e.IsMutating
	fn.IsStatic = e.IsStatic

	return sg
}
