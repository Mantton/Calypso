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
	table *SymbolTable
	ctx   *NodeContext
	// lhsType types.Type
	file    *ast.File
	fileSet *ast.FileSet
}

func New(mode CheckerMode, set *ast.FileSet) *Checker {
	return &Checker{
		depth:   0,
		mode:    mode,
		table:   NewSymbolTable(),
		fileSet: set,
	}
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
