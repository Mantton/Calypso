package compiler

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/ast"
)

func (c *Compiler) compileDeclaration(decl ast.Declaration) {
	fmt.Printf("Compiling %T\n", decl)
	switch decl := decl.(type) {
	case *ast.FunctionDeclaration:
		fmt.Println("---")
	case *ast.ConstantDeclaration:
		c.compileStatement(decl.Stmt)
	default:
		panic("UNKNOWN DECL")
	}
}
