package lirgen

import (
	"github.com/mantton/calypso/internal/calypso/lir"
	"github.com/mantton/calypso/internal/calypso/resolver"
	"github.com/mantton/calypso/internal/calypso/types"
)

func Generate(data *resolver.ResolvedData, tmap *types.PackageMap) error {

	mp := lir.NewPackageMap()
	for _, mod := range data.OrderedModules {
		path := mod.FSMod.Path
		tMod := tmap.Modules[path]
		mod := lir.NewModule(tMod)
		err := build(mod, mp)
		if err != nil {
			return err
		}
		mp.Modules[path] = mod
	}

	return nil
}
