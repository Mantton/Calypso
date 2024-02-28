package typechecker

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/types"
)

func (c *Checker) CheckFile(file *ast.File) *SymbolTable {
	c.enterScope() // global enter
	main := c.scope
	main.Parent = types.GlobalScope
	if len(file.Declarations) != 0 {
		for _, decl := range file.Declarations {
			c.checkDeclaration(decl)
		}
	}
	c.leaveScope() // global leave
	fmt.Println(main)

	for _, s := range c.table.scopes {
		fmt.Println(s)
	}

	return c.table

}
