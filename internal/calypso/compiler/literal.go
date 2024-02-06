package compiler

import (
	"github.com/mantton/calypso/internal/calypso/ast"
	"tinygo.org/x/go-llvm"
)

// Base Literals
func (c *Compiler) VisitIntegerLiteral(n *ast.IntegerLiteral) {
	t := c.context.Int32Type()
	c.stack.Push(llvm.ConstInt(t, uint64(n.Value), false))
}
func (c *Compiler) VisitFloatLiteral(n *ast.FloatLiteral) {
	panic("MISSING_IMPLEMENTATION")
}
func (c *Compiler) VisitStringLiteral(n *ast.StringLiteral) {
	panic("MISSING_IMPLEMENTATION")
}
func (c *Compiler) VisitBooleanLiteral(n *ast.BooleanLiteral) {
	panic("MISSING_IMPLEMENTATION")
}
func (c *Compiler) VisitNullLiteral(n *ast.NullLiteral) {
	panic("MISSING_IMPLEMENTATION")
}
func (c *Compiler) VisitVoidLiteral(n *ast.VoidLiteral) {
	panic("MISSING_IMPLEMENTATION")
}

// Complex Literals
func (c *Compiler) VisitCompositeLiteral(n *ast.CompositeLiteral) {
	panic("MISSING_IMPLEMENTATION")
}
func (c *Compiler) VisitArrayLiteral(n *ast.ArrayLiteral) {
	panic("MISSING_IMPLEMENTATION")
}
func (c *Compiler) VisitMapLiteral(n *ast.MapLiteral) {
	panic("MISSING_IMPLEMENTATION")
}
