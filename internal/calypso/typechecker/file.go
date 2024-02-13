package typechecker

import (
	"github.com/mantton/calypso/internal/calypso/ast"
)

func (c *Checker) CheckFile(file *ast.File) {
	c.enterScope() // global enter
	c.injectLiterals()
	if len(file.Declarations) != 0 {
		for _, decl := range file.Declarations {
			c.checkDeclaration(decl)
		}
	}

	c.leaveScope(true) // global leave

	file.Symbols = c.symbols

}
