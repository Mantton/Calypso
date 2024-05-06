package lir

import (
	"fmt"

	"github.com/mantton/calypso/internal/calypso/types"
)

// New Type For Array Type in LLVM
type StaticArray struct {
	OfType types.Type
	Count  int
}

func (t *StaticArray) Parent() types.Type { return t }

func (t *StaticArray) String() string {
	return fmt.Sprintf("[%d x %s]", t.Count, t.OfType)
}
