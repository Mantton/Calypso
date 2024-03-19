package typechecker

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/types"
)

func (c *Checker) Check() *SymbolTable {
	c.enterScope() // global enter
	main := c.scope
	main.Parent = types.GlobalScope
	if len(c.file.Declarations) != 0 {
		for _, decl := range c.file.Declarations {
			c.checkDeclaration(decl)
		}
	}
	c.leaveScope() // global leave
	fmt.Println(main)

	for _, s := range c.table.scopes {
		fmt.Println(s)
	}

	c.table.Main = main
	return c.table

}
