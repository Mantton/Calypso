package compiler

import (
	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/typechecker"
	"tinygo.org/x/go-llvm"
)

func (c *Compiler) compileFile(file *ast.File, symbols *typechecker.SymbolTable) string {
	// Data from File
	decls := file.Declarations
	name := file.ModuleName

	// Setup
	c.context = llvm.NewContext()
	c.module = c.context.NewModule(name)
	c.builder = c.context.NewBuilder()
	defer c.context.Dispose()
	defer c.module.Dispose()
	defer c.builder.Dispose()

	c.doInScope(func() {
		for _, decl := range decls {
			decl.Accept(c)
		}
	})

	return c.module.String()
}
