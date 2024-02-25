package t

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
	typ := types.NewStandard()
	s := types.NewTypeDef(d.Identifier.Value, typ)

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

		// Evalutate Function Signature
		sg := c.evaluateFunctionSignature(fn.Func)

		// Add method
		ok = typ.AddMethod(n, sg)

		// already defined in standard, add error
		if !ok {
			c.addError(fmt.Sprintf("`%s` is already defined in `%s`", n, s.Name()), fn.Range())
			continue
		}
	}
}
