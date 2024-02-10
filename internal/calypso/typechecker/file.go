package typechecker

import (
	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/symbols"
)

func (c *Checker) CheckFile(file *ast.File) *symbols.SymbolTable {
	c.enterScope() // global enter
	c.injectLiterals()
	if len(file.Declarations) != 0 {
		for _, decl := range file.Declarations {
			c.checkDeclaration(decl)
		}
	}

	c.leaveScope(true) // global leave

	return c.symbols
}
