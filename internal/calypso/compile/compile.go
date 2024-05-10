package compile

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/resolver"
)

const DEBUG = false

func CompilePackage(path string) error {
	// Resolve AST & Imports
	fmt.Println("\n\nAST GEN")
	packages, err := resolver.ParseAndResolve(path)
	if err != nil {
		return err
	}

	// fmt.Println("\n\nTypeCheck")
	// pkgMap, err := typechecker.CheckParsedData(data)
	// if err != nil {
	// 	return err
	// }

	// fmt.Println("\n\nLIR GEN")
	// exec, err := lirgen.Generate(data, pkgMap)

	// if err != nil {
	// 	return err
	// }

	// fmt.Println("\n\nLLVM-IR GEN")

	// llir.Compile(exec)
	return nil
}
