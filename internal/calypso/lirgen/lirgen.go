package lirgen

import (
	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/lir"
	"github.com/mantton/calypso/internal/calypso/types"
)

func Generate(packages []*ast.Package, tmap *types.PackageMap) (*lir.Executable, error) {

	exec := lir.NewExecutable()

	for _, pkg := range packages {
		err := genPackage(pkg, tmap, exec)

		if err != nil {
			return nil, err
		}
	}
	return exec, nil
}

func genPackage(p *ast.Package, mp *types.PackageMap, e *lir.Executable) error {

	// Add pkg
	pkg := lir.NewPackage(p)
	e.Packages[pkg.ID()] = pkg

	// add modules
	err := p.PerformInOrder(func(m *ast.Module) error {
		tMod := mp.Modules[m.ID()]
		mod := lir.NewModule(tMod)

		err := build(mod, e)

		if err != nil {
			return err
		}

		e.Modules[mod.ID()] = mod
		pkg.Modules[mod.ID()] = mod
		return nil
	})

	if err != nil {
		return err
	}

	return nil
}
