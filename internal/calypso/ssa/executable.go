package ssa

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/ast"
)

func NewExecutable(file *ast.File) *Executable {
	return &Executable{
		IncludedFile: file,
		Modules:      make(map[string]*Module),
	}
}

type builder struct {
	Mod   *Module
	Fn    *Function
	Block *Block
}

func (n *Executable) Build() {
	b := &builder{
		Mod: NewModule(n.IncludedFile, "main"),
	}

	for _, decl := range n.IncludedFile.Declarations {
		switch d := decl.(type) {
		case *ast.FunctionDeclaration:
			b.resolveFunction(d.Func)
		}
	}

	n.Modules["main"] = b.Mod
}

func (b builder) resolveFunction(d *ast.FunctionExpression) {

	fn := &Function{}
	b.Mod.Functions[d.Identifier.Value] = fn
	b.Fn = fn
	b.Block = &Block{}
	b.resolveBlockStatement(d.Body, fn)
	fn.Blocks = append(fn.Blocks, b.Block)
}

func (b builder) resolveStmt(n ast.Statement, fn *Function) {

	switch n := n.(type) {
	case *ast.VariableStatement:
		b.resolveVariableStmt(n, fn)
		return
	case *ast.IfStatement:
		return
	case *ast.ReturnStatement:
		b.resolveReturnStmt(n, fn)
		return
	case *ast.BlockStatement:
		panic("CANNOT BE CALLED DIRECTLY")
	case *ast.ExpressionStatement:
		v, ok := b.resolveExpr(n.Expr).(Instruction)

		if ok {
			b.Block.Instructions = append(b.Block.Instructions, v)

		}
		return
	}

	panic(fmt.Sprintf("unknown statement %T\n", n))
}

func (b *builder) resolveVariableStmt(n *ast.VariableStatement, fn *Function) {

	// val := b.resolveExpr(n.Value)

	// if n.IsGlobal {

	// }

	// instr := &Assign{
	// 	Target: &Variable{Name: n.Identifier.Value},
	// 	Value:  val,
	// }

	// b.Block.Instructions = append(b.Block.Instructions, instr)
}

func (b *builder) resolveReturnStmt(n *ast.ReturnStatement, fn *Function) {

	val := b.resolveExpr(n.Value)

	instr := &Return{
		Result: val,
	}

	b.Block.Instructions = append(b.Block.Instructions, instr)
}

func (b *builder) resolveBlockStatement(n *ast.BlockStatement, fn *Function) {

	for _, s := range n.Statements {
		b.resolveStmt(s, fn)
	}
}

func (b *builder) resolveExpr(n ast.Expression) Value {
	// switch n := n.(type) {
	// case *ast.IntegerLiteral:
	// 	return &Constant{
	// 		Value: n.Value,
	// 	}
	// case *ast.CallExpression:
	// 	return &Call{
	// 		Target:    "Foo",
	// 		Arguments: nil,
	// 	}
	// case *ast.IdentifierExpression:
	// 	return &Variable{
	// 		Name: n.Value,
	// 	}
	// case *ast.AssignmentExpression:
	// 	v := b.resolveExpr(n.Target).(*Variable)
	// 	return &Assign{
	// 		Target: v,
	// 		Value:  b.resolveExpr(n.Value),
	// 	}
	// case *ast.BinaryExpression:
	// 	lhs, rhs := b.resolveExpr(n.Left), b.resolveExpr(n.Right)
	// 	return &Binary{
	// 		Left:  lhs,
	// 		Op:    n.Op,
	// 		Right: rhs,
	// 	}
	// }

	panic(fmt.Sprintf("unknown expr %T\n", n))
}
