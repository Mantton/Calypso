package lir

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
	Target    *Function
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
