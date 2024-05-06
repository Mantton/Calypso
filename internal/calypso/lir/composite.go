package lir

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/types"
)

type Composite struct {
	Members    []types.Type
	Type       types.Type
	Name       string
	EnumParent *Composite
	IsAligned  bool
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
	return c.Type
}
