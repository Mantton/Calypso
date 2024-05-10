package lir

import "github.com/mantton/calypso/internal/calypso/ast"

// *

type Package struct {
	Modules map[int64]*Module
	AST     *ast.Package
}

func NewPackage(p *ast.Package) *Package {
	return &Package{
		Modules: map[int64]*Module{},
		AST:     p,
	}
}

func (p *Package) ID() int64 {
	return p.AST.ID()
}
