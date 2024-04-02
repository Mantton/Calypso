package lirgen

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/lir"
	"github.com/mantton/calypso/internal/calypso/token"
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
	case *ast.UnaryExpression:
		return b.evaluateUnaryExpression(e, fn)
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
		i.SetType(types.LookUp(types.Void))
	} else {
		i.SetType(typ)

	}

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

func (b *builder) evaluateUnaryExpression(n *ast.UnaryExpression, fn *lir.Function) lir.Value {

	rhs := b.evaluateExpression(n.Expr, fn)

	switch n.Op {
	case token.STAR:
		panic("todo: dereference")
	case token.AMP:
		panic("todo: get pointer Reference")
	case token.NOT:
		return b.evaluateLogicalNot(rhs)
	case token.MINUS:
		return b.evaluateArithmeticNegate(rhs)
	default:
		msg := fmt.Sprintf("unimplemented unary operand, %s", token.LookUp(n.Op))
		panic(msg)
	}
}

func (b *builder) evaluateBinaryExpression(n *ast.BinaryExpression, fn *lir.Function) lir.Value {
	lhs, rhs := b.evaluateExpression(n.Left, fn), b.evaluateExpression(n.Right, fn)
	switch n.Op {
	case token.PLUS:
		return b.evaluateArithmeticAddExpression(lhs, rhs)
	case token.MINUS:
		return b.evaluateArithmeticSubExpression(lhs, rhs)
	case token.QUO:
		return b.evaluateArithmeticDivExpression(lhs, rhs)
	case token.STAR:
		return b.evaluateArithmeticMulExpression(lhs, rhs)
	case token.PCT:
		return b.evaluateArithmeticRemExpression(lhs, rhs)
	case token.L_CHEVRON, token.R_CHEVRON, token.EQL, token.LEQ, token.GEQ:
		return b.evaluateArithmeticComparison(n.Op, lhs, rhs)

	default:
		msg := fmt.Sprintf("unimplemented binary operand, %s", token.LookUp(n.Op))
		panic(msg)
	}

}

func (b *builder) evaluateArithmeticAddExpression(lhs, rhs lir.Value) lir.Value {
	typ := lhs.Yields()

	if types.IsInteger(typ) {
		return &lir.Add{
			Left:  lhs,
			Right: rhs,
		}
	}

	if types.IsFloatingPoint(typ) {
		return &lir.FAdd{
			Left:  lhs,
			Right: rhs,
		}
	}

	panic("todo: implement operand calls")
}

func (b *builder) evaluateArithmeticSubExpression(lhs, rhs lir.Value) lir.Value {
	typ := lhs.Yields()

	if types.IsInteger(typ) {
		return &lir.Sub{
			Left:  lhs,
			Right: rhs,
		}
	}

	if types.IsFloatingPoint(typ) {
		return &lir.FSub{
			Left:  lhs,
			Right: rhs,
		}
	}

	panic("todo: implement operand calls")
}

func (b *builder) evaluateArithmeticMulExpression(lhs, rhs lir.Value) lir.Value {
	typ := lhs.Yields()

	if types.IsInteger(typ) {
		return &lir.Mul{
			Left:  lhs,
			Right: rhs,
		}
	}

	if types.IsFloatingPoint(typ) {
		return &lir.FMul{
			Left:  lhs,
			Right: rhs,
		}
	}

	panic("todo: implement operand calls")
}

func (b *builder) evaluateArithmeticDivExpression(lhs, rhs lir.Value) lir.Value {
	typ := lhs.Yields()

	if types.IsInteger(typ) {
		if types.IsUnsigned(typ) {
			return &lir.UDiv{
				Left:  lhs,
				Right: rhs,
			}
		}
		return &lir.SDiv{
			Left:  lhs,
			Right: rhs,
		}
	}

	if types.IsFloatingPoint(typ) {
		return &lir.FDiv{
			Left:  lhs,
			Right: rhs,
		}
	}

	panic("todo: implement operand calls")
}

func (b *builder) evaluateArithmeticRemExpression(lhs, rhs lir.Value) lir.Value {
	typ := lhs.Yields()

	if types.IsInteger(typ) {
		if types.IsUnsigned(typ) {
			return &lir.URem{
				Left:  lhs,
				Right: rhs,
			}
		}
		return &lir.SRem{
			Left:  lhs,
			Right: rhs,
		}
	}

	if types.IsFloatingPoint(typ) {
		return &lir.FRem{
			Left:  lhs,
			Right: rhs,
		}
	}

	panic("todo: implement operand calls")
}

func (b *builder) evaluateArithmeticComparison(op token.Token, lhs, rhs lir.Value) lir.Value {
	typ := lhs.Yields()

	if types.IsInteger(typ) {
		comp := lir.INVALID_ICOMP

		if types.IsUnsigned(typ) {
			comp = lir.UOpMap[op]
		} else {
			comp = lir.SOpMap[op]
		}

		if comp == lir.INVALID_ICOMP {
			panic("invalid comparison operand")
		}

		instr := &lir.ICmp{
			Left:       lhs,
			Right:      rhs,
			Comparison: comp,
		}

		return instr

	}

	if types.IsFloatingPoint(typ) {
		panic("todo: floating point comparisons")
	}

	panic("todo: implement operand calls")

}

func (b *builder) evaluateArithmeticNegate(rhs lir.Value) lir.Value {

	typ := rhs.Yields()

	if types.IsInteger(typ) {
		return &lir.INeg{
			Right: rhs,
		}
	}

	if types.IsFloatingPoint(typ) {
		return &lir.FNeg{
			Right: rhs,
		}
	}

	panic("negate on unsupported type")
}

func (b *builder) evaluateLogicalNot(rhs lir.Value) lir.Value {
	typ := rhs.Yields()

	if types.IsInteger(typ) {
		return &lir.ICmp{
			Left:       rhs,
			Comparison: lir.EQL,
			Right:      lir.NewConst(0, rhs.Yields()),
		}
	}

	if typ == types.LookUp(types.Bool) {
		return &lir.XOR{
			Left:  rhs,
			Right: lir.NewConst(true, typ),
		}
	}

	panic(fmt.Sprintf("unimplemented logical not %s", typ))
}
