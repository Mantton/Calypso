package ssa

import (
	"github.com/mantton/calypso/internal/calypso/symbols"
	"github.com/mantton/calypso/internal/calypso/token"
)

type Load struct {
	Address Value
}

type Allocate struct {
	TypeSymbol *symbols.SymbolInfo
	OnHeap     bool
}

type Store struct {
	Address Value
	Value   Value
}

type Call struct {
	Target    string
	Arguments []Value
}

type Return struct {
	Result Value
}

type Binary struct {
	Left  Value
	Op    token.Token
	Right Value
}

func (*Load) ssaInstr()     {}
func (*Allocate) ssaInstr() {}
func (*Store) ssaInstr()    {}
func (*Call) ssaInstr()     {}
func (*Return) ssaInstr()   {}
func (*Binary) ssaInstr()   {}

func (*Load) ssaNode()     {}
func (*Allocate) ssaNode() {}
func (*Store) ssaNode()    {}
func (*Call) ssaNode()     {}
func (*Return) ssaNode()   {}
func (*Binary) ssaNode()   {}

// Instructions That Yield Values
func (*Allocate) ssaVal() {}
func (*Binary) ssaVal()   {}
func (*Load) ssaVal()     {}
func (*Call) ssaVal()     {}
