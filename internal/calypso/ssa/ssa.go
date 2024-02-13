package ssa

import (
	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/symbols"
)

// * 1

// named members of a package
type Member interface {
	ssaMbr()
}

// an expression that yields a value
type Value interface {
	ssaVal()
}

// a statement that consumes a value and performs computation
type Instruction interface {
	ssaInstr()
}

// Either a node or value
type Node interface {
	ssaNode()
}

// * 2
type Executable struct {
	Modules      map[string]*Module
	IncludedFile *ast.File
}

// TODO: Package
type Package struct {
	Modules map[string]*Module
}

type Module struct {
	Members map[string]Member
	Files   []*ast.File
}

func NewModule(file *ast.File) *Module {
	return &Module{
		Members: make(map[string]Member),
	}
}

// * 3
// allocates mem for value and yields it's mem address
type Alloc struct {
	Name string
}

// Stores `Value` at mem address `Address`
type Store struct {
	Address Value
	Value   Value
}

type Load struct {
	Address Value
}

// represents a value known at compile time
type Constant struct {
	Value any
	Type  *symbols.SymbolInfo
}

type Global struct {
	Value Constant
	Name  string
}

type Return struct {
	Results Value
}

func (n *Alloc) ssaVal()   {}
func (n *Alloc) ssaInstr() {}

func (n *Store) ssaInstr()  {}
func (n *Load) ssaInstr()   {}
func (n *Constant) ssaVal() {}
func (n *Return) ssaInstr() {}
