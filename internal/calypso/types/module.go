package types

import "github.com/mantton/calypso/internal/calypso/ast"

type Package struct {
	name string
}

type Module struct {
	Table   *SymbolTable
	FileSet *ast.FileSet
	pkg     *Package
	name    string
}

func NewPackage(name string) *Package {
	return &Package{
		name: name,
	}
}

func NewModule(name string, pkg *Package) *Module {
	return &Module{
		name: name,
		pkg:  pkg,
	}
}

func (m *Module) Package() *Package {
	return m.pkg
}

func (m *Module) Name() string {
	return m.name
}

func (m *Package) Name() string {
	return m.name
}
