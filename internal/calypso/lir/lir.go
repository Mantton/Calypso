package lir

import (
	"github.com/mantton/calypso/internal/calypso/ast"
	"github.com/mantton/calypso/internal/calypso/types"
)

// * 1
// Either a node or value
type Node interface{}

// named members of a package
type Member interface {
	Node
}

// an expression that yields a value
type Value interface {
	Node
	Yields() types.Type
}

// a statement that consumes a value and performs computation
type Instruction interface {
	Node
}

// mix-in embeded by all SSA instructions that yield a value
type yielder struct {
	typ types.Type
}

func (y *yielder) SetType(typ types.Type) { y.typ = typ }
func (y *yielder) Yields() types.Type {
	return y.typ
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
