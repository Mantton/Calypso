package llir

import (
	"github.com/mantton/calypso/internal/calypso/lir"
	"tinygo.org/x/go-llvm"
)

type Locals map[lir.Value]llvm.Value

type builder struct {
	*compiler
	llvm.Builder
	lirFn        *lir.Function
	llvmFn       llvm.Value
	llvmFnType   llvm.Type
	locals       Locals
	blocks       map[*lir.Block]llvm.BasicBlock
	currentBlock llvm.BasicBlock
}

func newBuilder(fn *lir.Function, c *compiler, b llvm.Builder) *builder {
	// Initialize Function
	f, fnType := c.getFunction(fn.Type)
	return &builder{
		compiler:   c,
		lirFn:      fn,
		Builder:    b,
		locals:     make(Locals),
		blocks:     make(map[*lir.Block]llvm.BasicBlock),
		llvmFn:     f,
		llvmFnType: fnType,
	}
}

func (b *builder) buildFunction() {
	// Create Blocks
	var entry llvm.BasicBlock
	for i, block := range b.lirFn.Blocks {
		llvmBlock := b.context.AddBasicBlock(b.llvmFn, "")
		b.blocks[block] = llvmBlock

		if i == 0 {
			entry = llvmBlock
		}
	}

	b.SetInsertPointAtEnd(entry)

	// load params
	for i, v := range b.lirFn.Parameters {
		b.setValue(v, b.llvmFn.Param(i))
	}

	// Fill Blocks
	for _, ptr := range b.lirFn.Blocks {
		blk := b.blocks[ptr]
		b.SetInsertPointAtEnd(blk)
		b.currentBlock = blk
		for _, i := range ptr.Instructions {
			b.visitInstruction(i)
		}
	}

}
