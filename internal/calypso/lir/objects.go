package lir

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/types"
)

// * 2
type Executable struct {
	Modules map[string]*Module
}
type Package struct {
	Modules map[string]*Module
}

type Composite struct {
	Members          []types.Type
	UnderlyingType   types.Type
	UnderlyingSymbol types.Symbol
	Name             string
	EnumParent       *Composite
	IsAligned        bool
}

func (c *Composite) String() string {
	members := ""

	for i, m := range c.Members {
		members += m.String()

		if i != len(c.Members)-1 {
			members += ", "
		}

	}
	base := fmt.Sprintf("%s = { %s }", c.Name, members)
	return base

}

func (c *Composite) Yields() types.Type {
	return c.UnderlyingSymbol.Type()
}

// New Type For Array Type in LLVM
type StaticArray struct {
	OfType types.Type
	Count  int
}

func (t *StaticArray) Parent() types.Type { return t }

func (t *StaticArray) String() string {
	return fmt.Sprintf("[%d x %s]", t.Count, t.OfType)
}
