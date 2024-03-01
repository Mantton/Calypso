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
		// case *ast.ExtensionDeclaration:
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
	s := types.NewDefinedType(d.Identifier.Value, typ, nil)

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

	t, ok := xx.Type().(*types.DefinedType)

	if !ok {
		c.addError(fmt.Sprintf("%s is not a defined type", d.Target.Value), d.Target.Range())
		return
	}

	c.enterScope()
	defer c.leaveScope()
	for _, stmt := range d.Content {
		c.checkFunctionExpression(stmt.Func)
		n := stmt.Func.Identifier.Value
		fn, ok := c.find(n)

		if !ok {
			panic("wut")
		}

		if x := fn.(*types.Function); x != nil {
			ok := t.AddMethod(x.Name(), x)
			if !ok {
				panic("method already in type")
			}
			continue
		}

		panic("x is not a function")
	}

	for _, x := range s.Dna {
		n, ok := t.Methods[x.Name()]

		if !ok {
			c.addError(fmt.Sprintf("%s does not conform to `%s`, missing %s", d.Target.Value, s.Name, x.Name()), d.Target.Range())
			return
		}
		_, err := c.validate(x.Type(), n.Type())

		if err != nil {
			c.addError(err.Error(), d.Target.Range())
			return
		}
	}

}
