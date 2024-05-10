package types

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/ast"
)

type PackageMap struct {
	Packages map[int64]*Package
	Modules  map[int64]*Module
}

type Package struct {
	AST     *ast.Package
	modules map[int64]*Module
}

type Module struct {
	Scope        *Scope // the top level scope of the module
	Table        *SymbolTable
	pkg          *Package    // the package in which this module belongs to
	AST          *ast.Module // the ast module this module typed
	ParentModule *Module     // the parent mod
}

func NewPackageMap() *PackageMap {
	return &PackageMap{
		Packages: make(map[int64]*Package),
		Modules:  make(map[int64]*Module),
	}
}
func NewPackage(p *ast.Package) *Package {
	return &Package{
		AST: p,
	}
}

func (p *Package) AddModule(m *Module) {
	p.modules[m.AST.ID()] = m
}

func NewModule(m *ast.Module, p *Package) *Module {
	return &Module{
		AST:   m,
		pkg:   p,
		Table: NewSymbolTable(),
	}
}

func (m *Module) Package() *Package {
	return m.pkg
}

func (m *Module) Name() string {
	return m.AST.Name()
}
func (m *Module) ID() int64 {
	return m.AST.ID()
}

func (m *Package) Name() string {
	return m.AST.Name()
}

func (m *Package) ID() int64 {
	return m.AST.ID()
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
	return fmt.Sprintf("%s::%s", m.Package().Name(), m.Name())
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

func (e *Module) IsPublic() bool {
	return false
}
