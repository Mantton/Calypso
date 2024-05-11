package compile

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/lirgen"
	"github.com/mantton/calypso/internal/calypso/llir"
	"github.com/mantton/calypso/internal/calypso/resolver"
	"github.com/mantton/calypso/internal/calypso/typechecker"
)

const DEBUG = false

func CompilePackage(path string) error {
	// Resolve AST & Imports
	fmt.Println("\n\nAST GEN")
	packages, err := resolver.ParseAndResolve(path)
	if err != nil {
		return err
	}

	fmt.Println("\n\nTypeCheck")
	typedPackages, err := typechecker.CheckPackages(packages)
	if err != nil {
		return err
	}

	fmt.Println("\n\nLIR GEN")
	exec, err := lirgen.Generate(packages, typedPackages)

	if err != nil {
		return err
	}

	fmt.Println("\n\nLLVM-IR GEN")
	err = llir.Compile(exec)

	if err != nil {
		return err
	}
	return nil
}
