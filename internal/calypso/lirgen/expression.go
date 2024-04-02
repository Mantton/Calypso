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

	switch n.Op {
	case token.STAR:
		panic("todo: dereference")
	case token.AMP:
		panic("todo: get pointer Reference")
	case token.NOT:
		return b.evaluateLogicalNot(n, fn)
	case token.MINUS:
		return b.evaluateArithmeticNegate(n, fn)
	default:
		msg := fmt.Sprintf("unimplemented unary operand, %s", token.LookUp(n.Op))
		panic(msg)
	}
}

func (b *builder) evaluateBinaryExpression(n *ast.BinaryExpression, fn *lir.Function) lir.Value {
	switch n.Op {
	case token.PLUS:
		return b.evaluateArithmeticAddExpression(n, fn)
	case token.MINUS:
		return b.evaluateArithmeticSubExpression(n, fn)
	case token.QUO:
		return b.evaluateArithmeticDivExpression(n, fn)
	case token.STAR:
		return b.evaluateArithmeticMulExpression(n, fn)
	case token.PCT:
		return b.evaluateArithmeticRemExpression(n, fn)
	case token.L_CHEVRON, token.R_CHEVRON, token.EQL, token.LEQ, token.GEQ, token.NEQ:
		return b.evaluateArithmeticComparison(n.Op, n, fn)
	case token.BIT_SHIFT_LEFT, token.BIT_SHIFT_RIGHT, token.AMP, token.BAR, token.CARET:
		return b.evaluateBitOperation(n.Op, n, fn)
	case token.DOUBLE_AMP, token.DOUBLE_BAR:
		return b.evaluateBooleanOp(n.Op, n, fn)
	default:
		msg := fmt.Sprintf("unimplemented binary operand, %s", token.LookUp(n.Op))
		panic(msg)
	}

}

func (b *builder) evaluateArithmeticAddExpression(n *ast.BinaryExpression, fn *lir.Function) lir.Value {
	lhs, rhs := b.evaluateExpression(n.Left, fn), b.evaluateExpression(n.Right, fn)

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

func (b *builder) evaluateArithmeticSubExpression(n *ast.BinaryExpression, fn *lir.Function) lir.Value {
	lhs, rhs := b.evaluateExpression(n.Left, fn), b.evaluateExpression(n.Right, fn)

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

func (b *builder) evaluateArithmeticMulExpression(n *ast.BinaryExpression, fn *lir.Function) lir.Value {
	lhs, rhs := b.evaluateExpression(n.Left, fn), b.evaluateExpression(n.Right, fn)

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

func (b *builder) evaluateArithmeticDivExpression(n *ast.BinaryExpression, fn *lir.Function) lir.Value {
	lhs, rhs := b.evaluateExpression(n.Left, fn), b.evaluateExpression(n.Right, fn)

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

func (b *builder) evaluateArithmeticRemExpression(n *ast.BinaryExpression, fn *lir.Function) lir.Value {
	lhs, rhs := b.evaluateExpression(n.Left, fn), b.evaluateExpression(n.Right, fn)

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

func (b *builder) evaluateArithmeticComparison(op token.Token, n *ast.BinaryExpression, fn *lir.Function) lir.Value {
	lhs, rhs := b.evaluateExpression(n.Left, fn), b.evaluateExpression(n.Right, fn)

	typ := lhs.Yields()

	if types.IsInteger(typ) {
		comp := lir.INVALID_ICOMP

		if types.IsUnsigned(typ) {
			comp = lir.UOpMap[op]
		} else {
			comp = lir.SOpMap[op]
		}

		if comp == lir.INVALID_ICOMP {
			panic(fmt.Sprintf("invalid comparison operand, %s", op))
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

func (b *builder) evaluateArithmeticNegate(n *ast.UnaryExpression, fn *lir.Function) lir.Value {
	rhs := b.evaluateExpression(n.Expr, fn)

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

func (b *builder) evaluateLogicalNot(n *ast.UnaryExpression, fn *lir.Function) lir.Value {
	rhs := b.evaluateExpression(n.Expr, fn)

	typ := rhs.Yields()

	if types.IsInteger(typ) {
		return &lir.ICmp{
			Left:       rhs,
			Comparison: lir.EQL,
			Right:      lir.NewConst(0, rhs.Yields()),
		}
	}

	if types.IsBoolean(typ) {
		return &lir.XOR{
			Left:  rhs,
			Right: lir.NewConst(true, typ),
		}
	}

	panic(fmt.Sprintf("unimplemented logical not %s", typ))
}

func (b *builder) evaluateBitOperation(op token.Token, n *ast.BinaryExpression, fn *lir.Function) lir.Value {
	lhs, rhs := b.evaluateExpression(n.Left, fn), b.evaluateExpression(n.Right, fn)

	typ := lhs.Yields()

	if !types.IsInteger(typ) {
		panic("unsupported type")
	}

	switch op {
	case token.BIT_SHIFT_LEFT:
		return &lir.ShiftLeft{
			Left: lhs, Right: rhs,
		}
	case token.BIT_SHIFT_RIGHT:

		// For Unsigned use logical shift right
		if types.IsUnsigned(typ) {
			return &lir.LogicalShiftRight{
				Left: lhs, Right: rhs,
			}
		}

		// Else use arithmetic shift
		return &lir.ArithmeticShiftRight{
			Left: lhs, Right: rhs,
		}

	case token.AMP:
		return &lir.AND{
			Left: lhs, Right: rhs,
		}
	case token.BAR:
		return &lir.OR{
			Left: lhs, Right: rhs,
		}
	case token.CARET:
		return &lir.XOR{
			Left: lhs, Right: rhs,
		}
	default:
		panic(fmt.Sprintf("unimplemented bit operation %s", op))
	}
}

// Reference: https://en.wikipedia.org/wiki/Short-circuit_evaluation
func (b *builder) evaluateBooleanOp(op token.Token, n *ast.BinaryExpression, fn *lir.Function) lir.Value {
	lhs := b.evaluateExpression(n.Left, fn)
	typ := lhs.Yields()
	if !types.IsBoolean(typ) {
		panic("unsupported type")
	}

	prev := fn.CurrentBlock
	next := fn.NewBlock()
	done := fn.NewBlock()

	// 1 - Compare if LHS != false
	fn.CurrentBlock = prev
	cmp := &lir.ICmp{
		Left:       lhs,
		Right:      lir.NewConst(false, typ),
		Comparison: lir.NEQ,
	}

	// 2 - If LHS is not false, branch to resolve RHS else, branch to done block
	var br *lir.ConditionalBranch

	switch op {
	case token.DOUBLE_AMP:
		// && / AND Operation, only Branch to Next if condition LHS true
		br = &lir.ConditionalBranch{
			Condition:   cmp,
			Action:      next,
			Alternative: done,
		}

	case token.DOUBLE_BAR:
		// || / OR Operation, branch to done if LHS is true
		br = &lir.ConditionalBranch{
			Condition:   cmp,
			Action:      done,
			Alternative: next,
		}
	default:
		panic("unsupported operand")
	}

	fn.Emit(br)

	// 2 - Populate RHS Resolution Block
	// Create comparison checking if RHS != false
	fn.CurrentBlock = next
	rhs := b.evaluateExpression(n.Right, fn)
	cmp2 := &lir.ICmp{
		Left:       rhs,
		Right:      lir.NewConst(false, typ),
		Comparison: lir.NEQ,
	}

	br2 := &lir.Branch{
		Block: done,
	}

	fn.Emit(cmp2)
	fn.Emit(br2)

	// 3 - Done Block, Use Phi to pick value based on which block executed prior
	fn.CurrentBlock = done
	phi := &lir.PHI{
		Nodes: []*lir.PhiNode{
			{
				Value: lir.NewConst(false, typ),
				Block: prev,
			},
			{
				Value: cmp2,
				Block: next,
			},
		},
	}

	fn.Emit(phi)
	return phi
}
