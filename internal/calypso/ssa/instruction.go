package ssa

import "go/token"

type Load struct {
	Address *Address
}

type Allocate struct {
	Variable *Variable
	OnHeap   bool
}

type Store struct {
	Address *Address
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
