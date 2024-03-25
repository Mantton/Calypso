package typechecker

import (
	"github.com/mantton/calypso/internal/calypso/types"
)

func (c *Checker) Check() *SymbolTable {

	main := types.NewScope(types.GlobalScope)
	main.Parent = types.GlobalScope
	c.table.Main = main
	mainContext := NewContext(main, nil, nil)

	if len(c.file.Declarations) != 0 {
		for _, decl := range c.file.Declarations {
			c.checkDeclaration(decl, mainContext)
		}
	}

	c.table.DebugPrintScopes()
	return c.table
}
