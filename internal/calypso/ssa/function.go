package ssa

import "github.com/mantton/calypso/internal/calypso/symbols"

type Function struct {
	Symbol    *symbols.SymbolInfo
	Blocks    []*Block
	Locals    []*Allocate
	Variables map[string]Value

	Owner *Member

	CurrentBlock *Block
}

func (f *Function) ssaMbr() {}

func (f *Function) Emit(i Instruction) {
	if f.CurrentBlock == nil {
		return
	}

	f.CurrentBlock.Emit(i)
}

func (f *Function) NewBlock() *Block {
	b := &Block{
		Parent: f,
	}

	f.Blocks = append(f.Blocks, b)
	b.Index = len(f.Blocks) - 1

	return b
}

func NewFunction(s *symbols.SymbolInfo) *Function {
	return &Function{
		Symbol:    s,
		Variables: make(map[string]Value),
	}
}
