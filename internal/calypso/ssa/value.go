package ssa

import (
	"go/constant"

	"github.com/mantton/calypso/internal/calypso/types"
)

// represents a value known at compile time
type Constant struct {
	Value any
	typ   types.Type
}

type Global struct {
	Value *Constant
}

func (*Constant) ssaNode() {}
func (*Global) ssaNode()   {}

func (*Constant) ssaVal() {}
func (*Global) ssaVal()   {}

func (c *Constant) Type() types.Type { return c.typ }
func (c *Global) Type() types.Type   { return c.Value.Type() }

func NewConst(val constant.Value, typ types.Type) *Constant {
	return &Constant{
		Value: val,
		typ:   typ,
	}
}
