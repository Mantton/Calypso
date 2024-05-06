package lirgen

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/lir"
	"github.com/mantton/calypso/internal/calypso/token"
	"github.com/mantton/calypso/internal/calypso/types"
)

func (b *builder) evaluateExpression(n ast.Expression, fn *lir.Function, mod *lir.Module) lir.Value {
	fmt.Printf(
		"\tVisiting Expression: %T @ Line %d\n",
		n,
		n.Range().Start.Line,
	)

	switch e := n.(type) {
	case *ast.BooleanLiteral:
		return lir.NewConst(e.Value, types.LookUp(types.Bool))
	case *ast.StringLiteral:
		// TODO: Global Composites
		panic("string literals")
	case *ast.CharLiteral:
		return lir.NewConst(e.Value, types.LookUp(types.Char))
	case *ast.IntegerLiteral:
		typ := b.Mod.TModule.Table.GetNodeType(n)
		if typ == nil {
			typ = types.LookUp(types.Int)
		}
		return lir.NewConst(e.Value, typ)
	case *ast.FloatLiteral:
		typ := b.Mod.TModule.Table.GetNodeType(n)

		if typ == nil {
			typ = types.LookUp(types.Float)
		}

		return lir.NewConst(e.Value, typ)
	case *ast.NilLiteral:
		typ := b.Mod.TModule.Table.GetNodeType(n)
		if typ == nil {
			panic("unknown nullptr type")
		}
		return lir.NewConst(0, typ)
	case *ast.VoidLiteral:
		return lir.NewConst(0, types.LookUp(types.Void))
	case *ast.IdentifierExpression:
		return b.evaluateIdentifierExpression(e, fn, mod)
	case *ast.CallExpression:
		return b.evaluateCallExpression(e, fn, mod)
	case *ast.CallArgument:
		return b.evaluateExpression(e.Value, fn, mod)
	case *ast.AssignmentExpression:
		return b.evaluateAssignmentExpression(e, fn, mod)
	case *ast.BinaryExpression:
		return b.evaluateBinaryExpression(e, fn, mod)
	case *ast.GroupedExpression:
		return b.evaluateExpression(e.Expr, fn, mod)
	case *ast.UnaryExpression:
		return b.evaluateUnaryExpression(e, fn, mod)
	case *ast.ShorthandAssignmentExpression:
		return b.evaluateShortHandExpression(e, fn, mod)
	case *ast.CompositeLiteral:
		return b.evaluateCompositeLiteral(e, fn, mod)
	case *ast.FieldAccessExpression:
		return b.evaluateFieldAccessExpression(e, fn, mod, true)
	case *ast.SpecializationExpression:
		return b.evaluateSpecializationExpression(e, fn, mod)
	default:
		msg := fmt.Sprintf("unknown expr %T\n", e)
		panic(msg)
	}
}

func (b *builder) evaluateCallExpression(n *ast.CallExpression, fn *lir.Function, mod *lir.Module) lir.Value {

	val := b.evaluateExpression(n.Target, fn, mod)

	if val == nil {
		panic(fmt.Sprintf("unable to locate target function for: %s", n.Target))
	}

	var target *lir.Function
	var args []lir.Value

	switch val := val.(type) {
	case *lir.GenericFunction:
		// Find Target To Use
		X := b.Mod.TModule.Table.GetNodeType(n).(*types.SpecializedFunctionSignature)
		if types.IsGeneric(X) {
			// Target Is Generic, Specialize with Function Spec
			ssg := types.Instantiate(X, fn.Spec.Specialization()).(*types.SpecializedFunctionSignature)
			target = val.Specs[ssg.SymbolName()]
		} else {
			// Target is non generic, find
			target = val.Specs[X.SymbolName()]
		}

	case *lir.Function:
		target = val

	case *lir.UnionTypeInlineCreation:
		var args []lir.Value

		for _, p := range n.Arguments {
			v := b.evaluateExpression(p, fn, mod)
			args = append(args, v)
		}
		X := b.Mod.TModule.Table.GetNodeType(n)

		var ret types.Type
		switch X := X.(type) {
		case *types.SpecializedFunctionSignature:
			ret = X.Sg().Result.Type()
		case *types.FunctionSignature:
			ret = X
		}
		return b.emitUnionVariant(val, fn, args, ret)
	case *lir.Method:
		target = val.Fn
		args = append(args, val.Self)

	default:
		panic(fmt.Sprintf("unhandled call expression, %T", val))
	}

	if target == nil {
		panic("target function is nil")
	}

	// Add Call Graph Edge
	g := b.MP.CallGraph
	e := g.NewEdge(fn, target)
	g.SetEdge(e)

	for _, p := range n.Arguments {
		v := b.evaluateExpression(p, fn, mod)
		args = append(args, v)
	}

	i := &lir.Call{
		Target:    target,
		Arguments: args,
	}

	fn.Emit(i)
	return i
}

func (b *builder) evaluateIdentifierExpression(n *ast.IdentifierExpression, fn *lir.Function, mod *lir.Module) lir.Value {
	// Scoped Variable
	val, ok := fn.Variables[n.Value]

	if ok {
		switch val := val.(type) {
		case *lir.Allocate:
			// Return Pointer to Composites, dereference the rest
			typ := val.TypeOf.Parent()
			if types.IsStruct(typ) || types.IsUnionEnum(typ) {
				return val
			}

			i := &lir.Load{
				Address: val,
			}

			fn.Emit(i)
			return i
		case *lir.Constant:
			return val
		case *lir.Parameter:

			if !types.IsPointer(val.Yields()) {
				return val
			}

			i := &lir.Load{
				Address: val,
			}

			fn.Emit(i)
			return i
		case *lir.Load:
			return val
		case *lir.ExtractValue:
			return val
		default:
			panic(fmt.Sprintf("identifier found invalid type: %T", val))
		}
	}

	// Global Constant
	cons, ok := mod.GlobalConstants[n.Value]
	if ok {
		return cons
	}

	val = mod.Find(n.Value)

	if val != nil {
		return val
	}

	panic(fmt.Sprintf("unable to locate identifier, %s", n.Value))

}

func (b *builder) evaluateAssignmentExpression(n *ast.AssignmentExpression, fn *lir.Function, mod *lir.Module) lir.Value {
	a := b.evaluateAddressOfExpression(n.Target, fn, mod)
	v := b.evaluateExpression(n.Value, fn, mod)
	b.emitStore(fn, a, v)
	return nil
}

func (b *builder) evaluateUnaryExpression(n *ast.UnaryExpression, fn *lir.Function, mod *lir.Module) lir.Value {

	switch n.Op {
	case token.STAR:
		panic("todo: dereference")
	case token.AMP:
		panic("todo: get pointer Reference")
	case token.NOT:
		return b.evaluateLogicalNot(n, fn, mod)
	case token.MINUS:
		return b.evaluateArithmeticNegate(n, fn, mod)
	default:
		msg := fmt.Sprintf("unimplemented unary operand, %s", token.LookUp(n.Op))
		panic(msg)
	}
}

func (b *builder) evaluateBinaryExpression(n *ast.BinaryExpression, fn *lir.Function, mod *lir.Module) lir.Value {
	switch n.Op {
	case token.PLUS:
		return b.evaluateArithmeticAddExpression(n, fn, mod)
	case token.MINUS:
		return b.evaluateArithmeticSubExpression(n, fn, mod)
	case token.QUO:
		return b.evaluateArithmeticDivExpression(n, fn, mod)
	case token.STAR:
		return b.evaluateArithmeticMulExpression(n, fn, mod)
	case token.PCT:
		return b.evaluateArithmeticRemExpression(n, fn, mod)
	case token.L_CHEVRON, token.R_CHEVRON, token.EQL, token.LEQ, token.GEQ, token.NEQ:
		return b.evaluateArithmeticComparison(n.Op, n, fn, mod)
	case token.BIT_SHIFT_LEFT, token.BIT_SHIFT_RIGHT, token.AMP, token.BAR, token.CARET:
		return b.evaluateBitOperation(n.Op, n, fn, mod)
	case token.DOUBLE_AMP, token.DOUBLE_BAR:
		return b.evaluateBooleanOp(n.Op, n, fn, mod)
	default:
		msg := fmt.Sprintf("unimplemented binary operand, %s", token.LookUp(n.Op))
		panic(msg)
	}

}

func (b *builder) evaluateArithmeticAddExpression(n *ast.BinaryExpression, fn *lir.Function, mod *lir.Module) lir.Value {
	lhs, rhs := b.evaluateExpression(n.Left, fn, mod), b.evaluateExpression(n.Right, fn, mod)
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

	msg := fmt.Sprintf("TODO: Operand Calls, %s", typ)
	panic(msg)
}

func (b *builder) evaluateArithmeticSubExpression(n *ast.BinaryExpression, fn *lir.Function, mod *lir.Module) lir.Value {
	lhs, rhs := b.evaluateExpression(n.Left, fn, mod), b.evaluateExpression(n.Right, fn, mod)

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

	msg := fmt.Sprintf("TODO: Operand Calls, %s", typ)
	panic(msg)
}

func (b *builder) evaluateArithmeticMulExpression(n *ast.BinaryExpression, fn *lir.Function, mod *lir.Module) lir.Value {
	lhs, rhs := b.evaluateExpression(n.Left, fn, mod), b.evaluateExpression(n.Right, fn, mod)

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

	msg := fmt.Sprintf("TODO: Operand Calls, %s", typ)
	panic(msg)
}

func (b *builder) evaluateArithmeticDivExpression(n *ast.BinaryExpression, fn *lir.Function, mod *lir.Module) lir.Value {
	lhs, rhs := b.evaluateExpression(n.Left, fn, mod), b.evaluateExpression(n.Right, fn, mod)

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

func (b *builder) evaluateArithmeticRemExpression(n *ast.BinaryExpression, fn *lir.Function, mod *lir.Module) lir.Value {
	lhs, rhs := b.evaluateExpression(n.Left, fn, mod), b.evaluateExpression(n.Right, fn, mod)

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

func (b *builder) evaluateArithmeticComparison(op token.Token, n *ast.BinaryExpression, fn *lir.Function, mod *lir.Module) lir.Value {
	lhs, rhs := b.evaluateExpression(n.Left, fn, mod), b.evaluateExpression(n.Right, fn, mod)

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

	if types.IsPointer(typ) {

		instr := &lir.ICmp{
			Left:       lhs,
			Right:      rhs,
			Comparison: lir.EQL,
		}

		return instr
	}

	panic("todo: implement operand calls")

}

func (b *builder) evaluateArithmeticNegate(n *ast.UnaryExpression, fn *lir.Function, mod *lir.Module) lir.Value {
	rhs := b.evaluateExpression(n.Expr, fn, mod)
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

	panic(fmt.Sprintf("negate: unsupported type, %s", typ))
}

func (b *builder) evaluateLogicalNot(n *ast.UnaryExpression, fn *lir.Function, mod *lir.Module) lir.Value {
	rhs := b.evaluateExpression(n.Expr, fn, mod)

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

func (b *builder) evaluateBitOperation(op token.Token, n *ast.BinaryExpression, fn *lir.Function, mod *lir.Module) lir.Value {
	lhs, rhs := b.evaluateExpression(n.Left, fn, mod), b.evaluateExpression(n.Right, fn, mod)

	typ := lhs.Yields()

	if !types.IsInteger(typ) {
		panic(fmt.Sprintf("unsupported type, %s, operand: %s", typ, op))
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
func (b *builder) evaluateBooleanOp(op token.Token, n *ast.BinaryExpression, fn *lir.Function, mod *lir.Module) lir.Value {
	lhs := b.evaluateExpression(n.Left, fn, mod)
	typ := lhs.Yields()

	if !types.IsBoolean(typ) {
		panic(fmt.Sprintf("unsupported type, %s, operand: %s", typ, op))
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
	rhs := b.evaluateExpression(n.Right, fn, mod)
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

func (b *builder) evaluateAddressOfExpression(n ast.Expression, fn *lir.Function, mod *lir.Module) lir.Value {
	switch n := n.(type) {
	case *ast.IdentifierExpression:

		global, ok := mod.GlobalConstants[n.Value]

		if ok {
			return global
		}

		val, ok := fn.Variables[n.Value]

		if ok {
			return val
		}

		if mod, ok := mod.Imports[n.Value]; ok {
			return mod
		}

		val = mod.Find(n.Value)

		if val != nil {
			return val
		}

		panic(fmt.Sprintf("unknown identifier, %s", n.Value))
	case *ast.FieldAccessExpression:
		x := b.evaluateFieldAccessExpression(n, fn, mod, false)
		return x
	default:
		panic("unimplmented address of")
	}
}

func (b *builder) evaluateShortHandExpression(n *ast.ShorthandAssignmentExpression, fn *lir.Function, mod *lir.Module) lir.Value {
	var rhs lir.Value
	addr := b.evaluateAddressOfExpression(n.Target, fn, mod)

	switch n.Op {
	case token.PLUS_EQ:
		rhs = b.evaluateArithmeticAddExpression(&ast.BinaryExpression{
			Left:  n.Target,
			Right: n.Right,
		}, fn, mod)
	case token.MINUS_EQ:
		rhs = b.evaluateArithmeticSubExpression(&ast.BinaryExpression{
			Left:  n.Target,
			Right: n.Right,
		}, fn, mod)
	case token.QUO_EQ:
		rhs = b.evaluateArithmeticDivExpression(&ast.BinaryExpression{
			Left:  n.Target,
			Right: n.Right,
		}, fn, mod)
	case token.STAR_EQ:
		rhs = b.evaluateArithmeticMulExpression(&ast.BinaryExpression{
			Left:  n.Target,
			Right: n.Right,
		}, fn, mod)
	case token.PCT_EQ:
		rhs = b.evaluateArithmeticRemExpression(&ast.BinaryExpression{
			Left:  n.Target,
			Right: n.Right,
		}, fn, mod)
	case token.AMP_EQ:
		rhs = b.evaluateBitOperation(token.AMP, &ast.BinaryExpression{
			Left:  n.Target,
			Right: n.Right,
		}, fn, mod)
	case token.BAR_EQ:
		rhs = b.evaluateBitOperation(token.BAR, &ast.BinaryExpression{
			Left:  n.Target,
			Right: n.Right,
		}, fn, mod)
	case token.CARET_EQ:
		rhs = b.evaluateBitOperation(token.CARET, &ast.BinaryExpression{
			Left:  n.Target,
			Right: n.Right,
		}, fn, mod)
	case token.BIT_SHIFT_LEFT_EQ:
		rhs = b.evaluateBitOperation(token.BIT_SHIFT_LEFT, &ast.BinaryExpression{
			Left:  n.Target,
			Right: n.Right,
		}, fn, mod)
	case token.BIT_SHIFT_RIGHT_EQ:
		rhs = b.evaluateBitOperation(token.BIT_SHIFT_RIGHT, &ast.BinaryExpression{
			Left:  n.Target,
			Right: n.Right,
		}, fn, mod)

	default:
		panic("unimplemented shorthand expression")
	}

	b.emitStore(fn, addr, rhs)
	return lir.NewConst(nil, types.LookUp(types.Void))
}

func (b *builder) evaluateCompositeLiteral(n *ast.CompositeLiteral, fn *lir.Function, mod *lir.Module) lir.Value {
	val := b.evaluateExpression(n.Target, fn, mod)

	var composite *lir.Composite

	switch val := val.(type) {
	case *lir.Composite:
		composite = val
	case *lir.GenericType:
		A := b.Mod.TModule.Table.GetNodeType(n).(*types.SpecializedType)
		fmt.Println(A)
		composite = val.Specs[A.SymbolName()]
	}

	if composite == nil {
		panic("should be composite")
	}

	// ! Mandatory
	b.Mod.Composites[composite.Name] = composite
	b.MP.Composites[composite.Type] = composite

	addr := b.emitHeapAlloc(fn, composite.Yields())

	for _, field := range n.Body.Fields {
		index := types.GetFieldIndex(field.Key.Value, composite.Yields())
		value := b.evaluateExpression(field.Value, fn, mod)
		// Get Pointer to Property

		prop_ptr := &lir.GEP{
			Index:     index,
			Address:   addr,
			Composite: composite,
		}

		fn.Emit(prop_ptr)

		// Store
		store := &lir.Store{
			Address: prop_ptr,
			Value:   value,
		}

		fn.Emit(store)
	}

	return addr
}

func (b *builder) evaluateFieldAccessExpression(n *ast.FieldAccessExpression, fn *lir.Function, mod *lir.Module, load bool) lir.Value {
	// 1 - Evaluate Address or Type Reference
	target := b.evaluateAddressOfExpression(n.Target, fn, mod)

	var field string
	switch p := n.Field.(type) {
	case *ast.IdentifierExpression:
		field = p.Value
	default:
		if target, ok := target.(*lir.Module); ok {
			tgt := b.evaluateExpression(p, fn, target)
			return tgt
		}
	}

	// Handle Module
	switch target := (target).(type) {
	case *lir.Module:
		tgt := target.Find(field)
		if tgt == nil {
			panic(fmt.Sprintf("cannot find, %s", field))
		}
		return tgt
	}

	var targetType types.Type
	var base types.Type

	switch target := target.(type) {
	case *lir.GenericEnumReference:
		X := target.Type.Parent().(*types.Enum).FindVariant(field)

		// Unit
		if len(X.Fields) == 0 {
			return lir.NewConst(int64(X.Discriminant), types.LookUp(types.Int32))
		}

		return &lir.UnionTypeInlineCreation{
			Type:    target.Type,
			Variant: X,
		}
	case *lir.EnumReference:
		X := target.Type.Parent().(*types.Enum).FindVariant(field)
		// Unit
		if len(X.Fields) == 0 {
			return lir.NewConst(int64(X.Discriminant), types.LookUp(types.Int32))
		}
		return &lir.UnionTypeInlineCreation{
			Type:    target.Type.(types.Symbol),
			Variant: X,
		}
	default:
		base = target.Yields()
		targetType = base

		// Deref
		if types.IsPointer(targetType) {
			targetType = types.Dereference(targetType)
		}

		// Specialize
		if fn.Spec != nil {
			targetType = types.Instantiate(targetType, fn.Spec.Spec)
		}
		fmt.Println("\t\tAccessing", field, "on", targetType, fmt.Sprintf("%T", target))
	}

	symbol, symbolType := types.ResolveSymbol(targetType, field)
	parent := targetType.Parent()

	switch parent := parent.(type) {
	case *types.Struct:
		switch symbol := symbol.(type) {
		case *types.Var:
			// Composite Type
			composite := b.resolveCompositeOf(targetType, mod)

			if composite == nil {
				panic("unknown composite")
			}
			index := symbol.StructIndex

			// Invalid Field
			if index == -1 {
				panic("unknown field")
			}

			// Get Element Pointer of Field
			ptr := &lir.GEP{
				Index:     index,
				Address:   target,
				Composite: composite,
			}

			// If Should Load, return load instruction
			if load {
				return &lir.Load{
					Address: ptr,
				}
			}

			// return GEP instruction, yeilding ptr to field
			return ptr
		case *types.Function:
			tgt, ok := b.MP.Functions[symbolType]

			if !ok {
				panic("unable to locate function for symbol")
			}

			// Is Method Access
			if tgt.TFunction.Self != nil {
				return &lir.Method{
					Fn:   tgt,
					Self: target,
				}
			}

			return tgt
		default:
			panic("unhandled symbol type")
		}
	default:
		panic(fmt.Sprintf("unhandled symbol type, %s", parent))
	}
}

func (b *builder) evaluateSwitchConditionExpression(n ast.Expression, fn *lir.Function, mod *lir.Module) (lir.Value, *ast.CallExpression, *types.EnumVariant) {

	var target lir.Value
	var expr *ast.CallExpression
	switch n := n.(type) {
	case *ast.CallExpression:
		expr = n
		target = b.evaluateExpression(n.Target, fn, mod)
	case *ast.FieldAccessExpression:
		target = b.evaluateExpression(n, fn, mod)

	default:
		return b.evaluateExpression(n, fn, mod), nil, nil
	}

	f, ok := target.(*lir.Function)

	if !ok {
		return b.evaluateExpression(n, fn, mod), nil, nil
	}

	en := b.RFunctionEnums[f]

	if en == nil {

		return b.evaluateExpression(n, fn, mod), nil, nil
	}

	// Composite Enum
	dis := lir.NewConst(int64(en.Discriminant), types.LookUp(types.Int8))

	// for each argument create name:value
	return dis, expr, en
}

func (b *builder) evaluateEnumVariantTuple(fn *lir.Function, n *ast.CallExpression, v *types.EnumVariant, s types.Symbol, self lir.Value, mod *lir.Module) {
	composite := mod.Composites[EnumVariantSymbolName(v, s)]

	// 0 Index is Discriminant
	x := 1

	// If aligned 1 index is padding
	if composite.IsAligned {
		x += 1
	}

	// iterate through arguments
	for i, arg := range n.Arguments {
		idx := i + x
		ident, ok := arg.Value.(*ast.IdentifierExpression)

		if !ok {
			panic("expected identifier")
		}

		addr := &lir.GEP{
			Address:   self,
			Index:     idx,
			Composite: composite,
		}

		fn.Emit(addr)

		load := &lir.Load{
			Address: addr,
		}

		fn.Emit(load)
		fn.Variables[ident.Value] = load
	}
}

func (b *builder) evaluateSpecializationExpression(expr *ast.SpecializationExpression, fn *lir.Function, mod *lir.Module) lir.Value {
	A := b.evaluateExpression(expr.Expression, fn, mod)
	B := b.Mod.TModule.Table.GetNodeType(expr)
	var C lir.Value
	switch A := A.(type) {
	case *lir.GenericType:
		symbol := B.(*types.SpecializedType).SymbolName()
		C = A.Specs[symbol]
	case *lir.GenericFunction:
		// TODO:
		symbol := B.(*types.SpecializedFunctionSignature).SymbolName()
		C = A.Specs[symbol]
	}
	return C
}

func (b *builder) resolveCompositeOf(t types.Type, mod *lir.Module) *lir.Composite {
	switch t := t.(type) {
	case *types.Pointer:
		return b.resolveCompositeOf(t.PointerTo, mod)
	case *types.SpecializedType:
		c := mod.Composites[t.SymbolName()]

		if c != nil {
			return c
		}
	case *types.DefinedType:
		c := mod.Composites[t.SymbolName()]
		if c != nil {
			return c
		}
	}

	if x, ok := b.MP.Composites[t]; ok {
		mod.Composites[x.Name] = x
		return x
	}

	panic(fmt.Sprintf("unhandled type, %s", t))
}

func (b *builder) emitUnionVariant(n *lir.UnionTypeInlineCreation, fn *lir.Function, args []lir.Value, ret types.Type) lir.Value {
	composite := b.MP.Composites[n.Variant]

	// Allocate Base Type
	addr := &lir.Allocate{
		TypeOf: ret,
	}

	fn.Emit(addr)

	// GEP & Store of Fields
	for i := range n.Variant.Fields {

		ptr := &lir.GEP{
			Index:     i + 1, // First Position is always dicriminant
			Address:   addr,
			Composite: composite,
		}

		store := &lir.Store{
			Address: ptr,
			Value:   args[i],
		}

		fn.Emit(store)
	}

	// emit store of disciminant
	fn.Emit(&lir.Store{
		Address: addr,
		Value:   lir.NewConst(int64(n.Variant.Discriminant), types.LookUp(types.Int8)),
	})

	// emit return of pointer to foo
	fn.Emit(&lir.Return{
		Result: addr,
	})

	return fn
}
