package types

import (
	"github.com/mantton/calypso/internal/calypso/ast"
)

type PackageMap struct {
	Packages map[string]*Package
	Modules  map[string]*Module
}

type Package struct {
	AST     *ast.Package
	modules map[string]*Module
}

type Module struct {
	Table        *SymbolTable
	pkg          *Package
	AST          *ast.Module
	ParentModule *Module
}

func NewPackageMap() *PackageMap {
	return &PackageMap{
		Packages: make(map[string]*Package),
		Modules:  make(map[string]*Module),
	}
}
func NewPackage(p *ast.Package) *Package {
	return &Package{
		AST: p,
	}
}

func (p *Package) AddModule(m *Module) {
	p.modules[m.Name()] = m
}

func NewModule(m *ast.Module, p *Package) *Module {
	return &Module{
		AST: m,
		pkg: p,
	}
}

func (m *Module) Package() *Package {
	return m.pkg
}

func (m *Module) Name() string {
	return m.AST.Name()
}

func (m *Package) Name() string {
	return m.AST.Name()
}

func (m *Module) Type() Type {
	return m
}

func (m *Module) Parent() Type {
	return m
}
func (m *Module) Module() *Module {
	return m
}

func (m *Module) SymbolName() string {
	return "wut"
}

func (m *Module) String() string {
	return m.Name()
}

func (m *Module) IsVisible(from *Module) bool {
	switch m.AST.Visibility {
	case ast.PRIVATE:
		return true // TODO: Private Modules
	case ast.PUBLIC:
		return true
	}

	return false
}
