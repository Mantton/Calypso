package lirgen

import (
	"github.com/mantton/calypso/internal/calypso/lir"
	"github.com/mantton/calypso/internal/calypso/resolver"
	"github.com/mantton/calypso/internal/calypso/types"
)

func Generate(data *resolver.ResolvedData, tmap *types.PackageMap) (*lir.Executable, error) {

	mp := lir.NewExecutable()
	for _, mod := range data.OrderedModules {
		path := mod.Info.Path
		tMod := tmap.Modules[path]
		mod := lir.NewModule(tMod)
		err := build(mod, mp)
		if err != nil {
			return nil, err
		}
		mp.Modules[path] = mod
	}

	for _, astPkg := range data.Packages {
		lirPkg := lir.NewPackage(astPkg.Name())

		for _, mod := range astPkg.Modules {
			path := mod.Info.Path
			lirPkg.Modules[path] = mp.Modules[path]
		}

		mp.Packages[lirPkg.Name] = lirPkg
	}

	return mp, nil
}
