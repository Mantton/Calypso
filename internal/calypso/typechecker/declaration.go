package typechecker

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/lexer"
	"github.com/mantton/calypso/internal/calypso/types"
)

func (c *Checker) checkDeclaration(decl ast.Declaration) {
	func() {
		defer func() {
			if r := recover(); r != nil {

				if err, y := r.(lexer.Error); y {
					c.Errors.Add(err)
				} else {
					fmt.Println(c.scope)
					panic(r)
				}
			}
		}()
		fmt.Printf(
			"Checking Declaration: %T @ Line %d\n",
			decl,
			decl.Range().Start.Line,
		)
		switch decl := decl.(type) {
		case *ast.ConstantDeclaration:
			c.checkStatement(decl.Stmt)
		case *ast.FunctionDeclaration:
			c.checkExpression(decl.Func)
		case *ast.StatementDeclaration:
			c.checkStatement(decl.Stmt)
		case *ast.StandardDeclaration:
			c.checkStandardDeclaration(decl)
		case *ast.ConformanceDeclaration:
			c.checkConformanceDeclaration(decl)
		case *ast.ExtensionDeclaration:
			c.checkExtensionDeclaration(decl)

		// case *ast.TypeDeclaration:
		default:
			msg := fmt.Sprintf("declaration check not implemented, %T", decl)
			panic(msg)
		}
	}()
}

func (c *Checker) checkStandardDeclaration(d *ast.StandardDeclaration) {
	// declare type & it's definition
	typ := types.NewStandard(d.Identifier.Value)
	s := types.NewDefinedType(d.Identifier.Value, typ, nil, c.scope)

	// define in scope
	ok := c.define(s)

	if !ok {
		msg := fmt.Sprintf("`%s` is already defined.", d.Identifier.Value)
		c.addError(msg, d.Identifier.Range())
	}

	// Loop through statements in standard definition
	for _, expr := range d.Block.Statements {

		fn, ok := expr.(*ast.FunctionStatement)
		n := fn.Func.Identifier.Value

		if !ok {
			c.addError("Only Functions are allowed in a Standards body", expr.Range())
			continue
		}

		// Parser ensures only signatures & no function bodies in standard decl

		// evaluate Function Signature
		sg := c.evaluateFunctionSignature(fn.Func)

		f := types.NewFunction(n, sg)
		// Add method
		ok = typ.AddMethod(n, f)

		// already defined in standard, add error
		if !ok {
			c.addError(fmt.Sprintf("`%s` is already defined in `%s`", n, s.Name()), fn.Range())
			continue
		}
	}
}

func (c *Checker) checkConformanceDeclaration(d *ast.ConformanceDeclaration) {

	x, ok := c.find(d.Standard.Value)

	if !ok {
		c.addError(fmt.Sprintf("cannot find %s", d.Standard.Value), d.Standard.Range())
		return
	}

	s, ok := x.Type().Parent().(*types.Standard)

	if !ok {
		c.addError(fmt.Sprintf("%s is not a standard", d.Standard.Value), d.Standard.Range())
		return
	}

	xx, ok := c.find(d.Target.Value)

	if !ok {
		c.addError(fmt.Sprintf("cannot find %s", d.Target.Value), d.Target.Range())
		return
	}

	typ := types.AsDefined(xx.Type())

	if typ == nil {
		c.addError(fmt.Sprintf("%s is not a defined type", d.Target.Value), d.Target.Range())
		return
	}

	// add functions to type
	c.injectFunctionsInType(typ, d.Content)

	// Ensure All Functions of Standard are implemented
	for _, eFn := range s.Dna {

		// Get Implemented Method
		pFn, ok := typ.Methods[eFn.Name()]

		if !ok {
			c.addError(fmt.Sprintf("%s does not conform to `%s`, missing `%s`", d.Target.Value, s.Name, eFn.Name()), d.Target.Range())
			return
		}

		// Ensure Function Is of same signature
		_, err := c.validate(eFn.Sg(), pFn.Type())

		if err != nil {
			c.addError(err.Error(), d.Target.Range())
		}
	}

}

func (c *Checker) checkExtensionDeclaration(d *ast.ExtensionDeclaration) {

	// Find Symbol
	name := d.Identifier.Value
	sym, ok := c.find(name)

	if !ok {
		c.addError(fmt.Sprintf("cannot find %s", name), d.Identifier.Range())
		return
	}

	// Cast as Type
	typ := types.AsDefined(sym.Type())

	if typ == nil {
		c.addError(fmt.Sprintf("%s is not a type", name), d.Identifier.Range())
	}

	c.injectFunctionsInType(typ, d.Content)
}

func (c *Checker) injectFunctionsInType(typ *types.DefinedType, fns []*ast.FunctionStatement) {
	// Define Functions in Type Scope
	for _, stmt := range fns {

		// eval function
		sg := c.evaluateFunctionExpression(stmt.Func, typ)

		// error already reported
		if sg == unresolved {
			continue
		}

		fn := types.NewFunction(stmt.Func.Identifier.Value, sg.(*types.FunctionSignature))

		// Define in type scope
		ok := typ.Scope.Define(fn)
		if !ok {
			c.addError(fmt.Sprintf("%s is already defined in %s", fn.Name(), typ), stmt.Func.Identifier.Range())
		}

		// add to methods
		ok = typ.AddMethod(fn.Name(), fn)

		if !ok {
			panic("unreachable")
		}
	}
}
