package compiler

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/collections"
	"github.com/mantton/calypso/internal/calypso/typechecker"
	"tinygo.org/x/go-llvm"
)

type Compiler struct {
	context     llvm.Context
	module      llvm.Module
	builder     llvm.Builder
	namedValues map[string]llvm.Value
	stack       collections.Stack[llvm.Value]
}

func New() *Compiler {
	return &Compiler{}
}

func (c *Compiler) Compile(file *ast.File, sym *typechecker.SymbolTable) {
	c.stack = collections.Stack[llvm.Value]{}
	data := c.compileFile(file, sym)

	fmt.Println("\nDATA")
	fmt.Println(data)
}

func (c *Compiler) doInScope(fn func()) {

	prev := c.namedValues
	newS := make(map[string]llvm.Value)
	c.namedValues = newS
	fn()
	c.namedValues = prev
}
