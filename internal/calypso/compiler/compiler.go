package compiler

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/typechecker"
	"tinygo.org/x/go-llvm"
)

type Compiler struct {
	context llvm.Context
	module  llvm.Module
	builder llvm.Builder
}

func New() *Compiler {
	return &Compiler{}
}

func (c *Compiler) Compile(file *ast.File, sym *typechecker.SymbolTable) {
	data := c.compileFile(file, sym)

	fmt.Println("\nDATA")
	fmt.Println(data)
}
