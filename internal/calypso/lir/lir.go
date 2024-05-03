package lir

import (
	"github.com/mantton/calypso/internal/calypso/types"
	"gonum.org/v1/gonum/graph/simple"
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

type PackageMap struct {
	Modules   map[string]*Module
	CallGraph *simple.DirectedGraph
}

func NewPackageMap() *PackageMap {
	return &PackageMap{
		Modules:   make(map[string]*Module),
		CallGraph: simple.NewDirectedGraph(),
	}
}

func (p PackageMap) GetNestedFunctions(fn *Function) []*Function {
	var dependencies []*Function

	// Traverse the graph and called called functions
	g := p.CallGraph

	nodes := g.From(fn.ID())

	for nodes.Next() {
		t := nodes.Node().(*Function)
		dependencies = append(dependencies, t)

		dependencies = append(dependencies, p.GetNestedFunctions(t)...)
	}

	return dependencies
}
