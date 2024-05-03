package lir

import (
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

type PackageMap struct {
	Modules map[string]*Module
}

func NewPackageMap() *PackageMap {
	return &PackageMap{
		Modules: make(map[string]*Module),
	}
}
