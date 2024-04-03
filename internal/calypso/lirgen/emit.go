package lirgen

import (
	"github.com/mantton/calypso/internal/calypso/lir"
	"github.com/mantton/calypso/internal/calypso/types"
)

func (b *builder) emitStackAlloc(fn *lir.Function, t types.Type) *lir.Allocate {
	i := &lir.Allocate{}
	i.SetType(t)
	fn.Emit(i)
	return i
}

func (b *builder) emitHeapAlloc(fn *lir.Function, t types.Type) *lir.Allocate {
	i := &lir.Allocate{
		OnHeap: true,
	}

	i.SetType(t)
	fn.Emit(i)
	return i
}

func (b *builder) emitLocalVar(fn *lir.Function, n string, t types.Type, addr *lir.Allocate) *lir.Allocate {
	if addr == nil {
		addr := b.emitStackAlloc(fn, t)
		fn.Variables[n] = addr
		return addr
	}
	fn.Variables[n] = addr
	return addr
}

func (b *builder) emitStore(fn *lir.Function, addr lir.Value, val lir.Value) {
	i := &lir.Store{
		Address: addr,
		Value:   val,
	}

	fn.Emit(i)
}

func (b *builder) emitGlobalVar(m *lir.Module, c *lir.Constant, k string) {
	m.GlobalConstants[k] = &lir.Global{
		Value: c,
	}
}

func (b *builder) emitConstantVar(fn *lir.Function, c *lir.Constant, k string) {
	fn.Variables[k] = c
}
