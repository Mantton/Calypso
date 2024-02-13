package ssagen

import (
	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/ssa"
)

func Generate(file *ast.File) *ssa.Executable {
	exec := ssa.NewExecutable(file)
	return exec
}
