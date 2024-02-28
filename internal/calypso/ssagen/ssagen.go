package ssagen

import (
	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/ssa"
	"github.com/mantton/calypso/internal/calypso/typechecker"
)

func Generate(file *ast.File, table *typechecker.SymbolTable) *ssa.Executable {
	exec := ssa.NewExecutable(file)
	build(exec, table)
	return exec
}
