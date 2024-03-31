package lirgen

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/lir"
	"github.com/mantton/calypso/internal/calypso/types"
)

func (b *builder) evaluateExpression(n ast.Expression, fn *lir.Function) lir.Value {
	switch e := n.(type) {
	case *ast.BooleanLiteral:
		return lir.NewConst(e.Value, types.LookUp(types.Bool))
	case *ast.StringLiteral:
		// TODO: Global Composites
		panic("string literals")
	case *ast.CharLiteral:
		return lir.NewConst(e.Value, types.LookUp(types.Char))
	case *ast.IntegerLiteral:
		typ := b.Mod.TypeTable().GetNodeType(n)
		if typ == nil {
			typ = types.LookUp(types.Int)
		}
		return lir.NewConst(e.Value, typ)
	case *ast.FloatLiteral:
		typ := b.Mod.TypeTable().GetNodeType(n)

		if typ == nil {
			typ = types.LookUp(types.Float)
		}

		return lir.NewConst(e.Value, typ)
	case *ast.NilLiteral:
		typ := b.Mod.TypeTable().GetNodeType(n)
		if typ == nil {
			panic("unknown nullptr type")
		}

		return lir.NewConst(0, typ)
	case *ast.IdentifierExpression:
		return b.evaluateIdentifierExpression(e, fn)
	case *ast.CallExpression:
		return b.evaluateCallExpression(e, fn)
	case *ast.CallArgument:
		return b.evaluateExpression(e.Value, fn)
	case *ast.AssignmentExpression:
		return b.evaluateAssignmentExpression(e, fn)
	case *ast.BinaryExpression:
		return b.evaluateBinaryExpression(e, fn)
	case *ast.GroupedExpression:
		return b.evaluateExpression(e.Expr, fn)
	default:
		msg := fmt.Sprintf("unknown expr %T\n", e)
		panic(msg)
	}
}

func (b *builder) evaluateCallExpression(n *ast.CallExpression, fn *lir.Function) lir.Value {
	val := b.evaluateExpression(n.Target, fn)

	f, ok := val.(*lir.Function)

	if !ok {
		panic("target cannot be invoked")
	}

	var args []lir.Value

	for _, p := range n.Arguments {
		v := b.evaluateExpression(p, fn)
		args = append(args, v)
	}

	i := &lir.Call{
		Target:    f,
		Arguments: args,
	}

	typ := b.Mod.TModule.Table.GetNodeType(n)

	if typ == nil {
		panic("???")
	}

	i.SetType(typ)
	fn.Emit(i)
	return i
}

func (b *builder) evaluateIdentifierExpression(n *ast.IdentifierExpression, fn *lir.Function) lir.Value {

	// Global Constant
	cons, ok := b.Mod.GlobalConstants[n.Value]
	if ok {
		return cons
	}

	// Function
	f, ok := b.Mod.Functions[n.Value]
	if ok {
		return f
	}

	// Scoped Variable
	val, ok := fn.Variables[n.Value]

	if ok {
		switch val := val.(type) {
		case *lir.Allocate:
			i := &lir.Load{
				Address: val,
			}

			i.SetType(val.Yields())
			fn.Emit(i)
			return i
		case *lir.Constant, *lir.Parameter:
			return val
		default:
			panic(fmt.Sprintf("identifier found invalid type: %T", val))
		}
	}

	panic("unable to locate identifier")

}

func (b *builder) evaluateAssignmentExpression(n *ast.AssignmentExpression, fn *lir.Function) lir.Value {
	a := b.evaluateExpression(n.Target, fn)
	v := b.evaluateExpression(n.Target, fn)
	b.emitStore(fn, a, v)
	return nil
}

func (b *builder) evaluateBinaryExpression(n *ast.BinaryExpression, fn *lir.Function) lir.Value {
	lhs, rhs := b.evaluateExpression(n.Left, fn), b.evaluateExpression(n.Right, fn)
	i := &lir.Binary{
		Left:  lhs,
		Op:    n.Op,
		Right: rhs,
	}
	i.SetType(lhs.Yields())
	fn.Emit(i)
	return i
}
