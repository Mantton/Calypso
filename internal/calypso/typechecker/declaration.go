package typechecker

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/lexer"
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
		c.currentNode = decl

		switch decl := decl.(type) {
		case *ast.ConstantDeclaration:
			c.checkStatement(decl.Stmt)
		case *ast.FunctionDeclaration:
			c.checkExpression(decl.Func)
		case *ast.StatementDeclaration:
			c.checkStatement(decl.Stmt)
		case *ast.StandardDeclaration:
			c.checkStandardDeclaration(decl)
		case *ast.ExtensionDeclaration:
			c.checkExtension(decl)
		case *ast.TypeDeclaration:
			c.checkTypeDeclaration(decl)
		default:
			msg := fmt.Sprintf("declaration check not implemented, %T", decl)
			panic(msg)
		}
	}()
}

func (c *Checker) checkStandardDeclaration(d *ast.StandardDeclaration) {
	// TODO: Scope to Module/Package
	standard := newSymbolInfo(d.Identifier.Value, StandardSymbol)

	ok := c.define(standard)

	for _, expr := range d.Block.Statements {

		fn, ok := expr.(*ast.FunctionStatement)

		if !ok {
			c.addError("Only Functions are allowed in a Standards body", expr.Range())
			continue
		}

		// * NOTE: Parser already ensures the body is not part here

		// TODO: Parse Function Type
		fnDesc := newSymbolInfo(fn.Func.Identifier.Value, FunctionSymbol)
		// Add Property
		ok = standard.addProperty(fnDesc)

		if !ok {
			// already defined
			c.addError(fmt.Sprintf("`%s` is already defined in `%s`", fnDesc.Name, standard.Name), fn.Range())
			continue
		}
	}

	if !ok {
		msg := fmt.Sprintf("`%s` is already defined.", d.Identifier.Value)
		c.addError(msg, d.Identifier.Range())
	}

}

func (c *Checker) checkExtension(d *ast.ExtensionDeclaration) {

	s, ok := c.find(d.Identifier.Value)

	if !ok {
		msg := fmt.Sprintf("Unable to locate `%s`", d.Identifier.Value)
		c.addError(msg, d.Identifier.Range())
	}

	action := func(t *SymbolInfo, node ast.Node) {
		if s != nil {
			ok = s.addProperty(t)

			if !ok {
				c.addError(fmt.Sprintf("`%s` is already defined in `%s`", t.Name, s.Name), node.Range())
			}
		}
	}

	for _, fn := range d.Content {
		// TODO: Parse Function Type
		fnDesc := newSymbolInfo(fn.Func.Identifier.Value, FunctionSymbol)
		action(fnDesc, fn)
	}

}

func (c *Checker) checkTypeDeclaration(d *ast.TypeDeclaration) {

	s := &SymbolInfo{
		Name: d.Identifier.Value,
		Type: TypeSymbol,
	}

	ok := c.define(s)

	// Already Defined,
	if !ok {
		c.addError(fmt.Sprintf("`%s` is already defined", d.Identifier.Value), d.Identifier.Range())
	}

	// TODO: Check Generics
	e := c.evaluateTypeExpression(d.Value)
	s.ChildOf = e
}
