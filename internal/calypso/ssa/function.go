package ssa

import "github.com/mantton/calypso/internal/calypso/symbols"

type Function struct {
	Symbol *symbols.SymbolInfo
	Blocks []*Block

	Owner *Member
}

func (f *Function) ssaMbr() {}
