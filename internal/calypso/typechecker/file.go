package typechecker

import (
	"errors"

	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/types"
)

func (c *Checker) Check() (*SymbolTable, error) {

	main := types.NewScope(types.GlobalScope)
	main.Parent = types.GlobalScope
	c.table.Main = main

	for _, file := range c.fileSet.Files {
		c.checkFile(file)
	}

	c.table.DebugPrintScopes()

	if len(c.Errors) != 0 {
		return nil, errors.New(c.Errors.String())
	}

	return c.table, nil
}

func (c *Checker) checkFile(f *ast.File) {
	c.file = f
	mainContext := NewContext(c.table.Main, nil, nil)

	if len(f.Declarations) != 0 {
		for _, decl := range f.Declarations {
			c.checkDeclaration(decl, mainContext)
		}
	}
}
