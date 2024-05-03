package compile

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/fs"
	"github.com/mantton/calypso/internal/calypso/lirgen"
	"github.com/mantton/calypso/internal/calypso/resolver"
	"github.com/mantton/calypso/internal/calypso/typechecker"
)

const DEBUG = false

func CompilePackage(pkg *fs.LitePackage) error {
	// Resolve AST & Imports
	data, err := resolver.ParseAndResolve(pkg)
	if err != nil {
		return err
	}

	pkgMap, err := typechecker.CheckParsedData(data)
	if err != nil {
		return err
	}

	err = lirgen.Generate(data, pkgMap)

	if err != nil {
		return err
	}

	fmt.Println(pkgMap)
	return nil
}
