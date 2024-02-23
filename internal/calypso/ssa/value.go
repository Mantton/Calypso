package ssa

import (
	"github.com/mantton/calypso/internal/calypso/symbols"
)

// represents a value known at compile time
type Constant struct {
	Value any
	Type  *symbols.SymbolInfo
}

type Global struct {
	Value    Constant
	Variable *Variable
}

type Variable struct {
	Name string
}

type Address struct {
	Variable *Variable
	Anchor   *Address
	Offset   int
}

func (*Constant) ssaNode() {}
func (*Global) ssaNode()   {}
func (*Variable) ssaNode() {}
func (*Address) ssaNode()  {}

func (*Constant) ssaVal() {}
func (*Global) ssaVal()   {}
func (*Variable) ssaVal() {}
func (*Address) ssaVal()  {}
