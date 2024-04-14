package typechecker

import (
	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/lexer"
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
	file    *ast.File
	fileSet *ast.FileSet

	module *types.Module
}

func New(mode CheckerMode, set *ast.FileSet) *Checker {
	c := &Checker{
		depth:   0,
		mode:    mode,
		table:   types.NewSymbolTable(),
		fileSet: set,
	}

	m := types.NewModule(set.ModuleName, types.NewPackage("local"))
	m.Table = c.table
	m.FileSet = c.fileSet
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
