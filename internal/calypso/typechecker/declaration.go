package typechecker

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/types"
)

func (c *Checker) checkStandardDeclaration(d *ast.StandardDeclaration) {
	standard := c.resolve(d.Identifier, d, c.ctx.scope)

	if standard == nil {
		return
	}
	underlying := standard.Parent().(*types.Standard)

	if underlying == nil {
		return
	}

	scope := standard.GetScope()
	ctx := NewContext(scope, nil, nil)

	// Loop through statements in standard definition

	for _, expr := range d.Block.Statements {

		switch node := expr.(type) {

		case *ast.FunctionStatement:
			n := node.Func.Identifier.Value
			// Parser ensures only signatures & no function bodies in standard decl

			// evaluate Function Signature
			sg := c.registerFunctionSignatures(node.Func)

			f := types.NewFunction(n, sg, c.module)
			// Add method
			ok := underlying.AddMethod(n, f)

			// already defined in standard, add error
			if !ok {
				c.addError(fmt.Sprintf("`%s` is already defined in `%s`", n, standard.Name()), node.Range())
				continue
			}

		case *ast.TypeStatement:
			c.checkTypeStatement(node, ctx)

		default:
			c.addError("cannot use statement in standard declaration", node.Range())
			continue
		}

	}
}

func (c *Checker) checkConformanceDeclaration(d *ast.ConformanceDeclaration) {

	x, ok := c.GlobalFind(d.Standard.Value)

	if !ok {
		c.addError(fmt.Sprintf("cannot find %s", d.Standard.Value), d.Standard.Range())
		return
	}

	s, ok := x.Type().Parent().(*types.Standard)

	if !ok {
		c.addError(fmt.Sprintf("%s is not a standard", d.Standard.Value), d.Standard.Range())
		return
	}

	target, ok := c.GlobalFind(d.Target.Value)

	if !ok {
		c.addError(fmt.Sprintf("cannot find %s", d.Target.Value), d.Target.Range())
		return
	}

	typ := types.AsDefined(target.Type())

	if typ == nil {
		c.addError(fmt.Sprintf("%s is not a defined type", d.Target.Value), d.Target.Range())
		return
	}

	scope := typ.GetScope()
	ctx := NewContext(scope, nil, nil)
	// Inject types into scope
	for _, node := range d.Types {
		c.checkTypeStatement(node, ctx)
	}

	// ensure all required types have been injected
	for _, t := range s.Types {
		sym := scope.ResolveInCurrent(t.Name())

		if sym == nil {
			c.addError(fmt.Sprintf("%s does not conform to `%s`, missing `%s`", d.Target.Value, s.Name, t.Name()), d.Target.Range())
			return
		}

		switch sym.(type) {
		case types.Type:
			continue
		default:
			c.addError(fmt.Sprintf("\"%s\" does not conform to \"%s\", \"%s\" is defined but not a type", d.Target.Value, s.Name, t.Name()), d.Target.Range())
			return
		}

	}

	// add functions to type
	c.injectFunctionsInType(d.Signatures, typ)

	// Ensure All Functions of Standard are implemented
	for _, eFn := range s.Signature {

		// Get Implemented Method
		pFn := typ.ResolveMethod(eFn.Name())

		if pFn == nil {
			c.addError(fmt.Sprintf("%s does not conform to `%s`, missing `%s`", d.Target.Value, s.Name, eFn.Name()), d.Target.Range())
			return
		}

		// Ensure Function Is of same signature
		_, err := c.validate(eFn.Sg(), pFn)

		if err != nil {
			c.addError(err.Error(), d.Target.Range())
		}
	}

}

func (c *Checker) checkExtensionDeclaration(d *ast.ExtensionDeclaration) {

	// Find Symbol
	name := d.Identifier.Value
	sym, ok := c.GlobalFind(name)

	if !ok {
		c.addError(fmt.Sprintf("cannot find %s", name), d.Identifier.Range())
		return
	}

	// Cast as Type
	typ := types.AsDefined(sym.Type())

	if typ == nil {
		c.addError(fmt.Sprintf("%s is not a type", name), d.Identifier.Range())
		return
	}

	c.injectFunctionsInType(d.Content, typ)
}

func (c *Checker) injectFunctionsInType(fns []*ast.FunctionStatement, t *types.DefinedType) {
	// Define Functions in Type Scope
	for _, stmt := range fns {
		fn := stmt.Func
		sg := c.registerFunctionSignatures(fn)

		if sg.IsStatic {
			continue
		}

		// Inject `self`
		self := types.NewVar("self", t)
		self.Mutable = sg.IsMutating
		sg.Self = self

		// Default
		sg.Scope.Define(self)
	}
}
