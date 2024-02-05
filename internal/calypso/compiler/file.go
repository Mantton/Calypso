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

	for _, decl := range decls {
		c.compileDeclaration(decl)
	}

	// fnType := llvm.FunctionType(c.context.Int32Type(), []llvm.Type{}, false)
	// fn := llvm.AddFunction(c.module, "main", fnType)
	// entry := llvm.AddBasicBlock(fn, "entry")
	// c.builder.SetInsertPointAtEnd(entry)

	// // Construct
	// one := llvm.ConstInt(c.context.Int32Type(), 1, false)
	// sum := c.builder.CreateAdd(one, one, "sum")

	// c.builder.CreateRet(sum)
	// fmt.Println(c.module.String())

	return c.module.String()

}
