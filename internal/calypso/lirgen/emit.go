package lirgen

import (
	"github.com/mantton/calypso/internal/calypso/lir"
	"github.com/mantton/calypso/internal/calypso/types"
)

func emitStackAlloc(fn *lir.Function, t types.Type) *lir.Allocate {
	i := &lir.Allocate{}
	i.SetType(t)
	fn.Emit(i)
	return i
}

func emitHeapAlloc(fn *lir.Function, t types.Type) *lir.Allocate {
	i := &lir.Allocate{
		OnHeap: true,
	}

	i.SetType(t)
	fn.Emit(i)
	return i
}

func emitLocalVar(fn *lir.Function, v *types.Var) *lir.Allocate {
	addr := emitStackAlloc(fn, v.Type())
	fn.Variables[v.Name()] = addr
	return addr
}

func emitStore(f *lir.Function, addr lir.Value, val lir.Value) {
	i := &lir.Store{
		Address: addr,
		Value:   val,
	}

	f.Emit(i)
}

func emitGlobalVar(m *lir.Module, c *lir.Constant, k string) {
	m.GlobalConstants[k] = &lir.Global{
		Value: c,
	}
}

func emitConstantVar(fn *lir.Function, c *lir.Constant, k string) {
	fn.Variables[k] = c
}
