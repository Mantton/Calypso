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
	fmt.Println("\n\nAST GEN")
	data, err := resolver.ParseAndResolve(pkg)
	if err != nil {
		return err
	}

	fmt.Println("\n\nTypeCheck")
	pkgMap, err := typechecker.CheckParsedData(data)
	if err != nil {
		return err
	}

	fmt.Println("\n\nLIR GEN")
	err = lirgen.Generate(data, pkgMap)

	if err != nil {
		return err
	}

	fmt.Println(pkgMap)
	return nil
}
