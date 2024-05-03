package lir

import "github.com/mantton/calypso/internal/calypso/types"

type GenericFunction struct {
	Target *types.Function
	Specs  map[string]*Function
}

func NewGenericFunction(fn *types.Function) *GenericFunction {
	return &GenericFunction{
		Target: fn,
		Specs:  map[string]*Function{},
	}
}

func (fn *GenericFunction) Yields() types.Type {
	return fn.Target.Sg()
}
