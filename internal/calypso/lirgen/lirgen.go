package lirgen

import (
	"github.com/mantton/calypso/internal/calypso/lir"
	"github.com/mantton/calypso/internal/calypso/resolver"
	"github.com/mantton/calypso/internal/calypso/types"
)

func Generate(data *resolver.ResolvedData, tmap *types.PackageMap) (any, error) {

	for _, mod := range data.OrderedModules {
		tMod := tmap.Modules[mod.FSMod.Path]
		mod := lir.NewModule(tMod)
		err := build(mod)
		if err != nil {
			return nil, err
		}
	}

	return nil, nil
}
