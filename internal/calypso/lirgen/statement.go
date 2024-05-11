package lirgen

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/lir"
	"github.com/mantton/calypso/internal/calypso/types"
)

func (b *builder) visitStatement(node ast.Statement, fn *lir.Function) {

	if fn.CurrentBlock.Complete {
		fmt.Println("Statements after return statement are not executed")
		return
	}

	fmt.Printf(
		"Visiting Statement: %T @ Line %d\n",
		node,
		node.Range().Start.Line,
	)

	switch node := node.(type) {

	case *ast.VariableStatement:
		b.visitVariableStatement(node, fn)
	case *ast.ReturnStatement:
		b.visitReturnStatement(node, fn)
	case *ast.ExpressionStatement:
		b.visitExpressionStatement(node, fn)
	case *ast.BlockStatement:
		b.visitBlockStatement(node, fn)
	case *ast.IfStatement:
		b.visitIfStatement(node, fn)
	case *ast.SwitchStatement:
		b.visitSwitchStatement(node, fn)
	case *ast.WhileStatement:
		b.visitWhileStatement(node, fn)
	case *ast.BreakStatement, *ast.StructStatement, *ast.EnumStatement, *ast.TypeStatement:
		break
	case *ast.DereferenceAssignmentStatement:
		b.visitDerefAssignmentStatement(node, fn)
	default:
		msg := fmt.Sprintf("statement check not implemented, %T\n", node)
		panic(msg)
	}
}

func (b *builder) visitVariableStatement(n *ast.VariableStatement, fn *lir.Function) {
	val := b.evaluateExpression(n.Value, fn, b.Mod)

	if n.IsConstant {
		v, ok := val.(*lir.Constant)

		// constant is a compile time constant
		if ok {
			b.emitConstantVar(fn, v, n.Identifier.Value)
			return
		}
	}

	vAddr, _ := val.(*lir.Allocate)
	addr := b.emitLocalVar(fn, n.Identifier.Value, val.Yields(), vAddr)

	if vAddr == nil {
		b.emitStore(fn, addr, val)
	}
}

func (b *builder) visitReturnStatement(n *ast.ReturnStatement, fn *lir.Function) {
	val := b.evaluateExpression(n.Value, fn, b.Mod)

	if val.Yields() == types.LookUp(types.Void) {
		fn.Emit(&lir.ReturnVoid{})
	} else {
		i := &lir.Return{
			Result: val,
		}
		fn.Emit(i)
	}
	fn.CurrentBlock.Complete = true
}

func (b *builder) visitExpressionStatement(n *ast.ExpressionStatement, fn *lir.Function) {
	i, ok := b.evaluateExpression(n.Expr, fn, b.Mod).(lir.Instruction)

	if !ok {
		return
	}

	fn.Emit(i)
}

func (b *builder) visitBlockStatement(n *ast.BlockStatement, fn *lir.Function) {
	for _, s := range n.Statements {
		b.visitStatement(s, fn)
	}
}

func (b *builder) visitIfStatement(n *ast.IfStatement, fn *lir.Function) {

	cond := b.evaluateExpression(n.Condition, fn, b.Mod)

	br := &lir.ConditionalBranch{
		Condition: cond,
	}
	fn.Emit(br)

	// Generate Blocks
	then := fn.NewBlock()
	var elseBlock *lir.Block
	if n.Alternative != nil {
		elseBlock = fn.NewBlock()
	}
	done := fn.NewBlock()

	br.Action = then
	if elseBlock != nil {
		br.Alternative = elseBlock
	} else {
		br.Alternative = then
	}

	// Action
	fn.CurrentBlock = then
	b.visitBlockStatement(n.Action, fn)
	fn.Emit(&lir.Branch{
		Block: done,
	})

	// Alternative
	if n.Alternative != nil {
		fn.CurrentBlock = elseBlock
		b.visitBlockStatement(n.Alternative, fn)
		fn.Emit(&lir.Branch{
			Block: done,
		})
	}

	fn.CurrentBlock = done
}

func (b *builder) visitWhileStatement(n *ast.WhileStatement, fn *lir.Function) {

	// Setup Blocks
	loop := fn.NewBlock() // Checks the Condition
	body := fn.NewBlock() // Body of While loop
	done := fn.NewBlock() // Exit of while loop

	// Emit Condition
	fn.CurrentBlock = loop
	cond := b.evaluateExpression(n.Condition, fn, b.Mod)
	fn.Emit(&lir.ConditionalBranch{
		Condition:   cond,
		Action:      body,
		Alternative: done,
	})

	// Emit Body
	fn.CurrentBlock = body
	b.visitBlockStatement(n.Action, fn)
	fn.Emit(&lir.Branch{
		Block: loop,
	})

	fn.CurrentBlock = done
}

func (b *builder) visitSwitchStatement(n *ast.SwitchStatement, fn *lir.Function) {
	cond := b.evaluateExpression(n.Condition, fn, b.Mod)

	var T lir.Value

	symT := cond.Yields()

	if types.IsPointer(symT) {
		symT = types.Dereference(symT)
	}

	symbol := symT
	typ := cond.Yields().Parent()

	// Is Composite Enum
	if types.IsPointer(typ) {
		if en, ok := types.Dereference(typ).Parent().(*types.Enum); ok && en.IsUnion() {
			addr := &lir.AccessStructProperty{
				Address:   cond,
				Index:     0,
				Composite: b.MP.Composites[symbol],
			}

			T = &lir.Load{
				Address: addr,
			}

		} else {
			T = cond
		}

	} else {
		T = cond
	}

	instr := &lir.Switch{
		Value: T,
	}
	fn.Emit(instr)

	var defaultCase *ast.SwitchCaseExpression

	for _, cs := range n.Cases {
		if cs.IsDefault {
			defaultCase = cs
			continue
		}
		pair := &lir.SwitchValueBlock{}

		block := fn.NewBlock()
		pair.Block = block
		res := b.evaluateSwitchCaseCondition(cs, fn, cond)

		switch res := res.(type) {
		case *lir.EnumExpansionResult:
			pair.Value = res.Discriminant
			res.Emit(fn, cond)
		default:
			pair.Value = res
		}

		b.visitBlockStatement(cs.Action, fn)

		pair.EndBlock = fn.CurrentBlock

		instr.Blocks = append(instr.Blocks, pair)
	}

	// Declare Final BLock of Default Path & the "Done" Block after the switch
	var done *lir.Block
	var defBlock *lir.Block

	// if the default case not nil, set the done block to the default path, note that the "Done" & Switch Default path are mixed here
	if defaultCase != nil {
		instr.Done = fn.NewBlock()
		b.visitBlockStatement(defaultCase.Action, fn)
		defBlock = fn.CurrentBlock
	}

	done = fn.NewBlock()

	// No Default Case, set default to next block after switch statement, our "done" block
	if instr.Done == nil {
		instr.Done = done
	}

	// Loop through previous ending blocks and emit a branch to the done block if needed
	for _, pair := range instr.Blocks {
		if pair.EndBlock.Complete {
			continue
		}

		pair.EndBlock.Emit(&lir.Branch{
			Block: done,
		})
	}

	// if the default endblock is not nil & is not complete, alos emit a branch to the done block
	if defBlock != nil && !defBlock.Complete {
		defBlock.Emit(&lir.Branch{
			Block: done,
		})
	}

}

func (b *builder) visitDerefAssignmentStatement(n *ast.DereferenceAssignmentStatement, fn *lir.Function) {

	addr := b.evaluateAddressOfExpression(n.Target, fn, b.Mod)

	val := b.evaluateExpression(n.Value, fn, b.Mod)

	str := &lir.Store{
		Address: addr,
		Value:   val,
	}

	fn.Emit(str)
}
