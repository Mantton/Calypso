package typechecker

import (
	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/lexer"
	"github.com/mantton/calypso/internal/calypso/resolver"
	"github.com/mantton/calypso/internal/calypso/types"
)

type CheckerMode byte

const DEBUG = false

const (
	//  Standard Library, Certain Restrictions are lifted
	STD CheckerMode = iota

	// User Scripts, This is the standard language
	USER
)

type Checker struct {
	Errors lexer.ErrorList
	depth  int
	mode   CheckerMode
	// scope   *types.Scope
	// fn      *types.FunctionSignature
	table *types.SymbolTable
	ctx   *NodeContext
	// lhsType types.Type
	file      *ast.File
	astModule *ast.Module

	module *types.Module
	mp     *types.PackageMap
}

func New(mod *ast.Module, mp *types.PackageMap) *Checker {
	c := &Checker{
		depth:     0,
		mode:      USER,
		table:     types.NewSymbolTable(),
		astModule: mod,
		mp:        mp,
	}

	m := types.NewModule(mod, mp.Packages[mod.Package.FSPackage.Path])
	m.Table = c.table
	c.module = m
	return c
}

func (c *Checker) ParentScope() *types.Scope {
	return c.table.Main
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
		mp.Packages[tPkg.AST.FSPackage.Path] = tPkg
	}

	for _, m := range p.OrderedModules {
		c := New(m, mp)
		mod, err := c.Check()

		if err != nil {
			return nil, err
		}

		mp.Modules[mod.AST.FSMod.Path] = mod
	}

	return mp, nil
}

// TODO: Check cyclic function usage
