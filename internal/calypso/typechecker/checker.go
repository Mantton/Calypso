package typechecker

import (
	"errors"

	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/lexer"
	"github.com/mantton/calypso/internal/calypso/parser"
	"github.com/mantton/calypso/internal/calypso/resolver"
	"github.com/mantton/calypso/internal/calypso/types"
)

type Checker struct {
	Errors lexer.ErrorList
	depth  int
	ctx    *NodeContext
	file   *ast.File

	module *types.Module
	mp     *types.PackageMap
}

func New(mod *ast.Module, mp *types.PackageMap) *Checker {
	c := &Checker{
		depth: 0,
		mp:    mp,
	}

	m := types.NewModule(mod, mp.Packages[mod.Package.Info.Path])
	c.module = m
	return c
}

func (c *Checker) ParentScope() *types.Scope {
	return c.module.Scope
}

var unresolved types.Type

func init() {
	unresolved = types.LookUp(types.Unresolved)
}

func (c *Checker) GlobalDefine(s types.Symbol) error {
	return c.ParentScope().Define(s)
}

func (c *Checker) GlobalFind(n string) (types.Symbol, bool) {
	return c.ParentScope().Resolve(n, c.ParentScope())
}

func CheckParsedData(p *resolver.ResolvedData) (*types.PackageMap, error) {

	mp := types.NewPackageMap()

	for _, pkg := range p.Packages {
		tPkg := types.NewPackage(pkg)
		mp.Packages[tPkg.AST.Info.Path] = tPkg
	}

	for _, m := range p.OrderedModules {
		c := New(m, mp)
		mod, err := c.Check()

		if err != nil {
			return nil, err
		}

		mp.Modules[mod.AST.Info.Path] = mod
	}

	return mp, nil
}

func CheckString(str string) (*types.Module, error) {

	file, errs := parser.ParseString(str)

	if len(errs) != 0 {
		return nil, errors.New(errs.String())
	}

	m := &ast.Module{
		Set: &ast.FileSet{Files: []*ast.File{file}},
	}
	mp := types.NewPackageMap()
	c := New(m, mp)

	return c.Check()
}

// TODO: Check cyclic function usage
