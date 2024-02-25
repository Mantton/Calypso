package types

type Type interface {
	clyT()
	String() string
}

func IsGeneric(t Type) bool {
	_, ok := t.(*TypeParam)
	return ok
}
