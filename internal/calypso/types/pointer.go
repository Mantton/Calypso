package types

import "fmt"

type Pointer struct {
	PointerTo Type
}

func NewPointer(t Type) *Pointer {
	return &Pointer{PointerTo: t}
}

func (t *Pointer) Parent() Type { return t }

func (t *Pointer) String() string {

	f := "*%s"

	return fmt.Sprintf(f, t.PointerTo)
}

func Dereference(t Type) Type {

	ptr, ok := t.(*Pointer)

	if ok {
		return ptr.PointerTo
	}

	// return t
	panic(fmt.Sprintf("cannot dereference non pointer type, %s", t))
}
