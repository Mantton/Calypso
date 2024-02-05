package compiler

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/ast"
)

func (c *Compiler) compileStatement(stmt ast.Statement) {
	fmt.Printf("Compiling %T\n", stmt)

	switch stmt := stmt.(type) {
	case *ast.VariableStatement:
		c.compileVariableStatement(stmt)
	default:
		fmt.Printf("[COMPILER] Missing Compilation Implementation for %T", stmt)
	}
}

func (c *Compiler) compileVariableStatement(stmt *ast.VariableStatement) {

	value := c.evaluateExpression(stmt.Value)

	fmt.Println(value)
}
