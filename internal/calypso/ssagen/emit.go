package ssagen

import (
	"github.com/mantton/calypso/internal/calypso/ssa"
	"github.com/mantton/calypso/internal/calypso/symbols"
)

func emitStackAlloc(fn *ssa.Function, s *symbols.SymbolInfo) *ssa.Allocate {
	i := &ssa.Allocate{
		TypeSymbol: s.TypeDesc,
	}

	fn.Emit(i)
	return i
}

func emitHeapAlloc(fn *ssa.Function, s *symbols.SymbolInfo) *ssa.Allocate {
	i := &ssa.Allocate{
		TypeSymbol: s.TypeDesc,
		OnHeap:     true,
	}

	fn.Emit(i)
	return i
}

func emitLocalVar(fn *ssa.Function, s *symbols.SymbolInfo) *ssa.Allocate {
	addr := emitStackAlloc(fn, s.TypeDesc)
	fn.Variables[s.Name] = addr
	return addr
}

func emitStore(f *ssa.Function, addr ssa.Value, val ssa.Value) {
	i := &ssa.Store{
		Address: addr,
		Value:   val,
	}

	f.Emit(i)
}
