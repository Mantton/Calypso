package ssa

import (
	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/types"
)

// * 1

// named members of a package
type Member interface {
	Node
	ssaMbr()
}

// an expression that yields a value
type Value interface {
	Node
	ssaVal()
	Type() types.Type
}

// a statement that consumes a value and performs computation
type Instruction interface {
	Node
	ssaInstr()
}

// mix-in embeded by all SSA instructions that yield a value
type yielder struct {
	typ types.Type
}

func (y *yielder) SetType(typ types.Type) { y.typ = typ }
func (y *yielder) Type() types.Type {
	return y.typ
}

// Either a node or value
type Node interface {
	ssaNode()
}

// * 2
type Executable struct {
	Modules      map[string]*Module
	IncludedFile *ast.File
	Scope        *types.Scope
}

// TODO: Package
type Package struct {
	Modules map[string]*Module
}

type Module struct {
	Name            string
	Functions       map[string]*Function
	GlobalConstants map[string]*Global
	Composites      map[string]types.Symbol
	Files           []*ast.File
}

func NewModule(file *ast.File, name string) *Module {
	return &Module{
		Name:            name,
		Functions:       make(map[string]*Function),
		GlobalConstants: make(map[string]*Global),
		Composites:      make(map[string]types.Symbol),
	}
}
