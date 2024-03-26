package typechecker

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/types"
)

func (c *Checker) checkDeclaration(decl ast.Declaration, ctx *NodeContext) {
	fmt.Printf(
		"Checking Declaration: %T @ Line %d\n",
		decl,
		decl.Range().Start.Line,
	)
	switch decl := decl.(type) {
	case *ast.ConstantDeclaration:
		c.checkStatement(decl.Stmt, ctx)
	case *ast.FunctionDeclaration:
		c.checkExpression(decl.Func, ctx)
	case *ast.StatementDeclaration:
		c.checkStatement(decl.Stmt, ctx)
	case *ast.StandardDeclaration:
		c.checkStandardDeclaration(decl)
	case *ast.ConformanceDeclaration:
		c.checkConformanceDeclaration(decl)
	case *ast.ExtensionDeclaration:
		c.checkExtensionDeclaration(decl)
	case *ast.ExternDeclaration:
		break // Function Signature is registered in first loop through
	default:
		msg := fmt.Sprintf("declaration check not implemented, %T", decl)
		panic(msg)
	}
}

func (c *Checker) checkStandardDeclaration(d *ast.StandardDeclaration) {
	standard := c.resolve(d.Identifier, d, c.ctx.scope)
	scope := standard.GetScope()
	underlying := standard.Parent().(*types.Standard)
	// Loop through statements in standard definition
	ctx := NewContext(scope, nil, nil)
	for _, expr := range d.Block.Statements {

		switch node := expr.(type) {

		case *ast.FunctionStatement:
			n := node.Func.Identifier.Value
			// Parser ensures only signatures & no function bodies in standard decl

			// evaluate Function Signature
			sg := c.evaluateFunctionSignature(node.Func, ctx)

			f := types.NewFunction(n, sg)
			// Add method
			err := standard.AddMethod(n, f)

			// already defined in standard, add error
			if err != nil {
				c.addError(fmt.Sprintf("`%s` is already defined in `%s`", n, standard.Name()), node.Range())
				continue
			}

		case *ast.TypeStatement:
			c.checkTypeStatement(node, underlying, ctx)

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
		c.checkTypeStatement(node, nil, ctx)
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
	c.injectFunctionsInType(typ, d.Signatures)

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

	c.injectFunctionsInType(typ, d.Content)
}

func (c *Checker) injectFunctionsInType(typ *types.DefinedType, fns []*ast.FunctionStatement) {
	// Define Functions in Type Scope
	for _, stmt := range fns {
		c.evaluateFunctionExpression(stmt.Func, typ)
	}
}

func (c *Checker) checkExternDeclaration(n *ast.ExternDeclaration) {
	target := n.Target
	fmt.Printf("[DEBUG] External Target \"%s\"\n", target.Value)

	ctx := NewContext(c.ParentScope(), nil, nil)
	for _, node := range n.Signatures {
		// eval function
		fn := types.NewFunction(node.Func.Identifier.Value, nil)
		fn.Target = &types.FunctionTarget{
			Target: target.Value,
		}

		err := c.GlobalDefine(fn)

		// Define in type scope
		if err != nil {
			c.addError(
				err.Error(),
				node.Func.Identifier.Range())
			continue
		}

		// Resolve Function Body
		sg := c.evaluateFunctionSignature(node.Func, ctx)
		fn.SetSignature(sg)

		if types.IsGeneric(sg) {
			c.addError("external function cannot be generic", node.Func.Identifier.Range())
		}
	}
}
