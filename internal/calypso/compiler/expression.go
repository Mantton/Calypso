package compiler

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/ast"
	"tinygo.org/x/go-llvm"
)

func (c *Compiler) evaluateExpression(expr ast.Expression) llvm.Value {
	fmt.Printf("Compiling %T\n", expr)

	switch expr := expr.(type) {
	case *ast.IntegerLiteral:
		return llvm.ConstInt(c.context.Int32Type(), uint64(expr.Value), false)

	default:
		fmt.Printf("[COMPILER] Missing Compilation Implementation for %T", expr)
		panic("")

	}

}
