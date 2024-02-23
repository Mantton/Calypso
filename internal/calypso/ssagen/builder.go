package ssagen

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/ssa"
	"github.com/mantton/calypso/internal/calypso/symbols"
)

type builder struct {
	Mod *ssa.Module
	Fn  *ssa.Function
}

func build(e *ssa.Executable) {
	b := &builder{
		Mod: ssa.NewModule(e.IncludedFile, "main"),
	}

	for _, decl := range e.IncludedFile.Declarations {
		switch d := decl.(type) {
		case *ast.FunctionDeclaration:
			b.resolveFunction(d.Func)
		}
	}

	e.Modules["main"] = b.Mod
}

func (b builder) resolveFunction(d *ast.FunctionExpression) {

	fn := ssa.NewFunction(&symbols.SymbolInfo{Name: d.Identifier.Value})
	b.Mod.Functions[d.Identifier.Value] = fn
	b.Fn = fn
	fn.CurrentBlock = fn.NewBlock()
	b.resolveBlockStatement(d.Body, fn)
}

func (b builder) resolveStmt(n ast.Statement, fn *ssa.Function) {

	switch n := n.(type) {
	case *ast.VariableStatement:
		b.resolveVariableStmt(n, fn)
	// case *ast.IfStatement:
	case *ast.ReturnStatement:
		b.resolveReturnStmt(n, fn)
	case *ast.BlockStatement:
		panic("CANNOT BE CALLED DIRECTLY")
	case *ast.ExpressionStatement:
		i, ok := b.resolveExpr(n.Expr, fn).(ssa.Instruction)

		if ok {
			fn.Emit(i)
		}
		return
	default:
		panic(fmt.Sprintf("unknown statement %T\n", n))

	}

}

func (b *builder) resolveVariableStmt(n *ast.VariableStatement, fn *ssa.Function) {

	val := b.resolveExpr(n.Value, fn)

	// Global variables are constants known at compile-time
	if n.IsGlobal {

		return
	}

	// TODO: check if value is a constant
	// Constants known at compile time do not need to be allocated
	if n.IsConstant {
		return
	}

	// Variable, Allocate Memory & Store Address
	addr := emitLocalVar(fn, &symbols.SymbolInfo{Name: n.Identifier.Value, TypeDesc: &symbols.SymbolInfo{Name: "Test"}})

	// Store Data @ Addr

	emitStore(fn, addr, val)
}

func (b *builder) resolveReturnStmt(n *ast.ReturnStatement, fn *ssa.Function) {

	val := b.resolveExpr(n.Value, fn)

	i := &ssa.Return{
		Result: val,
	}

	fn.Emit(i)
}

func (b *builder) resolveBlockStatement(n *ast.BlockStatement, fn *ssa.Function) {

	for _, s := range n.Statements {
		b.resolveStmt(s, fn)
	}
}

func (b *builder) resolveExpr(n ast.Expression, fn *ssa.Function) ssa.Value {
	switch n := n.(type) {
	case *ast.IntegerLiteral:
		return &ssa.Constant{
			Value: n.Value,
		}
	case *ast.CallExpression:
		return &ssa.Call{
			Target:    "Foo",
			Arguments: nil,
		}
	case *ast.IdentifierExpression:
		addr := fn.Variables[n.Value]
		return &ssa.Load{
			Address: addr,
		}
	case *ast.AssignmentExpression:
		a := b.resolveExpr(n.Target, fn)
		v := b.resolveExpr(n.Value, fn)
		emitStore(fn, a, v)
		return nil
	case *ast.BinaryExpression:
		lhs, rhs := b.resolveExpr(n.Left, fn), b.resolveExpr(n.Right, fn)
		return &ssa.Binary{
			Left:  lhs,
			Op:    n.Op,
			Right: rhs,
		}

	default:
		panic(fmt.Sprintf("unknown expr %T\n", n))

	}
}
