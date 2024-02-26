package ssagen

import (
	"github.com/mantton/calypso/internal/calypso/ssa"
	"github.com/mantton/calypso/internal/calypso/types"
)

func emitStackAlloc(fn *ssa.Function, t types.Type) *ssa.Allocate {
	i := &ssa.Allocate{}
	i.SetType(t)
	fn.Emit(i)
	return i
}

func emitHeapAlloc(fn *ssa.Function, t types.Type) *ssa.Allocate {
	i := &ssa.Allocate{
		OnHeap: true,
	}

	i.SetType(t)
	fn.Emit(i)
	return i
}

func emitLocalVar(fn *ssa.Function, v *types.Var) *ssa.Allocate {
	addr := emitStackAlloc(fn, v.Type())
	fn.Variables[v.Name()] = addr
	return addr
}

func emitStore(f *ssa.Function, addr ssa.Value, val ssa.Value) {
	i := &ssa.Store{
		Address: addr,
		Value:   val,
	}

	f.Emit(i)
}

func emitGlobalVar(m *ssa.Module, c *ssa.Constant, k string) {
	m.GlobalConstants[k] = &ssa.Global{
		Value: c,
	}
}

func emitConstantVar(fn *ssa.Function, c *ssa.Constant, k string) {
	fn.Variables[k] = c
}
