package types

import "fmt"

type Pointer struct {
	PointerTo Type
}

func NewPointer(t Type) *Pointer {
	return &Pointer{PointerTo: t}
}

func (t *Pointer) clyT()        {}
func (t *Pointer) Parent() Type { return t }

func (t *Pointer) String() string {

	f := "*%s"

	return fmt.Sprintf(f, t.PointerTo)
}

func IsPointer(t Type) bool {
	return t.(*Pointer) != nil
}
