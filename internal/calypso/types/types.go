package types

type Type interface {
	clyT()
	String() string
	Parent() Type
}

func IsGeneric(t Type) bool {
	_, ok := t.(*TypeParam)

	if ok {
		return true
	}

	ptr, ok := t.(*Pointer)

	if ok && IsGeneric(ptr.PointerTo) {
		return true
	}

	_, ok = t.(*Instance)

	return ok
}
