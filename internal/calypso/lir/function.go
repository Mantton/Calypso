package lir

import (
	"github.com/mantton/calypso/internal/calypso/types"
)

type Parameter struct {
	Name   string
	Symbol types.Type
	Parent *Function
}

type Function struct {
	Symbol       *types.FunctionSignature
	Blocks       []*Block
	Locals       []*Allocate
	Variables    map[string]Value
	Parameters   []*Parameter
	Owner        *Member
	CurrentBlock *Block
}

func (f *Function) Yields() types.Type  { return f.Symbol }
func (f *Parameter) Yields() types.Type { return f.Symbol }

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

	f.CurrentBlock = b

	return b
}

func (f *Function) AddParameter(t *types.Var) {
	p := &Parameter{
		Name:   t.Name(),
		Symbol: t.Type(),
		Parent: f,
	}
	f.Parameters = append(f.Parameters, p)
	f.Variables[t.Name()] = p
}

func NewFunction(fn *types.Function) *Function {
	f := &Function{
		Symbol:    fn.Sg(),
		Variables: make(map[string]Value),
	}
	f.NewBlock()
	return f
}
