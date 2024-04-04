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
	Type         *types.Function
	Blocks       []*Block
	Locals       []*Allocate
	Variables    map[string]Value
	Parameters   []*Parameter
	Owner        *Member
	CurrentBlock *Block
	External     bool
}

func (f *Function) Signature() *types.FunctionSignature { return f.Type.Sg() }
func (f *Function) Name() string                        { return f.Type.Name() }
func (f *Function) Yields() types.Type                  { return f.Signature() }
func (f *Parameter) Yields() types.Type                 { return f.Symbol }

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

func (f *Function) AddSelf() {
	self := f.Signature().Self
	if self == nil || f.Signature().IsStatic {
		return
	}

	p := &Parameter{
		Name:   "self",
		Symbol: types.NewPointer(self.Type()),
		Parent: f,
	}
	f.Parameters = append(f.Parameters, p)
	f.Variables[p.Name] = p

}

func NewFunction(fn *types.Function) *Function {
	f := &Function{
		Variables: make(map[string]Value),
		Type:      fn,
	}
	f.NewBlock()
	return f
}
