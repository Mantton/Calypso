package builder

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/commands/utils"
	"github.com/mantton/calypso/internal/calypso/lir"
	"github.com/mantton/calypso/internal/calypso/lirgen"
	"github.com/mantton/calypso/internal/calypso/llir"
	"github.com/mantton/calypso/internal/calypso/parser"
	"github.com/mantton/calypso/internal/calypso/typechecker"
)

func CompileFileSet(set *utils.FileSet, mode typechecker.CheckerMode) error {

	astSet, err := parser.ParseFileSet(set)

	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("----------  TYPECHECKER ----------")
	c := typechecker.New(mode, astSet)
	mod, err := c.Check()

	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("----------  LIRGEN ----------")
	lirMod, err := lirgen.Generate(mod)
	if err != nil {
		return err
	}

	// return nil
	exec := lir.NewExecutable()
	exec.Modules[lirMod.Name()] = lirMod

	fmt.Println()
	fmt.Println("----------  LLVM-IR GEN ----------")
	llir.Compile(exec)
	return nil
}
