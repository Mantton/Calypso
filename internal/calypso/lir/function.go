package lir

import (
	"github.com/mantton/calypso/internal/calypso/types"
)

type Parameter struct {
	Name   string
	Symbol types.Type
	Parent *Function
	IsSelf bool
}

type Function struct {
	TFunction *types.Function
	Spec      *types.SpecializedFunctionSignature

	Blocks     []*Block
	Locals     []*Allocate
	Variables  map[string]Value
	Parameters []*Parameter

	Owner        *Member
	CurrentBlock *Block
	External     bool
	Name         string
}

func (f *Function) Signature() *types.FunctionSignature {
	if f.Spec != nil {
		return f.Spec.Sg()
	} else {
		return f.TFunction.Sg()
	}
}

func (f *Function) Yields() types.Type  { return f.Signature() }
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

func (f *Function) AddSelf() {
	self := f.TFunction.Self
	if self == nil || f.TFunction.IsStatic {
		return
	}

	var s types.Type
	// Mutable Self || Composite
	if f.TFunction.IsMutating || types.IsStruct(self.Type().Parent()) {
		// Mutating pass self as pointer
		s = types.NewPointer(self.Type())
	} else {
		s = self.Type()
	}

	p := &Parameter{
		Name:   "self",
		Symbol: s,
		Parent: f,
		IsSelf: true,
	}

	f.Parameters = append(f.Parameters, p)
	f.Variables[p.Name] = p
}

func NewFunction(fn *types.Function) *Function {
	f := &Function{
		Variables: make(map[string]Value),
		TFunction: fn,
	}
	f.NewBlock()
	return f
}
