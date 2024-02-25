package ssa

import "github.com/mantton/calypso/internal/calypso/types"

// represents a value known at compile time
type Constant struct {
	Value any
	Type  types.Type
}

type Global struct {
	Value Constant
}

func (*Constant) ssaNode() {}
func (*Global) ssaNode()   {}

func (*Constant) ssaVal() {}
func (*Global) ssaVal()   {}
