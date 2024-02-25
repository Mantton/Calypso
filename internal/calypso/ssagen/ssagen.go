package ssagen

import (
	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/ssa"
	"github.com/mantton/calypso/internal/calypso/types"
)

func Generate(file *ast.File, scope *types.Scope) *ssa.Executable {
	exec := ssa.NewExecutable(file)
	exec.Scope = scope
	build(exec)
	return exec
}
