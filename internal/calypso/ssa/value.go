package ssa

import "github.com/mantton/calypso/internal/calypso/types"

// represents a value known at compile time
type Constant struct {
	Value any
	Typ   types.Type
}

type Global struct {
	Value *Constant
}

func (*Constant) ssaNode() {}
func (*Global) ssaNode()   {}

func (*Constant) ssaVal() {}
func (*Global) ssaVal()   {}

func (c *Constant) Type() types.Type { return c.Typ }
func (c *Global) Type() types.Type   { return c.Value.Type() }
