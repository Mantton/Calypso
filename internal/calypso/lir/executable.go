package lir

import (
	"github.com/mantton/calypso/internal/calypso/types"
	"gonum.org/v1/gonum/graph/simple"
)

type Executable struct {

	// Outputs
	Packages map[string]*Package
	Modules  map[string]*Module

	// Composites & Functions
	Composites map[types.Type]*Composite
	Functions  map[types.Type]*Function

	// Graphs
	CallGraph *simple.DirectedGraph
}

func NewExecutable() *Executable {
	return &Executable{
		Modules:    make(map[string]*Module),
		Packages:   make(map[string]*Package),
		Composites: make(map[types.Type]*Composite),
		Functions:  make(map[types.Type]*Function),
		CallGraph:  simple.NewDirectedGraph(),
	}
}

func (p Executable) GetNestedFunctions(fn *Function) []*Function {
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
