package lir

import (
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

func (c *Constant) Yields() types.Type { return c.typ }
func (c *Global) Yields() types.Type   { return c.Value.Yields() }

func NewConst(val any, typ types.Type) *Constant {
	return &Constant{
		Value: val,
		typ:   typ,
	}
}
