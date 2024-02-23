package ssa

type Block struct {
	Index        int
	Instructions []Instruction
	Parent       *Function
}

func (b *Block) Emit(i Instruction) {
	b.Instructions = append(b.Instructions, i)
}
