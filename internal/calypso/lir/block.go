package lir

type Block struct {
	Index        int
	Instructions []Instruction
	Parent       *Function
	Complete     bool
}

func (b *Block) Emit(i Instruction) {

	if b.Complete {
		return
	}

	b.Instructions = append(b.Instructions, i)

	if _, ok := i.(*Return); ok {
		b.Complete = true
	}
}
