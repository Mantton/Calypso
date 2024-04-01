package lirgen

import (
	"github.com/mantton/calypso/internal/calypso/lir"
	"github.com/mantton/calypso/internal/calypso/typechecker"
)

func Generate(mod *typechecker.Module) (*lir.Module, error) {
	m := lir.NewModule(mod)
	err := build(m)

	if err != nil {
		return nil, err
	}

	return m, nil
}