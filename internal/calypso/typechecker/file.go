package t

import (
	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/types"
)

func (c *Checker) CheckFile(file *ast.File) *types.Scope {
	c.enterScope() // global enter
	main := c.scope
	main.Parent = types.GlobalScope
	if len(file.Declarations) != 0 {
		for _, decl := range file.Declarations {
			c.checkDeclaration(decl)
		}
	}
	c.leaveScope() // global leave
	return main

}
