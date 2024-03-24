package lir

import (
	"github.com/mantton/calypso/internal/calypso/ast"
)

func NewExecutable(file *ast.File) *Executable {
	return &Executable{
		IncludedFile: file,
		Modules:      make(map[string]*Module),
	}
}
