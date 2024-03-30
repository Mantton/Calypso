package lir

import (
	"github.com/mantton/calypso/internal/calypso/types"
)

// * 1
// Either a node or value
type Node interface{}

// named members of a package
type Member interface {
	Node
}

// an expression that yields a value
type Value interface {
	Node
	Yields() types.Type
}

// a statement that consumes a value and performs computation
type Instruction interface {
	Node
}

// mix-in embeded by all SSA instructions that yield a value
type yielder struct {
	typ types.Type
}

func (y *yielder) SetType(typ types.Type) { y.typ = typ }
func (y *yielder) Yields() types.Type {
	return y.typ
}
