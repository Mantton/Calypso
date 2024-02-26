package ssa

import (
	"github.com/mantton/calypso/internal/calypso/token"
)

type Load struct {
	yielder
	Address Value
}

type Allocate struct {
	yielder
	OnHeap bool
}

type Store struct {
	Address Value
	Value   Value
}

type Call struct {
	yielder
	Target    string
	Arguments []Value
}

type Return struct {
	Result Value
}

type Binary struct {
	yielder
	Left  Value
	Op    token.Token
	Right Value
}

type Branch struct {
	Condition   Value
	Action      *Block // Then
	Alternative *Block // Else
}

type Jump struct {
	Block *Block
}

func (*Load) ssaInstr()     {}
func (*Allocate) ssaInstr() {}
func (*Store) ssaInstr()    {}
func (*Call) ssaInstr()     {}
func (*Return) ssaInstr()   {}
func (*Binary) ssaInstr()   {}
func (*Branch) ssaInstr()   {}
func (*Jump) ssaInstr()     {}

func (*Load) ssaNode()     {}
func (*Allocate) ssaNode() {}
func (*Store) ssaNode()    {}
func (*Call) ssaNode()     {}
func (*Return) ssaNode()   {}
func (*Binary) ssaNode()   {}
func (*Branch) ssaNode()   {}
func (*Jump) ssaNode()     {}

// Instructions That Yield Values
func (*Allocate) ssaVal() {}
func (*Binary) ssaVal()   {}
func (*Load) ssaVal()     {}
func (*Call) ssaVal()     {}
