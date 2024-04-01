package llir

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/lir"
)

func (b *builder) visitInstruction(i lir.Instruction) {
	switch i := i.(type) {
	case lir.Value:
		val := b.createValue(i)
		b.setValue(i, val)

	case *lir.Return:
		v := b.getValue(i.Result)
		b.CreateRet(v)

	case *lir.Store:
		v := b.getValue(i.Value)
		a := b.getValue(i.Address)
		b.CreateStore(v, a)
	case *lir.Branch:
		v := b.getValue(i.Condition)
		x := b.blocks[i.Action]
		y := b.blocks[i.Alternative]

		b.CreateCondBr(v, x, y)
	case *lir.Jump:
		x := b.blocks[i.Block]
		b.CreateBr(x)

	default:

		panic(fmt.Sprintf("TODO: NOT IMPLEMENTED: %T", i))
	}

}
