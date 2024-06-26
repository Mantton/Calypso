package lir

import (
	"sync/atomic"

	"github.com/mantton/calypso/internal/calypso/types"
)

type Parameter struct {
	Name   string
	Symbol types.Type
	Parent *Function
	IsSelf bool
}

type Function struct {
	TFunction  *types.Function
	Spec       *types.SpecializedFunctionSignature
	id         int64
	Blocks     []*Block
	Locals     []*Allocate
	Variables  map[string]Value
	Parameters []*Parameter

	Owner        *Member
	CurrentBlock *Block
	External     bool
	Name         string
}

func (f *Function) IsIntrinsic() bool {
	return f.TFunction.Module().AST.IsSTD() && f.TFunction.Module().Name() == "intrinsic"
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
	var self types.Type

	if f.TFunction.Self != nil {
		self = f.TFunction.Self.Type()
	}

	if self != nil && f.Spec != nil {
		self = types.Instantiate(self, f.Spec.Spec)
	}

	if self == nil || f.TFunction.IsStatic {
		return
	}

	var s types.Type
	// Mutable Self || Composite
	if f.TFunction.IsMutating || types.IsStruct(self.Parent()) {
		// Mutating pass self as pointer
		s = types.NewPointer(self)
	} else {
		s = self
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

var tick int64

func NewFunction(fn *types.Function) *Function {
	newID := atomic.AddInt64(&tick, 1)

	f := &Function{
		Variables: make(map[string]Value),
		TFunction: fn,
		id:        newID,
	}

	f.NewBlock()
	return f
}

func (fn *Function) ID() int64 {
	return fn.id
}
