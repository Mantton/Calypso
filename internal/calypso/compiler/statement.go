package compiler

import (
	"github.com/mantton/calypso/internal/calypso/ast"
	"tinygo.org/x/go-llvm"
)

func (c *Compiler) VisitIfStatement(n *ast.IfStatement) {
	panic("MISSING_IMPLEMENTATION")
}
func (c *Compiler) VisitExpressionStatement(n *ast.ExpressionStatement) {
	panic("MISSING_IMPLEMENTATION")
}
func (c *Compiler) VisitWhileStatement(n *ast.WhileStatement) {
	panic("MISSING_IMPLEMENTATION")
}
func (c *Compiler) VisitReturnStatement(n *ast.ReturnStatement) {

	n.Value.Accept(c)

	val, ok := c.stack.Pop()

	if !ok {
		panic("expected a value")
	}

	if val.Type() == c.context.VoidType() {
		c.builder.CreateRetVoid()
	} else {
		c.builder.CreateRet(val)
	}
}
func (c *Compiler) VisitBlockStatement(n *ast.BlockStatement) {
	for _, stmt := range n.Statements {
		stmt.Accept(c)
	}
}

func (c *Compiler) VisitVariableStatement(n *ast.VariableStatement) {
	key := n.Identifier.Value
	n.Value.Accept(c)
	value, ok := c.stack.Pop()
	t := value.Type()

	if !ok {
		panic("expected value from stack")
	}

	if n.IsGlobal {
		// Globals are always constant
		g := llvm.AddGlobal(c.module, t, "")
		g.SetInitializer(value)
	} else {
		ptr := c.builder.CreateAlloca(t, "")
		c.builder.CreateStore(value, ptr)
		c.namedValues[key] = ptr
	}

}
func (c *Compiler) VisitFunctionStatement(n *ast.FunctionStatement) {
	panic("MISSING_IMPLEMENTATION")
}
func (c *Compiler) VisitAliasStatement(n *ast.AliasStatement) {
	panic("MISSING_IMPLEMENTATION")
}
func (c *Compiler) VisitStructStatement(n *ast.StructStatement) {
	panic("MISSING_IMPLEMENTATION")
}
