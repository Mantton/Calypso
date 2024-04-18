package compile

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/fs"
	"github.com/mantton/calypso/internal/calypso/resolver"
)

type compiler struct {
	pkg *fs.LitePackage
}

func newCompiler(p *fs.LitePackage) *compiler {
	return &compiler{
		pkg: p,
	}
}

func CompilePackage(pkg *fs.LitePackage) error {

	// Resolve Imports
	data, err := resolver.ParseAndResolve(pkg)
	if err != nil {
		return err
	}

	fmt.Println("Packages", data.Packages)
	fmt.Println("Module Order", data.ModuleOrder)
	return nil
}
