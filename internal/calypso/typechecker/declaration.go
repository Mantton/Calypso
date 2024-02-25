package t

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
		// case *ast.StatementDeclaration:
		// case *ast.StandardDeclaration:
		// case *ast.ExtensionDeclaration:
		// case *ast.TypeDeclaration:
		default:
			msg := fmt.Sprintf("declaration check not implemented, %T", decl)
			panic(msg)
		}
	}()
}
