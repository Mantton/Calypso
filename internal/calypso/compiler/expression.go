package compiler

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/token"
	"tinygo.org/x/go-llvm"
)

// Base Expressions
func (c *Compiler) VisitFunctionExpression(n *ast.FunctionExpression) {

	fnType := llvm.FunctionType(c.context.VoidType(), nil, false)
	fn := llvm.AddFunction(c.module, n.Identifier.Value, fnType)
	entry := llvm.AddBasicBlock(fn, "entry")

	c.builder.SetInsertPointAtEnd(entry)

	c.doInScope(func() {
		n.Body.Accept(c)
	})
}
func (c *Compiler) VisitIdentifierExpression(n *ast.IdentifierExpression) {
	c.stack.Push(c.namedValues[n.Value])
}

// Complex Expressions
func (c *Compiler) VisitGroupedExpression(n *ast.GroupedExpression) {
	panic("MISSING_IMPLEMENTATION")
}
func (c *Compiler) VisitCallExpression(n *ast.CallExpression) {
	panic("MISSING_IMPLEMENTATION")
}
func (c *Compiler) VisitUnaryExpression(n *ast.UnaryExpression) {
	panic("MISSING_IMPLEMENTATION")
}
func (c *Compiler) VisitBinaryExpression(n *ast.BinaryExpression) {

	n.Left.Accept(c)
	lhs, ok := c.stack.Pop()

	if !ok {
		panic("expected value")
	}

	n.Right.Accept(c)

	rhs, ok := c.stack.Pop()
	if !ok {
		panic("expected value")
	}

	// fmt.Println(lhs.Type())

	// if lhs.Type() != rhs.Type() {
	// 	panic("[Compiler Binary Expr] TYPE MISMATCH")
	// }

	x := lhs.Type().IsNil()

	fmt.Println(x)

	switch n.Op {
	case token.ADD:
		sum := c.builder.CreateAdd(lhs, rhs, "")
		c.stack.Push(sum)
		return
	}
	// fmt.Println(lhs.Type())
	fmt.Println(lhs, rhs)
	panic("MISSING_IMPLEMENTATION")
}
func (c *Compiler) VisitAssignmentExpression(n *ast.AssignmentExpression) {
	panic("MISSING_IMPLEMENTATION")
}
func (c *Compiler) VisitIndexExpression(n *ast.IndexExpression) {
	panic("MISSING_IMPLEMENTATION")
}
func (c *Compiler) VisitPropertyExpression(n *ast.PropertyExpression) {
	panic("MISSING_IMPLEMENTATION")
}
func (c *Compiler) VisitKeyValueExpression(n *ast.KeyValueExpression) {
	panic("MISSING_IMPLEMENTATION")
}
func (c *Compiler) VisitCompositeLiteralBodyClause(n *ast.CompositeLiteralBodyClause) {
	panic("MISSING_IMPLEMENTATION")
}
