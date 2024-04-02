package llir

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/lir"
)

func (b *builder) visitInstruction(i lir.Instruction) {
	switch i := i.(type) {
	case lir.Value:
		b.visitYieldingInstruction(i)
	case *lir.Return:
		b.visitReturnInstruction(i)
	case *lir.ReturnVoid:
		b.visitReturnVoidInstruction(i)
	case *lir.Store:
		b.visitStoreInstruction(i)
	case *lir.ConditionalBranch:
		b.visitConditionalBranchInstruction(i)
	case *lir.Branch:
		b.visitBranchInstruction(i)
	default:
		msg := fmt.Sprintf("[LLIR] Instruction Not Implemented, %T", i)
		panic(msg)
	}
}

func (b *builder) visitYieldingInstruction(i lir.Value) {
	b.getValue(i)
}

func (b *builder) visitReturnInstruction(i *lir.Return) {
	v := b.getValue(i.Result)
	b.CreateRet(v)
}

func (b *builder) visitReturnVoidInstruction(*lir.ReturnVoid) {
	b.CreateRetVoid()
}

func (b *builder) visitStoreInstruction(i *lir.Store) {
	v := b.getValue(i.Value)
	a := b.getValue(i.Address)
	b.CreateStore(v, a)
}

func (b *builder) visitConditionalBranchInstruction(i *lir.ConditionalBranch) {
	v := b.getValue(i.Condition)
	x := b.blocks[i.Action]
	y := b.blocks[i.Alternative]

	b.CreateCondBr(v, x, y)
}

func (b *builder) visitBranchInstruction(i *lir.Branch) {
	x := b.blocks[i.Block]
	b.CreateBr(x)
}
