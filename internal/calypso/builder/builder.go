package builder

import (
	"github.com/mantton/calypso/internal/calypso/commands/utils"
	"github.com/mantton/calypso/internal/calypso/parser"
	"github.com/mantton/calypso/internal/calypso/typechecker"
)

func CompileFileSet(set *utils.FileSet, mode typechecker.CheckerMode) error {

	astSet, err := parser.ParseFileSet(set)

	if err != nil {
		return err
	}

	c := typechecker.New(mode, astSet)
	_, err = c.Check()

	if err != nil {
		return err
	}

	return nil
}
