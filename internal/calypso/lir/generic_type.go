package lir

import "github.com/mantton/calypso/internal/calypso/types"

type GenericType struct {
	Target *types.DefinedType
	Specs  map[string]*Composite
}

func NewGenericType(def *types.DefinedType) *GenericType {
	return &GenericType{
		Target: def,
		Specs:  map[string]*Composite{},
	}
}

func (fn *GenericType) Yields() types.Type {
	return fn.Target
}
